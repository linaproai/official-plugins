// Package backend wires the linapro-monitor-online source plugin into the host plugin registry.
package backend

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/pluginhost"
	monitoronlineplugin "lina-plugin-linapro-monitor-online"
	monitorcontroller "lina-plugin-linapro-monitor-online/backend/internal/controller/monitor"
	monitorsvc "lina-plugin-linapro-monitor-online/backend/internal/service/monitor"
)

// linapro-monitor-online plugin constants.
const (
	// pluginID is the immutable identifier published by the embedded source plugin.
	pluginID = "linapro-monitor-online"
)

// init registers the linapro-monitor-online source plugin and its host callbacks.
func init() {
	plugin := pluginhost.NewSourcePlugin(pluginID)
	plugin.Assets().UseEmbeddedFiles(monitoronlineplugin.EmbeddedFiles)
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

// registerRoutes binds online-user governance routes through the published host middleware set.
func registerRoutes(ctx context.Context, registrar pluginhost.HTTPRegistrar) error {
	var (
		routes      = registrar.Routes()
		middlewares = routes.Middlewares()
		services    = registrar.Services()
	)
	if services == nil ||
		services.BizCtx() == nil ||
		services.Sessions() == nil ||
		services.Admin() == nil ||
		services.Admin().Sessions() == nil ||
		services.TenantFilter() == nil {
		return gerror.New("linapro-monitor-online routes require host bizctx, session domain, session admin, and tenant-filter services")
	}
	monitorSvc := monitorsvc.New(
		services.BizCtx(),
		services.TenantFilter(),
		services.Sessions(),
		services.Admin().Sessions(),
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
			group.Group("/", func(group pluginhost.RouteGroup) {
				group.Middleware(
					middlewares.Auth(),
					middlewares.Tenancy(),
					middlewares.Permission(),
				)
				group.Bind(monitorcontroller.NewV1(monitorSvc))
			})
		})
	})
	return nil
}
