// Package backend wires the linapro-monitor-loginlog source plugin into the host plugin registry.
package backend

import (
	"context"
	"strconv"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	plugincontract "lina-core/pkg/plugin/capability/contract"
	"lina-core/pkg/plugin/pluginhost"
	monitorloginlogplugin "lina-plugin-linapro-monitor-loginlog"
	loginlogcontroller "lina-plugin-linapro-monitor-loginlog/backend/internal/controller/loginlog"
	loginlogsvc "lina-plugin-linapro-monitor-loginlog/backend/internal/service/loginlog"
)

// linapro-monitor-loginlog plugin constants.
const (
	// pluginID is the immutable identifier published by the embedded source plugin.
	pluginID = "linapro-monitor-loginlog"
	// logRetentionDaysKey is the host protected runtime parameter shared by log cleanup jobs.
	logRetentionDaysKey = "sys.log.retentionDays"
	// loginLogCleanupCronName identifies the login-log cleanup cron declaration.
	loginLogCleanupCronName = "login-log-cleanup"
	// loginLogCleanupCronDisplayName is the English source title for the cleanup cron.
	loginLogCleanupCronDisplayName = "Login Log Cleanup"
	// loginLogCleanupCronDescription is the English source description for the cleanup cron.
	loginLogCleanupCronDescription = "Cleans up expired login audit logs for the linapro-monitor-loginlog plugin."
)

// loginLogRetentionCleaner is the plugin service subset needed by the cleanup cron.
type loginLogRetentionCleaner interface {
	// CleanupExpired hard-deletes login logs older than the given retention period.
	CleanupExpired(ctx context.Context, retentionDays int) (int, error)
}

// init registers the linapro-monitor-loginlog source plugin and its host callbacks.
func init() {
	plugin := pluginhost.NewSourcePlugin(pluginID)
	plugin.Assets().UseEmbeddedFiles(monitorloginlogplugin.EmbeddedFiles)
	if err := plugin.HTTP().RegisterRoutes(
		pluginhost.ExtensionPointHTTPRouteRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerRoutes,
	); err != nil {
		panic(err)
	}
	if err := plugin.Hooks().RegisterHook(
		pluginhost.ExtensionPointAuthLoginSucceeded,
		pluginhost.CallbackExecutionModeAsync,
		handleAuthEvent,
	); err != nil {
		panic(err)
	}
	if err := plugin.Hooks().RegisterHook(
		pluginhost.ExtensionPointAuthLoginFailed,
		pluginhost.CallbackExecutionModeAsync,
		handleAuthEvent,
	); err != nil {
		panic(err)
	}
	if err := plugin.Hooks().RegisterHook(
		pluginhost.ExtensionPointAuthLogoutSucceeded,
		pluginhost.CallbackExecutionModeAsync,
		handleAuthEvent,
	); err != nil {
		panic(err)
	}
	if err := plugin.Cron().RegisterCron(
		pluginhost.ExtensionPointCronRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerCleanupCron,
	); err != nil {
		panic(err)
	}
	if err := pluginhost.RegisterSourcePlugin(plugin); err != nil {
		panic(err)
	}
}

// registerRoutes binds login-log governance routes through the published host middleware set.
func registerRoutes(ctx context.Context, registrar pluginhost.HTTPRegistrar) error {
	var (
		routes      = registrar.Routes()
		middlewares = routes.Middlewares()
		services    = registrar.Services()
	)
	if services == nil || services.I18n() == nil || services.TenantFilter() == nil {
		return gerror.New("linapro-monitor-loginlog routes require host i18n and tenant-filter services")
	}
	loginLogSvc := loginlogsvc.New(services.I18n(), services.TenantFilter())
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
				group.Bind(loginlogcontroller.NewV1(loginLogSvc))
			})
		})
	})
	return nil
}

// handleAuthEvent persists one host authentication lifecycle event into the login-log table owned by this plugin.
func handleAuthEvent(ctx context.Context, payload pluginhost.HookPayload) error {
	services := payload.Services()
	if services == nil || services.I18n() == nil || services.TenantFilter() == nil {
		return gerror.New("linapro-monitor-loginlog hook requires host i18n and tenant-filter services")
	}
	values := payload.Values()
	status, _ := pluginhost.HookPayloadIntValue(values, pluginhost.HookPayloadKeyStatus)
	message := pluginhost.HookPayloadStringValue(values, pluginhost.HookPayloadKeyReason)
	if message == "" {
		message = pluginhost.HookPayloadStringValue(values, pluginhost.HookPayloadKeyMessage)
	}

	return loginlogsvc.New(services.I18n(), services.TenantFilter()).Create(ctx, loginlogsvc.CreateInput{
		UserName: pluginhost.HookPayloadStringValue(values, pluginhost.HookPayloadKeyUserName),
		Status:   status,
		Ip:       pluginhost.HookPayloadStringValue(values, pluginhost.HookPayloadKeyIP),
		Browser:  pluginhost.HookPayloadStringValue(values, pluginhost.HookPayloadKeyBrowser),
		Os:       pluginhost.HookPayloadStringValue(values, pluginhost.HookPayloadKeyOS),
		Msg:      message,
	})
}

// registerCleanupCron contributes the plugin-owned login-log retention cleanup job.
func registerCleanupCron(ctx context.Context, registrar pluginhost.CronRegistrar) error {
	services := registrar.Services()
	if services == nil || services.HostConfig() == nil {
		return gerror.New("linapro-monitor-loginlog cleanup cron requires host config service")
	}
	cleaner := loginlogsvc.New(nil, nil)
	return registrar.AddWithMetadata(
		ctx,
		"# 27 3 * * *",
		loginLogCleanupCronName,
		loginLogCleanupCronDisplayName,
		loginLogCleanupCronDescription,
		func(ctx context.Context) error {
			return cleanupExpiredLoginLogs(ctx, registrar.IsPrimaryNode(), services.HostConfig(), cleaner)
		},
	)
}

// cleanupExpiredLoginLogs runs primary-node login-log retention cleanup.
func cleanupExpiredLoginLogs(
	ctx context.Context,
	primaryNode bool,
	hostConfigSvc plugincontract.HostConfigService,
	cleaner loginLogRetentionCleaner,
) error {
	if !primaryNode {
		return nil
	}
	if hostConfigSvc == nil {
		return gerror.New("linapro-monitor-loginlog cleanup requires host config service")
	}
	if cleaner == nil {
		return gerror.New("linapro-monitor-loginlog cleanup requires login-log service")
	}
	retentionDays, err := requiredLogRetentionDays(ctx, hostConfigSvc)
	if err != nil {
		return err
	}
	_, err = cleaner.CleanupExpired(ctx, retentionDays)
	return err
}

// requiredLogRetentionDays reads the required host log-retention parameter.
func requiredLogRetentionDays(ctx context.Context, hostConfigSvc plugincontract.HostConfigService) (int, error) {
	value, err := hostConfigSvc.Get(ctx, logRetentionDaysKey)
	if err != nil {
		return 0, err
	}
	if value == nil || value.IsNil() || strings.TrimSpace(value.String()) == "" {
		return 0, gerror.Newf("linapro-monitor-loginlog cleanup requires %s", logRetentionDaysKey)
	}
	retentionDays, err := strconv.Atoi(strings.TrimSpace(value.String()))
	if err != nil {
		return 0, gerror.Wrapf(err, "parse linapro-monitor-loginlog cleanup %s failed", logRetentionDaysKey)
	}
	if retentionDays <= 0 {
		return 0, gerror.Newf("linapro-monitor-loginlog cleanup requires positive %s", logRetentionDaysKey)
	}
	return retentionDays, nil
}
