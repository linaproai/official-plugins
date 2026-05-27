// Package v1 defines Google OIDC settings request and response DTOs.
package v1

import "github.com/gogf/gf/v2/frame/g"

// GetSettingsReq defines the request to retrieve Google OIDC settings.
type GetSettingsReq struct {
	g.Meta `path:"/plugin/linapro-oidc-google/settings" method:"get" tags:"Google OIDC" summary:"Get Google OIDC settings" dc:"Retrieve the current Google OIDC login plugin configuration including client ID, redirect URI, and enabled status. The client secret is masked in the response." permission:"linapro-oidc-google:settings:view"`
}

// GetSettingsRes is the Google OIDC settings response.
type GetSettingsRes struct {
	ClientID     string `json:"clientId" dc:"Google OAuth2 client ID" eg:"123456789-abc.apps.googleusercontent.com"`
	ClientSecret string `json:"clientSecret" dc:"Google OAuth2 client secret, masked after first save" eg:"GOCSPX-***"`
	RedirectURI  string `json:"redirectUri" dc:"OAuth2 redirect URI registered in Google Console" eg:"https://example.com/api/v1/auth/google/callback"`
	EnableBackendRedirect bool   `json:"enableBackendRedirect" dc:"Whether to enable post-login backend redirect by state" eg:"true"`
	DefaultBackendRedirect string `json:"defaultBackendRedirect" dc:"Default redirect URL after login when state does not match any rule" eg:"/dashboard"`
	BackendRedirects       string `json:"backendRedirects" dc:"JSON mapping of state to redirect URL, e.g. {\"dashboard\":\"/dashboard\",\"profile\":\"/dashboard/profile\"}" eg:"{\"dashboard\":\"/dashboard\"}"`
	Enabled      bool   `json:"enabled" dc:"Whether Google login is currently enabled" eg:"true"`
}

// SaveSettingsReq defines the request to save Google OIDC settings.
//
// ClientSecret is intentionally optional: the response masks the stored
// secret to avoid leaking it to the admin UI, and the admin form posts an
// empty value when the operator does not want to rotate the secret. The
// backend keeps the existing stored secret when ClientSecret is empty.
type SaveSettingsReq struct {
	g.Meta       `path:"/plugin/linapro-oidc-google/settings" method:"put" tags:"Google OIDC" summary:"Save Google OIDC settings" dc:"Save Google OIDC login plugin configuration. Leave client secret empty to keep the stored value unchanged." permission:"linapro-oidc-google:settings:edit"`
	ClientID     string `json:"clientId" v:"required" dc:"Google OAuth2 client ID from Google Cloud Console" eg:"123456789-abc.apps.googleusercontent.com"`
	ClientSecret string `json:"clientSecret" dc:"Google OAuth2 client secret from Google Cloud Console; leave empty to keep the stored secret unchanged" eg:"GOCSPX-xxxxxxxxxxxxxxxxxxxxxxxx"`
	RedirectURI  string `json:"redirectUri" v:"required" dc:"OAuth2 redirect URI, must be registered in Google Console" eg:"https://example.com/api/v1/auth/google/callback"`
	EnableBackendRedirect bool   `json:"enableBackendRedirect" dc:"Whether to enable post-login backend redirect by state" eg:"true"`
	DefaultBackendRedirect string `json:"defaultBackendRedirect" dc:"Default redirect URL after login when state does not match any rule" eg:"/dashboard"`
	BackendRedirects       string `json:"backendRedirects" dc:"JSON mapping of state to redirect URL, e.g. {\"dashboard\":\"/dashboard\",\"profile\":\"/dashboard/profile\"}" eg:"{\"dashboard\":\"/dashboard\"}"`
	Enabled      bool   `json:"enabled" dc:"Whether to enable Google login on the workbench login page" eg:"true"`
}

// SaveSettingsRes is the Google OIDC settings save response.
type SaveSettingsRes struct{}
