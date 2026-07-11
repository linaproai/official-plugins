// Package v1 declares the linapro-oidc-generic settings API DTOs.

package v1

// SettingsItem is the settings projection returned by the get endpoint.
type SettingsItem struct {
	ConnectionKey          string `json:"connectionKey" dc:"Stable connection key used in provider encoding; v1 is fixed to default" eg:"default"`
	DisplayName            string `json:"displayName" dc:"Login button display name" eg:"Company SSO"`
	Issuer                 string `json:"issuer" dc:"OIDC issuer URL" eg:"https://keycloak.example.com/realms/demo"`
	ClientId               string `json:"clientId" dc:"OAuth 2.0 client ID" eg:"linapro-web"`
	ClientSecretMasked     string `json:"clientSecretMasked" dc:"Masked client secret indicator" eg:"************"`
	ClientSecretConfigured bool   `json:"clientSecretConfigured" dc:"True when a client secret is stored" eg:"true"`
	RedirectUrl            string `json:"redirectUrl" dc:"Optional override callback URL; empty derives from request host" eg:""`
	Scopes                 string `json:"scopes" dc:"Space-separated OIDC scopes; openid is always required" eg:"openid email profile"`
	AllowAutoProvision     bool   `json:"allowAutoProvision" dc:"Allow host auto-provision for unlinked identities; default false" eg:"false"`
	DefaultBackendRedirect string `json:"defaultBackendRedirect" dc:"SPA landing path after login" eg:"/dashboard/analytics"`
}
