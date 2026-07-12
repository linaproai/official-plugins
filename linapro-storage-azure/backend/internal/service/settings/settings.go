// Package settings persists linapro-storage-azure admin settings through host sys_config.
package settings

import (
	"context"
	"strings"

	"lina-core/pkg/plugin/capability/hostconfigcap"
)

// SecretMask is returned instead of plaintext secrets.
const SecretMask = "************"

// Plugin-scoped sys_config keys.
const (
	ConfigKeyAccountName hostconfigcap.SysConfigKey = "plugin.linapro-storage-azure.account_name"
	ConfigKeyAccountKey  hostconfigcap.SysConfigKey = "plugin.linapro-storage-azure.account_key"
	ConfigKeyContainer   hostconfigcap.SysConfigKey = "plugin.linapro-storage-azure.container"
	ConfigKeyEndpoint    hostconfigcap.SysConfigKey = "plugin.linapro-storage-azure.endpoint"
	ConfigKeyPathPrefix  hostconfigcap.SysConfigKey = "plugin.linapro-storage-azure.path_prefix"
)

// Snapshot is the raw settings used by the storage provider factory.
type Snapshot struct {
	AccountName string
	AccountKey  string
	Container   string
	Endpoint    string
	PathPrefix  string
}

// Projection is the masked admin projection.
type Projection struct {
	AccountName          string
	AccountKeyMasked     string
	AccountKeyConfigured bool
	Container            string
	Endpoint             string
	PathPrefix           string
}

// SaveInput is the admin save payload.
type SaveInput struct {
	AccountName string
	AccountKey  string
	Container   string
	Endpoint    string
	PathPrefix  string
}

// Service is the settings surface.
type Service interface {
	// Get returns the masked settings projection for admin pages.
	Get(ctx context.Context) (*Projection, error)
	// Save persists settings and returns the masked projection.
	Save(ctx context.Context, in SaveInput) (*Projection, error)
	// Load returns the raw settings snapshot including secrets.
	Load(ctx context.Context) (*Snapshot, error)
	// ValidateReady reports whether the snapshot has required fields for serving.
	ValidateReady(snapshot *Snapshot) error
}

var _ Service = (*serviceImpl)(nil)

type serviceImpl struct {
	sysConfigSvc hostconfigcap.SysConfigService
}

// New creates a settings service.
func New(sysConfigSvc hostconfigcap.SysConfigService) Service {
	return &serviceImpl{sysConfigSvc: sysConfigSvc}
}

// NormalizePathPrefix trims slashes from optional prefixes.
func NormalizePathPrefix(raw string) string {
	return strings.Trim(strings.ReplaceAll(strings.TrimSpace(raw), "\\", "/"), "/")
}
