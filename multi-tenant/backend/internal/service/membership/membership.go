// Package membership implements user-to-tenant membership management for the
// multi-tenant source plugin.
package membership

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"

	plugincontract "lina-core/pkg/pluginservice/contract"
	pkgtenantcap "lina-core/pkg/tenantcap"
)

// Service defines tenant membership operations and the host user-scope provider seam.
type Service interface {
	// List queries tenant members by page using explicit tenant/user/status filters.
	// It returns database errors and does not mutate membership state.
	List(ctx context.Context, in ListInput) (*ListOutput, error)
	// Add adds a user to one tenant after tenant lifecycle and cardinality checks.
	// It returns the membership ID or a business/database error.
	Add(ctx context.Context, in AddInput) (int64, error)
	// Update updates membership status fields for a visible membership row and
	// returns business/database errors when the row or target tenant is invalid.
	Update(ctx context.Context, in UpdateInput) error
	// Remove deletes one visible membership row by ID. Missing or out-of-scope
	// memberships are reported as business errors.
	Remove(ctx context.Context, id int64) error
	// Current returns the current user's membership in ctx's tenant or nil when
	// the user is not actively assigned.
	Current(ctx context.Context) (*Entity, error)
	// GetByUserAndTenant returns one membership for a user and tenant, enforcing
	// plugin membership visibility rules and returning nil when absent.
	GetByUserAndTenant(ctx context.Context, userID int64, tenantID int64) (*Entity, error)
	// ListUserTenants returns enabled tenant memberships for one user. It is used
	// by resolver/provider flows and returns only active tenant projections.
	ListUserTenants(ctx context.Context, userID int64) ([]*TenantInfo, error)
	// ApplyUserTenantScope constrains user rows by active current-tenant membership.
	// It returns the possibly modified model, whether a scope was applied, and errors.
	ApplyUserTenantScope(ctx context.Context, model *gdb.Model, userIDColumn string) (*gdb.Model, bool, error)
	// ApplyUserTenantFilter constrains platform user-list rows to a requested tenant.
	// It returns the possibly modified model, whether a filter was applied, and errors.
	ApplyUserTenantFilter(ctx context.Context, model *gdb.Model, userIDColumn string, tenantID pkgtenantcap.TenantID) (*gdb.Model, bool, error)
	// ListUserTenantProjections returns tenant ownership labels for visible users
	// without changing host user data or i18n resources.
	ListUserTenantProjections(ctx context.Context, userIDs []int) (map[int]*pkgtenantcap.UserTenantProjection, error)
	// ResolveUserTenantAssignment validates requested memberships and returns a
	// host write plan. It does not persist changes itself.
	ResolveUserTenantAssignment(ctx context.Context, requested []pkgtenantcap.TenantID, mode pkgtenantcap.UserTenantAssignmentMode) (*pkgtenantcap.UserTenantAssignmentPlan, error)
	// ReplaceUserTenantAssignments rewrites one user's active tenant ownership rows
	// from a previously resolved plan and returns business/database errors.
	ReplaceUserTenantAssignments(ctx context.Context, userID int, plan *pkgtenantcap.UserTenantAssignmentPlan) error
	// EnsureUsersInTenant verifies every user has active membership in the tenant,
	// returning a business error when any requested user is outside scope.
	EnsureUsersInTenant(ctx context.Context, userIDs []int, tenantID pkgtenantcap.TenantID) error
	// ValidateStartupConsistency returns user-membership startup consistency failures.
	// It is read-only and returns database errors separately from warning strings.
	ValidateStartupConsistency(ctx context.Context) ([]string, error)
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	bizCtxSvc plugincontract.BizCtxService
}

// New creates and returns a new membership Service instance.
func New(bizCtxSvc plugincontract.BizCtxService) Service {
	return &serviceImpl{bizCtxSvc: bizCtxSvc}
}

// Entity is the service-layer membership projection.
type Entity struct {
	Id       int64  `json:"id" orm:"id"`
	UserID   int64  `json:"userId" orm:"user_id"`
	TenantID int64  `json:"tenantId" orm:"tenant_id"`
	Status   int    `json:"status" orm:"status"`
	JoinedAt string `json:"joinedAt" orm:"joined_at"`
	Username string `json:"username" orm:"username"`
	Nickname string `json:"nickname" orm:"nickname"`
}

// TenantInfo is a user-facing tenant projection.
type TenantInfo struct {
	Id     int64  `json:"id" orm:"id"`
	Code   string `json:"code" orm:"code"`
	Name   string `json:"name" orm:"name"`
	Status string `json:"status" orm:"status"`
}

// userTenantProjectionRow is the joined membership and tenant display
// projection consumed by host user-list screens.
type userTenantProjectionRow struct {
	UserID     int    `json:"userId" orm:"user_id"`
	TenantID   int    `json:"tenantId" orm:"tenant_id"`
	TenantName string `json:"tenantName" orm:"tenant_name"`
}

// membershipTenantRow is the joined membership and tenant lifecycle
// projection used for login and tenant-switch authorization.
type membershipTenantRow struct {
	Id       int64  `json:"id" orm:"id"`
	Status   int    `json:"status" orm:"status"`
	TenantID int64  `json:"tenantId" orm:"tenant_id"`
	TStatus  string `json:"tenantStatus" orm:"tenant_status"`
}

// ListInput defines member list filters.
type ListInput struct {
	PageNum  int
	PageSize int
	TenantID int64
	UserID   int64
	Status   int
}

// ListOutput defines member list output.
type ListOutput struct {
	List  []*Entity
	Total int
}

// AddInput defines membership creation fields.
type AddInput struct {
	TenantID int64
	UserID   int64
}

// UpdateInput defines membership update fields.
type UpdateInput struct {
	Id     int64
	Status *int
}

// sysUserTenantRow is the sys_user tenant projection needed by membership rules.
type sysUserTenantRow struct {
	Id       int64 `json:"id" orm:"id"`
	TenantID int64 `json:"tenantId" orm:"tenant_id"`
}
