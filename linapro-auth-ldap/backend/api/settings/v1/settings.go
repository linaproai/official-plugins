package v1

// SettingsItem is the masked admin projection.
type SettingsItem struct {
	ConnectionKey          string `json:"connectionKey" dc:"Fixed connection key" eg:"default"`
	DisplayName            string `json:"displayName" dc:"Login entry display name" eg:"Directory Login"`
	Host                   string `json:"host" dc:"LDAP host" eg:"ldap.example.com"`
	Port                   string `json:"port" dc:"LDAP port" eg:"636"`
	TlsMode                string `json:"tlsMode" dc:"ldaps, starttls, or plain (localhost only)" eg:"ldaps"`
	BindDn                 string `json:"bindDn" dc:"Optional service bind DN" eg:"cn=search,dc=example,dc=com"`
	BindPasswordMasked     string `json:"bindPasswordMasked" dc:"Masked bind password" eg:"************"`
	BindPasswordConfigured bool   `json:"bindPasswordConfigured" dc:"Whether bind password is stored" eg:"true"`
	BaseDn                 string `json:"baseDn" dc:"Search base DN" eg:"ou=people,dc=example,dc=com"`
	UserFilter             string `json:"userFilter" dc:"Search filter with {username}" eg:"(uid={username})"`
	UserDnTemplate         string `json:"userDnTemplate" dc:"DN template with {username}" eg:"uid={username},ou=people,dc=example,dc=com"`
	SubjectAttr            string `json:"subjectAttr" dc:"Stable subject attribute" eg:"entryUUID"`
	EmailAttr              string `json:"emailAttr" dc:"Email attribute" eg:"mail"`
	DisplayNameAttr        string `json:"displayNameAttr" dc:"Display name attribute" eg:"cn"`
	AllowAutoProvision     bool   `json:"allowAutoProvision" dc:"JIT provision; default false" eg:"false"`
}
