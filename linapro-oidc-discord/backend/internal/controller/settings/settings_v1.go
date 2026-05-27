// Package settings implements the Discord OAuth2 plugin settings controller.
package settings

import (
	"context"

	v1 "lina-plugin-linapro-oidc-discord/backend/api/settings/v1"
	configsvc "lina-plugin-linapro-oidc-discord/backend/internal/service/config"
)

// ControllerV1 handles Discord OAuth2 plugin settings endpoints.
type ControllerV1 struct {
	// settingsSvc reads and writes the typed Discord OAuth2 settings.
	settingsSvc *configsvc.Service
}

// NewV1 creates a ControllerV1 bound to the supplied settings service.
func NewV1(settingsSvc *configsvc.Service) *ControllerV1 {
	return &ControllerV1{settingsSvc: settingsSvc}
}

// GetSettings returns the current Discord OAuth2 settings with the stored
// client secret replaced by its masked projection so admin UIs cannot leak
// the raw secret.
func (c *ControllerV1) GetSettings(ctx context.Context, _ *v1.GetSettingsReq) (res *v1.GetSettingsRes, err error) {
	settings, err := c.settingsSvc.Get(ctx)
	if err != nil {
		return nil, err
	}
	maskedSecret, err := c.settingsSvc.GetMaskedClientSecret(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.GetSettingsRes{
		ClientID:               settings.ClientID,
		ClientSecret:           maskedSecret,
		RedirectURI:            settings.RedirectURI,
		EnableBackendRedirect:  settings.EnableBackendRedirect,
		DefaultBackendRedirect: settings.DefaultBackendRedirect,
		BackendRedirects:       settings.BackendRedirects,
		Enabled:                settings.Enabled,
	}, nil
}

// SaveSettings stores the Discord OAuth2 settings. Leaving ClientSecret
// empty preserves the stored secret because the DTO is documented as
// "leave empty to keep current".
func (c *ControllerV1) SaveSettings(ctx context.Context, req *v1.SaveSettingsReq) (res *v1.SaveSettingsRes, err error) {
	err = c.settingsSvc.Save(ctx, &configsvc.Settings{
		ClientID:               req.ClientID,
		ClientSecret:           req.ClientSecret,
		RedirectURI:            req.RedirectURI,
		EnableBackendRedirect:  req.EnableBackendRedirect,
		DefaultBackendRedirect: req.DefaultBackendRedirect,
		BackendRedirects:       req.BackendRedirects,
		Enabled:                req.Enabled,
	})
	if err != nil {
		return nil, err
	}
	return &v1.SaveSettingsRes{}, nil
}
