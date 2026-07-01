// Package backend wires the linapro-ops-demo-guard source plugin into the host plugin registry.
package backend

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/pluginhost"
	democontrolplugin "lina-plugin-linapro-ops-demo-guard"
	middlewaresvc "lina-plugin-linapro-ops-demo-guard/backend/internal/service/middleware"
)

// linapro-ops-demo-guard plugin constants.
const (
	// pluginID is the immutable identifier published by the embedded source plugin.
	pluginID = "linapro-ops-demo-guard"
)

// init registers the embedded linapro-ops-demo-guard source plugin and its host callbacks.
func init() {
	plugin := pluginhost.NewDeclarations(pluginID)
	plugin.Assets().UseEmbeddedFiles(democontrolplugin.EmbeddedFiles)
	if err := plugin.Lifecycle().RegisterBeforeInstallHandler(beforeInstall); err != nil {
		panic(err)
	}
	if err := plugin.HTTP().RegisterRoutes(
		pluginhost.ExtensionPointHTTPRouteRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerGlobalMiddleware,
	); err != nil {
		panic(err)
	}
	if err := pluginhost.RegisterSourcePlugin(plugin); err != nil {
		panic(err)
	}
}

// beforeInstall prevents accidental management-page installation of demo mode.
func beforeInstall(_ context.Context, input pluginhost.SourcePluginLifecycleInput) (bool, string, error) {
	if input != nil && input.StartupAutoEnable() {
		return true, "", nil
	}
	return false, middlewaresvc.CodeDemoControlInstallManualDenied.MessageKey(), nil
}

// registerGlobalMiddleware binds the demo read-only guard into the host-wide
// system request chain published to source plugins.
func registerGlobalMiddleware(_ context.Context, registrar pluginhost.HTTPRegistrar) error {
	services := registrar.Services()
	if services == nil || services.I18n() == nil || services.Plugins() == nil || services.Plugins().State() == nil {
		return gerror.New("linapro-ops-demo-guard middleware requires host i18n and plugin-state services")
	}
	guardSvc := middlewaresvc.New(services.I18n(), services.Plugins())
	return registrar.GlobalMiddlewares().Bind("/*", guardSvc.Guard)
}
