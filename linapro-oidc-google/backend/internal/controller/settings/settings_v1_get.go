// settings_v1_get.go implements the get-settings endpoint of the
// linapro-oidc-google settings controller.

package settings

import (
	"context"

	v1 "lina-plugin-linapro-oidc-google/backend/api/settings/v1"
)

// GetSettings returns the persisted settings projection with the client
// secret masked.
func (c *ControllerV1) GetSettings(ctx context.Context, _ *v1.GetSettingsReq) (res *v1.GetSettingsRes, err error) {
	projection, err := c.settingsSvc.Get(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.GetSettingsRes{Settings: projectSettingsItem(projection)}, nil
}
