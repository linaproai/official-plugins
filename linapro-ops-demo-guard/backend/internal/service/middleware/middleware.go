// Package middleware implements the linapro-ops-demo-guard request-guard middleware.
package middleware

import (
	"context"

	"github.com/gogf/gf/v2/net/ghttp"
)

// demoControlPluginID is the immutable source-plugin identifier for this middleware.
const demoControlPluginID = "linapro-ops-demo-guard"

// Service defines the linapro-ops-demo-guard middleware service contract.
type Service interface {
	// Guard enforces the demo-mode read-only policy on API requests.
	Guard(request *ghttp.Request)
}

// EnablementReader defines the host plugin-state capability needed by the guard.
type EnablementReader interface {
	// IsEnabledAuthoritative reports whether the given plugin is currently
	// installed and enabled after bypassing process-local platform snapshots.
	// The demo guard uses the authoritative path because it controls
	// whole-system write protection.
	IsEnabledAuthoritative(ctx context.Context, pluginID string) bool
}

// Translator defines the runtime translation capability needed by the guard.
type Translator interface {
	// Translate returns the localized value for one runtime i18n key and fallback text.
	Translate(ctx context.Context, key string, fallback string) string
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	i18nSvc          Translator       // i18nSvc resolves plugin runtime translations.
	enablementReader EnablementReader // enablementReader checks whether linapro-ops-demo-guard is active.
}

// New creates and returns a new linapro-ops-demo-guard middleware service.
func New(i18nSvc Translator, enablementReader EnablementReader) Service {
	return &serviceImpl{
		i18nSvc:          i18nSvc,
		enablementReader: enablementReader,
	}
}
