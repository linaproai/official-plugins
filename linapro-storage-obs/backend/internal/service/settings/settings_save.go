package settings

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/util/gconv"

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
	var (
		nextSecret  = resolveNextSecret(current.SecretAccessKey, in.SecretAccessKey)
		accessKeyID = strings.TrimSpace(in.AccessKeyID)
		region      = strings.TrimSpace(in.Region)
		bucket      = strings.TrimSpace(in.Bucket)
		endpoint    = strings.TrimSpace(in.Endpoint)
		pathPrefix  = NormalizePathPrefix(in.PathPrefix)
	)

	items := []hostconfigcap.SetSysConfigValueItem{
		{Key: ConfigKeyAccessKeyID, Value: accessKeyID},
		{Key: ConfigKeyRegion, Value: region},
		{Key: ConfigKeyBucket, Value: bucket},
		{Key: ConfigKeyEndpoint, Value: endpoint},
		{Key: ConfigKeyPathPrefix, Value: pathPrefix},
	}
	if nextSecret != current.SecretAccessKey {
		items = append(items, hostconfigcap.SetSysConfigValueItem{Key: ConfigKeySecretAccessKey, Value: nextSecret})
	}
	if err := s.setValues(ctx, items); err != nil {
		return nil, err
	}
	logger.Infof(ctx, "linapro-storage-obs settings saved accessKeySet=%t regionSet=%t bucketSet=%t secretSet=%t",
		accessKeyID != "", region != "", bucket != "", nextSecret != "")
	return projectFromSnapshot(&Snapshot{
		AccessKeyID:     accessKeyID,
		SecretAccessKey: nextSecret,
		Region:          region,
		Bucket:          bucket,
		Endpoint:        endpoint,
		PathPrefix:      pathPrefix,
	}), nil
}

func (s *serviceImpl) setValues(ctx context.Context, items []hostconfigcap.SetSysConfigValueItem) error {
	if err := s.sysConfigSvc.BatchSetValue(ctx, items, &hostconfigcap.SetSysConfigValueOptions{
		SystemManageable: gconv.PtrBool(false),
	}); err != nil {
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
