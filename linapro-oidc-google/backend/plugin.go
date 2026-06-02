// Package backend wires the linapro-oidc-google source plugin into the host
// plugin registry. The plugin no longer owns a private SQL config table;
// all runtime settings flow through the host's PluginSettingsService and
// are stored in sys_config under the "linapro-oidc-google." namespace.
package backend

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/capability/authprovider"
	"lina-core/pkg/plugin/pluginhost"
	googleplugin "lina-plugin-linapro-oidc-google"
	oauthctrl "lina-plugin-linapro-oidc-google/backend/internal/controller/oauth"
	settingsctrl "lina-plugin-linapro-oidc-google/backend/internal/controller/settings"
	configsvc "lina-plugin-linapro-oidc-google/backend/internal/service/config"
	googleprovider "lina-plugin-linapro-oidc-google/backend/internal/service/provider"
)

// pluginID is the canonical source-plugin identifier for the Google OIDC
// auth provider. Declared here so the lifecycle registration block stays
// independent of the configsvc package import order.
const pluginID = "linapro-oidc-google"

// init registers the plugin lifecycle and HTTP routes. The auth provider
// is registered later inside registerRoutes because it needs the host's
// PluginSettingsService, which is only available through the HTTPRegistrar.
func init() {
	plugin := pluginhost.NewSourcePlugin(pluginID)
	plugin.Assets().UseEmbeddedFiles(googleplugin.EmbeddedFiles)
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
//
// The ctx parameter is required by the host HTTPRegistrar callback contract
// but is unused during route registration because all I/O happens per-request
// later. It is left as a blank identifier so the linter does not need a
// `_ = ctx` discard at the end of the function body.
func registerRoutes(_ context.Context, registrar pluginhost.HTTPRegistrar) error {
	routes := registrar.Routes()
	middlewares := routes.Middlewares()
	hostServices := registrar.Services()
	if hostServices == nil {
		return gerror.New("linapro-oidc-google routes require host services")
	}
	authSvc := hostServices.Auth()
	if authSvc == nil {
		return gerror.New("linapro-oidc-google routes require host auth service")
	}
	settingsClient := hostServices.PluginSettings()
	if settingsClient == nil {
		return gerror.New("linapro-oidc-google routes require host plugin settings service")
	}
	pluginState := hostServices.PluginState()
	if pluginState == nil {
		return gerror.New("linapro-oidc-google routes require host plugin state service")
	}
	settingsSvc := configsvc.New(settingsClient)
	settingsCtrl := settingsctrl.NewV1(settingsSvc)
	oauthCtrl := oauthctrl.New(authSvc, settingsSvc, pluginState)
	authprovider.RegisterProvider(googleprovider.New(settingsSvc))
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
	// configuration. The redirect URI registered with Google must match
	// this path; the settings page surfaces it automatically.
	routes.Group("/api/v1/auth", func(group pluginhost.RouteGroup) {
		group.Middleware(
			middlewares.NeverDoneCtx(),
			middlewares.CORS(),
			middlewares.Ctx(),
		)
		group.GET("/google", oauthCtrl.StartLogin)
		group.GET("/google/callback", oauthCtrl.HandleCallback)
	})
	return nil
}
