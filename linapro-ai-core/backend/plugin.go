// Package backend wires the linapro-ai-core source plugin into the host plugin registry.
package backend

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/capability/ai/aitext"
	"lina-core/pkg/plugin/capability/contract"
	"lina-core/pkg/plugin/pluginhost"
	aicore "lina-plugin-linapro-ai-core"
	invocationcontroller "lina-plugin-linapro-ai-core/backend/internal/controller/invocation"
	modelcontroller "lina-plugin-linapro-ai-core/backend/internal/controller/model"
	providercontroller "lina-plugin-linapro-ai-core/backend/internal/controller/provider"
	tiercontroller "lina-plugin-linapro-ai-core/backend/internal/controller/tier"
	aisvc "lina-plugin-linapro-ai-core/backend/internal/service/ai"
)

const (
	// pluginID is the immutable identifier published by the embedded source plugin.
	pluginID = aitext.ProviderPluginID
)

var (
	smartCenterMu         sync.Mutex
	smartCenterService    aisvc.Service
	smartCenterHTTPClient = &http.Client{Timeout: 60 * time.Second}
)

// init registers the linapro-ai-core source plugin, route bindings, and text AI provider.
func init() {
	plugin := pluginhost.NewSourcePlugin(pluginID)
	plugin.Assets().UseEmbeddedFiles(aicore.EmbeddedFiles)
	if err := aitext.Provide(pluginID, provideAIText); err != nil {
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

// provideAIText creates the framework text AI provider from host-published services.
func provideAIText(_ context.Context, env aitext.ProviderEnv) (aitext.Provider, error) {
	if env.BizCtx == nil || env.Cache == nil {
		return nil, gerror.New("linapro-ai-core provider requires host biz-context and cache services")
	}
	return smartCenter(env.BizCtx, env.Cache), nil
}

// registerRoutes binds Smart Center management routes through host middleware.
func registerRoutes(ctx context.Context, registrar pluginhost.HTTPRegistrar) error {
	routes := registrar.Routes()
	if routes == nil {
		return gerror.New("linapro-ai-core routes require host route registrar")
	}
	middlewares := routes.Middlewares()
	services := registrar.Services()
	if middlewares == nil || services == nil || services.BizCtx() == nil || services.Cache() == nil {
		return gerror.New("linapro-ai-core routes require host middlewares, biz-context service, and cache service")
	}
	aiService := smartCenter(services.BizCtx(), services.Cache())
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
					providercontroller.NewV1(aiService),
					modelcontroller.NewV1(aiService),
					tiercontroller.NewV1(aiService),
					invocationcontroller.NewV1(aiService),
				)
			})
		})
	})
	return nil
}

// smartCenter returns the shared Smart Center service so management writes and
// framework provider calls observe the same tier-resolution cache.
func smartCenter(bizCtxSvc contract.BizCtxService, cacheSvc contract.CacheService) aisvc.Service {
	smartCenterMu.Lock()
	defer smartCenterMu.Unlock()
	if smartCenterService == nil {
		smartCenterService = aisvc.New(bizCtxSvc, cacheSvc, smartCenterHTTPClient)
	}
	return smartCenterService
}
