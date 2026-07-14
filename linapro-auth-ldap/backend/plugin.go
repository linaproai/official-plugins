// Package backend wires linapro-auth-ldap into the host plugin registry.
package backend

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/logger"
	"lina-core/pkg/plugin/pluginhost"
	pluginauthldap "lina-plugin-linapro-auth-ldap"
	loginctrl "lina-plugin-linapro-auth-ldap/backend/internal/controller/login"
	settingsctrl "lina-plugin-linapro-auth-ldap/backend/internal/controller/settings"
	ldapauthsvc "lina-plugin-linapro-auth-ldap/backend/internal/service/ldapauth"
	settingssvc "lina-plugin-linapro-auth-ldap/backend/internal/service/settings"
	"lina-plugin-linapro-extlogin-core/backend/cap/extidcap"
)

const pluginID = "linapro-auth-ldap"
const portalGroupPath = "/portal/linapro-auth-ldap"

func init() {
	plugin := pluginhost.NewDeclarations(pluginID)
	plugin.Assets().UseEmbeddedFiles(pluginauthldap.EmbeddedFiles)
	if err := plugin.Providers().ProvideExternalIdentity(ldapauthsvc.Provider); err != nil {
		panic(err)
	}
	extidcap.RegisterProviderDescriptor(extidcap.ProviderDescriptor{
		ID:          ldapauthsvc.Provider,
		DisplayName: "LDAP",
		Icon:        "mdi:folder-account-outline",
		Order:       40,
		PluginID:    pluginID,
		Protocols:   []string{"ldap"},
		Capabilities: extidcap.ProviderCapabilities{
			Login: true, Bind: true, Unbind: true, AutoProvision: true,
			Email: true, Avatar: false, OneTap: false,
		},
		SubjectStrategy: extidcap.SubjectStrategy("custom"),
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
	if services == nil || services.Auth() == nil || services.Auth().ExternalLogin() == nil {
		return gerror.New("linapro-auth-ldap requires host external-login")
	}
	hostConfigSvc := services.HostConfig()
	if hostConfigSvc == nil || hostConfigSvc.SysConfig() == nil {
		return gerror.New("linapro-auth-ldap requires host sys_config")
	}
	logger.Infof(ctx, "linapro-auth-ldap registering routes group=%s", portalGroupPath)
	var (
		pluginSettings     = settingssvc.New(hostConfigSvc.SysConfig())
		loginSvc           = ldapauthsvc.New(services.Auth().ExternalLogin(), pluginSettings, nil)
		loginController    = loginctrl.NewV1(loginSvc)
		settingsController = settingsctrl.NewV1(pluginSettings)
	)

	// Public portal login (JSON bind).
	routes.Group(portalGroupPath, func(group pluginhost.RouteGroup) {
		group.Middleware(
			middlewares.NeverDoneCtx(),
			middlewares.HandlerResponse(),
			middlewares.CORS(),
			middlewares.RequestBodyLimit(),
			middlewares.Ctx(),
		)
		group.Bind(loginController)
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
