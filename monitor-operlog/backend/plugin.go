// Package backend wires the monitor-operlog source plugin into the host plugin registry.
package backend

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/pluginhost"
	monitoroperlogplugin "lina-plugin-monitor-operlog"
	operlogcontroller "lina-plugin-monitor-operlog/backend/internal/controller/operlog"
	middlewaresvc "lina-plugin-monitor-operlog/backend/internal/service/middleware"
	operlogsvc "lina-plugin-monitor-operlog/backend/internal/service/operlog"
)

// monitor-operlog plugin constants.
const (
	// pluginID is the immutable identifier published by the embedded source plugin.
	pluginID = "monitor-operlog"
)

// init registers the monitor-operlog source plugin and its host callbacks.
func init() {
	plugin := pluginhost.NewSourcePlugin(pluginID)
	plugin.Assets().UseEmbeddedFiles(monitoroperlogplugin.EmbeddedFiles)
	if err := plugin.HTTP().RegisterRoutes(
		pluginhost.ExtensionPointHTTPRouteRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerRoutes,
	); err != nil {
		panic(err)
	}
	if err := pluginhost.RegisterSourcePlugin(plugin); err != nil {
		panic(err)
	}
}

// registerRoutes binds operation-log governance routes and audit middleware through the published host HTTP registrars.
func registerRoutes(ctx context.Context, registrar pluginhost.HTTPRegistrar) error {
	hostServices := registrar.HostServices()
	if hostServices == nil ||
		hostServices.APIDoc() == nil ||
		hostServices.BizCtx() == nil ||
		hostServices.I18n() == nil ||
		hostServices.Route() == nil ||
		hostServices.TenantFilter() == nil {
		return gerror.New("monitor-operlog routes require host apidoc, bizctx, i18n, route, and tenant-filter services")
	}
	operLogSvc := operlogsvc.New(hostServices.APIDoc(), hostServices.I18n(), hostServices.TenantFilter())
	auditMiddlewareSvc := middlewaresvc.New(hostServices.Route(), hostServices.BizCtx(), operLogSvc)
	if err := registrar.GlobalMiddlewares().Bind("/*", auditMiddlewareSvc.Audit); err != nil {
		return err
	}

	var (
		routes      = registrar.Routes()
		middlewares = routes.Middlewares()
	)
	routes.Group("/api/v1", func(group pluginhost.RouteGroup) {
		group.Middleware(
			middlewares.NeverDoneCtx(),
			middlewares.HandlerResponse(),
			middlewares.CORS(),
			middlewares.RequestBodyLimit(),
			middlewares.Ctx(),
		)
		group.Group("/", func(group pluginhost.RouteGroup) {
			group.Middleware(
				middlewares.Auth(),
				middlewares.Tenancy(),
				middlewares.Permission(),
			)
			group.Bind(operlogcontroller.NewV1(operLogSvc))
		})
	})
	return nil
}
