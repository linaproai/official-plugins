package settings

import (
	"context"

	v1 "lina-plugin-linapro-auth-ldap/backend/api/settings/v1"
)

func (c *ControllerV1) GetSettings(ctx context.Context, _ *v1.GetSettingsReq) (*v1.GetSettingsRes, error) {
	p, err := c.settingsSvc.Get(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.GetSettingsRes{Settings: project(p)}, nil
}
