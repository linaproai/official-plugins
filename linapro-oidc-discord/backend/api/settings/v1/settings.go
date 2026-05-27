// Package v1 defines Discord OAuth2 settings request and response DTOs.
package v1

import "github.com/gogf/gf/v2/frame/g"

// GetSettingsReq defines the request to retrieve Discord OAuth2 settings.
type GetSettingsReq struct {
	g.Meta `path:"/plugin/linapro-oidc-discord/settings" method:"get" tags:"Discord OAuth2" summary:"Get Discord OAuth2 settings" dc:"Retrieve the current Discord OAuth2 login plugin configuration. The client secret is masked in the response." permission:"linapro-oidc-discord:settings:view"`
}

// GetSettingsRes is the Discord OAuth2 settings response.
type GetSettingsRes struct {
	ClientID     string `json:"clientId" dc:"Discord application client ID" eg:"1234567890123456789"`
	ClientSecret string `json:"clientSecret" dc:"Discord application client secret, masked after first save" eg:"abc***xyz"`
	RedirectURI  string `json:"redirectUri" dc:"OAuth2 redirect URI registered in Discord Developer Portal" eg:"https://example.com/api/v1/auth/discord/callback"`
	EnableBackendRedirect bool   `json:"enableBackendRedirect" dc:"Whether to enable post-login backend redirect by state" eg:"true"`
	DefaultBackendRedirect string `json:"defaultBackendRedirect" dc:"Default redirect URL after login when state does not match any rule" eg:"/dashboard"`
	BackendRedirects       string `json:"backendRedirects" dc:"JSON mapping of state to redirect URL, e.g. {\"dashboard\":\"/dashboard\",\"profile\":\"/dashboard/profile\"}" eg:"{\"dashboard\":\"/dashboard\"}"`
	Enabled      bool   `json:"enabled" dc:"Whether Discord login is currently enabled" eg:"true"`
}

// SaveSettingsReq defines the request to save Discord OAuth2 settings.
//
// ClientSecret is intentionally optional: the response masks the stored
// secret to avoid leaking it to the admin UI, and the admin form posts an
// empty value when the operator does not want to rotate the secret. The
// backend keeps the existing stored secret when ClientSecret is empty.
type SaveSettingsReq struct {
	g.Meta       `path:"/plugin/linapro-oidc-discord/settings" method:"put" tags:"Discord OAuth2" summary:"Save Discord OAuth2 settings" dc:"Save Discord OAuth2 login plugin configuration. Leave client secret empty to keep the stored value unchanged." permission:"linapro-oidc-discord:settings:edit"`
	ClientID     string `json:"clientId" v:"required" dc:"Discord application client ID from Discord Developer Portal" eg:"1234567890123456789"`
	ClientSecret string `json:"clientSecret" dc:"Discord application client secret from Discord Developer Portal; leave empty to keep the stored secret unchanged" eg:"xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"`
	RedirectURI  string `json:"redirectUri" v:"required" dc:"OAuth2 redirect URI, must be registered in Discord Developer Portal" eg:"https://example.com/api/v1/auth/discord/callback"`
	EnableBackendRedirect bool   `json:"enableBackendRedirect" dc:"Whether to enable post-login backend redirect by state" eg:"true"`
	DefaultBackendRedirect string `json:"defaultBackendRedirect" dc:"Default redirect URL after login when state does not match any rule" eg:"/dashboard"`
	BackendRedirects       string `json:"backendRedirects" dc:"JSON mapping of state to redirect URL, e.g. {\"dashboard\":\"/dashboard\",\"profile\":\"/dashboard/profile\"}" eg:"{\"dashboard\":\"/dashboard\"}"`
	Enabled      bool   `json:"enabled" dc:"Whether to enable Discord login on the workbench login page" eg:"true"`
}

// SaveSettingsRes is the Discord OAuth2 settings save response.
type SaveSettingsRes struct{}
