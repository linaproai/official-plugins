// Package settings persists linapro-storage-aws admin settings through host sys_config.
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
	ConfigKeyAccessKeyID     hostconfigcap.SysConfigKey = "plugin.linapro-storage-aws.access_key_id"
	ConfigKeySecretAccessKey hostconfigcap.SysConfigKey = "plugin.linapro-storage-aws.secret_access_key"
	ConfigKeyRegion          hostconfigcap.SysConfigKey = "plugin.linapro-storage-aws.region"
	ConfigKeyBucket          hostconfigcap.SysConfigKey = "plugin.linapro-storage-aws.bucket"
	ConfigKeyPathPrefix      hostconfigcap.SysConfigKey = "plugin.linapro-storage-aws.path_prefix"
)

// Snapshot is the raw settings used by the storage provider factory.
type Snapshot struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	Bucket          string
	PathPrefix      string
}

// Projection is the masked admin projection.
type Projection struct {
	AccessKeyID               string
	SecretAccessKeyMasked     string
	SecretAccessKeyConfigured bool
	Region                    string
	Bucket                    string
	PathPrefix                string
}

// SaveInput is the admin save payload.
type SaveInput struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	Bucket          string
	PathPrefix      string
}

// Service is the settings surface.
type Service interface {
	Get(ctx context.Context) (*Projection, error)
	Save(ctx context.Context, in SaveInput) (*Projection, error)
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
