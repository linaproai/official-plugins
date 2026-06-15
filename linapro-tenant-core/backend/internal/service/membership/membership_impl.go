// membership_impl.go implements tenant membership lookup and mutation for the
// linapro-tenant-core plugin. It maintains user-to-tenant relations in plugin-owned
// tables and preserves platform membership semantics required by host tenant
// context resolution.

package membership

import (
	"context"
	"strconv"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/capability/usercap"
	"lina-plugin-linapro-tenant-core/backend/internal/dao"
	"lina-plugin-linapro-tenant-core/backend/internal/model/do"
	"lina-plugin-linapro-tenant-core/backend/internal/service/shared"
)

const membershipCapabilityPluginID = "linapro-tenant-core"

// List queries tenant members by page.
func (s *serviceImpl) List(ctx context.Context, in ListInput) (*ListOutput, error) {
	in.TenantID = s.effectiveTenantID(ctx, in.TenantID)
	model := membershipListModel(ctx, in)
	total, err := model.Count()
	if err != nil {
		return nil, err
	}
	list := make([]*Entity, 0)
	if err = membershipListModel(ctx, in).
		Fields("m.*").
		Page(in.PageNum, in.PageSize).
		OrderDesc("m.id").
		Scan(&list); err != nil {
		return nil, err
	}
	if err = s.hydrateUserLabels(ctx, list); err != nil {
		return nil, err
	}
	return &ListOutput{List: list, Total: total}, nil
}

// Add adds a user to one tenant.
func (s *serviceImpl) Add(ctx context.Context, in AddInput) (int64, error) {
	in.TenantID = s.effectiveTenantID(ctx, in.TenantID)
	if err := s.ensureUserCanJoinTenant(ctx, in.UserID, []tenantcap.TenantID{tenantcap.TenantID(in.TenantID)}); err != nil {
		return 0, err
	}
	if err := s.ensureTenantAcceptsMembership(ctx, in.TenantID); err != nil {
		return 0, err
	}
	count, err := shared.Model(ctx, shared.TableMembership).
		Where("user_id", in.UserID).
		Where("tenant_id", in.TenantID).
		Count()
	if err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, bizerr.NewCode(CodeMembershipExists)
	}

	bizCtx := s.bizCtxSvc.Current(ctx)
	userID := int64(bizCtx.UserID)
	return shared.Model(ctx, shared.TableMembership).Data(do.UserMembership{
		UserId:    in.UserID,
		TenantId:  in.TenantID,
		Status:    shared.MembershipStatusEnabled,
		CreatedBy: userID,
		UpdatedBy: userID,
	}).InsertAndGetId()
}

// Update updates membership status fields.
func (s *serviceImpl) Update(ctx context.Context, in UpdateInput) error {
	tenantID := s.effectiveTenantID(ctx, shared.PlatformTenantID)
	if _, err := s.getVisible(ctx, in.Id, tenantID); err != nil {
		return err
	}
	bizCtx := s.bizCtxSvc.Current(ctx)
	data := do.UserMembership{UpdatedBy: int64(bizCtx.UserID)}
	if in.Status != nil {
		data.Status = *in.Status
	}
	result, err := s.visibleMembershipModel(ctx, tenantID).Where("id", in.Id).OmitNilData().Data(data).Update()
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return bizerr.NewCode(CodeMembershipNotFound)
	}
	return nil
}

// Remove deletes one membership.
func (s *serviceImpl) Remove(ctx context.Context, id int64) error {
	tenantID := s.effectiveTenantID(ctx, shared.PlatformTenantID)
	if _, err := s.getVisible(ctx, id, tenantID); err != nil {
		return err
	}
	result, err := s.visibleMembershipModel(ctx, tenantID).Where("id", id).Delete()
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return bizerr.NewCode(CodeMembershipNotFound)
	}
	return nil
}

// Current returns the current user's membership in the current tenant.
func (s *serviceImpl) Current(ctx context.Context) (*Entity, error) {
	bizCtx := s.bizCtxSvc.Current(ctx)
	userID := int64(bizCtx.UserID)
	tenantID := int64(bizCtx.TenantID)
	if userID <= 0 || tenantID <= shared.PlatformTenantID {
		return nil, bizerr.NewCode(CodeMembershipNotFound)
	}
	return s.GetByUserAndTenant(ctx, userID, tenantID)
}

// GetByUserAndTenant returns one membership for a user and tenant.
func (s *serviceImpl) GetByUserAndTenant(ctx context.Context, userID int64, tenantID int64) (*Entity, error) {
	var row *membershipTenantRow
	err := shared.Model(ctx, shared.TableMembership).
		As("m").
		InnerJoin(shared.TableTenant+" t", "t.id = m.tenant_id AND t.deleted_at IS NULL").
		Fields("m.id, m.status, m.tenant_id, t.status AS tenant_status").
		Where("m.user_id", userID).
		Where("m.tenant_id", tenantID).
		Where("m.status", shared.MembershipStatusEnabled).
		Scan(&row)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, bizerr.NewCode(CodeMembershipNotFound)
	}
	if row.TStatus != string(shared.TenantStatusActive) {
		return nil, bizerr.NewCode(CodeTenantUnavailable)
	}
	return &Entity{Id: row.Id, UserID: userID, TenantID: row.TenantID, Status: row.Status}, nil
}

// ListUserTenants returns enabled tenant memberships for one user.
func (s *serviceImpl) ListUserTenants(ctx context.Context, userID int64) ([]*TenantInfo, error) {
	list := make([]*TenantInfo, 0)
	err := shared.Model(ctx, shared.TableMembership).As("m").
		InnerJoin(shared.TableTenant+" t", "t.id = m.tenant_id").
		Fields("t.id, t.code, t.name, t.status").
		Where("m.user_id", userID).
		Where("m.status", shared.MembershipStatusEnabled).
		Where("t.status", string(shared.TenantStatusActive)).
		OrderAsc("m.id").
		Scan(&list)
	return list, err
}

// ApplyUserTenantScope constrains user rows by active current-tenant membership.
func (s *serviceImpl) ApplyUserTenantScope(
	ctx context.Context,
	model *gdb.Model,
	userIDColumn string,
) (*gdb.Model, bool, error) {
	bizCtx := s.bizCtxSvc.Current(ctx)
	tenantID := int64(bizCtx.TenantID)
	if model == nil || tenantID <= shared.PlatformTenantID {
		return model, false, nil
	}
	return model.WhereIn(userIDColumn, activeMembershipUserModel(ctx, tenantID)), false, nil
}

// ApplyUserTenantFilter constrains platform user-list rows to a requested tenant.
func (s *serviceImpl) ApplyUserTenantFilter(
	ctx context.Context,
	model *gdb.Model,
	userIDColumn string,
	tenantID tenantcap.TenantID,
) (*gdb.Model, bool, error) {
	bizCtx := s.bizCtxSvc.Current(ctx)
	if model == nil ||
		tenantID <= tenantcap.PLATFORM ||
		bizCtx.TenantID != int(shared.PlatformTenantID) {
		return model, false, nil
	}
	return model.WhereIn(userIDColumn, activeMembershipUserModel(ctx, int64(tenantID))), false, nil
}

// ListUserTenantProjections returns tenant ownership labels for visible users.
func (s *serviceImpl) ListUserTenantProjections(
	ctx context.Context,
	userIDs []int,
) (map[int]*tenantcap.UserTenantProjection, error) {
	result := make(map[int]*tenantcap.UserTenantProjection)
	if len(userIDs) == 0 {
		return result, nil
	}

	var rows []*userTenantProjectionRow
	model := shared.Model(ctx, shared.TableMembership).As("m").
		InnerJoin(shared.TableTenant+" t", "t.id = m.tenant_id AND t.deleted_at IS NULL").
		Fields("m.user_id, m.tenant_id, t.name AS tenant_name").
		WhereIn("m.user_id", userIDs).
		Where("m.status", shared.MembershipStatusEnabled).
		Where("t.status", string(shared.TenantStatusActive)).
		OrderAsc("m.id")
	bizCtx := s.bizCtxSvc.Current(ctx)
	if tenantID := int64(bizCtx.TenantID); tenantID > shared.PlatformTenantID {
		model = model.Where("m.tenant_id", tenantID)
	}
	if err := model.Scan(&rows); err != nil {
		return nil, err
	}
	for _, row := range rows {
		if row == nil {
			continue
		}
		item := result[row.UserID]
		if item == nil {
			item = &tenantcap.UserTenantProjection{}
			result[row.UserID] = item
		}
		item.TenantIDs = append(item.TenantIDs, tenantcap.TenantID(row.TenantID))
		item.TenantNames = append(item.TenantNames, row.TenantName)
	}
	return result, nil
}

// ResolveUserTenantAssignment validates requested memberships and returns a host write plan.
func (s *serviceImpl) ResolveUserTenantAssignment(
	ctx context.Context,
	requested []tenantcap.TenantID,
	mode tenantcap.UserTenantAssignmentMode,
) (*tenantcap.UserTenantAssignmentPlan, error) {
	normalized := normalizeTenantIDs(requested)
	bizCtx := s.bizCtxSvc.Current(ctx)
	currentTenantID := int64(bizCtx.TenantID)
	if currentTenantID > shared.PlatformTenantID {
		if mode == tenantcap.UserTenantAssignmentUpdate {
			return &tenantcap.UserTenantAssignmentPlan{}, nil
		}
		return &tenantcap.UserTenantAssignmentPlan{
			TenantIDs:     []tenantcap.TenantID{tenantcap.TenantID(currentTenantID)},
			ShouldReplace: true,
			PrimaryTenant: tenantcap.TenantID(currentTenantID),
		}, nil
	}
	if err := s.ensureTenantIDsAcceptMembership(ctx, normalized); err != nil {
		return nil, err
	}
	return &tenantcap.UserTenantAssignmentPlan{
		TenantIDs:     normalized,
		ShouldReplace: mode == tenantcap.UserTenantAssignmentUpdate || len(normalized) > 0,
		PrimaryTenant: firstTenantIDOrPlatform(normalized),
	}, nil
}

// ReplaceUserTenantAssignments rewrites one user's active tenant ownership rows.
func (s *serviceImpl) ReplaceUserTenantAssignments(
	ctx context.Context,
	userID int,
	plan *tenantcap.UserTenantAssignmentPlan,
) error {
	if userID <= 0 || plan == nil {
		return nil
	}
	normalized := normalizeTenantIDs(plan.TenantIDs)
	if len(normalized) > 0 {
		if err := s.ensureUserCanJoinTenant(ctx, int64(userID), normalized); err != nil {
			return err
		}
	}
	if len(normalized) == 0 {
		if _, err := dao.UserMembership.Ctx(ctx).Unscoped().Where("user_id", userID).Delete(); err != nil {
			return err
		}
		return nil
	}
	if err := s.ensureTenantIDsAcceptMembership(ctx, normalized); err != nil {
		return err
	}

	if _, err := dao.UserMembership.Ctx(ctx).Unscoped().Where("user_id", userID).Delete(); err != nil {
		return err
	}
	bizCtx := s.bizCtxSvc.Current(ctx)
	operatorUserID := int64(bizCtx.UserID)
	for _, tenantID := range normalized {
		if _, err := dao.UserMembership.Ctx(ctx).Data(do.UserMembership{
			UserId:    userID,
			TenantId:  int64(tenantID),
			Status:    shared.MembershipStatusEnabled,
			CreatedBy: operatorUserID,
			UpdatedBy: operatorUserID,
		}).Insert(); err != nil {
			return err
		}
	}
	return nil
}

// EnsureUsersInTenant verifies every user has active membership in the tenant.
func (s *serviceImpl) EnsureUsersInTenant(
	ctx context.Context,
	userIDs []int,
	tenantID tenantcap.TenantID,
) error {
	normalized := normalizeUserIDs(userIDs)
	if len(normalized) == 0 || tenantID <= tenantcap.PLATFORM {
		return nil
	}
	count, err := activeMembershipUserModel(ctx, int64(tenantID)).WhereIn("user_id", normalized).Count()
	if err != nil {
		return err
	}
	if count != len(normalized) {
		return bizerr.NewCode(tenantcap.CodeTenantForbidden, bizerr.P("tenantId", int(tenantID)))
	}
	return nil
}

// ValidateStartupConsistency returns user-membership startup consistency failures.
func (s *serviceImpl) ValidateStartupConsistency(ctx context.Context) ([]string, error) {
	if s == nil || s.users == nil {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", "user"))
	}
	memberRows := make([]*Entity, 0)
	if err := shared.Model(ctx, shared.TableMembership).
		Fields("user_id").
		Where("status", shared.MembershipStatusEnabled).
		Limit(200).
		Scan(&memberRows); err != nil {
		return nil, err
	}
	userIDs := make([]usercap.UserID, 0, len(memberRows))
	for _, row := range memberRows {
		if row != nil && row.UserID > 0 {
			userIDs = append(userIDs, usercap.UserID(strconv.FormatInt(row.UserID, 10)))
		}
	}
	out, err := s.users.BatchGet(ctx, s.capabilityContext(ctx, "membership.startup_consistency"), userIDs)
	if err != nil {
		return nil, err
	}
	details := make([]string, 0)
	if out != nil {
		for _, row := range out.Items {
			if row == nil || row.TenantID != capmodel.DomainID(strconv.FormatInt(shared.PlatformTenantID, 10)) {
				continue
			}
			id, _ := strconv.Atoi(string(row.ID))
			details = append(details, "platform user "+row.Username+"("+gconv.String(id)+") must not have active tenant membership")
			if len(details) >= 10 {
				break
			}
		}
	}
	return details, nil
}

// getVisible retrieves one membership by primary key within the visible tenant.
func (s *serviceImpl) getVisible(ctx context.Context, id int64, tenantID int64) (*Entity, error) {
	var item *Entity
	if err := s.visibleMembershipModel(ctx, tenantID).Where("id", id).Scan(&item); err != nil {
		return nil, err
	}
	if item == nil {
		return nil, bizerr.NewCode(CodeMembershipNotFound)
	}
	return item, nil
}

// ensureUserCanJoinTenant enforces platform-user and cardinality membership rules.
func (s *serviceImpl) ensureUserCanJoinTenant(
	ctx context.Context,
	userID int64,
	replacementTenantIDs []tenantcap.TenantID,
) error {
	if s == nil || s.users == nil {
		return bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", "user"))
	}
	userDomainID := usercap.UserID(strconv.FormatInt(userID, 10))
	out, err := s.users.BatchGet(ctx, s.capabilityContext(ctx, "membership.user_visibility"), []usercap.UserID{userDomainID})
	if err != nil {
		return err
	}
	user := (*usercap.UserProjection)(nil)
	if out != nil {
		user = out.Items[userDomainID]
	}
	if user == nil {
		return bizerr.NewCode(tenantcap.CodeTenantForbidden, bizerr.P("tenantId", int(firstTenantIDOrPlatform(replacementTenantIDs))))
	}
	if user.TenantID == capmodel.DomainID(strconv.FormatInt(shared.PlatformTenantID, 10)) && len(replacementTenantIDs) > 0 {
		return bizerr.NewCode(CodePlatformMembershipForbidden)
	}
	if membershipCardinalityMode() != shared.SingleCardinality {
		return nil
	}
	if len(replacementTenantIDs) > 1 {
		return bizerr.NewCode(CodeSingleCardinalityViolation)
	}
	return nil
}

// ensureTenantAcceptsMembership verifies membership writes target an active tenant.
func (s *serviceImpl) ensureTenantAcceptsMembership(ctx context.Context, tenantID int64) error {
	if tenantID <= shared.PlatformTenantID {
		return bizerr.NewCode(CodeTenantUnavailable)
	}
	var tenant *struct {
		Status string `json:"status" orm:"status"`
	}
	if err := shared.Model(ctx, shared.TableTenant).Fields("status").Where("id", tenantID).Scan(&tenant); err != nil {
		return err
	}
	if tenant == nil || tenant.Status != string(shared.TenantStatusActive) {
		return bizerr.NewCode(CodeTenantUnavailable)
	}
	return nil
}

// ensureTenantIDsAcceptMembership verifies all requested tenant IDs are active.
func (s *serviceImpl) ensureTenantIDsAcceptMembership(ctx context.Context, tenantIDs []tenantcap.TenantID) error {
	for _, tenantID := range tenantIDs {
		if err := s.ensureTenantAcceptsMembership(ctx, int64(tenantID)); err != nil {
			return err
		}
	}
	return nil
}

// effectiveTenantID returns the current request tenant as the authority for
// tenant-scoped member operations. Platform requests may still specify a tenant
// explicitly for administration and E2E setup.
func (s *serviceImpl) effectiveTenantID(ctx context.Context, requestedTenantID int64) int64 {
	currentTenantID := int64(s.bizCtxSvc.Current(ctx).TenantID)
	if currentTenantID > shared.PlatformTenantID {
		return currentTenantID
	}
	return requestedTenantID
}

// visibleMembershipModel builds a membership model constrained by tenant when
// the current request is tenant-scoped.
func (s *serviceImpl) visibleMembershipModel(ctx context.Context, tenantID int64) *gdb.Model {
	model := shared.Model(ctx, shared.TableMembership)
	if tenantID > shared.PlatformTenantID {
		model = model.Where("tenant_id", tenantID)
	}
	return model
}

// membershipCardinalityMode returns the code-owned tenant cardinality default.
// The default is multi, meaning one global user identity may belong to multiple
// tenants. A future management setting can switch this to single without
// reintroducing host config-file coupling.
func membershipCardinalityMode() string {
	return shared.DefaultCardinality
}

// membershipListModel builds the shared member list query without a projection
// so Count can generate valid SQL.
func membershipListModel(ctx context.Context, in ListInput) *gdb.Model {
	model := shared.Model(ctx, shared.TableMembership).As("m")
	if in.TenantID > 0 {
		model = model.Where("m.tenant_id", in.TenantID)
	}
	if in.UserID > 0 {
		model = model.Where("m.user_id", in.UserID)
	}
	if in.Status >= 0 {
		model = model.Where("m.status", in.Status)
	}
	return model
}

// hydrateUserLabels resolves current-page user names through usercap in one
// batch and leaves labels empty for missing or invisible users.
func (s *serviceImpl) hydrateUserLabels(ctx context.Context, list []*Entity) error {
	if len(list) == 0 {
		return nil
	}
	if s == nil || s.users == nil {
		return bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", "user"))
	}
	ids := make([]usercap.UserID, 0, len(list))
	seen := make(map[usercap.UserID]struct{}, len(list))
	for _, item := range list {
		if item == nil || item.UserID <= 0 {
			continue
		}
		id := usercap.UserID(strconv.FormatInt(item.UserID, 10))
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	out, err := s.users.BatchGet(ctx, s.capabilityContext(ctx, "membership.list"), ids)
	if err != nil {
		return err
	}
	if out == nil {
		return nil
	}
	for _, item := range list {
		if item == nil {
			continue
		}
		projection := out.Items[usercap.UserID(strconv.FormatInt(item.UserID, 10))]
		if projection == nil {
			continue
		}
		item.Username = projection.Username
		item.Nickname = projection.Nickname
	}
	return nil
}

// capabilityContext creates plugin-visible metadata for tenant membership
// calls into host-owned domain capabilities.
func (s *serviceImpl) capabilityContext(ctx context.Context, resource string) capmodel.CapabilityContext {
	current := bizctxcap.CurrentContext{}
	if s != nil && s.bizCtxSvc != nil {
		current = s.bizCtxSvc.Current(ctx)
	}
	actorID := current.ActingUserID
	if actorID == 0 {
		actorID = current.UserID
	}
	actor := capmodel.CapabilityActor{
		Type:   capmodel.ActorTypeUser,
		UserID: int64(actorID),
		Name:   current.Username,
	}
	if actorID == 0 {
		actor = capmodel.CapabilityActor{
			Type:         capmodel.ActorTypeSystem,
			Name:         membershipCapabilityPluginID,
			SystemReason: "tenant membership user projection",
		}
	}
	return capmodel.CapabilityContext{
		PluginID:    membershipCapabilityPluginID,
		Actor:       actor,
		TenantID:    capmodel.DomainID(strconv.Itoa(current.TenantID)),
		Source:      capmodel.CapabilitySourceHTTP,
		SystemCall:  actor.Type == capmodel.ActorTypeSystem,
		Resource:    resource,
		RequestedAt: time.Now(),
	}
}

// activeMembershipUserModel returns the subquery for active users in one tenant.
func activeMembershipUserModel(ctx context.Context, tenantID int64) *gdb.Model {
	return dao.UserMembership.Ctx(ctx).
		Fields("user_id").
		Where("tenant_id", tenantID).
		Where("status", shared.MembershipStatusEnabled)
}

// normalizeTenantIDs returns positive unique tenant IDs while preserving order.
func normalizeTenantIDs(tenantIDs []tenantcap.TenantID) []tenantcap.TenantID {
	normalized := make([]tenantcap.TenantID, 0, len(tenantIDs))
	seen := make(map[tenantcap.TenantID]struct{}, len(tenantIDs))
	for _, tenantID := range tenantIDs {
		if tenantID <= tenantcap.PLATFORM {
			continue
		}
		if _, ok := seen[tenantID]; ok {
			continue
		}
		seen[tenantID] = struct{}{}
		normalized = append(normalized, tenantID)
	}
	return normalized
}

// firstTenantIDOrPlatform returns the primary tenant for user assignment plans.
func firstTenantIDOrPlatform(tenantIDs []tenantcap.TenantID) tenantcap.TenantID {
	if len(tenantIDs) == 0 {
		return tenantcap.PLATFORM
	}
	return tenantIDs[0]
}

// normalizeUserIDs removes invalid and duplicate user IDs while preserving order.
func normalizeUserIDs(userIDs []int) []int {
	normalized := make([]int, 0, len(userIDs))
	seen := make(map[int]struct{}, len(userIDs))
	for _, userID := range userIDs {
		if userID <= 0 {
			continue
		}
		if _, ok := seen[userID]; ok {
			continue
		}
		seen[userID] = struct{}{}
		normalized = append(normalized, userID)
	}
	return normalized
}
