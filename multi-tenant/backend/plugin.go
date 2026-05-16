// Package backend wires the multi-tenant source plugin into the host plugin registry.
package backend

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/pluginhost"
	plugincontract "lina-core/pkg/pluginservice/contract"
	pkgtenantcap "lina-core/pkg/tenantcap"
	multitenant "lina-plugin-multi-tenant"
	authcontroller "lina-plugin-multi-tenant/backend/internal/controller/auth"
	platformcontroller "lina-plugin-multi-tenant/backend/internal/controller/platform"
	tenantcontroller "lina-plugin-multi-tenant/backend/internal/controller/tenant"
	"lina-plugin-multi-tenant/backend/internal/service/impersonate"
	"lina-plugin-multi-tenant/backend/internal/service/lifecycleprecondition"
	"lina-plugin-multi-tenant/backend/internal/service/membership"
	"lina-plugin-multi-tenant/backend/internal/service/provider"
	"lina-plugin-multi-tenant/backend/internal/service/resolver"
	"lina-plugin-multi-tenant/backend/internal/service/resolverconfig"
	tenantsvc "lina-plugin-multi-tenant/backend/internal/service/tenant"
	"lina-plugin-multi-tenant/backend/internal/service/tenantplugin"
)

// multi-tenant plugin constants.
const (
	// pluginID is the immutable identifier published by the embedded source plugin.
	pluginID = "multi-tenant"
)

// init registers the multi-tenant source plugin and its host callbacks.
func init() {
	plugin := pluginhost.NewSourcePlugin(pluginID)
	plugin.Assets().UseEmbeddedFiles(multitenant.EmbeddedFiles)
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

// beforeDisable enforces multi-tenant plugin disable preconditions.
func beforeDisable(ctx context.Context, input pluginhost.SourcePluginLifecycleInput) (bool, string, error) {
	precondition, err := newLifecyclePrecondition()
	if err != nil {
		return false, lifecycleprecondition.ReasonDisableTenantsExist, err
	}
	return precondition.BeforeDisable(ctx, input)
}

// beforeUninstall enforces multi-tenant plugin uninstall preconditions.
func beforeUninstall(ctx context.Context, input pluginhost.SourcePluginLifecycleInput) (bool, string, error) {
	precondition, err := newLifecyclePrecondition()
	if err != nil {
		return false, lifecycleprecondition.ReasonUninstallTenantsExist, err
	}
	return precondition.BeforeUninstall(ctx, input)
}

// beforeTenantDelete runs multi-tenant tenant deletion preconditions.
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
// checker from the same explicit dependencies used by route registration.
func newLifecyclePrecondition() (*lifecycleprecondition.Checker, error) {
	return lifecycleprecondition.New(newTenantService(nil)), nil
}

// newTenantService creates the plugin-owned tenant service used by lifecycle
// preconditions and HTTP route controllers.
func newTenantService(bizCtx plugincontract.BizCtxService) tenantsvc.Service {
	tenantPluginSvc := tenantplugin.New(bizCtx)
	return tenantsvc.New(bizCtx, resolverconfig.New(), tenantPluginSvc)
}

// registerRoutes binds multi-tenant routes through the published host middleware set.
func registerRoutes(ctx context.Context, registrar pluginhost.HTTPRegistrar) error {
	hostServices := registrar.HostServices()
	if hostServices == nil ||
		hostServices.Auth() == nil ||
		hostServices.BizCtx() == nil ||
		hostServices.Config() == nil {
		return gerror.New("multi-tenant routes require host auth, bizctx, and config services")
	}
	var (
		membershipSvc     = membership.New(hostServices.BizCtx())
		tenantPluginSvc   = tenantplugin.New(hostServices.BizCtx())
		tenantSvc         = tenantsvc.New(hostServices.BizCtx(), resolverconfig.New(), tenantPluginSvc)
		resolverConfigSvc = resolverconfig.New()
		resolverSvc       = resolver.New(hostServices.BizCtx(), membershipSvc)
		providerSvc       = provider.New(membershipSvc, resolverSvc, resolverConfigSvc)
	)
	pkgtenantcap.RegisterProvider(providerSvc)
	var (
		impersonateSvc = impersonate.New(hostServices.BizCtx(), hostServices.Config(), tenantSvc)
		routes         = registrar.Routes()
		middlewares    = routes.Middlewares()
	)
	routes.Group("/api/v1", func(group pluginhost.RouteGroup) {
		authCtrl := authcontroller.NewV1(hostServices.Auth(), membershipSvc, providerSvc)
		group.Middleware(
			middlewares.NeverDoneCtx(),
			middlewares.HandlerResponse(),
			middlewares.CORS(),
			middlewares.RequestBodyLimit(),
			middlewares.Ctx(),
		)
		group.Group("/", func(group pluginhost.RouteGroup) {
			group.Bind(
				authCtrl.SelectTenant,
			)
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
	return nil
}
