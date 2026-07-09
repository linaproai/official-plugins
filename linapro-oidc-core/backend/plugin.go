// Package backend wires the linapro-oidc-core source plugin into the host
// plugin registry. It declares the external-identity provider engine factory
// (resolve/provision/bind storage behind the host external-login seam) and
// registers the current-user identity binding API. OAuth protocol plugins such
// as linapro-oidc-google and linapro-oidc-discord depend on this plugin and
// keep calling the host externallogin seam unchanged.
package backend

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/capability/authcap/externallogin/externalidentityspi"
	"lina-core/pkg/plugin/pluginhost"
	pluginoidccore "lina-plugin-linapro-oidc-core"
	identityctrl "lina-plugin-linapro-oidc-core/backend/internal/controller/identity"
	identitysvc "lina-plugin-linapro-oidc-core/backend/internal/service/identity"
)

// pluginID is the immutable identifier declared by the embedded plugin.yaml.
const pluginID = "linapro-oidc-core"

// init registers the linapro-oidc-core source plugin, declares the
// external-identity provider engine factory, and registers HTTP routes.
func init() {
	plugin := pluginhost.NewDeclarations(pluginID)
	plugin.Assets().UseEmbeddedFiles(pluginoidccore.EmbeddedFiles)
	// ProvideExternalIdentityProvider declares the engine factory only. The
	// host manager lazily constructs the provider gated by plugin enablement:
	// disabling this plugin immediately fails external login closed. This
	// plugin deliberately declares no provider-ID ownership: it never calls
	// LoginByVerifiedIdentity; provider IDs stay owned by the OAuth plugins.
	// This is an allowed top-level static registration panic.
	if err := plugin.Providers().ProvideExternalIdentityProvider(newProvider); err != nil {
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

// newProvider creates the linapro-oidc-core external-identity provider from
// host-published services during lazy capability activation.
func newProvider(_ context.Context, env externalidentityspi.ProviderEnv) (externalidentityspi.Provider, error) {
	if env.Users == nil {
		return nil, gerror.New("linapro-oidc-core provider requires host user capability")
	}
	return identitysvc.New(env.Users)
}

// registerRoutes registers the current-user identity binding API under the
// plugin API prefix, guarded by Auth+Tenancy+Permission middlewares. All
// endpoints are strictly self-isolated to the current session user (design
// D3), so they carry no extra permission nodes beyond authentication.
func registerRoutes(_ context.Context, registrar pluginhost.HTTPRegistrar) error {
	var (
		routes      = registrar.Routes()
		middlewares = routes.Middlewares()
		services    = registrar.Services()
	)
	if services == nil || services.Users() == nil || services.BizCtx() == nil {
		return gerror.New("linapro-oidc-core routes require host user and bizctx capabilities")
	}
	identitySvc, err := identitysvc.New(services.Users())
	if err != nil {
		return err
	}
	identityController := identityctrl.NewV1(identitySvc, services.BizCtx())
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
					identityController,
				)
			})
		})
	})
	return routes.Err()
}
