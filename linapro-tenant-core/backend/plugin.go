// Package backend wires the linapro-tenant-core source plugin into the host plugin registry.
package backend

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
	"lina-core/pkg/plugin/pluginhost"
	multitenant "lina-plugin-linapro-tenant-core"
	authcontroller "lina-plugin-linapro-tenant-core/backend/internal/controller/auth"
	platformcontroller "lina-plugin-linapro-tenant-core/backend/internal/controller/platform"
	tenantcontroller "lina-plugin-linapro-tenant-core/backend/internal/controller/tenant"
	"lina-plugin-linapro-tenant-core/backend/internal/provider/tenantadapter"
	"lina-plugin-linapro-tenant-core/backend/internal/service/impersonate"
	"lina-plugin-linapro-tenant-core/backend/internal/service/lifecycleprecondition"
	"lina-plugin-linapro-tenant-core/backend/internal/service/membership"
	"lina-plugin-linapro-tenant-core/backend/internal/service/provider"
	"lina-plugin-linapro-tenant-core/backend/internal/service/resolver"
	"lina-plugin-linapro-tenant-core/backend/internal/service/resolverconfig"
	tenantsvc "lina-plugin-linapro-tenant-core/backend/internal/service/tenant"
	"lina-plugin-linapro-tenant-core/backend/internal/service/tenantplugin"
)

// linapro-tenant-core plugin constants.
const (
	// pluginID is the immutable identifier published by the embedded source plugin.
	pluginID = "linapro-tenant-core"
)

// init registers the linapro-tenant-core source plugin and its host callbacks.
func init() {
	plugin := pluginhost.NewDeclarations(pluginID)
	plugin.Assets().UseEmbeddedFiles(multitenant.EmbeddedFiles)
	if err := plugin.Providers().ProvideTenant(provideTenant); err != nil {
		panic(err)
	}
	if err := plugin.HTTP().RegisterRoutes(
		pluginhost.ExtensionPointHTTPRouteRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerRoutes,
	); err != nil {
		panic(err)
	}
	if err := plugin.Lifecycle().RegisterBeforeDisableHandler(beforeDisable); err != nil {
		panic(err)
	}
	if err := plugin.Lifecycle().RegisterBeforeUninstallHandler(beforeUninstall); err != nil {
		panic(err)
	}
	if err := plugin.Lifecycle().RegisterBeforeTenantDeleteHandler(beforeTenantDelete); err != nil {
		panic(err)
	}
	if err := pluginhost.RegisterSourcePlugin(plugin); err != nil {
		panic(err)
	}
}

// provideTenant creates the linapro-tenant-core tenant capability adapter from
// host-published services during framework capability activation.
func provideTenant(_ context.Context, env tenantspi.ProviderEnv) (tenantspi.Provider, error) {
	if env.BizCtx == nil {
		return nil, gerror.New("linapro-tenant-core provider requires host bizctx service")
	}
	return tenantadapter.New(env.BizCtx, env.Tenant, env.Users, env.Plugins)
}

// beforeDisable enforces linapro-tenant-core plugin disable preconditions.
func beforeDisable(ctx context.Context, input pluginhost.SourcePluginLifecycleInput) (bool, string, error) {
	precondition, err := newLifecyclePrecondition()
	if err != nil {
		return false, lifecycleprecondition.ReasonDisableTenantsExist, err
	}
	return precondition.BeforeDisable(ctx, input)
}

// beforeUninstall enforces linapro-tenant-core plugin uninstall preconditions.
func beforeUninstall(ctx context.Context, input pluginhost.SourcePluginLifecycleInput) (bool, string, error) {
	precondition, err := newLifecyclePrecondition()
	if err != nil {
		return false, lifecycleprecondition.ReasonUninstallTenantsExist, err
	}
	return precondition.BeforeUninstall(ctx, input)
}

// beforeTenantDelete runs linapro-tenant-core tenant deletion preconditions.
func beforeTenantDelete(
	ctx context.Context,
	input pluginhost.SourcePluginTenantLifecycleInput,
) (bool, string, error) {
	precondition, err := newLifecyclePrecondition()
	if err != nil {
		return false, "", err
	}
	return precondition.BeforeTenantDelete(ctx, input)
}

// newLifecyclePrecondition creates the plugin-owned lifecycle precondition
// checker from the tenant-counting dependency it requires.
func newLifecyclePrecondition() (*lifecycleprecondition.Checker, error) {
	return lifecycleprecondition.New(tenantsvc.ExistingCounter{})
}

// registerRoutes binds linapro-tenant-core routes through the published host middleware set.
func registerRoutes(ctx context.Context, registrar pluginhost.HTTPRegistrar) error {
	var (
		routes      = registrar.Routes()
		middlewares = routes.Middlewares()
		services    = registrar.Services()
	)
	if services == nil {
		return gerror.New("linapro-tenant-core routes require host auth, authz, bizctx, user, plugin lifecycle, and plugin capability services")
	}
	tenantDomain := services.Tenant()
	plugins := services.Plugins()
	if services.Auth() == nil ||
		services.Auth().Token() == nil ||
		services.Auth().Authz() == nil ||
		services.BizCtx() == nil ||
		services.Users() == nil ||
		tenantDomain == nil ||
		tenantDomain.Plugins() == nil ||
		plugins == nil ||
		plugins.Lifecycle() == nil {
		return gerror.New("linapro-tenant-core routes require host auth, authz, bizctx, user, plugin lifecycle, and plugin capability services")
	}
	var (
		membershipSvc     = membership.New(services.BizCtx(), services.Users())
		resolverConfigSvc = resolverconfig.New()
		tenantPluginSvc   = tenantplugin.New(services.BizCtx(), tenantDomain, plugins)
		tenantSvc         = tenantsvc.New(services.BizCtx(), resolverConfigSvc, tenantPluginSvc, plugins)
		resolverSvc       = resolver.New(services.BizCtx(), membershipSvc)
	)
	providerSvc, err := provider.New(membershipSvc, resolverSvc, resolverConfigSvc, tenantPluginSvc)
	if err != nil {
		return err
	}
	var (
		impersonateSvc = impersonate.New(services.Auth().Token(), services.Auth().Authz(), services.BizCtx(), tenantSvc, services.Users())
		authCtrl       = authcontroller.NewV1(services.Auth().Token(), membershipSvc, providerSvc)
	)
	routes.Group(routes.APIPrefix(), func(group pluginhost.RouteGroup) {
		group.Group("/api/v1", func(group pluginhost.RouteGroup) {
			group.Middleware(
				middlewares.NeverDoneCtx(),
				middlewares.HandlerResponse(),
				middlewares.CORS(),
				middlewares.RequestBodyLimit(),
				middlewares.Ctx(),
			)
			// SelectTenant is public; LoginTenants/SwitchTenant require auth.
			// Split method registration across middleware groups (same pattern as host).
			group.Group("/", func(group pluginhost.RouteGroup) {
				group.Bind(authCtrl.SelectTenant)
			})
			group.Group("/", func(group pluginhost.RouteGroup) {
				group.Middleware(
					middlewares.Auth(),
					middlewares.Tenancy(),
					middlewares.Permission(),
				)
				group.Bind(
					authCtrl.LoginTenants,
					authCtrl.SwitchTenant,
					platformcontroller.NewV1(tenantSvc, impersonateSvc),
					tenantcontroller.NewV1(tenantPluginSvc),
				)
			})
		})
	})
	return nil
}
