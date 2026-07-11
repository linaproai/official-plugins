// Package login implements the HTTP entry points for the Google OIDC login
// flow exposed by the linapro-oidc-google plugin. The controller is
// intentionally thin; all orchestration lives in the oauth service so the
// route handlers only translate between GoFrame request objects and service
// inputs/outputs.
package login

import (
	oauthsvc "lina-plugin-linapro-oidc-google/backend/internal/service/oauth"
	settingssvc "lina-plugin-linapro-oidc-google/backend/internal/service/settings"
)

// ControllerV1 is the linapro-oidc-google login controller. It handles the
// browser-facing login-start and callback routes registered on the portal
// route group.
type ControllerV1 struct {
	oauthSvc    oauthsvc.Service    // oauthSvc orchestrates the Google OIDC login flow.
	settingsSvc settingssvc.Service // settingsSvc supplies SSO delivery rules and the SPA landing path at request time.
}

// NewV1 creates and returns one Google OIDC login controller instance.
func NewV1(oauthSvc oauthsvc.Service, settingsSvc settingssvc.Service) *ControllerV1 {
	return &ControllerV1{oauthSvc: oauthSvc, settingsSvc: settingsSvc}
}
