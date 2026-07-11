package v1

import "github.com/gogf/gf/v2/frame/g"

type SaveSettingsReq struct {
	g.Meta             `path:"/settings" method:"put" tags:"LDAP Login" summary:"Save LDAP settings" dc:"Persist LDAP directory settings. Empty/masked bind password keeps previous. Auto-provision defaults off." permission:"linapro-auth-ldap:settings:update"`
	DisplayName        string `json:"displayName" v:"max-length:128"`
	Host               string `json:"host" v:"max-length:256"`
	Port               string `json:"port" v:"max-length:16"`
	TlsMode            string `json:"tlsMode" v:"max-length:16"`
	BindDn             string `json:"bindDn" v:"max-length:512"`
	BindPassword       string `json:"bindPassword" v:"max-length:512"`
	BaseDn             string `json:"baseDn" v:"max-length:512"`
	UserFilter         string `json:"userFilter" v:"max-length:512"`
	UserDnTemplate     string `json:"userDnTemplate" v:"max-length:512"`
	SubjectAttr        string `json:"subjectAttr" v:"max-length:128"`
	EmailAttr          string `json:"emailAttr" v:"max-length:128"`
	DisplayNameAttr    string `json:"displayNameAttr" v:"max-length:128"`
	AllowAutoProvision bool   `json:"allowAutoProvision"`
}

type SaveSettingsRes struct {
	Settings *SettingsItem `json:"settings"`
}
