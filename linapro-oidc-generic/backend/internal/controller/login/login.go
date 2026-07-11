// Package login implements HTTP entry points for generic OIDC login.

package login

import (
	oauthsvc "lina-plugin-linapro-oidc-generic/backend/internal/service/oauth"
	settingssvc "lina-plugin-linapro-oidc-generic/backend/internal/service/settings"
)

// ControllerV1 handles browser-facing login-start and callback routes.
type ControllerV1 struct {
	oauthSvc    oauthsvc.Service
	settingsSvc settingssvc.Service
}

// NewV1 creates a login controller.
func NewV1(oauthSvc oauthsvc.Service, settingsSvc settingssvc.Service) *ControllerV1 {
	return &ControllerV1{oauthSvc: oauthSvc, settingsSvc: settingsSvc}
}
