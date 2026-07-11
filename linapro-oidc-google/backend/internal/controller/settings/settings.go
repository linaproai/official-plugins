// Package settings implements the linapro-oidc-google settings HTTP
// controller. The controller stays thin: it forwards DTO values to the
// settings service and projects the masked settings result back into the
// response DTO.
package settings

import (
	settingssvc "lina-plugin-linapro-oidc-google/backend/internal/service/settings"
)

// ControllerV1 is the linapro-oidc-google settings controller.
type ControllerV1 struct {
	settingsSvc settingssvc.Service // settingsSvc persists and projects the plugin settings.
}

// NewV1 creates and returns a new linapro-oidc-google settings controller instance.
func NewV1(settingsSvc settingssvc.Service) *ControllerV1 {
	return &ControllerV1{settingsSvc: settingsSvc}
}
