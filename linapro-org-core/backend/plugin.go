// Package backend wires the linapro-org-core source plugin into the host plugin registry.
package backend

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/capability/orgcap/orgspi"
	"lina-core/pkg/plugin/pluginhost"
	orgcenter "lina-plugin-linapro-org-core"
	deptcontroller "lina-plugin-linapro-org-core/backend/internal/controller/dept"
	postcontroller "lina-plugin-linapro-org-core/backend/internal/controller/post"
	"lina-plugin-linapro-org-core/backend/internal/provider/orgcapadapter"
	deptsvc "lina-plugin-linapro-org-core/backend/internal/service/dept"
	postsvc "lina-plugin-linapro-org-core/backend/internal/service/post"
)

// linapro-org-core plugin constants.
const (
	// pluginID is the immutable identifier published by the embedded source plugin.
	pluginID = "linapro-org-core"
)

// init registers the linapro-org-core source plugin and its host callbacks.
func init() {
	plugin := pluginhost.NewDeclarations(pluginID)
	plugin.Assets().UseEmbeddedFiles(orgcenter.EmbeddedFiles)
	if err := plugin.Providers().ProvideOrg(provideOrg); err != nil {
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

// provideOrg creates the linapro-org-core organization capability adapter from
// host-published services during framework capability activation.
func provideOrg(_ context.Context, env orgspi.ProviderEnv) (orgspi.Provider, error) {
	if env.TenantFilter == nil {
		return nil, gerror.New("linapro-org-core provider requires host tenant-filter service")
	}
	if env.Users == nil {
		return nil, gerror.New("linapro-org-core provider requires host user capability service")
	}
	return orgcapadapter.New(env.TenantFilter, env.Users), nil
}

// registerRoutes binds department and post management routes through the published host middleware set.
func registerRoutes(ctx context.Context, registrar pluginhost.HTTPRegistrar) error {
	var (
		routes      = registrar.Routes()
		middlewares = routes.Middlewares()
		services    = registrar.Services()
	)
	if services == nil || services.I18n() == nil || services.TenantFilter() == nil || services.Users() == nil {
		return gerror.New("linapro-org-core routes require host i18n, tenant-filter, and user capability services")
	}
	deptSvc := deptsvc.New(services.TenantFilter(), services.Users())
	postSvc := postsvc.New(services.I18n(), services.TenantFilter())
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
					deptcontroller.NewV1(deptSvc),
					postcontroller.NewV1(postSvc),
				)
			})
		})
	})
	return nil
}
