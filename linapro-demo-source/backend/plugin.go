// Package backend wires the source demo plugin into the host plugin registry.
package backend

import (
	"context"
	"strconv"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/pkg/logger"
	"lina-core/pkg/plugin/pluginhost"
	plugindemosource "lina-plugin-linapro-demo-source"
	democtrl "lina-plugin-linapro-demo-source/backend/internal/controller/demo"
	demosvc "lina-plugin-linapro-demo-source/backend/internal/service/demo"
)

// Source demo plugin constants.
const (
	// pluginID is the immutable identifier published by the embedded demo plugin.
	pluginID = "linapro-demo-source"
	// sourcePluginEchoInspectionName identifies the demo source-plugin job.
	sourcePluginEchoInspectionName = "source-plugin-echo-inspection"
	// sourcePluginEchoInspectionDisplayName is the English source title for the demo job.
	sourcePluginEchoInspectionDisplayName = "Source Plugin Echo Inspection"
	// sourcePluginEchoInspectionDescription is the English source description for the demo job.
	sourcePluginEchoInspectionDescription = "Runs a lightweight source-plugin inspection task for scheduler integration validation."
)

// init registers the embedded source demo plugin and its host callbacks.
func init() {
	plugin := pluginhost.NewDeclarations(pluginID)
	plugin.Assets().UseEmbeddedFiles(plugindemosource.EmbeddedFiles)
	if err := registerLifecycleDebugHandlers(plugin); err != nil {
		panic(err)
	}
	if err := plugin.Lifecycle().RegisterUninstallHandler(
		func(ctx context.Context, input pluginhost.SourcePluginUninstallInput) error {
			logSourceUninstallLifecycle(ctx, pluginhost.LifecycleHookUninstall, input)
			if !input.PurgeStorageData() {
				return nil
			}
			if input.Services() == nil || input.Services().Storage() == nil {
				return gerror.New("linapro-demo-source uninstall requires host storage service")
			}
			return demosvc.PurgeStorageData(ctx, input.Services().Storage())
		},
	); err != nil {
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
		registerBuiltinJobs,
	); err != nil {
		panic(err)
	}
	if err := pluginhost.RegisterSourcePlugin(plugin); err != nil {
		panic(err)
	}
}

// registerLifecycleDebugHandlers wires source-plugin lifecycle callbacks for
// demonstrating the host lifecycle flow in development logs.
func registerLifecycleDebugHandlers(plugin pluginhost.Declarations) error {
	if err := plugin.Lifecycle().RegisterBeforeInstallHandler(
		func(ctx context.Context, input pluginhost.SourcePluginLifecycleInput) (bool, string, error) {
			logSourceLifecycle(ctx, pluginhost.LifecycleHookBeforeInstall, input)
			return true, "", nil
		},
	); err != nil {
		return err
	}
	if err := plugin.Lifecycle().RegisterAfterInstallHandler(
		func(ctx context.Context, input pluginhost.SourcePluginLifecycleInput) error {
			logSourceLifecycle(ctx, pluginhost.LifecycleHookAfterInstall, input)
			return nil
		},
	); err != nil {
		return err
	}
	if err := plugin.Lifecycle().RegisterBeforeUpgradeHandler(
		func(ctx context.Context, input pluginhost.SourcePluginUpgradeInput) (bool, string, error) {
			logSourceUpgradeLifecycle(ctx, pluginhost.LifecycleHookBeforeUpgrade, input)
			return true, "", nil
		},
	); err != nil {
		return err
	}
	if err := plugin.Lifecycle().RegisterUpgradeHandler(
		func(ctx context.Context, input pluginhost.SourcePluginUpgradeInput) error {
			logSourceUpgradeLifecycle(ctx, pluginhost.LifecycleHookUpgrade, input)
			return nil
		},
	); err != nil {
		return err
	}
	if err := plugin.Lifecycle().RegisterAfterUpgradeHandler(
		func(ctx context.Context, input pluginhost.SourcePluginUpgradeInput) error {
			logSourceUpgradeLifecycle(ctx, pluginhost.LifecycleHookAfterUpgrade, input)
			return nil
		},
	); err != nil {
		return err
	}
	if err := plugin.Lifecycle().RegisterBeforeDisableHandler(
		func(ctx context.Context, input pluginhost.SourcePluginLifecycleInput) (bool, string, error) {
			logSourceLifecycle(ctx, pluginhost.LifecycleHookBeforeDisable, input)
			return true, "", nil
		},
	); err != nil {
		return err
	}
	if err := plugin.Lifecycle().RegisterAfterDisableHandler(
		func(ctx context.Context, input pluginhost.SourcePluginLifecycleInput) error {
			logSourceLifecycle(ctx, pluginhost.LifecycleHookAfterDisable, input)
			return nil
		},
	); err != nil {
		return err
	}
	if err := plugin.Lifecycle().RegisterBeforeUninstallHandler(
		func(ctx context.Context, input pluginhost.SourcePluginLifecycleInput) (bool, string, error) {
			logSourceLifecycle(ctx, pluginhost.LifecycleHookBeforeUninstall, input)
			return true, "", nil
		},
	); err != nil {
		return err
	}
	if err := plugin.Lifecycle().RegisterAfterUninstallHandler(
		func(ctx context.Context, input pluginhost.SourcePluginLifecycleInput) error {
			logSourceLifecycle(ctx, pluginhost.LifecycleHookAfterUninstall, input)
			return nil
		},
	); err != nil {
		return err
	}
	if err := plugin.Lifecycle().RegisterBeforeTenantDisableHandler(
		func(ctx context.Context, input pluginhost.SourcePluginTenantLifecycleInput) (bool, string, error) {
			logSourceTenantLifecycle(ctx, pluginhost.LifecycleHookBeforeTenantDisable, input)
			return true, "", nil
		},
	); err != nil {
		return err
	}
	if err := plugin.Lifecycle().RegisterAfterTenantDisableHandler(
		func(ctx context.Context, input pluginhost.SourcePluginTenantLifecycleInput) error {
			logSourceTenantLifecycle(ctx, pluginhost.LifecycleHookAfterTenantDisable, input)
			return nil
		},
	); err != nil {
		return err
	}
	if err := plugin.Lifecycle().RegisterBeforeTenantDeleteHandler(
		func(ctx context.Context, input pluginhost.SourcePluginTenantLifecycleInput) (bool, string, error) {
			logSourceTenantLifecycle(ctx, pluginhost.LifecycleHookBeforeTenantDelete, input)
			return true, "", nil
		},
	); err != nil {
		return err
	}
	if err := plugin.Lifecycle().RegisterAfterTenantDeleteHandler(
		func(ctx context.Context, input pluginhost.SourcePluginTenantLifecycleInput) error {
			logSourceTenantLifecycle(ctx, pluginhost.LifecycleHookAfterTenantDelete, input)
			return nil
		},
	); err != nil {
		return err
	}
	if err := plugin.Lifecycle().RegisterBeforeInstallModeChangeHandler(
		func(ctx context.Context, input pluginhost.SourcePluginInstallModeChangeInput) (bool, string, error) {
			logSourceInstallModeLifecycle(ctx, pluginhost.LifecycleHookBeforeInstallModeChange, input)
			return true, "", nil
		},
	); err != nil {
		return err
	}
	if err := plugin.Lifecycle().RegisterAfterInstallModeChangeHandler(
		func(ctx context.Context, input pluginhost.SourcePluginInstallModeChangeInput) error {
			logSourceInstallModeLifecycle(ctx, pluginhost.LifecycleHookAfterInstallModeChange, input)
			return nil
		},
	); err != nil {
		return err
	}
	return nil
}

// logSourceLifecycle logs a generic source-plugin lifecycle callback.
func logSourceLifecycle(ctx context.Context, operation pluginhost.LifecycleHook, input pluginhost.SourcePluginLifecycleInput) {
	logger.Infof(
		ctx,
		"linapro-demo-source lifecycle operation=%s plugin=%s",
		operation.String(),
		input.PluginID(),
	)
}

// logSourceUpgradeLifecycle logs an upgrade-related source-plugin lifecycle callback.
func logSourceUpgradeLifecycle(ctx context.Context, operation pluginhost.LifecycleHook, input pluginhost.SourcePluginUpgradeInput) {
	logger.Infof(
		ctx,
		"linapro-demo-source lifecycle operation=%s plugin=%s fromVersion=%s toVersion=%s",
		operation.String(),
		input.PluginID(),
		input.FromVersion(),
		input.ToVersion(),
	)
}

// logSourceUninstallLifecycle logs the source-plugin uninstall cleanup callback.
func logSourceUninstallLifecycle(
	ctx context.Context,
	operation pluginhost.LifecycleHook,
	input pluginhost.SourcePluginUninstallInput,
) {
	logger.Infof(
		ctx,
		"linapro-demo-source lifecycle operation=%s plugin=%s purgeStorageData=%s",
		operation.String(),
		input.PluginID(),
		strconv.FormatBool(input.PurgeStorageData()),
	)
}

// logSourceTenantLifecycle logs a tenant-scoped source-plugin lifecycle callback.
func logSourceTenantLifecycle(
	ctx context.Context,
	operation pluginhost.LifecycleHook,
	input pluginhost.SourcePluginTenantLifecycleInput,
) {
	logger.Infof(
		ctx,
		"linapro-demo-source lifecycle operation=%s tenantId=%d",
		operation.String(),
		input.TenantID(),
	)
}

// logSourceInstallModeLifecycle logs an install-mode source-plugin lifecycle callback.
func logSourceInstallModeLifecycle(
	ctx context.Context,
	operation pluginhost.LifecycleHook,
	input pluginhost.SourcePluginInstallModeChangeInput,
) {
	logger.Infof(
		ctx,
		"linapro-demo-source lifecycle operation=%s plugin=%s fromMode=%s toMode=%s",
		operation.String(),
		input.PluginID(),
		input.FromMode(),
		input.ToMode(),
	)
}

// registerRoutes binds the demo plugin HTTP routes using the published host
// middleware directory so plugin traffic follows the same governance chain as
// host-owned APIs.
func registerRoutes(ctx context.Context, registrar pluginhost.HTTPRegistrar) error {
	var (
		routes      = registrar.Routes()
		middlewares = routes.Middlewares()
		services    = registrar.Services()
	)
	if services == nil || services.I18n() == nil || services.TenantFilter() == nil || services.Storage() == nil {
		return gerror.New("linapro-demo-source routes require host i18n, tenant-filter, and storage services")
	}
	demoSvc := demosvc.New(services.I18n(), services.TenantFilter(), services.Storage())
	demoController := democtrl.NewV1(demoSvc)
	routes.Group("/portal/linapro-demo-source", func(group pluginhost.RouteGroup) {
		group.Middleware(
			middlewares.NeverDoneCtx(),
			middlewares.CORS(),
			middlewares.RequestBodyLimit(),
		)
		group.GET("/ping", servePortalPing)
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
				group.Bind(demoController.Ping)
			})

			group.Group("/", func(group pluginhost.RouteGroup) {
				group.Middleware(
					middlewares.Auth(),
					middlewares.Tenancy(),
					middlewares.Permission(),
				)
				group.Bind(
					demoController.Summary,
					demoController.ListRecords,
					demoController.GetRecord,
					demoController.CreateRecord,
					demoController.UpdateRecord,
					demoController.DeleteRecord,
					demoController.DownloadAttachment,
				)
			})
		})
	})
	return nil
}

// servePortalPing returns a plugin-owned public route response outside the
// reserved /x API namespace so route-boundary E2E can verify host fallback
// does not claim source-plugin public paths.
func servePortalPing(request *ghttp.Request) {
	request.Response.Header().Set("Content-Type", "text/plain; charset=utf-8")
	request.Response.Write("linapro-demo-source-public-pong")
}

// registerBuiltinJobs contributes one plugin-owned builtin scheduled job so
// the host can project source-plugin job registrations into unified
// scheduled-job management.
func registerBuiltinJobs(ctx context.Context, registrar pluginhost.JobsRegistrar) error {
	return registrar.AddWithMetadata(
		ctx,
		"# */15 * * * *",
		sourcePluginEchoInspectionName,
		sourcePluginEchoInspectionDisplayName,
		sourcePluginEchoInspectionDescription,
		func(ctx context.Context) error {
			return nil
		},
	)
}
