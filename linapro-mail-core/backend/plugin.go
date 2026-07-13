// Package backend wires the linapro-mail-core source plugin into the host
// plugin registry, including global transport kind conflict detection and
// management routes.
package backend

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/logger"
	"lina-core/pkg/plugin/capability/plugincap"
	"lina-core/pkg/plugin/pluginhost"
	mailcore "lina-plugin-linapro-mail-core"
	"lina-plugin-linapro-mail-core/backend/cap/mailcap"
	"lina-plugin-linapro-mail-core/backend/cap/mailcap/spi"
	accountctrl "lina-plugin-linapro-mail-core/backend/internal/controller/account"
	connectionctrl "lina-plugin-linapro-mail-core/backend/internal/controller/connection"
	settingsctrl "lina-plugin-linapro-mail-core/backend/internal/controller/settings"
	mailsvc "lina-plugin-linapro-mail-core/backend/internal/service/mail"
)

const pluginID = mailcap.OwnerPluginID

// init registers the embedded linapro-mail-core source plugin.
func init() {
	plugin := pluginhost.NewDeclarations(pluginID)
	plugin.Assets().UseEmbeddedFiles(mailcore.EmbeddedFiles)
	if err := plugin.Lifecycle().RegisterGlobalBeforeEnableHandler(globalBeforeEnable); err != nil {
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

// globalBeforeEnable vetoes enabling a second transport plugin for the same kind.
func globalBeforeEnable(
	ctx context.Context,
	input pluginhost.SourcePluginGlobalLifecycleInput,
) (bool, string, error) {
	return spi.GlobalBeforeEnableVeto(ctx, input, enablementFromInput(input))
}

// stateEnablement adapts plugincap.StateService for SPI enablement checks.
type stateEnablement struct {
	state plugincap.StateService
}

// IsProviderEnabled reports whether the transport plugin is provider-enabled.
func (e stateEnablement) IsProviderEnabled(ctx context.Context, id string) bool {
	if e.state == nil {
		return false
	}
	enabled, err := e.state.IsProviderEnabled(ctx, plugincap.PluginID(id))
	if err != nil {
		return false
	}
	return enabled
}

// enablementFromInput builds SPI enablement from global lifecycle host services.
func enablementFromInput(input pluginhost.SourcePluginGlobalLifecycleInput) spi.EnablementReader {
	if input == nil || input.Services() == nil || input.Services().Plugins() == nil {
		return nil
	}
	state := input.Services().Plugins().State()
	if state == nil {
		return nil
	}
	return stateEnablement{state: state}
}

// registerRoutes binds Connection/Account management APIs.
func registerRoutes(ctx context.Context, registrar pluginhost.HTTPRegistrar) error {
	routes := registrar.Routes()
	middlewares := routes.Middlewares()
	services := registrar.Services()
	if services == nil || services.Plugins() == nil || services.Plugins().State() == nil {
		return gerror.New("linapro-mail-core routes require host plugin state capability")
	}
	if services.I18n() == nil {
		return gerror.New("linapro-mail-core routes require host i18n capability")
	}
	logger.Infof(ctx, "linapro-mail-core registering mail management routes")
	mailService := mailsvc.New(services.Plugins().State())
	// Bridge host notify email channel to mail-core; fail closed if already set.
	if err := mailsvc.ProvideNotifyEmailDelivery(mailService); err != nil {
		logger.Warningf(ctx, "linapro-mail-core notify email bridge registration skipped: %v", err)
	}
	connectionController := connectionctrl.NewV1(mailService)
	accountController := accountctrl.NewV1(mailService)
	settingsController := settingsctrl.NewV1(mailService, services.I18n())

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
					connectionController,
					accountController,
					settingsController,
				)
			})
		})
	})
	return routes.Err()
}
