// Package settings implements admin settings HTTP handlers for linapro-storage-aws.
package settings

import (
	objstore "lina-plugin-linapro-storage-aws/backend/internal/service/objstore"
	settingssvc "lina-plugin-linapro-storage-aws/backend/internal/service/settings"
)

// ControllerV1 binds settings endpoints.
type ControllerV1 struct {
	settingsSvc settingssvc.Service
	tester      objstore.ConnectionTester
}

// NewV1 creates the settings controller.
func NewV1(settingsSvc settingssvc.Service, tester objstore.ConnectionTester) *ControllerV1 {
	return &ControllerV1{settingsSvc: settingsSvc, tester: tester}
}
