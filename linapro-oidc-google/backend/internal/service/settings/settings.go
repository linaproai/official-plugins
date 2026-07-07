// Package settings implements the linapro-oidc-google admin settings service.
// It persists Google OIDC client credentials and the redirect URL through the
// host sys_config seam, and exposes both a masked projection for the admin
// settings page and a raw Snapshot used by the OAuth login orchestration at
// request time.
package settings

import (
	"context"

	"lina-core/pkg/plugin/capability/hostconfigcap"
)

// Plugin-scoped sys_config keys owned by the linapro-oidc-google plugin.
// Rows are seeded by the plugin manifest SQL at platform tenant scope
// (`tenant_id = 0`) so the host SysConfig seam can update them in place.
const (
	// ConfigKeyClientID is the sys_config key holding the Google client ID.
	ConfigKeyClientID hostconfigcap.SysConfigKey = "plugin.linapro-oidc-google.client_id"
	// ConfigKeyClientSecret is the sys_config key holding the Google client secret.
	ConfigKeyClientSecret hostconfigcap.SysConfigKey = "plugin.linapro-oidc-google.client_secret"
	// ConfigKeyRedirectURL is the sys_config key holding the Google redirect URL.
	ConfigKeyRedirectURL hostconfigcap.SysConfigKey = "plugin.linapro-oidc-google.redirect_url"
	// ConfigKeyEnableBackendRedirect is the sys_config key holding the SSO
	// token-delivery enablement flag ("1" enabled, anything else disabled).
	ConfigKeyEnableBackendRedirect hostconfigcap.SysConfigKey = "plugin.linapro-oidc-google.enable_backend_redirect"
	// ConfigKeyDefaultBackendRedirect is the sys_config key holding the SPA
	// landing path used after a normal (non-SSO) external login.
	ConfigKeyDefaultBackendRedirect hostconfigcap.SysConfigKey = "plugin.linapro-oidc-google.default_backend_redirect"
	// ConfigKeyBackendRedirects is the sys_config key holding the JSON object
	// that maps business state keys to third-party SSO receiver URLs.
	ConfigKeyBackendRedirects hostconfigcap.SysConfigKey = "plugin.linapro-oidc-google.backend_redirects"
)

// SecretMask is the fixed indicator returned to the client whenever a client
// secret is currently stored. It intentionally does not embed characters from
// the real secret so the projection carries no plaintext material.
const SecretMask = "************"

// Snapshot describes the resolved settings values used by the OAuth login
// flow. Fields hold empty strings when unset; the OAuth service is responsible
// for layering safe defaults on top of unset fields.
type Snapshot struct {
	// ClientID is the raw stored Google client ID.
	ClientID string
	// ClientSecret is the raw stored Google client secret. Callers must never
	// project this value to a public API response.
	ClientSecret string
	// RedirectURL is the raw stored Google redirect URL.
	RedirectURL string
	// EnableBackendRedirect reports whether SSO token delivery to third-party
	// receiver URLs is enabled.
	EnableBackendRedirect bool
	// DefaultBackendRedirect is the SPA landing path used after a normal
	// (non-SSO) external login. Empty keeps the host default landing.
	DefaultBackendRedirect string
	// BackendRedirects is the raw JSON object mapping business state keys to
	// third-party SSO receiver URLs.
	BackendRedirects string
}

// Projection is the masked settings projection returned to the admin settings
// page. It never carries the raw client secret; ClientSecretMasked is either
// SecretMask or an empty string, and ClientSecretConfigured reflects whether a
// non-empty raw secret is currently persisted.
type Projection struct {
	// ClientID is the raw stored Google client ID.
	ClientID string
	// ClientSecretMasked reports SecretMask when a stored secret exists and an
	// empty string when no secret is stored yet.
	ClientSecretMasked string
	// ClientSecretConfigured reports whether a non-empty raw secret is stored.
	ClientSecretConfigured bool
	// RedirectURL is the raw stored Google redirect URL.
	RedirectURL string
	// EnableBackendRedirect reports whether SSO token delivery is enabled.
	EnableBackendRedirect bool
	// DefaultBackendRedirect is the SPA landing path after a normal login.
	DefaultBackendRedirect string
	// BackendRedirects is the raw JSON object of state key to receiver URL.
	BackendRedirects string
}

// SaveInput carries the caller-supplied settings values submitted through the
// admin settings API. An empty ClientSecret or a value equal to SecretMask
// signals that the previously stored secret must be kept.
type SaveInput struct {
	// ClientID replaces the stored client ID.
	ClientID string
	// ClientSecret replaces the stored client secret unless empty or masked.
	ClientSecret string
	// RedirectURL replaces the stored redirect URL.
	RedirectURL string
	// EnableBackendRedirect replaces the stored SSO delivery enablement flag.
	EnableBackendRedirect bool
	// DefaultBackendRedirect replaces the stored SPA landing path.
	DefaultBackendRedirect string
	// BackendRedirects replaces the stored state-key-to-receiver JSON object.
	// A non-empty value must parse as a JSON object of string values.
	BackendRedirects string
}

// Service defines the linapro-oidc-google settings service surface used by
// both the admin settings HTTP API and the OAuth login orchestration.
type Service interface {
	// Get returns the masked projection consumed by the admin settings page.
	Get(ctx context.Context) (*Projection, error)
	// Save persists the caller-supplied settings values, honoring the
	// "empty or masked secret keeps existing" contract.
	Save(ctx context.Context, in SaveInput) (*Projection, error)
	// Load returns the raw Snapshot consumed by the OAuth login flow. The
	// snapshot must be re-read on every OAuth request so admin edits take
	// effect without a restart.
	Load(ctx context.Context) (*Snapshot, error)
}

// Interface compliance assertion for the default settings service implementation.
var _ Service = (*serviceImpl)(nil)

// serviceImpl reads and writes plugin-scoped settings through the host
// sys_config seam. It holds an explicit reference to the seam so wiring stays
// visible and testable.
type serviceImpl struct {
	// sysConfigSvc is the host sys_config service used for persisted reads and writes.
	sysConfigSvc hostconfigcap.SysConfigService
}

// New creates and returns a new settings service instance bound to the
// supplied host sys_config service.
func New(sysConfigSvc hostconfigcap.SysConfigService) Service {
	return &serviceImpl{sysConfigSvc: sysConfigSvc}
}
