// Package backend wires the linapro-oidc-google source plugin into the host
// plugin registry. It declares the external-identity provider ID this plugin
// owns and registers the browser-facing OIDC login routes on the plugin's
// portal route group.
package backend

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/logger"
	"lina-core/pkg/plugin/pluginhost"
	pluginoidcgoogle "lina-plugin-linapro-oidc-google"
	loginctrl "lina-plugin-linapro-oidc-google/backend/internal/controller/login"
	settingsctrl "lina-plugin-linapro-oidc-google/backend/internal/controller/settings"
	oauthsvc "lina-plugin-linapro-oidc-google/backend/internal/service/oauth"
	settingssvc "lina-plugin-linapro-oidc-google/backend/internal/service/settings"
)

// pluginID is the immutable identifier declared by the embedded plugin.yaml.
const pluginID = "linapro-oidc-google"

// portalGroupPath is the browser-facing route group mounted outside the
// reserved /x API namespace. Login-start and callback routes must be public
// so unauthenticated browsers can complete the OIDC handshake.
const portalGroupPath = "/portal/linapro-oidc-google"

// init registers the linapro-oidc-google source plugin, declares its
// external-identity provider ownership, and registers HTTP routes.
func init() {
	plugin := pluginhost.NewDeclarations(pluginID)
	plugin.Assets().UseEmbeddedFiles(pluginoidcgoogle.EmbeddedFiles)
	// ProvideExternalIdentity establishes provider ownership so the host can
	// reject LoginByVerifiedIdentity requests that claim providers this plugin
	// does not own. This is an allowed top-level static registration panic.
	if err := plugin.Providers().ProvideExternalIdentity(oauthsvc.Provider); err != nil {
		panic(err)
	}
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

// registerRoutes registers the plugin HTTP routes: two PUBLIC browser-facing
// OIDC routes on the portal route group (login-start redirects to Google, the
// callback exchanges the code with the host external-login seam), plus the
// PROTECTED admin settings API under the plugin API prefix guarded by
// Auth+Tenancy+Permission middlewares.
func registerRoutes(ctx context.Context, registrar pluginhost.HTTPRegistrar) error {
	var (
		routes      = registrar.Routes()
		middlewares = routes.Middlewares()
		services    = registrar.Services()
	)
	if services == nil {
		return gerror.New("linapro-oidc-google routes require host external-login capability")
	}
	authSvc := services.Auth()
	if authSvc == nil || authSvc.ExternalLogin() == nil {
		return gerror.New("linapro-oidc-google routes require host external-login sub capability")
	}
	hostConfigSvc := services.HostConfig()
	if hostConfigSvc == nil || hostConfigSvc.SysConfig() == nil {
		return gerror.New("linapro-oidc-google routes require host sys_config capability")
	}
	logger.Infof(ctx, "linapro-oidc-google registering portal routes group=%s", portalGroupPath)
	pluginSettingsSvc := settingssvc.New(hostConfigSvc.SysConfig())
	// The shared config resolver feeds both the login orchestration and the
	// production HTTP identity verifier so they read identical request-time
	// settings (client credentials, endpoints, derived callback URL).
	configResolver := oauthsvc.NewConfigResolver(pluginSettingsSvc, oauthsvc.DefaultConfig())
	loginSvc := oauthsvc.New(
		authSvc.ExternalLogin(),
		configResolver,
		oauthsvc.NewHTTPIdentityVerifier(configResolver),
		oauthsvc.NewHMACStateCodec(),
	)
	loginController := loginctrl.NewV1(loginSvc, pluginSettingsSvc)
	settingsController := settingsctrl.NewV1(pluginSettingsSvc)
	routes.Group(portalGroupPath, func(group pluginhost.RouteGroup) {
		group.Middleware(
			middlewares.NeverDoneCtx(),
			middlewares.CORS(),
			middlewares.RequestBodyLimit(),
			middlewares.Ctx(),
		)
		group.GET("/login", loginController.Start)
		group.GET("/callback", loginController.Callback)
	})
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
				group.Bind(
					settingsController.GetSettings,
					settingsController.SaveSettings,
				)
			})
		})
	})
	return routes.Err()
}
