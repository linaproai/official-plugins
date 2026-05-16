// Package tenant implements tenant CRUD and lifecycle state transitions for
// the multi-tenant source plugin.
package tenant

import (
	"context"
	"time"

	plugincontract "lina-core/pkg/pluginservice/contract"
	"lina-plugin-multi-tenant/backend/internal/service/resolverconfig"
	"lina-plugin-multi-tenant/backend/internal/service/shared"
	"lina-plugin-multi-tenant/backend/internal/service/tenantplugin"
)

// Tenant code validation and tombstone retention constants.
const (
	tenantCodeMinLength      = 2
	tenantCodeMaxLength      = 32
	tenantTombstoneRetention = 30 * 24 * time.Hour
)

// Service defines tenant CRUD and lifecycle management for plugin-owned tenant rows.
type Service interface {
	// List queries tenants with pagination and code/name/status filters. It is
	// read-only and returns database errors.
	List(ctx context.Context, in ListInput) (*ListOutput, error)
	// Get retrieves one tenant by primary key or returns CodeTenantNotFound when absent.
	Get(ctx context.Context, id int64) (*Entity, error)
	// Create creates one tenant, validates reserved codes, records creator metadata
	// from ctx, and provisions built-in tenant plugin state. It returns business or
	// database errors from validation, insert, or provisioning.
	Create(ctx context.Context, in CreateInput) (int64, error)
	// Update updates tenant basic fields for an existing tenant and returns
	// validation or database errors.
	Update(ctx context.Context, in UpdateInput) error
	// ChangeStatus performs an allowed lifecycle status transition and rejects
	// invalid transitions with business errors.
	ChangeStatus(ctx context.Context, id int64, status shared.TenantStatus) error
	// Delete soft-deletes a tenant after plugin lifecycle preconditions allow it.
	// It notifies host plugin lifecycle hooks but does not alter i18n resources.
	Delete(ctx context.Context, id int64) error
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	bizCtxSvc          plugincontract.BizCtxService
	resolverConfigSvc  resolverconfig.Service
	tenantPluginSvc    tenantplugin.Service
	pluginLifecycleSvc plugincontract.PluginLifecycleService
}

// New creates and returns a new tenant Service instance.
func New(
	bizCtxSvc plugincontract.BizCtxService,
	resolverConfigSvc resolverconfig.Service,
	tenantPluginSvc tenantplugin.Service,
	pluginLifecycleSvc plugincontract.PluginLifecycleService,
) Service {
	return &serviceImpl{
		bizCtxSvc:          bizCtxSvc,
		resolverConfigSvc:  resolverConfigSvc,
		tenantPluginSvc:    tenantPluginSvc,
		pluginLifecycleSvc: pluginLifecycleSvc,
	}
}

// Entity is the service-layer tenant projection.
type Entity struct {
	Id        int64  `json:"id" orm:"id"`
	Code      string `json:"code" orm:"code"`
	Name      string `json:"name" orm:"name"`
	Status    string `json:"status" orm:"status"`
	Remark    string `json:"remark" orm:"remark"`
	CreatedBy int64  `json:"createdBy" orm:"created_by"`
	UpdatedBy int64  `json:"updatedBy" orm:"updated_by"`
	CreatedAt string `json:"createdAt" orm:"created_at"`
	UpdatedAt string `json:"updatedAt" orm:"updated_at"`
}

// tenantCodeRow is the code lookup projection that includes soft-deleted rows.
type tenantCodeRow struct {
	Id        int64      `json:"id" orm:"id"`
	Code      string     `json:"code" orm:"code"`
	DeletedAt *time.Time `json:"deletedAt" orm:"deleted_at"`
}

// ListInput defines tenant list filters.
type ListInput struct {
	PageNum  int
	PageSize int
	Code     string
	Name     string
	Status   string
}

// ListOutput defines tenant list output.
type ListOutput struct {
	List  []*Entity
	Total int
}

// CreateInput defines tenant creation fields.
type CreateInput struct {
	Code   string
	Name   string
	Remark string
}

// UpdateInput defines tenant update fields.
type UpdateInput struct {
	Id     int64
	Name   *string
	Remark *string
}

// tenantInsertData is a typed insert payload for plugin_multi_tenant_tenant.
type tenantInsertData struct {
	Code      string `orm:"code"`
	Name      string `orm:"name"`
	Status    string `orm:"status"`
	Remark    string `orm:"remark"`
	CreatedBy int64  `orm:"created_by"`
	UpdatedBy int64  `orm:"updated_by"`
}

// tenantUpdateData is a typed update payload for plugin_multi_tenant_tenant.
type tenantUpdateData struct {
	Name      any   `orm:"name,omitempty"`
	Status    any   `orm:"status,omitempty"`
	Remark    any   `orm:"remark,omitempty"`
	UpdatedBy int64 `orm:"updated_by"`
}
