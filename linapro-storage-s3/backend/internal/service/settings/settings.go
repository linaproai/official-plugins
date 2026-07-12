// Package settings persists linapro-storage-s3 admin settings through host sys_config.
package settings

import (
	"context"
	"strings"

	"lina-core/pkg/plugin/capability/hostconfigcap"
)

// SecretMask is returned instead of plaintext secrets.
const SecretMask = "************"

// DefaultSigningRegion is used for SigV4 when the admin leaves region empty.
// Self-hosted S3 endpoints (MinIO, R2, etc.) typically ignore the physical region.
const DefaultSigningRegion = "us-east-1"

// Plugin-scoped sys_config keys.
const (
	ConfigKeyAccessKeyID     hostconfigcap.SysConfigKey = "plugin.linapro-storage-s3.access_key_id"
	ConfigKeySecretAccessKey hostconfigcap.SysConfigKey = "plugin.linapro-storage-s3.secret_access_key"
	ConfigKeyRegion          hostconfigcap.SysConfigKey = "plugin.linapro-storage-s3.region"
	ConfigKeyBucket          hostconfigcap.SysConfigKey = "plugin.linapro-storage-s3.bucket"
	ConfigKeyEndpoint        hostconfigcap.SysConfigKey = "plugin.linapro-storage-s3.endpoint"
	ConfigKeyPathPrefix      hostconfigcap.SysConfigKey = "plugin.linapro-storage-s3.path_prefix"
	ConfigKeyForcePathStyle  hostconfigcap.SysConfigKey = "plugin.linapro-storage-s3.force_path_style"
)

// Snapshot is the raw settings used by the storage provider factory.
type Snapshot struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	Bucket          string
	Endpoint        string
	PathPrefix      string
	ForcePathStyle  bool
}

// Projection is the masked admin projection.
type Projection struct {
	AccessKeyID               string
	SecretAccessKeyMasked     string
	SecretAccessKeyConfigured bool
	Region                    string
	Bucket                    string
	Endpoint                  string
	PathPrefix                string
	ForcePathStyle            bool
}

// SaveInput is the admin save payload.
type SaveInput struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	Bucket          string
	Endpoint        string
	PathPrefix      string
	ForcePathStyle  bool
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

// EffectiveRegion returns a non-empty region for the AWS SDK signing scope.
func EffectiveRegion(region string) string {
	if trimmed := strings.TrimSpace(region); trimmed != "" {
		return trimmed
	}
	return DefaultSigningRegion
}
