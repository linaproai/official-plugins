// GetSettings handler for Azure Blob storage settings.
package settings

import (
	"context"

	v1 "lina-plugin-linapro-storage-azure/backend/api/settings/v1"
	settingssvc "lina-plugin-linapro-storage-azure/backend/internal/service/settings"
)

// GetSettings returns masked settings.
func (c *ControllerV1) GetSettings(ctx context.Context, _ *v1.GetSettingsReq) (*v1.GetSettingsRes, error) {
	projection, err := c.settingsSvc.Get(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.GetSettingsRes{Settings: projectItem(projection)}, nil
}

func projectItem(p *settingssvc.Projection) *v1.SettingsItem {
	if p == nil {
		return &v1.SettingsItem{}
	}
	return &v1.SettingsItem{
		AccountName:          p.AccountName,
		AccountKeyMasked:     p.AccountKeyMasked,
		AccountKeyConfigured: p.AccountKeyConfigured,
		Container:            p.Container,
		Endpoint:             p.Endpoint,
		PathPrefix:           p.PathPrefix,
	}
}
