// Settings save path: persist Azure Blob settings with empty-secret keep semantics.
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
		nextSecret  = resolveNextSecret(current.AccountKey, in.AccountKey)
		accountName = strings.TrimSpace(in.AccountName)
		container   = strings.TrimSpace(in.Container)
		endpoint    = strings.TrimSpace(in.Endpoint)
		pathPrefix  = NormalizePathPrefix(in.PathPrefix)
	)

	items := []hostconfigcap.SetSysConfigValueItem{
		{Key: ConfigKeyAccountName, Value: accountName},
		{Key: ConfigKeyContainer, Value: container},
		{Key: ConfigKeyEndpoint, Value: endpoint},
		{Key: ConfigKeyPathPrefix, Value: pathPrefix},
	}
	if nextSecret != current.AccountKey {
		items = append(items, hostconfigcap.SetSysConfigValueItem{Key: ConfigKeyAccountKey, Value: nextSecret})
	}
	if err := s.setValues(ctx, items); err != nil {
		return nil, err
	}
	logger.Infof(ctx, "linapro-storage-azure settings saved accountSet=%t containerSet=%t secretSet=%t",
		accountName != "", container != "", nextSecret != "")
	return projectFromSnapshot(&Snapshot{
		AccountName: accountName,
		AccountKey:  nextSecret,
		Container:   container,
		Endpoint:    endpoint,
		PathPrefix:  pathPrefix,
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
