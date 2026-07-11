// Package backend wires the linapro-extlogin-core source plugin into the host
// plugin registry. It declares the external-identity provider engine factory
// (resolve/provision/bind storage behind the host external-login seam) and
// registers the current-user identity binding API. OAuth protocol plugins such
// as linapro-oidc-google and linapro-oidc-discord depend on this plugin and
// keep calling the host extlogin seam unchanged.
//
// Runtime stores for SPA handoff, the protocol provider catalog, and the
// identity domain service are process-shared: handoff/catalog are bound into
// extidcap package facades for protocol plugins; the identity domain is shared
// between the host SPI provider factory and HTTP controllers so ticket stores
// and linkage orchestration stay single-instance. Multi-node cluster sharing is
// a later hardening step.
package backend

import (
	"context"
	"sync"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/capability/authcap/extlogin/extidspi"
	"lina-core/pkg/plugin/capability/usercap"
	"lina-core/pkg/plugin/pluginhost"
	pluginextlogincore "lina-plugin-linapro-extlogin-core"
	"lina-plugin-linapro-extlogin-core/backend/cap/extidcap"
	handoffctrl "lina-plugin-linapro-extlogin-core/backend/internal/controller/handoff"
	identityctrl "lina-plugin-linapro-extlogin-core/backend/internal/controller/identity"
	handoffsvc "lina-plugin-linapro-extlogin-core/backend/internal/service/handoff"
	identitysvc "lina-plugin-linapro-extlogin-core/backend/internal/service/identity"
	catalogsvc "lina-plugin-linapro-extlogin-core/backend/internal/service/providercatalog"
)

// pluginID is the immutable identifier declared by the embedded plugin.yaml.
const pluginID = "linapro-extlogin-core"

// Process-shared runtime stores for this owner plugin instance.
var (
	sharedHandoff  = handoffsvc.New()
	sharedCatalog  = catalogsvc.New()
	identityMu     sync.Mutex
	sharedIdentity *identitysvc.Service
)

// init registers the linapro-extlogin-core source plugin, declares the
// external-identity provider engine factory, and registers HTTP routes.
func init() {
	// Bind public facades before other source plugins may call
	// RegisterProviderDescriptor from their init (order is not guaranteed;
	// the facade also buffers pre-bind registrations).
	extidcap.BindHandoffService(sharedHandoff)
	extidcap.BindCatalogService(sharedCatalog)

	plugin := pluginhost.NewDeclarations(pluginID)
	plugin.Assets().UseEmbeddedFiles(pluginextlogincore.EmbeddedFiles)
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

// sharedIdentityDomain returns the single process-local identity domain used by
// both the host SPI provider and HTTP controllers. Tickets and linkage writes
// must not split across independently constructed service graphs.
func sharedIdentityDomain(users usercap.Service) (*identitysvc.Service, error) {
	if users == nil {
		return nil, gerror.New("linapro-extlogin-core identity domain requires host user capability")
	}
	identityMu.Lock()
	defer identityMu.Unlock()
	if sharedIdentity != nil {
		return sharedIdentity, nil
	}
	svc, err := identitysvc.New(users)
	if err != nil {
		return nil, err
	}
	sharedIdentity = svc
	return sharedIdentity, nil
}

// newProvider creates the linapro-extlogin-core external-identity provider from
// host-published services during lazy capability activation.
func newProvider(_ context.Context, env extidspi.ProviderEnv) (extidspi.Provider, error) {
	return sharedIdentityDomain(env.Users)
}

// registerRoutes registers:
//  1. PUBLIC handoff exchange (SPA redeems one-time codes; no session required)
//  2. PROTECTED current-user identity bind/list/unbind APIs
//
// The identity controller receives extidcap.Service (wide domain entry), not the
// concrete internal service type. Handoff continues to use the process-shared
// store bound into extidcap at init.
func registerRoutes(_ context.Context, registrar pluginhost.HTTPRegistrar) error {
	var (
		routes      = registrar.Routes()
		middlewares = routes.Middlewares()
		services    = registrar.Services()
	)
	if services == nil || services.Users() == nil || services.BizCtx() == nil {
		return gerror.New("linapro-extlogin-core routes require host user and bizctx capabilities")
	}
	// *identitysvc.Service implements extidcap.Service; inject the wide contract
	// shared with the SPI provider path.
	identityDomain, err := sharedIdentityDomain(services.Users())
	if err != nil {
		return err
	}
	identityController := identityctrl.NewV1(identityDomain, services.BizCtx())
	handoffController := handoffctrl.NewV1(sharedHandoff)
	routes.Group(routes.APIPrefix(), func(group pluginhost.RouteGroup) {
		group.Group("/api/v1", func(group pluginhost.RouteGroup) {
			group.Middleware(
				middlewares.NeverDoneCtx(),
				middlewares.HandlerResponse(),
				middlewares.CORS(),
				middlewares.RequestBodyLimit(),
				middlewares.Ctx(),
			)
			// Public: SPA exchanges handoff without an existing session.
			group.Bind(handoffController)
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
