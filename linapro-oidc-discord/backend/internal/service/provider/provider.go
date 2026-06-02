// Package provider implements the Discord OAuth2 auth provider registered
// with the host so the workbench login page can discover a "Continue with
// Discord" entry. The public login-entry projection is static by design;
// redirect rules and token-delivery settings are read only by authenticated
// settings APIs and OAuth callback handlers.
package provider

import (
	"context"

	"lina-core/pkg/plugin/capability/authprovider"

	configsvc "lina-plugin-linapro-oidc-discord/backend/internal/service/config"
)

// providerID is the stable identifier published to the host auth provider
// registry.
const providerID = "discord"

// Provider implements authprovider.Provider for Discord login.
type Provider struct {
	// settingsSvc is retained for construction compatibility; public discovery
	// must not read it because /auth/providers is anonymous and high traffic.
	settingsSvc *configsvc.Service
}

// New constructs a Provider bound to the supplied settings service.
func New(settingsSvc *configsvc.Service) *Provider {
	return &Provider{settingsSvc: settingsSvc}
}

// ProviderID returns the stable Discord provider identifier.
func (p *Provider) ProviderID() string {
	return providerID
}

// PluginID returns the owning source-plugin identifier.
func (p *Provider) PluginID() string {
	return configsvc.PluginID
}

// Kind returns the provider kind so the host can route the login entry to
// the right UI component.
func (p *Provider) Kind() authprovider.Kind {
	return authprovider.KindOAuth2
}

// LoginEntry returns only the public login button metadata rendered on the
// workbench login page. It deliberately avoids reading plugin settings so the
// anonymous provider list has bounded database access and never exposes SSO
// redirect-rule state keys or receiver URLs.
func (p *Provider) LoginEntry(_ context.Context) (*authprovider.LoginEntry, error) {
	return staticLoginEntry(), nil
}

// staticLoginEntry builds the immutable parts of the login entry so
// LoginEntry stays focused on the runtime-fetched fields.
func staticLoginEntry() *authprovider.LoginEntry {
	return &authprovider.LoginEntry{
		ProviderID:   providerID,
		PluginID:     configsvc.PluginID,
		Kind:         authprovider.KindOAuth2,
		Name:         "Discord",
		Description:  "Sign in with a Discord account",
		Icon:         "logos:discord-icon",
		EntryURL:     "/api/v1/auth/discord",
		DisplayOrder: 20,
	}
}
