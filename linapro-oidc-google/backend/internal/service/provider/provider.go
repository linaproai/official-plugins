// Package provider implements the Google OIDC auth provider registered with
// the host so the workbench login page can discover a "Continue with Google"
// entry. The provider's runtime metadata (backend redirect toggle, default
// redirect URL, JSON rule map) is sourced from the host's shared plugin
// settings store on every login-entry read so operator changes take effect
// without restarting the plugin.
package provider

import (
	"context"

	"lina-core/pkg/plugin/capability/authprovider"

	configsvc "lina-plugin-linapro-oidc-google/backend/internal/service/config"
)

// providerID is the stable identifier published to the host auth provider
// registry. It matches the OAuth client identifier convention so it can be
// reused inside backend-redirect rule keys.
const providerID = "google"

// Provider implements authprovider.Provider for Google login. The provider
// is intentionally state-less besides the settings dependency so the host
// can re-read live configuration on every request.
type Provider struct {
	// settingsSvc reads the typed Google OIDC settings on demand.
	settingsSvc *configsvc.Service
}

// New constructs a Provider bound to the supplied settings service. Passing
// nil is allowed at static registration time (the LoginEntry call returns
// an error until the settings backend is wired during route registration).
func New(settingsSvc *configsvc.Service) *Provider {
	return &Provider{settingsSvc: settingsSvc}
}

// ProviderID returns the stable Google provider identifier.
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
	return authprovider.KindOIDC
}

// LoginEntry returns the login entry metadata rendered on the workbench
// login page. The metadata reflects the current plugin settings so admin
// toggles take effect immediately.
func (p *Provider) LoginEntry(ctx context.Context) (*authprovider.LoginEntry, error) {
	if p == nil || p.settingsSvc == nil {
		return staticLoginEntry(), nil
	}
	settings, err := p.settingsSvc.Get(ctx)
	if err != nil {
		return nil, err
	}
	entry := staticLoginEntry()
	entry.BackendRedirectEnabled = settings.EnableBackendRedirect
	entry.BackendRedirectDefault = settings.DefaultBackendRedirect
	entry.BackendRedirectRules = settings.BackendRedirects
	return entry, nil
}

// staticLoginEntry builds the immutable parts of the login entry so
// LoginEntry stays focused on the runtime-fetched fields.
func staticLoginEntry() *authprovider.LoginEntry {
	return &authprovider.LoginEntry{
		ProviderID:   providerID,
		PluginID:     configsvc.PluginID,
		Kind:         authprovider.KindOIDC,
		Name:         "Google",
		Description:  "Sign in with a Google account",
		Icon:         "logos:google-icon",
		EntryURL:     "/api/v1/auth/google",
		DisplayOrder: 10,
	}
}
