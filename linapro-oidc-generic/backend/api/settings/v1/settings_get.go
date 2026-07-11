// settings_get.go declares get-settings request/response DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// GetSettingsReq is the request for querying generic OIDC settings.
type GetSettingsReq struct {
	g.Meta `path:"/settings" method:"get" tags:"Auth Login / OIDC" summary:"Query Generic OIDC settings" dc:"Return persisted linapro-oidc-generic settings. Client secret is always masked." permission:"linapro-oidc-generic:settings:view"`
}

// GetSettingsRes is the response for querying generic OIDC settings.
type GetSettingsRes struct {
	Settings *SettingsItem `json:"settings" dc:"Persisted settings with client secret masked"`
}
