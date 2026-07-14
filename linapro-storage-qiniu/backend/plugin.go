// Package backend wires linapro-storage-qiniu into the host plugin registry and storagecap provider registry.
package backend

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/logger"
	"lina-core/pkg/plugin/capability/storagecap"
	"lina-core/pkg/plugin/pluginhost"
	pluginroot "lina-plugin-linapro-storage-qiniu"
	settingsctrl "lina-plugin-linapro-storage-qiniu/backend/internal/controller/settings"
	"lina-plugin-linapro-storage-qiniu/backend/internal/service/objstore"
	settingssvc "lina-plugin-linapro-storage-qiniu/backend/internal/service/settings"
)

const pluginID = "linapro-storage-qiniu"

func init() {
	if err := storagecap.Provide(pluginID, objstore.Factory); err != nil {
		panic(err)
	}
	plugin := pluginhost.NewDeclarations(pluginID)
	plugin.Assets().UseEmbeddedFiles(pluginroot.EmbeddedFiles)
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

func registerRoutes(ctx context.Context, registrar pluginhost.HTTPRegistrar) error {
	var (
		routes      = registrar.Routes()
		middlewares = routes.Middlewares()
		services    = registrar.Services()
	)
	if services == nil || services.HostConfig() == nil || services.HostConfig().SysConfig() == nil {
		return gerror.New("linapro-storage-qiniu routes require host sys_config capability")
	}
	logger.Infof(ctx, "linapro-storage-qiniu registering settings routes")
	pluginSettings := settingssvc.New(services.HostConfig().SysConfig())
	objstore.Global.Configure(pluginSettings)
	settingsController := settingsctrl.NewV1(pluginSettings, objstore.Tester{})

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
				group.Bind(settingsController)
			})
		})
	})
	return routes.Err()
}
