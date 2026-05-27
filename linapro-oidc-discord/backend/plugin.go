// Package backend wires the linapro-oidc-discord source plugin into the
// host plugin registry. The plugin no longer owns a private SQL config
// table; all runtime settings flow through the host's PluginSettingsService
// and are stored in sys_config under the "linapro-oidc-discord." namespace.
package backend

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/capability/authprovider"
	"lina-core/pkg/plugin/pluginhost"
	discordplugin "lina-plugin-linapro-oidc-discord"
	oauthctrl "lina-plugin-linapro-oidc-discord/backend/internal/controller/oauth"
	settingsctrl "lina-plugin-linapro-oidc-discord/backend/internal/controller/settings"
	configsvc "lina-plugin-linapro-oidc-discord/backend/internal/service/config"
	discordprovider "lina-plugin-linapro-oidc-discord/backend/internal/service/provider"
)

// pluginID is the canonical source-plugin identifier for the Discord
// OAuth2 auth provider.
const pluginID = "linapro-oidc-discord"

// init registers the plugin lifecycle and HTTP routes. The auth provider
// is registered later inside registerRoutes because it needs the host's
// PluginSettingsService.
func init() {
	plugin := pluginhost.NewSourcePlugin(pluginID)
	plugin.Assets().UseEmbeddedFiles(discordplugin.EmbeddedFiles)
	if err := plugin.HTTP().RegisterRoutes(pluginhost.ExtensionPointHTTPRouteRegister, pluginhost.CallbackExecutionModeBlocking, registerRoutes); err != nil {
		panic(err)
	}
	if err := pluginhost.RegisterSourcePlugin(plugin); err != nil {
		panic(err)
	}
}

// registerRoutes binds the plugin's HTTP endpoints. It also constructs the
// settings service from the host's namespaced key-value store and registers
// the auth provider so the workbench login page can discover the entry.
func registerRoutes(ctx context.Context, registrar pluginhost.HTTPRegistrar) error {
	routes := registrar.Routes()
	middlewares := routes.Middlewares()
	hostServices := registrar.Services()
	if hostServices == nil {
		return gerror.New("linapro-oidc-discord routes require host services")
	}
	authSvc := hostServices.Auth()
	if authSvc == nil {
		return gerror.New("linapro-oidc-discord routes require host auth service")
	}
	settingsClient := hostServices.PluginSettings()
	if settingsClient == nil {
		return gerror.New("linapro-oidc-discord routes require host plugin settings service")
	}
	settingsSvc := configsvc.New(settingsClient)
	settingsCtrl := settingsctrl.NewV1(settingsSvc)
	oauthCtrl := oauthctrl.New(authSvc, settingsSvc)
	authprovider.RegisterProvider(discordprovider.New(settingsSvc))
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
				group.Bind(settingsCtrl.GetSettings, settingsCtrl.SaveSettings)
			})
		})
	})
	// OAuth routes live under /api/v1/auth/* so deployments that already
	// proxy /api/* to the backend pick them up without extra reverse proxy
	// configuration. The redirect URI registered with Discord must match
	// this path; the settings page surfaces it automatically.
	routes.Group("/api/v1/auth", func(group pluginhost.RouteGroup) {
		group.Middleware(
			middlewares.NeverDoneCtx(),
			middlewares.CORS(),
			middlewares.Ctx(),
		)
		group.GET("/discord", oauthCtrl.StartLogin)
		group.GET("/discord/callback", oauthCtrl.HandleCallback)
	})
	_ = ctx
	return nil
}
