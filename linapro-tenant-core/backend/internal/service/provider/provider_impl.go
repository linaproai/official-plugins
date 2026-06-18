// provider_impl.go implements the host tenant-capability provider backed by
// linapro-tenant-core plugin tables. It injects tenant filters, membership checks, and
// platform fallback behavior so host services can remain decoupled from plugin
// storage details.

package provider

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/tenantcap"
	"lina-plugin-linapro-tenant-core/backend/internal/dao"
	"lina-plugin-linapro-tenant-core/backend/internal/model/do"
	"lina-plugin-linapro-tenant-core/backend/internal/model/entity"
	"lina-plugin-linapro-tenant-core/backend/internal/service/resolverconfig"
	"lina-plugin-linapro-tenant-core/backend/internal/service/shared"
)

const (
	// membershipTableAlias is the provider SQL alias for membership rows.
	membershipTableAlias = "m"
	// tenantTableAlias is the provider SQL alias for tenant rows.
	tenantTableAlias = "t"
)

// ResolveTenant resolves a tenant from request metadata.
func (p *Provider) ResolveTenant(ctx context.Context, request *ghttp.Request) (*tenantcap.ResolverResult, error) {
	config, err := p.resolverConfigSvc.Get(ctx)
	if err != nil {
		return nil, err
	}
	result, err := p.resolverSvc.Resolve(ctx, request, resolverconfig.ToResolverConfig(config))
	if err != nil {
		return nil, err
	}
	return &tenantcap.ResolverResult{
		TenantID:        tenantcap.TenantID(result.TenantID),
		Matched:         true,
		ActingAsTenant:  result.ActingAsTenant,
		IsImpersonation: result.ActingAsTenant,
	}, nil
}

// ValidateUserInTenant validates one user belongs to one tenant.
func (p *Provider) ValidateUserInTenant(ctx context.Context, userID int, tenantID tenantcap.TenantID) error {
	_, err := p.membershipSvc.GetByUserAndTenant(ctx, int64(userID), int64(tenantID))
	return err
}

// ListUserTenants returns tenant options for one user.
func (p *Provider) ListUserTenants(ctx context.Context, userID int) ([]tenantcap.TenantInfo, error) {
	tenants, err := p.membershipSvc.ListUserTenants(ctx, int64(userID))
	if err != nil {
		return nil, err
	}
	result := make([]tenantcap.TenantInfo, 0, len(tenants))
	for _, item := range tenants {
		if item == nil {
			continue
		}
		result = append(result, tenantcap.TenantInfo{
			ID:     tenantcap.TenantID(item.Id),
			Code:   item.Code,
			Name:   item.Name,
			Status: item.Status,
		})
	}
	return result, nil
}

// BatchListUserTenants returns active tenant memberships for visible users.
func (p *Provider) BatchListUserTenants(ctx context.Context, userIDs []int) (map[int][]tenantcap.TenantInfo, error) {
	result := make(map[int][]tenantcap.TenantInfo)
	normalized := normalizePositiveUserIDs(userIDs)
	if len(normalized) == 0 {
		return result, nil
	}
	membershipCols := dao.UserMembership.Columns()
	tenantCols := dao.Tenant.Columns()
	rows := make([]*userTenantInfoRow, 0)
	err := shared.Model(ctx, shared.TableMembership).As(membershipTableAlias).
		InnerJoin(
			shared.TableTenant+" "+tenantTableAlias,
			qualifiedProviderColumn(tenantTableAlias, tenantCols.Id)+" = "+
				qualifiedProviderColumn(membershipTableAlias, membershipCols.TenantId)+" AND "+
				qualifiedProviderColumn(tenantTableAlias, tenantCols.DeletedAt)+" IS NULL",
		).
		Fields(
			providerSelectColumn(membershipTableAlias, membershipCols.UserId, membershipCols.UserId),
			providerSelectColumn(tenantTableAlias, tenantCols.Id, tenantCols.Id),
			providerSelectColumn(tenantTableAlias, tenantCols.Code, tenantCols.Code),
			providerSelectColumn(tenantTableAlias, tenantCols.Name, tenantCols.Name),
			providerSelectColumn(tenantTableAlias, tenantCols.Status, tenantCols.Status),
		).
		WhereIn(qualifiedProviderColumn(membershipTableAlias, membershipCols.UserId), normalized).
		Where(qualifiedProviderColumn(membershipTableAlias, membershipCols.Status), shared.MembershipStatusEnabled).
		Where(qualifiedProviderColumn(tenantTableAlias, tenantCols.Status), string(shared.TenantStatusActive)).
		OrderAsc(qualifiedProviderColumn(membershipTableAlias, membershipCols.Id)).
		Scan(&rows)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		if row == nil {
			continue
		}
		result[row.UserID] = append(result[row.UserID], tenantcap.TenantInfo{
			ID:     tenantcap.TenantID(row.ID),
			Code:   row.Code,
			Name:   row.Name,
			Status: row.Status,
		})
	}
	return result, nil
}

// CurrentTenantInfo returns one tenant projection visible in the current context.
func (p *Provider) CurrentTenantInfo(ctx context.Context, tenantID tenantcap.TenantID) (*tenantcap.TenantInfo, error) {
	if tenantID <= tenantcap.PLATFORM {
		return &tenantcap.TenantInfo{ID: tenantcap.PLATFORM, Code: "platform", Name: "Platform", Status: string(shared.TenantStatusActive)}, nil
	}
	result, err := p.BatchGetTenants(ctx, []tenantcap.TenantID{tenantID})
	if err != nil {
		return nil, err
	}
	if result == nil || result.Items[tenantID] == nil {
		return nil, bizerr.NewCode(tenantcap.CodeTenantForbidden, bizerr.P("tenantId", int(tenantID)))
	}
	return result.Items[tenantID], nil
}

// BatchGetTenants returns visible tenant projections and opaque missing IDs.
func (p *Provider) BatchGetTenants(
	ctx context.Context,
	tenantIDs []tenantcap.TenantID,
) (*capmodel.BatchResult[*tenantcap.TenantInfo, tenantcap.TenantID], error) {
	result := &capmodel.BatchResult[*tenantcap.TenantInfo, tenantcap.TenantID]{
		Items:      make(map[tenantcap.TenantID]*tenantcap.TenantInfo),
		MissingIDs: make([]tenantcap.TenantID, 0),
	}
	normalized := normalizeTenantIDs(tenantIDs)
	if len(normalized) == 0 {
		return result, nil
	}
	if containsTenantID(normalized, tenantcap.PLATFORM) {
		result.Items[tenantcap.PLATFORM] = &tenantcap.TenantInfo{ID: tenantcap.PLATFORM, Code: "platform", Name: "Platform", Status: string(shared.TenantStatusActive)}
	}
	filtered := make([]int64, 0, len(normalized))
	for _, tenantID := range normalized {
		if tenantID > tenantcap.PLATFORM {
			filtered = append(filtered, int64(tenantID))
		}
	}
	if len(filtered) > 0 {
		tenantCols := dao.Tenant.Columns()
		rows := make([]*entity.Tenant, 0, len(filtered))
		if err := shared.Model(ctx, shared.TableTenant).
			Fields(tenantCols.Id, tenantCols.Code, tenantCols.Name, tenantCols.Status).
			WhereIn(tenantCols.Id, filtered).
			Scan(&rows); err != nil {
			return nil, err
		}
		for _, row := range rows {
			if row == nil {
				continue
			}
			result.Items[tenantcap.TenantID(row.Id)] = toTenantInfo(row)
		}
	}
	for _, tenantID := range normalized {
		if _, ok := result.Items[tenantID]; !ok {
			result.MissingIDs = append(result.MissingIDs, tenantID)
		}
	}
	return result, nil
}

// SearchTenants returns bounded tenant candidates visible to the caller.
func (p *Provider) SearchTenants(
	ctx context.Context,
	input tenantcap.SearchInput,
) (*capmodel.PageResult[*tenantcap.TenantInfo], error) {
	pageNum, pageSize := normalizeTenantSearchPage(input.Page)
	tenantCols := dao.Tenant.Columns()
	model := shared.Model(ctx, shared.TableTenant)
	if keyword := strings.TrimSpace(input.Keyword); keyword != "" {
		filter := model.Builder().
			WhereLike(tenantCols.Code, "%"+keyword+"%").
			WhereOrLike(tenantCols.Name, "%"+keyword+"%")
		model = model.Where(filter)
	}
	if code := strings.TrimSpace(input.Code); code != "" {
		model = model.WhereLike(tenantCols.Code, "%"+code+"%")
	}
	if name := strings.TrimSpace(input.Name); name != "" {
		model = model.WhereLike(tenantCols.Name, "%"+name+"%")
	}
	if status := strings.TrimSpace(input.Status); status != "" {
		model = model.Where(tenantCols.Status, status)
	}
	total, err := model.Count()
	if err != nil {
		return nil, err
	}
	rows := make([]*entity.Tenant, 0)
	if err = model.Page(pageNum, pageSize).OrderDesc(tenantCols.Id).Scan(&rows); err != nil {
		return nil, err
	}
	items := make([]*tenantcap.TenantInfo, 0, len(rows))
	for _, row := range rows {
		if row != nil {
			items = append(items, toTenantInfo(row))
		}
	}
	return &capmodel.PageResult[*tenantcap.TenantInfo]{Items: items, Total: total}, nil
}

// EnsureTenantsVisible validates that every tenant identifier is visible.
func (p *Provider) EnsureTenantsVisible(ctx context.Context, tenantIDs []tenantcap.TenantID) error {
	normalized := normalizeTenantIDs(tenantIDs)
	if len(normalized) == 0 {
		return nil
	}
	result, err := p.BatchGetTenants(ctx, normalized)
	if err != nil {
		return err
	}
	if result == nil || len(result.MissingIDs) > 0 {
		tenantID := tenantcap.TenantID(0)
		if result != nil && len(result.MissingIDs) > 0 {
			tenantID = result.MissingIDs[0]
		}
		return bizerr.NewCode(tenantcap.CodeTenantForbidden, bizerr.P("tenantId", int(tenantID)))
	}
	return nil
}

// SwitchTenant validates one user can switch to a target tenant.
func (p *Provider) SwitchTenant(ctx context.Context, userID int, target tenantcap.TenantID) error {
	return p.ValidateUserInTenant(ctx, userID, target)
}

// ApplyUserTenantScope constrains user rows by active current-tenant membership.
func (p *Provider) ApplyUserTenantScope(
	ctx context.Context,
	model *gdb.Model,
	userIDColumn string,
) (*gdb.Model, bool, error) {
	return p.membershipSvc.ApplyUserTenantScope(ctx, model, userIDColumn)
}

// ApplyUserTenantFilter constrains platform user-list rows to a requested tenant.
func (p *Provider) ApplyUserTenantFilter(
	ctx context.Context,
	model *gdb.Model,
	userIDColumn string,
	tenantID tenantcap.TenantID,
) (*gdb.Model, bool, error) {
	return p.membershipSvc.ApplyUserTenantFilter(ctx, model, userIDColumn, tenantID)
}

// ListUserTenantProjections returns tenant ownership labels for visible users.
func (p *Provider) ListUserTenantProjections(
	ctx context.Context,
	userIDs []int,
) (map[int]*tenantcap.UserTenantProjection, error) {
	return p.membershipSvc.ListUserTenantProjections(ctx, userIDs)
}

// ResolveUserTenantAssignment validates requested memberships and returns a host write plan.
func (p *Provider) ResolveUserTenantAssignment(
	ctx context.Context,
	requested []tenantcap.TenantID,
	mode tenantcap.UserTenantAssignmentMode,
) (*tenantcap.UserTenantAssignmentPlan, error) {
	return p.membershipSvc.ResolveUserTenantAssignment(ctx, requested, mode)
}

// ReplaceUserTenantAssignments rewrites one user's active tenant ownership rows.
func (p *Provider) ReplaceUserTenantAssignments(
	ctx context.Context,
	userID int,
	plan *tenantcap.UserTenantAssignmentPlan,
) error {
	return p.membershipSvc.ReplaceUserTenantAssignments(ctx, userID, plan)
}

// EnsureUsersInTenant verifies every user has active membership in the tenant.
func (p *Provider) EnsureUsersInTenant(ctx context.Context, userIDs []int, tenantID tenantcap.TenantID) error {
	return p.membershipSvc.EnsureUsersInTenant(ctx, userIDs, tenantID)
}

// ValidateStartupConsistency returns user-membership startup consistency failures.
func (p *Provider) ValidateStartupConsistency(ctx context.Context) ([]string, error) {
	return p.membershipSvc.ValidateStartupConsistency(ctx)
}

// ProvisionAutoEnabledTenantPlugins provisions platform-approved tenant
// plugins for every existing active tenant. The host calls this during startup
// after plugin.autoEnable has enabled tenant-scoped plugins and after the
// linapro-tenant-core provider has registered through source-plugin route callbacks.
func (p *Provider) ProvisionAutoEnabledTenantPlugins(ctx context.Context) error {
	if p == nil || p.tenantPluginSvc == nil {
		return nil
	}
	var rows []*entity.Tenant
	err := dao.Tenant.Ctx(ctx).
		Where(do.Tenant{Status: string(shared.TenantStatusActive)}).
		OrderAsc(dao.Tenant.Columns().Id).
		Scan(&rows)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if row == nil || row.Id <= shared.PlatformTenantID {
			continue
		}
		if err = p.tenantPluginSvc.ProvisionForTenant(ctx, row.Id); err != nil {
			return err
		}
	}
	return nil
}

// userTenantInfoRow is the batch user-to-tenant projection row.
type userTenantInfoRow struct {
	UserID int    `json:"userId" orm:"user_id"`
	ID     int64  `json:"id" orm:"id"`
	Code   string `json:"code" orm:"code"`
	Name   string `json:"name" orm:"name"`
	Status string `json:"status" orm:"status"`
}

// toTenantInfo converts a plugin tenant row to the public capability DTO.
func toTenantInfo(row *entity.Tenant) *tenantcap.TenantInfo {
	if row == nil {
		return nil
	}
	return &tenantcap.TenantInfo{
		ID:     tenantcap.TenantID(row.Id),
		Code:   row.Code,
		Name:   row.Name,
		Status: row.Status,
	}
}

// qualifiedProviderColumn returns an aliased table column expression.
func qualifiedProviderColumn(tableAlias string, column string) string {
	return tableAlias + "." + column
}

// providerSelectColumn returns a qualified column with a stable scan alias.
func providerSelectColumn(tableAlias string, column string, scanAlias string) string {
	return qualifiedProviderColumn(tableAlias, column) + " AS " + scanAlias
}

// normalizeTenantSearchPage applies provider-side tenant search bounds.
func normalizeTenantSearchPage(page capmodel.PageRequest) (int, int) {
	pageNum := page.PageNum
	if pageNum <= 0 {
		pageNum = 1
	}
	pageSize := page.PageSize
	if pageSize <= 0 {
		pageSize = page.Limit
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > tenantcap.MaxTenantSearchPageSize {
		pageSize = tenantcap.MaxTenantSearchPageSize
	}
	return pageNum, pageSize
}

// normalizeTenantIDs removes duplicate valid tenant IDs.
func normalizeTenantIDs(ids []tenantcap.TenantID) []tenantcap.TenantID {
	result := make([]tenantcap.TenantID, 0, len(ids))
	seen := make(map[tenantcap.TenantID]struct{}, len(ids))
	for _, id := range ids {
		if id < tenantcap.PLATFORM {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

// normalizePositiveUserIDs removes duplicate positive user identifiers.
func normalizePositiveUserIDs(ids []int) []int {
	result := make([]int, 0, len(ids))
	seen := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

// containsTenantID reports whether the slice contains one tenant ID.
func containsTenantID(ids []tenantcap.TenantID, target tenantcap.TenantID) bool {
	for _, id := range ids {
		if id == target {
			return true
		}
	}
	return false
}
