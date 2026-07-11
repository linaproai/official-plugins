// settings_save.go declares save-settings request/response DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// SaveSettingsReq is the request for saving generic OIDC settings.
type SaveSettingsReq struct {
	g.Meta                 `path:"/settings" method:"put" tags:"Generic OIDC Login" summary:"Save Generic OIDC settings" dc:"Persist linapro-oidc-generic settings. Empty or masked client secret keeps the previous value. Auto-provision defaults off." permission:"linapro-oidc-generic:settings:update"`
	DisplayName            string `json:"displayName" v:"max-length:128" dc:"Login button display name" eg:"Company SSO"`
	Issuer                 string `json:"issuer" v:"max-length:512" dc:"OIDC issuer URL" eg:"https://keycloak.example.com/realms/demo"`
	ClientId               string `json:"clientId" v:"max-length:256" dc:"OAuth 2.0 client ID" eg:"linapro-web"`
	ClientSecret           string `json:"clientSecret" v:"max-length:512" dc:"Client secret; empty or mask keeps previous" eg:""`
	RedirectUrl            string `json:"redirectUrl" v:"max-length:512" dc:"Optional callback URL override" eg:""`
	Scopes                 string `json:"scopes" v:"max-length:512" dc:"Space-separated scopes" eg:"openid email profile"`
	AllowAutoProvision     bool   `json:"allowAutoProvision" dc:"Allow auto-provision; default false" eg:"false"`
	DefaultBackendRedirect string `json:"defaultBackendRedirect" v:"max-length:512" dc:"SPA landing path" eg:""`
}

// SaveSettingsRes is the response after saving settings.
type SaveSettingsRes struct {
	Settings *SettingsItem `json:"settings" dc:"Fresh masked projection after save"`
}
