// Package config implements typed access to the Discord OAuth2 plugin's
// runtime configuration. The persistence backend is the host's shared
// PluginSettingsService (which stores rows in sys_config keyed by
// "<pluginID>.<settingKey>"), so this package only owns the typed Settings
// shape, the per-field setting keys, and the parse/serialize logic.
package config

import (
	"context"
	"strconv"

	"github.com/gogf/gf/v2/errors/gerror"

	plugincontract "lina-core/pkg/plugin/capability/contract"
)

// PluginID is the canonical source-plugin identifier whose settings this
// package owns.
const PluginID = "linapro-oidc-discord"

// Settings is the typed projection of the Discord OAuth2 plugin
// configuration.
type Settings struct {
	// ClientID is the Discord application client identifier.
	ClientID string
	// ClientSecret is the Discord application client secret. Empty values
	// returned by Get mean "no secret stored". Empty values passed to Save
	// mean "preserve the stored secret unchanged".
	ClientSecret string
	// RedirectURI is the OAuth2 redirect URI registered in Discord
	// Developer Portal.
	RedirectURI string
	// EnableBackendRedirect controls whether the provider appends a state
	// parameter to the login entry URL so the callback can route the user
	// to a state-specific redirect URL after login.
	EnableBackendRedirect bool
	// DefaultBackendRedirect is the fallback URL used when the runtime
	// state does not match any configured redirect rule.
	DefaultBackendRedirect string
	// BackendRedirects is the JSON object mapping state keys to redirect
	// URLs, stored verbatim because the host treats it as opaque.
	BackendRedirects string
	// Enabled reports whether Discord login is currently enabled.
	Enabled bool
}

// Setting-key constants identify each field inside the plugin's sys_config
// namespace.
const (
	keyClientID               = "clientId"
	keyClientSecret           = "clientSecret"
	keyRedirectURI            = "redirectUri"
	keyEnableBackendRedirect  = "enableBackendRedirect"
	keyDefaultBackendRedirect = "defaultBackendRedirect"
	keyBackendRedirects       = "backendRedirects"
	keyEnabled                = "enabled"
)

// defaultDefaultBackendRedirect is the built-in fallback redirect URL used
// when no operator value is configured.
const defaultDefaultBackendRedirect = "/dashboard"

// Service reads and writes Discord OAuth2 plugin settings through the
// shared host settings store.
type Service struct {
	// settings is the host-provided namespaced key-value store.
	settings plugincontract.PluginSettingsService
}

// New constructs a Service bound to the host-published settings backend.
func New(settings plugincontract.PluginSettingsService) *Service {
	return &Service{settings: settings}
}

// Get returns the current Discord OAuth2 settings, applying built-in
// defaults for missing values.
func (s *Service) Get(ctx context.Context) (*Settings, error) {
	if s == nil || s.settings == nil {
		return nil, gerror.New("discord oauth2 settings service is not initialized")
	}
	values, err := s.settings.List(ctx, PluginID)
	if err != nil {
		return nil, err
	}
	out := &Settings{
		ClientID:               values[keyClientID],
		ClientSecret:           values[keyClientSecret],
		RedirectURI:            values[keyRedirectURI],
		EnableBackendRedirect:  parseBool(values[keyEnableBackendRedirect], false),
		DefaultBackendRedirect: stringOrDefault(values[keyDefaultBackendRedirect], defaultDefaultBackendRedirect),
		BackendRedirects:       values[keyBackendRedirects],
		Enabled:                parseBool(values[keyEnabled], false),
	}
	return out, nil
}

// Save persists the supplied Discord OAuth2 settings. ClientSecret follows
// the secret-style "leave empty to preserve" convention because admin UIs
// mask the stored secret on read; if the operator leaves the field blank
// we must not overwrite the persisted secret with an empty value.
func (s *Service) Save(ctx context.Context, in *Settings) error {
	if s == nil || s.settings == nil {
		return gerror.New("discord oauth2 settings service is not initialized")
	}
	if in == nil {
		return gerror.New("discord oauth2 settings save requires non-nil payload")
	}
	if err := s.settings.SetString(ctx, PluginID, keyClientID, in.ClientID); err != nil {
		return err
	}
	if err := s.settings.SetSecret(ctx, PluginID, keyClientSecret, in.ClientSecret); err != nil {
		return err
	}
	if err := s.settings.SetString(ctx, PluginID, keyRedirectURI, in.RedirectURI); err != nil {
		return err
	}
	if err := s.settings.SetString(ctx, PluginID, keyEnableBackendRedirect, formatBool(in.EnableBackendRedirect)); err != nil {
		return err
	}
	if err := s.settings.SetString(ctx, PluginID, keyDefaultBackendRedirect, in.DefaultBackendRedirect); err != nil {
		return err
	}
	if err := s.settings.SetString(ctx, PluginID, keyBackendRedirects, in.BackendRedirects); err != nil {
		return err
	}
	if err := s.settings.SetString(ctx, PluginID, keyEnabled, formatBool(in.Enabled)); err != nil {
		return err
	}
	return nil
}

// GetMaskedClientSecret returns a non-reversible masked projection of the
// stored secret suitable for displaying in admin UIs.
func (s *Service) GetMaskedClientSecret(ctx context.Context) (string, error) {
	if s == nil || s.settings == nil {
		return "", nil
	}
	return s.settings.GetMaskedSecret(ctx, PluginID, keyClientSecret)
}

// parseBool tolerantly parses a stored bool value, treating malformed
// values as the supplied default to keep callers from breaking on legacy
// rows or admin-UI typos.
func parseBool(raw string, defaultValue bool) bool {
	if raw == "" {
		return defaultValue
	}
	parsed, err := strconv.ParseBool(raw)
	if err != nil {
		return defaultValue
	}
	return parsed
}

// formatBool serializes a bool value into the canonical "true"/"false"
// string stored in sys_config so reads stay deterministic.
func formatBool(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

// stringOrDefault returns fallback when the supplied value is empty.
func stringOrDefault(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
