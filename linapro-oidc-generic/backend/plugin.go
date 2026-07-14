// Package backend wires the linapro-oidc-generic source plugin into the host
// plugin registry. It declares external-identity provider ownership and
// registers browser-facing OIDC login routes plus admin settings APIs.
package backend

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/logger"
	"lina-core/pkg/plugin/pluginhost"
	"lina-plugin-linapro-extlogin-core/backend/cap/extidcap"
	pluginoidcgeneric "lina-plugin-linapro-oidc-generic"
	loginctrl "lina-plugin-linapro-oidc-generic/backend/internal/controller/login"
	settingsctrl "lina-plugin-linapro-oidc-generic/backend/internal/controller/settings"
	oauthsvc "lina-plugin-linapro-oidc-generic/backend/internal/service/oauth"
	settingssvc "lina-plugin-linapro-oidc-generic/backend/internal/service/settings"
)

const pluginID = "linapro-oidc-generic"
const portalGroupPath = "/portal/linapro-oidc-generic"

func init() {
	plugin := pluginhost.NewDeclarations(pluginID)
	plugin.Assets().UseEmbeddedFiles(pluginoidcgeneric.EmbeddedFiles)
	if err := plugin.Providers().ProvideExternalIdentity(oauthsvc.Provider); err != nil {
		panic(err)
	}
	extidcap.RegisterProviderDescriptor(extidcap.ProviderDescriptor{
		ID:          oauthsvc.Provider,
		DisplayName: "Generic OIDC",
		Icon:        "mdi:shield-key-outline",
		Order:       30,
		PluginID:    pluginID,
		Protocols:   []string{"oidc", "oauth2"},
		Capabilities: extidcap.ProviderCapabilities{
			Login: true, Bind: true, Unbind: true, AutoProvision: true,
			Email: true, Avatar: false, OneTap: false,
		},
		SubjectStrategy: extidcap.SubjectStrategyOIDCSub,
		LoginEntryPath:  portalGroupPath + "/login",
	})
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
	if services == nil {
		return gerror.New("linapro-oidc-generic routes require host external-login capability")
	}
	authSvc := services.Auth()
	if authSvc == nil || authSvc.ExternalLogin() == nil {
		return gerror.New("linapro-oidc-generic routes require host external-login sub capability")
	}
	hostConfigSvc := services.HostConfig()
	if hostConfigSvc == nil || hostConfigSvc.SysConfig() == nil {
		return gerror.New("linapro-oidc-generic routes require host sys_config capability")
	}
	logger.Infof(ctx, "linapro-oidc-generic registering portal routes group=%s", portalGroupPath)
	var (
		pluginSettingsSvc = settingssvc.New(hostConfigSvc.SysConfig())
		configResolver    = oauthsvc.NewConfigResolver(pluginSettingsSvc, oauthsvc.DefaultConfig())
		loginSvc          = oauthsvc.New(
			authSvc.ExternalLogin(),
			configResolver,
			oauthsvc.NewHMACStateCodec(),
		)
		loginController    = loginctrl.NewV1(loginSvc, pluginSettingsSvc)
		settingsController = settingsctrl.NewV1(pluginSettingsSvc)
	)
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
				group.Bind(settingsController)
			})
		})
	})
	return routes.Err()
}
