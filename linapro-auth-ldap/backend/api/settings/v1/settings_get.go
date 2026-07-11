package v1

import "github.com/gogf/gf/v2/frame/g"

type GetSettingsReq struct {
	g.Meta `path:"/settings" method:"get" tags:"LDAP Login" summary:"Query LDAP settings" dc:"Return masked LDAP directory settings." permission:"linapro-auth-ldap:settings:view"`
}

type GetSettingsRes struct {
	Settings *SettingsItem `json:"settings"`
}
