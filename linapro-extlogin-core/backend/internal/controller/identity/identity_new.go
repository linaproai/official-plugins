// identity_new.go defines the current-user external-identity HTTP controller
// and its constructor. The controller depends on the plugin-owned extidcap.Service
// wide entry (not internal/service concrete types) so bind/list/unbind go through
// Linkage() and stay aligned with the published domain contract.

package identity

import (
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-plugin-linapro-extlogin-core/backend/api/identity"
	"lina-plugin-linapro-extlogin-core/backend/cap/extidcap"
)

// ControllerV1 handles the current-user external identity binding API.
type ControllerV1 struct {
	extidSvc  extidcap.Service  // plugin-owned external-identity domain entry
	bizCtxSvc bizctxcap.Service // current request business context
}

// NewV1 creates and returns a new external identity controller instance.
// Callers must pass a non-nil extidcap.Service and bizctxcap.Service.
func NewV1(extidSvc extidcap.Service, bizCtxSvc bizctxcap.Service) identity.IIdentityV1 {
	return &ControllerV1{extidSvc: extidSvc, bizCtxSvc: bizCtxSvc}
}
