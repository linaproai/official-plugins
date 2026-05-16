// tenant_impl.go implements tenant CRUD and lifecycle transitions for the
// multi-tenant plugin. It validates reserved codes, coordinates built-in tenant
// plugin provisioning, and calls host lifecycle preconditions before destructive
// tenant operations.

package tenant

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/bizerr"
	"lina-plugin-multi-tenant/backend/internal/dao"
	"lina-plugin-multi-tenant/backend/internal/service/shared"
)

// List queries tenants with pagination and filters.
func (s *serviceImpl) List(ctx context.Context, in ListInput) (*ListOutput, error) {
	model := shared.Model(ctx, shared.TableTenant)
	if in.Code != "" {
		model = model.WhereLike("code", "%"+in.Code+"%")
	}
	if in.Name != "" {
		model = model.WhereLike("name", "%"+in.Name+"%")
	}
	if in.Status != "" {
		model = model.Where("status", in.Status)
	}

	total, err := model.Count()
	if err != nil {
		return nil, err
	}

	list := make([]*Entity, 0)
	if err = model.Page(in.PageNum, in.PageSize).OrderDesc("id").Scan(&list); err != nil {
		return nil, err
	}
	return &ListOutput{List: list, Total: total}, nil
}

// Get retrieves one tenant by primary key.
func (s *serviceImpl) Get(ctx context.Context, id int64) (*Entity, error) {
	var item *Entity
	if err := shared.Model(ctx, shared.TableTenant).Where("id", id).Scan(&item); err != nil {
		return nil, err
	}
	if item == nil {
		return nil, bizerr.NewCode(CodeTenantNotFound)
	}
	return item, nil
}

// Create creates one tenant and provisions built-in tenant plugin state.
func (s *serviceImpl) Create(ctx context.Context, in CreateInput) (int64, error) {
	code := strings.TrimSpace(in.Code)
	if err := validateTenantCode(code); err != nil {
		return 0, err
	}
	reserved, err := s.isReservedTenantCode(ctx, code)
	if err != nil {
		return 0, err
	}
	if reserved {
		return 0, bizerr.NewCode(CodeTenantCodeReserved)
	}
	if err := s.ensureCodeAvailable(ctx, code); err != nil {
		return 0, err
	}

	bizCtx := s.bizCtxSvc.Current(ctx)
	userID := int64(bizCtx.UserID)

	var id int64
	err = dao.Tenant.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		insertID, err := shared.Model(ctx, shared.TableTenant).Data(tenantInsertData{
			Code:      code,
			Name:      in.Name,
			Status:    string(shared.TenantStatusActive),
			Remark:    in.Remark,
			CreatedBy: userID,
			UpdatedBy: userID,
		}).InsertAndGetId()
		if err != nil {
			return err
		}
		id = insertID
		return s.tenantPluginSvc.ProvisionForTenant(ctx, id)
	})
	if err != nil {
		return 0, err
	}
	return id, nil
}

// Update updates tenant basic fields.
func (s *serviceImpl) Update(ctx context.Context, in UpdateInput) error {
	if _, err := s.Get(ctx, in.Id); err != nil {
		return err
	}
	bizCtx := s.bizCtxSvc.Current(ctx)
	data := tenantUpdateData{UpdatedBy: int64(bizCtx.UserID)}
	if in.Name != nil {
		data.Name = *in.Name
	}
	if in.Remark != nil {
		data.Remark = *in.Remark
	}
	_, err := shared.Model(ctx, shared.TableTenant).OmitNilData().Where("id", in.Id).Data(data).Update()
	return err
}

// ChangeStatus performs a lifecycle status transition.
func (s *serviceImpl) ChangeStatus(ctx context.Context, id int64, status shared.TenantStatus) error {
	if !isValidStatus(status) {
		return bizerr.NewCode(CodeTenantInvalidStatus)
	}
	item, err := s.Get(ctx, id)
	if err != nil {
		return err
	}
	current := shared.TenantStatus(item.Status)
	if !isStatusTransitionAllowed(current, status) {
		return bizerr.NewCode(
			CodeTenantStatusTransitionInvalid,
			bizerr.P("from", current),
			bizerr.P("to", status),
		)
	}
	bizCtx := s.bizCtxSvc.Current(ctx)
	_, err = shared.Model(ctx, shared.TableTenant).Where("id", id).Data(tenantUpdateData{
		Status:    string(status),
		UpdatedBy: int64(bizCtx.UserID),
	}).Update()
	if err != nil {
		return err
	}
	return nil
}

// Delete soft-deletes a tenant after lifecycle precondition checks pass.
func (s *serviceImpl) Delete(ctx context.Context, id int64) error {
	if _, err := s.Get(ctx, id); err != nil {
		return err
	}
	if err := s.ensureTenantDeletePreconditionAllowed(ctx, id); err != nil {
		return err
	}
	_, err := shared.Model(ctx, shared.TableTenant).Where("id", id).Delete()
	if err != nil {
		return err
	}
	if s.pluginLifecycleSvc != nil {
		s.pluginLifecycleSvc.NotifyTenantDeleted(ctx, int(id))
	}
	return nil
}

// ensureTenantDeletePreconditionAllowed asks the host lifecycle service whether
// a tenant can be deleted.
func (s *serviceImpl) ensureTenantDeletePreconditionAllowed(ctx context.Context, tenantID int64) error {
	if s.pluginLifecycleSvc == nil {
		return nil
	}
	if err := s.pluginLifecycleSvc.EnsureTenantDeleteAllowed(ctx, int(tenantID)); err != nil {
		return bizerr.WrapCode(err, CodeTenantDeletePreconditionVetoed, bizerr.P("tenantId", tenantID))
	}
	return nil
}

// isReservedTenantCode reports whether code is blocked by resolver subdomain reservations.
func (s *serviceImpl) isReservedTenantCode(ctx context.Context, code string) (bool, error) {
	reservedCodes := defaultReservedTenantCodes()
	if s != nil && s.resolverConfigSvc != nil {
		config, err := s.resolverConfigSvc.Get(ctx)
		if err != nil {
			return false, err
		}
		if len(config.ReservedSubdomains) > 0 {
			reservedCodes = config.ReservedSubdomains
		}
	}
	for _, reserved := range reservedCodes {
		if code == reserved {
			return true, nil
		}
	}
	return false, nil
}

// ensureCodeAvailable enforces active uniqueness and soft-delete tombstone retention.
func (s *serviceImpl) ensureCodeAvailable(ctx context.Context, code string) error {
	var row *tenantCodeRow
	err := shared.Model(ctx, shared.TableTenant).Unscoped().
		Fields("id", "code", "deleted_at").
		Where("code", code).
		Scan(&row)
	if err != nil {
		return err
	}
	if row == nil || row.Id == 0 {
		return nil
	}
	if row.DeletedAt == nil || row.DeletedAt.IsZero() {
		return bizerr.NewCode(CodeTenantCodeExists)
	}
	if time.Since(*row.DeletedAt) < tenantTombstoneRetention {
		return bizerr.NewCode(CodeTenantCodeReserved)
	}
	_, err = shared.Model(ctx, shared.TableTenant).Unscoped().Where("id", row.Id).Delete()
	return err
}

var tenantCodePattern = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]*[a-z0-9])$`)

// isValidStatus reports whether status is a supported tenant lifecycle status.
func isValidStatus(status shared.TenantStatus) bool {
	switch status {
	case shared.TenantStatusActive, shared.TenantStatusSuspended:
		return true
	default:
		return false
	}
}

// validateTenantCode verifies the stable public tenant-code contract.
func validateTenantCode(code string) error {
	if len(code) < tenantCodeMinLength || len(code) > tenantCodeMaxLength {
		return bizerr.NewCode(CodeTenantCodeInvalid)
	}
	if !tenantCodePattern.MatchString(code) {
		return bizerr.NewCode(CodeTenantCodeInvalid)
	}
	return nil
}

// defaultReservedTenantCodes returns built-in labels that must never become tenant codes.
func defaultReservedTenantCodes() []string {
	return []string{"www", "api", "admin", "static", "docs"}
}

// isStatusTransitionAllowed enforces the tenant lifecycle state machine.
func isStatusTransitionAllowed(current shared.TenantStatus, next shared.TenantStatus) bool {
	switch current {
	case shared.TenantStatusActive:
		return next == shared.TenantStatusSuspended
	case shared.TenantStatusSuspended:
		return next == shared.TenantStatusActive
	default:
		return false
	}
}
