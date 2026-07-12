package settings

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/hostconfigcap"
)

func allSettingsKeys() []hostconfigcap.SysConfigKey {
	return []hostconfigcap.SysConfigKey{
		ConfigKeyAccessKeyID,
		ConfigKeySecretAccessKey,
		ConfigKeyRegion,
		ConfigKeyBucket,
		ConfigKeyEndpoint,
		ConfigKeyPathPrefix,
		ConfigKeyForcePathStyle,
	}
}

// Get returns the masked settings projection.
func (s *serviceImpl) Get(ctx context.Context) (*Projection, error) {
	snapshot, err := s.Load(ctx)
	if err != nil {
		return nil, err
	}
	return projectFromSnapshot(snapshot), nil
}

// Load returns the raw settings snapshot.
func (s *serviceImpl) Load(ctx context.Context) (*Snapshot, error) {
	if s == nil || s.sysConfigSvc == nil {
		return nil, bizerr.WrapCode(gerror.New("settings: host sys_config service is unavailable"), CodeStorageUnavailable)
	}
	result, err := s.sysConfigSvc.BatchGet(ctx, allSettingsKeys())
	if err != nil {
		if bizerr.Is(err, capmodel.CodeCapabilityUnavailable) {
			return nil, bizerr.WrapCode(err, CodeStorageUnavailable)
		}
		return nil, bizerr.WrapCode(err, CodeReadFailed)
	}
	snapshot := &Snapshot{}
	if result != nil {
		if info := result.Items[ConfigKeyAccessKeyID]; info != nil {
			snapshot.AccessKeyID = info.Value
		}
		if info := result.Items[ConfigKeySecretAccessKey]; info != nil {
			snapshot.SecretAccessKey = info.Value
		}
		if info := result.Items[ConfigKeyRegion]; info != nil {
			snapshot.Region = info.Value
		}
		if info := result.Items[ConfigKeyBucket]; info != nil {
			snapshot.Bucket = info.Value
		}
		if info := result.Items[ConfigKeyEndpoint]; info != nil {
			snapshot.Endpoint = info.Value
		}
		if info := result.Items[ConfigKeyPathPrefix]; info != nil {
			snapshot.PathPrefix = info.Value
		}
		if info := result.Items[ConfigKeyForcePathStyle]; info != nil {
			snapshot.ForcePathStyle = info.Value == "1"
		}
	}
	return snapshot, nil
}

func projectFromSnapshot(snapshot *Snapshot) *Projection {
	projection := &Projection{}
	if snapshot == nil {
		return projection
	}
	projection.AccessKeyID = snapshot.AccessKeyID
	projection.Region = snapshot.Region
	projection.Bucket = snapshot.Bucket
	projection.Endpoint = snapshot.Endpoint
	projection.PathPrefix = snapshot.PathPrefix
	projection.ForcePathStyle = snapshot.ForcePathStyle
	if snapshot.SecretAccessKey != "" {
		projection.SecretAccessKeyConfigured = true
		projection.SecretAccessKeyMasked = SecretMask
	}
	return projection
}

// ValidateReady ensures required S3 protocol fields are present.
// Endpoint and bucket are required; region is optional (defaults at client build).
func (s *serviceImpl) ValidateReady(snapshot *Snapshot) error {
	if snapshot == nil {
		return bizerr.NewCode(CodeConfigInvalid)
	}
	if strings.TrimSpace(snapshot.AccessKeyID) == "" ||
		strings.TrimSpace(snapshot.SecretAccessKey) == "" ||
		strings.TrimSpace(snapshot.Endpoint) == "" ||
		strings.TrimSpace(snapshot.Bucket) == "" {
		return bizerr.NewCode(CodeConfigInvalid)
	}
	return nil
}
