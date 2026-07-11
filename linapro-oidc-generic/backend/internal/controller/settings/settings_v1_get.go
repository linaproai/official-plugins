// settings_v1_get.go implements get-settings.

package settings

import (
	"context"

	v1 "lina-plugin-linapro-oidc-generic/backend/api/settings/v1"
)

// GetSettings returns masked settings projection.
func (c *ControllerV1) GetSettings(ctx context.Context, _ *v1.GetSettingsReq) (res *v1.GetSettingsRes, err error) {
	projection, err := c.settingsSvc.Get(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.GetSettingsRes{Settings: projectSettingsItem(projection)}, nil
}
