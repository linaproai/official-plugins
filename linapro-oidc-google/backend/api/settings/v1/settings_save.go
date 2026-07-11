// settings_save.go declares the save-settings request and response DTOs for
// the linapro-oidc-google settings API.

package v1

import "github.com/gogf/gf/v2/frame/g"

// SaveSettingsReq is the request for saving the linapro-oidc-google settings.
type SaveSettingsReq struct {
	g.Meta       `path:"/settings" method:"put" tags:"Auth Login / Google" summary:"Save Google OIDC settings" dc:"Persist the linapro-oidc-google settings to sys_config. An empty or masked client secret keeps the previously stored value; a non-empty non-masked value replaces it." permission:"linapro-oidc-google:settings:update"`
	ClientId     string `json:"clientId" v:"max-length:256" dc:"Google OAuth 2.0 client ID; empty string clears the stored value" eg:"1234567890-abc.apps.googleusercontent.com"`
	ClientSecret string `json:"clientSecret" v:"max-length:512" dc:"Google OAuth 2.0 client secret; empty string or the masked indicator keeps the previously stored value, any other value replaces it" eg:"GOCSPX-secret-value"`
	RedirectUrl  string `json:"redirectUrl" v:"max-length:512" dc:"Fully-qualified callback URL registered with Google; empty string clears the stored value" eg:"https://your-host/portal/linapro-oidc-google/callback"`
	// EnableBackendRedirect toggles SSO token delivery to third-party receivers.
	EnableBackendRedirect bool `json:"enableBackendRedirect" dc:"Enable SSO token delivery to third-party receiver URLs matched by state key" eg:"false"`
	// DefaultBackendRedirect sets the SPA landing path after a normal login.
	DefaultBackendRedirect string `json:"defaultBackendRedirect" v:"max-length:512" dc:"Workspace landing path after a normal external login; empty string keeps the host default" eg:"/dashboard/analytics"`
	// BackendRedirects sets the state-key-to-receiver-URL JSON object.
	BackendRedirects string `json:"backendRedirects" v:"max-length:4096" dc:"JSON object mapping business state keys to third-party SSO receiver URLs; empty string clears all rules" eg:"{\"crm\":\"https://crm.example.com/sso\"}"`
	// AllowAutoProvision toggles host auto-provisioning for unlinked identities.
	AllowAutoProvision bool `json:"allowAutoProvision" dc:"Allow the host to auto-provision a least-privilege platform user when a verified Google identity has no linked local account" eg:"false"`
	// EnableOneTap toggles the embeddable One Tap login endpoint.
	EnableOneTap bool `json:"enableOneTap" dc:"Enable the embeddable Google One Tap endpoint that accepts GSI ID Token credentials" eg:"false"`
}

// SaveSettingsRes is the response for saving the linapro-oidc-google settings.
type SaveSettingsRes struct {
	// Settings carries the fresh masked projection after the save applies.
	Settings *SettingsItem `json:"settings" dc:"Fresh settings projection after the save applies, with client secret masked" eg:"{}"`
}
