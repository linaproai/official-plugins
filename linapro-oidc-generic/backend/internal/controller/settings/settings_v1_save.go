// settings_v1_save.go implements save-settings and projection helper.

package settings

import (
	"context"

	v1 "lina-plugin-linapro-oidc-generic/backend/api/settings/v1"
	settingssvc "lina-plugin-linapro-oidc-generic/backend/internal/service/settings"
)

// SaveSettings persists settings and returns the fresh masked projection.
func (c *ControllerV1) SaveSettings(ctx context.Context, req *v1.SaveSettingsReq) (res *v1.SaveSettingsRes, err error) {
	projection, err := c.settingsSvc.Save(ctx, settingssvc.SaveInput{
		DisplayName:            req.DisplayName,
		Issuer:                 req.Issuer,
		ClientID:               req.ClientId,
		ClientSecret:           req.ClientSecret,
		RedirectURL:            req.RedirectUrl,
		Scopes:                 req.Scopes,
		AllowAutoProvision:     req.AllowAutoProvision,
		DefaultBackendRedirect: req.DefaultBackendRedirect,
	})
	if err != nil {
		return nil, err
	}
	return &v1.SaveSettingsRes{Settings: projectSettingsItem(projection)}, nil
}

func projectSettingsItem(projection *settingssvc.Projection) *v1.SettingsItem {
	if projection == nil {
		return &v1.SettingsItem{}
	}
	return &v1.SettingsItem{
		ConnectionKey:          projection.ConnectionKey,
		DisplayName:            projection.DisplayName,
		Issuer:                 projection.Issuer,
		ClientId:               projection.ClientID,
		ClientSecretMasked:     projection.ClientSecretMasked,
		ClientSecretConfigured: projection.ClientSecretConfigured,
		RedirectUrl:            projection.RedirectURL,
		Scopes:                 projection.Scopes,
		AllowAutoProvision:     projection.AllowAutoProvision,
		DefaultBackendRedirect: projection.DefaultBackendRedirect,
	}
}
