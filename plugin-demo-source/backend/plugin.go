// Package backend wires the source demo plugin into the host plugin registry.
package backend

import (
	"context"
	"strconv"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/logger"
	"lina-core/pkg/pluginhost"
	plugindemosource "lina-plugin-demo-source"
	democtrl "lina-plugin-demo-source/backend/internal/controller/demo"
	demosvc "lina-plugin-demo-source/backend/internal/service/demo"
)

// Source demo plugin constants.
const (
	// pluginID is the immutable identifier published by the embedded demo plugin.
	pluginID = "plugin-demo-source"
	// sourcePluginEchoInspectionName identifies the demo source-plugin cron.
	sourcePluginEchoInspectionName = "source-plugin-echo-inspection"
	// sourcePluginEchoInspectionDisplayName is the English source title for the demo cron.
	sourcePluginEchoInspectionDisplayName = "Source Plugin Echo Inspection"
	// sourcePluginEchoInspectionDescription is the English source description for the demo cron.
	sourcePluginEchoInspectionDescription = "Runs a lightweight source-plugin inspection task for scheduler integration validation."
)

// init registers the embedded source demo plugin and its host callbacks.
func init() {
	plugin := pluginhost.NewSourcePlugin(pluginID)
	plugin.Assets().UseEmbeddedFiles(plugindemosource.EmbeddedFiles)
	if err := registerLifecycleDebugHandlers(plugin); err != nil {
		panic(err)
	}
	if err := plugin.Lifecycle().RegisterUninstallHandler(func(ctx context.Context, input pluginhost.SourcePluginUninstallInput) error {
		logSourceUninstallLifecycle(ctx, "Uninstall", input)
		if !input.PurgeStorageData() {
			return nil
		}
		return demosvc.PurgeStorageData(ctx)
	}); err != nil {
		panic(err)
	}
	if err := plugin.HTTP().RegisterRoutes(
		pluginhost.ExtensionPointHTTPRouteRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerRoutes,
	); err != nil {
		panic(err)
	}
	if err := plugin.Cron().RegisterCron(
		pluginhost.ExtensionPointCronRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerBuiltinCrons,
	); err != nil {
		panic(err)
	}
	if err := pluginhost.RegisterSourcePlugin(plugin); err != nil {
		panic(err)
	}
}

// registerLifecycleDebugHandlers wires source-plugin lifecycle callbacks for
// demonstrating the host lifecycle flow in development logs.
func registerLifecycleDebugHandlers(plugin pluginhost.SourcePlugin) error {
	if err := plugin.Lifecycle().RegisterBeforeInstallHandler(func(ctx context.Context, input pluginhost.SourcePluginLifecycleInput) (bool, string, error) {
		logSourceLifecycle(ctx, "BeforeInstall", input)
		return true, "", nil
	}); err != nil {
		return err
	}
	if err := plugin.Lifecycle().RegisterAfterInstallHandler(func(ctx context.Context, input pluginhost.SourcePluginLifecycleInput) error {
		logSourceLifecycle(ctx, "AfterInstall", input)
		return nil
	}); err != nil {
		return err
	}
	if err := plugin.Lifecycle().RegisterBeforeUpgradeHandler(func(ctx context.Context, input pluginhost.SourcePluginUpgradeInput) (bool, string, error) {
		logSourceUpgradeLifecycle(ctx, "BeforeUpgrade", input)
		return true, "", nil
	}); err != nil {
		return err
	}
	if err := plugin.Lifecycle().RegisterUpgradeHandler(func(ctx context.Context, input pluginhost.SourcePluginUpgradeInput) error {
		logSourceUpgradeLifecycle(ctx, "Upgrade", input)
		return nil
	}); err != nil {
		return err
	}
	if err := plugin.Lifecycle().RegisterAfterUpgradeHandler(func(ctx context.Context, input pluginhost.SourcePluginUpgradeInput) error {
		logSourceUpgradeLifecycle(ctx, "AfterUpgrade", input)
		return nil
	}); err != nil {
		return err
	}
	if err := plugin.Lifecycle().RegisterBeforeDisableHandler(func(ctx context.Context, input pluginhost.SourcePluginLifecycleInput) (bool, string, error) {
		logSourceLifecycle(ctx, "BeforeDisable", input)
		return true, "", nil
	}); err != nil {
		return err
	}
	if err := plugin.Lifecycle().RegisterAfterDisableHandler(func(ctx context.Context, input pluginhost.SourcePluginLifecycleInput) error {
		logSourceLifecycle(ctx, "AfterDisable", input)
		return nil
	}); err != nil {
		return err
	}
	if err := plugin.Lifecycle().RegisterBeforeUninstallHandler(func(ctx context.Context, input pluginhost.SourcePluginLifecycleInput) (bool, string, error) {
		logSourceLifecycle(ctx, "BeforeUninstall", input)
		return true, "", nil
	}); err != nil {
		return err
	}
	if err := plugin.Lifecycle().RegisterAfterUninstallHandler(func(ctx context.Context, input pluginhost.SourcePluginLifecycleInput) error {
		logSourceLifecycle(ctx, "AfterUninstall", input)
		return nil
	}); err != nil {
		return err
	}
	if err := plugin.Lifecycle().RegisterBeforeTenantDisableHandler(func(ctx context.Context, input pluginhost.SourcePluginTenantLifecycleInput) (bool, string, error) {
		logSourceTenantLifecycle(ctx, "BeforeTenantDisable", input)
		return true, "", nil
	}); err != nil {
		return err
	}
	if err := plugin.Lifecycle().RegisterAfterTenantDisableHandler(func(ctx context.Context, input pluginhost.SourcePluginTenantLifecycleInput) error {
		logSourceTenantLifecycle(ctx, "AfterTenantDisable", input)
		return nil
	}); err != nil {
		return err
	}
	if err := plugin.Lifecycle().RegisterBeforeTenantDeleteHandler(func(ctx context.Context, input pluginhost.SourcePluginTenantLifecycleInput) (bool, string, error) {
		logSourceTenantLifecycle(ctx, "BeforeTenantDelete", input)
		return true, "", nil
	}); err != nil {
		return err
	}
	if err := plugin.Lifecycle().RegisterAfterTenantDeleteHandler(func(ctx context.Context, input pluginhost.SourcePluginTenantLifecycleInput) error {
		logSourceTenantLifecycle(ctx, "AfterTenantDelete", input)
		return nil
	}); err != nil {
		return err
	}
	if err := plugin.Lifecycle().RegisterBeforeInstallModeChangeHandler(func(ctx context.Context, input pluginhost.SourcePluginInstallModeChangeInput) (bool, string, error) {
		logSourceInstallModeLifecycle(ctx, "BeforeInstallModeChange", input)
		return true, "", nil
	}); err != nil {
		return err
	}
	if err := plugin.Lifecycle().RegisterAfterInstallModeChangeHandler(func(ctx context.Context, input pluginhost.SourcePluginInstallModeChangeInput) error {
		logSourceInstallModeLifecycle(ctx, "AfterInstallModeChange", input)
		return nil
	}); err != nil {
		return err
	}
	return nil
}

// logSourceLifecycle logs a generic source-plugin lifecycle callback.
func logSourceLifecycle(ctx context.Context, operation string, input pluginhost.SourcePluginLifecycleInput) {
	logger.Infof(
		ctx,
		"plugin-demo-source lifecycle operation=%s plugin=%s",
		operation,
		input.PluginID(),
	)
}

// logSourceUpgradeLifecycle logs an upgrade-related source-plugin lifecycle callback.
func logSourceUpgradeLifecycle(ctx context.Context, operation string, input pluginhost.SourcePluginUpgradeInput) {
	logger.Infof(
		ctx,
		"plugin-demo-source lifecycle operation=%s plugin=%s fromVersion=%s toVersion=%s",
		operation,
		input.PluginID(),
		input.FromVersion(),
		input.ToVersion(),
	)
}

// logSourceUninstallLifecycle logs the source-plugin uninstall cleanup callback.
func logSourceUninstallLifecycle(ctx context.Context, operation string, input pluginhost.SourcePluginUninstallInput) {
	logger.Infof(
		ctx,
		"plugin-demo-source lifecycle operation=%s plugin=%s purgeStorageData=%s",
		operation,
		input.PluginID(),
		strconv.FormatBool(input.PurgeStorageData()),
	)
}

// logSourceTenantLifecycle logs a tenant-scoped source-plugin lifecycle callback.
func logSourceTenantLifecycle(ctx context.Context, operation string, input pluginhost.SourcePluginTenantLifecycleInput) {
	logger.Infof(
		ctx,
		"plugin-demo-source lifecycle operation=%s tenantId=%d",
		operation,
		input.TenantID(),
	)
}

// logSourceInstallModeLifecycle logs an install-mode source-plugin lifecycle callback.
func logSourceInstallModeLifecycle(
	ctx context.Context,
	operation string,
	input pluginhost.SourcePluginInstallModeChangeInput,
) {
	logger.Infof(
		ctx,
		"plugin-demo-source lifecycle operation=%s plugin=%s fromMode=%s toMode=%s",
		operation,
		input.PluginID(),
		input.FromMode(),
		input.ToMode(),
	)
}

// registerRoutes binds the demo plugin HTTP routes using the published host
// middleware directory so plugin traffic follows the same governance chain as
// host-owned APIs.
func registerRoutes(ctx context.Context, registrar pluginhost.HTTPRegistrar) error {
	hostServices := registrar.HostServices()
	if hostServices == nil || hostServices.I18n() == nil || hostServices.TenantFilter() == nil {
		return gerror.New("plugin-demo-source routes require host i18n and tenant-filter services")
	}
	demoSvc := demosvc.New(hostServices.I18n(), hostServices.TenantFilter())
	var (
		routes         = registrar.Routes()
		middlewares    = routes.Middlewares()
		demoController = democtrl.NewV1(demoSvc)
	)
	routes.Group("/api/v1", func(group pluginhost.RouteGroup) {
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
	return nil
}

// registerBuiltinCrons contributes one plugin-owned builtin scheduled job so
// the host can project source-plugin cron registrations into unified
// scheduled-job management.
func registerBuiltinCrons(ctx context.Context, registrar pluginhost.CronRegistrar) error {
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
