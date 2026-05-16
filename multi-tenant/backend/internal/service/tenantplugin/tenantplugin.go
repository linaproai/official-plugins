// Package tenantplugin implements tenant-scoped plugin enablement governance.
package tenantplugin

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	plugincontract "lina-core/pkg/pluginservice/contract"
)

const (
	// pluginInstalledYes is the host sys_plugin installed flag for installed plugins.
	pluginInstalledYes = 1
	// pluginStatusEnabled is the host sys_plugin status flag for enabled plugins.
	pluginStatusEnabled = 1
	// pluginTypeSource is the host sys_plugin type for source plugins.
	pluginTypeSource = "source"
	// pluginHostStateEnabled is the stable enabled lifecycle state name.
	pluginHostStateEnabled = "enabled"
	// pluginScopeNaturePlatformOnly marks platform-only plugins.
	pluginScopeNaturePlatformOnly = "platform_only"
	// pluginScopeNatureTenantAware marks tenant-aware plugins.
	pluginScopeNatureTenantAware = "tenant_aware"
	// pluginInstallModeGlobal marks globally enabled plugins.
	pluginInstallModeGlobal = "global"
	// pluginInstallModeTenantScoped marks tenant-controlled plugins.
	pluginInstallModeTenantScoped = "tenant_scoped"
	// tenantEnablementStateKey is the sys_plugin_state key for tenant plugin enablement.
	tenantEnablementStateKey = "__tenant_enabled__"
	// tenantPluginEnabledValue stores enabled tenant plugin state for diagnostics.
	tenantPluginEnabledValue = "enabled"
	// tenantPluginDisabledValue stores disabled tenant plugin state for diagnostics.
	tenantPluginDisabledValue = "disabled"
	// tableSysCacheRevision is the host shared cache-revision table.
	tableSysCacheRevision = "sys_cache_revision"
	// pluginRuntimeCacheDomain coordinates plugin runtime and menu derived caches.
	pluginRuntimeCacheDomain = "plugin-runtime"
	// pluginRuntimeCacheScopeGlobal invalidates every plugin-runtime cache scope.
	pluginRuntimeCacheScopeGlobal = "global"
	// tenantPluginRuntimeChangeReason records tenant plugin enablement changes.
	tenantPluginRuntimeChangeReason = "tenant_plugin_enablement_changed"
)

// Service defines tenant plugin-governance operations and cache revision updates.
type Service interface {
	// List returns tenant-controllable plugins with current tenant enablement for
	// ctx's tenant. It is read-only and returns database errors.
	List(ctx context.Context) (*ListOutput, error)
	// SetEnabled updates one tenant plugin enablement row for ctx's tenant, runs
	// lifecycle guards, and bumps the shared plugin-runtime cache revision.
	SetEnabled(ctx context.Context, pluginID string, enabled bool) error
	// ProvisionForTenant provisions default tenant plugin enablement for a new
	// tenant and bumps runtime cache revision through the shared revision table.
	ProvisionForTenant(ctx context.Context, tenantID int64) error
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	bizCtxSvc          plugincontract.BizCtxService
	pluginLifecycleSvc plugincontract.PluginLifecycleService
}

// New creates and returns a tenant plugin governance service.
func New(
	bizCtxSvc plugincontract.BizCtxService,
	pluginLifecycleSvc plugincontract.PluginLifecycleService,
) Service {
	return &serviceImpl{
		bizCtxSvc:          bizCtxSvc,
		pluginLifecycleSvc: pluginLifecycleSvc,
	}
}

// Entity is the tenant plugin-governance projection.
type Entity struct {
	Id            string
	Name          string
	Version       string
	Type          string
	Description   string
	Installed     int
	Enabled       int
	ScopeNature   string
	InstallMode   string
	TenantEnabled int
}

// ListOutput defines tenant plugin list output.
type ListOutput struct {
	List  []*Entity
	Total int
}

// pluginRuntimeCacheRevisionDO is the local DO payload for sys_cache_revision writes.
type pluginRuntimeCacheRevisionDO struct {
	g.Meta   `orm:"table:sys_cache_revision, do:true"`
	Id       any // Primary key ID
	TenantId any // Revision tenant scope, 0 means platform/global
	Domain   any // Cache domain
	Scope    any // Cache invalidation scope
	Revision any // Monotonic shared revision
	Reason   any // Latest change reason
}

// pluginRuntimeCacheRevisionRow projects one sys_cache_revision row.
type pluginRuntimeCacheRevisionRow struct {
	Id       int64 `json:"id" orm:"id"`
	Revision int64 `json:"revision" orm:"revision"`
}
