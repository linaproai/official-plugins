// settings_get.go declares the get-settings request and response DTOs for the
// linapro-oidc-discord settings API.

package v1

import "github.com/gogf/gf/v2/frame/g"

// GetSettingsReq is the request for querying the linapro-oidc-discord settings.
type GetSettingsReq struct {
	g.Meta `path:"/settings" method:"get" tags:"Auth Login / Discord" summary:"Query Discord OIDC settings" dc:"Return the persisted linapro-oidc-discord settings for the admin settings page. The client secret is always returned as a masked indicator; plaintext secrets are never sent to the client." permission:"linapro-oidc-discord:settings:view"`
}

// GetSettingsRes is the response for querying the linapro-oidc-discord settings.
type GetSettingsRes struct {
	// Settings carries the projected settings row read from sys_config.
	Settings *SettingsItem `json:"settings" dc:"Persisted settings values with client secret masked" eg:"{}"`
}
