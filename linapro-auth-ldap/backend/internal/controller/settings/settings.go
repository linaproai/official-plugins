package settings

import settingssvc "lina-plugin-linapro-auth-ldap/backend/internal/service/settings"

type ControllerV1 struct {
	settingsSvc settingssvc.Service
}

func NewV1(settingsSvc settingssvc.Service) *ControllerV1 {
	return &ControllerV1{settingsSvc: settingsSvc}
}
