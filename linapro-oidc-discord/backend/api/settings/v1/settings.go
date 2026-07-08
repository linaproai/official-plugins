// Package v1 declares the linapro-oidc-discord settings API DTOs. Settings
// values are persisted through the host sys_config seam and read by the OAuth
// login orchestration at request time.

package v1

// SettingsItem is the settings projection returned by the get endpoint. The
// client secret is always masked; the raw value is never returned by the API.
type SettingsItem struct {
	// ClientId is the Discord OAuth 2.0 client ID issued by the Discord Developer Portal.
	ClientId string `json:"clientId" dc:"Discord OAuth 2.0 client ID issued by the Discord Developer Portal" eg:"1234567890123456789"`
	// ClientSecretMasked is the masked client secret indicator. It is never the
	// plaintext value: a non-empty stored secret returns a fixed mask, an empty
	// stored secret returns an empty string.
	ClientSecretMasked string `json:"clientSecretMasked" dc:"Masked client secret indicator; non-empty stored value returns a fixed mask, empty stored value returns empty string. Plaintext secrets are never returned." eg:"************"`
	// ClientSecretConfigured reports whether a client secret is currently
	// stored, so the UI can distinguish a masked value from an unset value.
	ClientSecretConfigured bool `json:"clientSecretConfigured" dc:"True when a client secret is currently stored; used by the UI to distinguish masked from unset" eg:"true"`
	// RedirectUrl is the callback URL registered with Discord.
	RedirectUrl string `json:"redirectUrl" dc:"Fully-qualified callback URL registered with Discord; must resolve to the plugin callback route" eg:"https://your-host/portal/linapro-oidc-discord/callback"`
	// EnableBackendRedirect reports whether SSO token delivery to third-party
	// receiver URLs is enabled.
	EnableBackendRedirect bool `json:"enableBackendRedirect" dc:"True when SSO token delivery to third-party receiver URLs is enabled" eg:"false"`
	// DefaultBackendRedirect is the SPA landing path used after a normal login.
	DefaultBackendRedirect string `json:"defaultBackendRedirect" dc:"Workspace landing path used after a normal external login; empty keeps the host default" eg:"/dashboard/analytics"`
	// BackendRedirects is the JSON object mapping state keys to receiver URLs.
	BackendRedirects string `json:"backendRedirects" dc:"JSON object mapping business state keys to third-party SSO receiver URLs" eg:"{\"crm\":\"https://crm.example.com/sso\"}"`
	// AllowAutoProvision reports whether unlinked verified identities may be
	// auto-provisioned as least-privilege platform users.
	AllowAutoProvision bool `json:"allowAutoProvision" dc:"True when the host may auto-provision a least-privilege platform user for an unlinked verified Discord identity" eg:"false"`
}
