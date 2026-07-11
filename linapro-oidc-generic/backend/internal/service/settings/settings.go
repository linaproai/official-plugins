// Package settings implements the linapro-oidc-generic admin settings service.
// It persists OIDC client credentials and issuer configuration through the
// host sys_config seam, and exposes a masked projection for the admin page
// plus a raw Snapshot used by the OAuth login orchestration at request time.
package settings

import (
	"context"

	"lina-core/pkg/plugin/capability/hostconfigcap"
)

// ConnectionKeyDefault is the v1 fixed connection key used in provider
// encoding (oidc:default). Multi-connection management is out of scope.
const ConnectionKeyDefault = "default"

// Plugin-scoped sys_config keys owned by linapro-oidc-generic.
const (
	ConfigKeyConnectionKey          hostconfigcap.SysConfigKey = "plugin.linapro-oidc-generic.connection_key"
	ConfigKeyDisplayName            hostconfigcap.SysConfigKey = "plugin.linapro-oidc-generic.display_name"
	ConfigKeyIssuer                 hostconfigcap.SysConfigKey = "plugin.linapro-oidc-generic.issuer"
	ConfigKeyClientID               hostconfigcap.SysConfigKey = "plugin.linapro-oidc-generic.client_id"
	ConfigKeyClientSecret           hostconfigcap.SysConfigKey = "plugin.linapro-oidc-generic.client_secret"
	ConfigKeyRedirectURL            hostconfigcap.SysConfigKey = "plugin.linapro-oidc-generic.redirect_url"
	ConfigKeyScopes                 hostconfigcap.SysConfigKey = "plugin.linapro-oidc-generic.scopes"
	ConfigKeyAllowAutoProvision     hostconfigcap.SysConfigKey = "plugin.linapro-oidc-generic.allow_auto_provision"
	ConfigKeyDefaultBackendRedirect hostconfigcap.SysConfigKey = "plugin.linapro-oidc-generic.default_backend_redirect"
)

// SecretMask is the fixed indicator returned whenever a client secret is stored.
const SecretMask = "************"

// Snapshot is the raw settings snapshot for OAuth orchestration.
type Snapshot struct {
	ConnectionKey          string
	DisplayName            string
	Issuer                 string
	ClientID               string
	ClientSecret           string
	RedirectURL            string
	Scopes                 string
	AllowAutoProvision     bool
	DefaultBackendRedirect string
}

// Projection is the masked settings projection for the admin settings page.
type Projection struct {
	ConnectionKey          string
	DisplayName            string
	Issuer                 string
	ClientID               string
	ClientSecretMasked     string
	ClientSecretConfigured bool
	RedirectURL            string
	Scopes                 string
	AllowAutoProvision     bool
	DefaultBackendRedirect string
}

// SaveInput carries admin-submitted settings values.
type SaveInput struct {
	DisplayName            string
	Issuer                 string
	ClientID               string
	ClientSecret           string
	RedirectURL            string
	Scopes                 string
	AllowAutoProvision     bool
	DefaultBackendRedirect string
}

// Service defines the settings service surface.
type Service interface {
	Get(ctx context.Context) (*Projection, error)
	Save(ctx context.Context, in SaveInput) (*Projection, error)
	Load(ctx context.Context) (*Snapshot, error)
}

var _ Service = (*serviceImpl)(nil)

type serviceImpl struct {
	sysConfigSvc hostconfigcap.SysConfigService
}

// New creates a settings service bound to the host sys_config seam.
func New(sysConfigSvc hostconfigcap.SysConfigService) Service {
	return &serviceImpl{sysConfigSvc: sysConfigSvc}
}
