// Package settings implements linapro-auth-ldap admin settings via host sys_config.
package settings

import (
	"context"
	"strings"

	"lina-core/pkg/plugin/capability/hostconfigcap"
)

const ConnectionKeyDefault = "default"
const SecretMask = "************"

// TLS modes for directory connections.
const (
	TLSModeLDAPS    = "ldaps"
	TLSModeStartTLS = "starttls"
	TLSModePlain    = "plain"
)

const (
	ConfigKeyConnectionKey      hostconfigcap.SysConfigKey = "plugin.linapro-auth-ldap.connection_key"
	ConfigKeyDisplayName        hostconfigcap.SysConfigKey = "plugin.linapro-auth-ldap.display_name"
	ConfigKeyHost               hostconfigcap.SysConfigKey = "plugin.linapro-auth-ldap.host"
	ConfigKeyPort               hostconfigcap.SysConfigKey = "plugin.linapro-auth-ldap.port"
	ConfigKeyTLSMode            hostconfigcap.SysConfigKey = "plugin.linapro-auth-ldap.tls_mode"
	ConfigKeyBindDN             hostconfigcap.SysConfigKey = "plugin.linapro-auth-ldap.bind_dn"
	ConfigKeyBindPassword       hostconfigcap.SysConfigKey = "plugin.linapro-auth-ldap.bind_password"
	ConfigKeyBaseDN             hostconfigcap.SysConfigKey = "plugin.linapro-auth-ldap.base_dn"
	ConfigKeyUserFilter         hostconfigcap.SysConfigKey = "plugin.linapro-auth-ldap.user_filter"
	ConfigKeyUserDNTemplate     hostconfigcap.SysConfigKey = "plugin.linapro-auth-ldap.user_dn_template"
	ConfigKeySubjectAttr        hostconfigcap.SysConfigKey = "plugin.linapro-auth-ldap.subject_attr"
	ConfigKeyEmailAttr          hostconfigcap.SysConfigKey = "plugin.linapro-auth-ldap.email_attr"
	ConfigKeyDisplayNameAttr    hostconfigcap.SysConfigKey = "plugin.linapro-auth-ldap.display_name_attr"
	ConfigKeyAllowAutoProvision hostconfigcap.SysConfigKey = "plugin.linapro-auth-ldap.allow_auto_provision"
)

// Snapshot is the raw settings for login orchestration.
type Snapshot struct {
	ConnectionKey      string
	DisplayName        string
	Host               string
	Port               string
	TLSMode            string
	BindDN             string
	BindPassword       string
	BaseDN             string
	UserFilter         string
	UserDNTemplate     string
	SubjectAttr        string
	EmailAttr          string
	DisplayNameAttr    string
	AllowAutoProvision bool
}

// Projection is the masked admin projection.
type Projection struct {
	ConnectionKey          string
	DisplayName            string
	Host                   string
	Port                   string
	TLSMode                string
	BindDN                 string
	BindPasswordMasked     string
	BindPasswordConfigured bool
	BaseDN                 string
	UserFilter             string
	UserDNTemplate         string
	SubjectAttr            string
	EmailAttr              string
	DisplayNameAttr        string
	AllowAutoProvision     bool
}

// SaveInput is the admin save payload.
type SaveInput struct {
	DisplayName        string
	Host               string
	Port               string
	TLSMode            string
	BindDN             string
	BindPassword       string
	BaseDN             string
	UserFilter         string
	UserDNTemplate     string
	SubjectAttr        string
	EmailAttr          string
	DisplayNameAttr    string
	AllowAutoProvision bool
}

// Service is the settings surface.
type Service interface {
	Get(ctx context.Context) (*Projection, error)
	Save(ctx context.Context, in SaveInput) (*Projection, error)
	Load(ctx context.Context) (*Snapshot, error)
}

var _ Service = (*serviceImpl)(nil)

type serviceImpl struct {
	sysConfigSvc hostconfigcap.SysConfigService
}

// New creates a settings service.
func New(sysConfigSvc hostconfigcap.SysConfigService) Service {
	return &serviceImpl{sysConfigSvc: sysConfigSvc}
}

// NormalizeTLSMode returns a supported TLS mode or empty when unknown.
func NormalizeTLSMode(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case TLSModeLDAPS, "ssl", "tls":
		return TLSModeLDAPS
	case TLSModeStartTLS, "start_tls":
		return TLSModeStartTLS
	case TLSModePlain, "none", "off":
		return TLSModePlain
	default:
		return ""
	}
}

// IsLocalHost reports whether host is loopback.
func IsLocalHost(host string) bool {
	h := strings.ToLower(strings.TrimSpace(host))
	return h == "localhost" || h == "127.0.0.1" || h == "::1" || h == "[::1]"
}
