// Package backend wires the linapro-monitor-server source plugin into the host plugin registry.
package backend

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/capability/plugincap"
	"lina-core/pkg/plugin/pluginhost"
	monitorserverplugin "lina-plugin-linapro-monitor-server"
	servercontroller "lina-plugin-linapro-monitor-server/backend/internal/controller/monitor"
	monitorconfig "lina-plugin-linapro-monitor-server/backend/internal/service/config"
	monitorsvc "lina-plugin-linapro-monitor-server/backend/internal/service/monitor"
)

// linapro-monitor-server plugin constants.
const (
	// pluginID is the immutable identifier published by the embedded source plugin.
	pluginID = "linapro-monitor-server"
	// serviceMonitorCollectorName identifies the server metric collection job.
	serviceMonitorCollectorName = "service-monitor-collector"
	// serviceMonitorCollectorDisplayName is the English source title for the collection job.
	serviceMonitorCollectorDisplayName = "Server Monitor Collection"
	// serviceMonitorCollectorDescription is the English source description for the collection job.
	serviceMonitorCollectorDescription = "Collects server runtime metrics for the linapro-monitor-server plugin."
	// serviceMonitorCleanupName identifies the server metric cleanup job.
	serviceMonitorCleanupName = "service-monitor-cleanup"
	// serviceMonitorCleanupDisplayName is the English source title for the cleanup job.
	serviceMonitorCleanupDisplayName = "Server Monitor Cleanup"
	// serviceMonitorCleanupDescription is the English source description for the cleanup job.
	serviceMonitorCleanupDescription = "Cleans up expired server runtime metric snapshots for the linapro-monitor-server plugin."
)

// sharedMonitorSvc is the plugin-owned server monitor service shared by HTTP,
// Jobs, and startup hooks so sampling state is not split across callbacks.
var sharedMonitorSvc = monitorsvc.New()

// init registers the linapro-monitor-server source plugin and its host callbacks.
func init() {
	plugin := pluginhost.NewDeclarations(pluginID)
	plugin.Assets().UseEmbeddedFiles(monitorserverplugin.EmbeddedFiles)
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
		registerBuiltinJobs,
	); err != nil {
		panic(err)
	}
	if err := plugin.Hooks().RegisterHook(
		pluginhost.ExtensionPointSystemStarted,
		pluginhost.CallbackExecutionModeAsync,
		collectOnSystemStarted,
	); err != nil {
		panic(err)
	}
	if err := pluginhost.RegisterSourcePlugin(plugin); err != nil {
		panic(err)
	}
}

// registerRoutes binds server-monitor query routes through the published host middleware set.
func registerRoutes(ctx context.Context, registrar pluginhost.HTTPRegistrar) error {
	var (
		routes      = registrar.Routes()
		middlewares = routes.Middlewares()
	)
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
				group.Bind(servercontroller.NewV1(sharedMonitorSvc))
			})
		})
	})
	return nil
}

// registerBuiltinJobs contributes managed job definitions for server-monitor collection and cleanup.
func registerBuiltinJobs(ctx context.Context, registrar pluginhost.JobsRegistrar) error {
	services := registrar.Services()
	if services == nil {
		return gerror.New("linapro-monitor-server job requires plugin config service")
	}
	plugins := services.Plugins()
	if plugins == nil {
		return gerror.New("linapro-monitor-server job requires plugin config service")
	}
	monitorCfg, err := monitorconfig.Load(ctx, plugins)
	if err != nil {
		return err
	}
	interval := monitorCfg.Interval

	if err := registrar.AddWithMetadata(
		ctx,
		"@every "+interval.String(),
		serviceMonitorCollectorName,
		serviceMonitorCollectorDisplayName,
		serviceMonitorCollectorDescription,
		func(ctx context.Context) error {
			return collectSnapshot(ctx, sharedMonitorSvc)
		},
	); err != nil {
		return err
	}
	return registrar.AddWithMetadata(
		ctx,
		"# * * * * *",
		serviceMonitorCleanupName,
		serviceMonitorCleanupDisplayName,
		serviceMonitorCleanupDescription,
		func(ctx context.Context) error {
			return cleanupSnapshots(ctx, registrar.IsPrimaryNode(), plugins, sharedMonitorSvc)
		},
	)
}

// collectOnSystemStarted performs one eager collection after host startup so the page has an initial snapshot.
func collectOnSystemStarted(ctx context.Context, payload pluginhost.HookPayload) error {
	sharedMonitorSvc.CollectAndStore(ctx)
	return nil
}

// collectSnapshot writes one fresh monitoring snapshot.
func collectSnapshot(ctx context.Context, monitorSvc monitorsvc.Service) error {
	monitorSvc.CollectAndStore(ctx)
	return nil
}

// cleanupSnapshots removes expired monitoring snapshots.
func cleanupSnapshots(
	ctx context.Context,
	primaryNode bool,
	plugins plugincap.Service,
	monitorSvc monitorsvc.Service,
) error {
	if !primaryNode {
		return nil
	}

	if plugins == nil {
		return gerror.New("linapro-monitor-server cleanup requires plugin service")
	}
	monitorCfg, err := monitorconfig.Load(ctx, plugins)
	if err != nil {
		return err
	}

	_, err = monitorSvc.CleanupStale(ctx, monitorCfg.Interval*time.Duration(monitorCfg.RetentionMultiplier))
	return err
}
