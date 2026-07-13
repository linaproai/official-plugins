package v1

import "github.com/gogf/gf/v2/frame/g"

// GetSettingsReq loads the single platform mail account settings projection.
type GetSettingsReq struct {
	g.Meta `path:"/mail/settings" method:"get" tags:"Mail Settings" summary:"Get platform mail settings" dc:"Return the single platform mail account settings with masked secrets for the admin page." permission:"linapro-mail-core:settings:view"`
}

// GetSettingsRes carries the masked settings projection.
type GetSettingsRes struct {
	g.Meta   `mime:"application/json"`
	Settings *SettingsItem `json:"settings" dc:"Masked platform mail settings"`
}
