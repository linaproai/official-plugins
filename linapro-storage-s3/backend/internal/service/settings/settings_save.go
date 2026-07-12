package settings

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/hostconfigcap"
)

// Save persists settings and returns the masked projection.
func (s *serviceImpl) Save(ctx context.Context, in SaveInput) (*Projection, error) {
	if s == nil || s.sysConfigSvc == nil {
		return nil, bizerr.WrapCode(gerror.New("settings: host sys_config service is unavailable"), CodeStorageUnavailable)
	}
	current, err := s.Load(ctx)
	if err != nil {
		return nil, err
	}
	nextSecret := resolveNextSecret(current.SecretAccessKey, in.SecretAccessKey)
	accessKeyID := strings.TrimSpace(in.AccessKeyID)
	region := strings.TrimSpace(in.Region)
	bucket := strings.TrimSpace(in.Bucket)
	endpoint := strings.TrimSpace(in.Endpoint)
	pathPrefix := NormalizePathPrefix(in.PathPrefix)

	if err := s.setValue(ctx, ConfigKeyAccessKeyID, accessKeyID); err != nil {
		return nil, err
	}
	if err := s.setValue(ctx, ConfigKeyRegion, region); err != nil {
		return nil, err
	}
	if err := s.setValue(ctx, ConfigKeyBucket, bucket); err != nil {
		return nil, err
	}
	if err := s.setValue(ctx, ConfigKeyEndpoint, endpoint); err != nil {
		return nil, err
	}
	if err := s.setValue(ctx, ConfigKeyPathPrefix, pathPrefix); err != nil {
		return nil, err
	}
	forcePathStyle := ""
	if in.ForcePathStyle {
		forcePathStyle = "1"
	}
	if err := s.setValue(ctx, ConfigKeyForcePathStyle, forcePathStyle); err != nil {
		return nil, err
	}
	if nextSecret != current.SecretAccessKey {
		if err := s.setValue(ctx, ConfigKeySecretAccessKey, nextSecret); err != nil {
			return nil, err
		}
	}
	logger.Infof(ctx, "linapro-storage-s3 settings saved accessKeySet=%t regionSet=%t bucketSet=%t secretSet=%t",
		accessKeyID != "", region != "", bucket != "", nextSecret != "")
	return projectFromSnapshot(&Snapshot{
		AccessKeyID:     accessKeyID,
		SecretAccessKey: nextSecret,
		Region:          region,
		Bucket:          bucket,
		Endpoint:        endpoint,
		PathPrefix:      pathPrefix,
		ForcePathStyle:  in.ForcePathStyle,
	}), nil
}

func (s *serviceImpl) setValue(ctx context.Context, key hostconfigcap.SysConfigKey, value string) error {
	if err := s.sysConfigSvc.SetValue(ctx, key, value); err != nil {
		if bizerr.Is(err, capmodel.CodeCapabilityUnavailable) {
			return bizerr.WrapCode(err, CodeStorageUnavailable)
		}
		return bizerr.WrapCode(err, CodeSaveFailed)
	}
	return nil
}

func resolveNextSecret(current string, submitted string) string {
	trimmed := strings.TrimSpace(submitted)
	if trimmed == "" || trimmed == SecretMask {
		return current
	}
	return trimmed
}
