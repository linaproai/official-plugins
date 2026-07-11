// Package settings implements the linapro-oidc-generic settings HTTP controller.

package settings

import (
	settingssvc "lina-plugin-linapro-oidc-generic/backend/internal/service/settings"
)

// ControllerV1 is the settings controller.
type ControllerV1 struct {
	settingsSvc settingssvc.Service
}

// NewV1 creates a settings controller.
func NewV1(settingsSvc settingssvc.Service) *ControllerV1 {
	return &ControllerV1{settingsSvc: settingsSvc}
}
