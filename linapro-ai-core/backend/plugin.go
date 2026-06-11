// Package backend wires the linapro-ai-core source plugin into the host plugin registry.
package backend

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/capability/aicap/aitext"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/cachecap"
	"lina-core/pkg/plugin/capability/hostconfigcap"
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
	// logRetentionDaysKey is the host protected runtime parameter shared by log cleanup jobs.
	logRetentionDaysKey = "sys.log.retentionDays"
	// invocationLogCleanupJobName identifies the AI invocation-log cleanup job declaration.
	invocationLogCleanupJobName = "ai-invocation-log-cleanup"
	// invocationLogCleanupJobDisplayName is the English source title for the cleanup job.
	invocationLogCleanupJobDisplayName = "AI Invocation Log Cleanup"
	// invocationLogCleanupJobDescription is the English source description for the cleanup job.
	invocationLogCleanupJobDescription = "Cleans up expired AI invocation logs for the linapro-ai-core plugin."
)

var (
	smartCenterMu         sync.Mutex
	smartCenterService    aisvc.Service
	smartCenterHTTPClient = &http.Client{Timeout: 60 * time.Second}
)

// invocationRetentionCleaner is the Smart Center service subset needed by the cleanup job.
type invocationRetentionCleaner interface {
	// CleanupExpiredInvocations hard-deletes invocation logs older than the retention period.
	CleanupExpiredInvocations(ctx context.Context, retentionDays int) (int, error)
}

// init registers the linapro-ai-core source plugin, route bindings, and text AI provider.
func init() {
	plugin := pluginhost.NewDeclarations(pluginID)
	plugin.Assets().UseEmbeddedFiles(aicore.EmbeddedFiles)
	if err := plugin.Providers().ProvideAIText(provideAIText); err != nil {
		panic(err)
	}
	if err := plugin.HTTP().RegisterRoutes(
		pluginhost.ExtensionPointHTTPRouteRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerRoutes,
	); err != nil {
		panic(err)
	}
	if err := plugin.Jobs().RegisterJobs(
		pluginhost.ExtensionPointJobsRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerCleanupJob,
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

// registerCleanupJob contributes the plugin-owned AI invocation retention cleanup job.
func registerCleanupJob(ctx context.Context, registrar pluginhost.JobsRegistrar) error {
	services := registrar.Services()
	if services == nil ||
		services.HostConfig() == nil ||
		services.BizCtx() == nil ||
		services.Cache() == nil {
		return gerror.New("linapro-ai-core cleanup job requires host config, biz-context, and cache services")
	}
	aiService := smartCenter(services.BizCtx(), services.Cache())
	return registrar.AddWithMetadata(
		ctx,
		"# 37 3 * * *",
		invocationLogCleanupJobName,
		invocationLogCleanupJobDisplayName,
		invocationLogCleanupJobDescription,
		func(ctx context.Context) error {
			return cleanupExpiredInvocations(ctx, registrar.IsPrimaryNode(), services.HostConfig(), aiService)
		},
	)
}

// cleanupExpiredInvocations runs primary-node AI invocation retention cleanup.
func cleanupExpiredInvocations(
	ctx context.Context,
	primaryNode bool,
	hostConfigSvc hostconfigcap.Service,
	cleaner invocationRetentionCleaner,
) error {
	if !primaryNode {
		return nil
	}
	if hostConfigSvc == nil {
		return gerror.New("linapro-ai-core cleanup requires host config service")
	}
	if cleaner == nil {
		return gerror.New("linapro-ai-core cleanup requires Smart Center service")
	}
	retentionDays, err := requiredLogRetentionDays(ctx, hostConfigSvc)
	if err != nil {
		return err
	}
	_, err = cleaner.CleanupExpiredInvocations(ctx, retentionDays)
	return err
}

// requiredLogRetentionDays reads the required host log-retention parameter.
func requiredLogRetentionDays(ctx context.Context, hostConfigSvc hostconfigcap.Service) (int, error) {
	value, err := hostConfigSvc.Get(ctx, logRetentionDaysKey)
	if err != nil {
		return 0, err
	}
	if value == nil || value.IsNil() || strings.TrimSpace(value.String()) == "" {
		return 0, gerror.Newf("linapro-ai-core cleanup requires %s", logRetentionDaysKey)
	}
	retentionDays, err := strconv.Atoi(strings.TrimSpace(value.String()))
	if err != nil {
		return 0, gerror.Wrapf(err, "parse linapro-ai-core cleanup %s failed", logRetentionDaysKey)
	}
	if retentionDays <= 0 {
		return 0, gerror.Newf("linapro-ai-core cleanup requires positive %s", logRetentionDaysKey)
	}
	return retentionDays, nil
}

// smartCenter returns the shared Smart Center service so management writes and
// framework provider calls observe the same tier-resolution cache.
func smartCenter(bizCtxSvc bizctxcap.Service, cacheSvc cachecap.Service) aisvc.Service {
	smartCenterMu.Lock()
	defer smartCenterMu.Unlock()
	if smartCenterService == nil {
		smartCenterService = aisvc.New(bizCtxSvc, cacheSvc, smartCenterHTTPClient)
	}
	return smartCenterService
}
