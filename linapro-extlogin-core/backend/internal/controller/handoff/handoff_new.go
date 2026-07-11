// handoff_new.go defines the public login-handoff exchange controller and its
// constructor. The handoff store is injected so create/exchange share one
// owner-plugin instance bound at startup.
package handoff

import (
	handoffapi "lina-plugin-linapro-extlogin-core/backend/api/handoff"
	"lina-plugin-linapro-extlogin-core/backend/cap/extidcap"
)

// ControllerV1 handles public login-handoff exchange.
type ControllerV1 struct {
	handoffs extidcap.HandoffService
}

// NewV1 creates the handoff controller with an explicit handoff service dependency.
func NewV1(handoffs extidcap.HandoffService) handoffapi.IHandoffV1 {
	return &ControllerV1{handoffs: handoffs}
}
