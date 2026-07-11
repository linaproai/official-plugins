// settings_v1_save.go implements the save-settings endpoint of the
// linapro-oidc-discord settings controller and the shared projection helper.

package settings

import (
	"context"

	v1 "lina-plugin-linapro-oidc-discord/backend/api/settings/v1"
	settingssvc "lina-plugin-linapro-oidc-discord/backend/internal/service/settings"
)

// SaveSettings persists the submitted settings values and returns the fresh
// masked projection. An empty or masked client secret keeps the previously
// stored value.
func (c *ControllerV1) SaveSettings(ctx context.Context, req *v1.SaveSettingsReq) (res *v1.SaveSettingsRes, err error) {
	projection, err := c.settingsSvc.Save(ctx, settingssvc.SaveInput{
		ClientID:               req.ClientId,
		ClientSecret:           req.ClientSecret,
		RedirectURL:            req.RedirectUrl,
		EnableBackendRedirect:  req.EnableBackendRedirect,
		DefaultBackendRedirect: req.DefaultBackendRedirect,
		BackendRedirects:       req.BackendRedirects,
		AllowAutoProvision:     req.AllowAutoProvision,
	})
	if err != nil {
		return nil, err
	}
	return &v1.SaveSettingsRes{Settings: projectSettingsItem(projection)}, nil
}

// projectSettingsItem maps the service projection into the API DTO shape.
func projectSettingsItem(projection *settingssvc.Projection) *v1.SettingsItem {
	if projection == nil {
		return &v1.SettingsItem{}
	}
	return &v1.SettingsItem{
		ClientId:               projection.ClientID,
		ClientSecretMasked:     projection.ClientSecretMasked,
		ClientSecretConfigured: projection.ClientSecretConfigured,
		RedirectUrl:            projection.RedirectURL,
		EnableBackendRedirect:  projection.EnableBackendRedirect,
		DefaultBackendRedirect: projection.DefaultBackendRedirect,
		BackendRedirects:       projection.BackendRedirects,
		AllowAutoProvision:     projection.AllowAutoProvision,
	}
}
