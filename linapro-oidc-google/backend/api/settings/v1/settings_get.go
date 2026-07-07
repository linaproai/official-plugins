// settings_get.go declares the get-settings request and response DTOs for the
// linapro-oidc-google settings API.

package v1

import "github.com/gogf/gf/v2/frame/g"

// GetSettingsReq is the request for querying the linapro-oidc-google settings.
type GetSettingsReq struct {
	g.Meta `path:"/plugins/linapro-oidc-google/settings" method:"get" tags:"Google OIDC Login" summary:"Query Google OIDC settings" dc:"Return the persisted linapro-oidc-google settings for the admin settings page. The client secret is always returned as a masked indicator; plaintext secrets are never sent to the client." permission:"linapro-oidc-google:settings:view"`
}

// GetSettingsRes is the response for querying the linapro-oidc-google settings.
type GetSettingsRes struct {
	// Settings carries the projected settings row read from sys_config.
	Settings *SettingsItem `json:"settings" dc:"Persisted settings values with client secret masked" eg:"{}"`
}
