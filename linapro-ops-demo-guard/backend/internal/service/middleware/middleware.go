// Package middleware implements the linapro-ops-demo-guard request-guard middleware.
package middleware

import (
	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/pkg/plugin/capability/i18ncap"
	"lina-core/pkg/plugin/capability/plugincap"
)

// demoControlPluginID is the immutable source-plugin identifier for this middleware.
const demoControlPluginID = "linapro-ops-demo-guard"

// Service defines the linapro-ops-demo-guard middleware service contract.
type Service interface {
	// Guard enforces the demo-mode read-only policy on API requests.
	Guard(request *ghttp.Request)
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	i18nSvc    i18ncap.Service   // i18nSvc resolves plugin runtime translations.
	pluginsSvc plugincap.Service // pluginsSvc checks whether linapro-ops-demo-guard is active.
}

// New creates and returns a new linapro-ops-demo-guard middleware service.
func New(i18nSvc i18ncap.Service, pluginsSvc plugincap.Service) Service {
	return &serviceImpl{
		i18nSvc:    i18nSvc,
		pluginsSvc: pluginsSvc,
	}
}
