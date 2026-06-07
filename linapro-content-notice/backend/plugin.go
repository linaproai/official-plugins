// Package backend wires the linapro-content-notice source plugin into the host plugin registry.
package backend

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/pluginhost"
	contentnotice "lina-plugin-linapro-content-notice"
	noticecontroller "lina-plugin-linapro-content-notice/backend/internal/controller/notice"
	noticesvc "lina-plugin-linapro-content-notice/backend/internal/service/notice"
)

// linapro-content-notice plugin constants.
const (
	// pluginID is the immutable identifier published by the embedded source plugin.
	pluginID = "linapro-content-notice"
)

// init registers the linapro-content-notice source plugin and its host callbacks.
func init() {
	plugin := pluginhost.NewSourcePlugin(pluginID)
	plugin.Assets().UseEmbeddedFiles(contentnotice.EmbeddedFiles)
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

// registerRoutes binds notice-management routes through the published host middleware set.
func registerRoutes(ctx context.Context, registrar pluginhost.HTTPRegistrar) error {
	var (
		routes      = registrar.Routes()
		middlewares = routes.Middlewares()
		services    = registrar.Services()
	)
	if services == nil ||
		services.BizCtx() == nil ||
		services.Admin() == nil ||
		services.Admin().Notifications() == nil ||
		services.TenantFilter() == nil ||
		services.Users() == nil {
		return gerror.New("linapro-content-notice routes require host bizctx, notification admin, tenant-filter, and user capability services")
	}
	noticeSvc := noticesvc.New(
		services.BizCtx(),
		services.Admin().Notifications(),
		services.TenantFilter(),
		services.Users(),
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
				group.Bind(noticecontroller.NewV1(noticeSvc))
			})
		})
	})

	return nil
}
