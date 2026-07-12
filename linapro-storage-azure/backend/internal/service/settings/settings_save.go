// Settings save path: persist Azure Blob settings with empty-secret keep semantics.
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
	nextSecret := resolveNextSecret(current.AccountKey, in.AccountKey)
	accountName := strings.TrimSpace(in.AccountName)
	container := strings.TrimSpace(in.Container)
	endpoint := strings.TrimSpace(in.Endpoint)
	pathPrefix := NormalizePathPrefix(in.PathPrefix)

	if err := s.setValue(ctx, ConfigKeyAccountName, accountName); err != nil {
		return nil, err
	}
	if err := s.setValue(ctx, ConfigKeyContainer, container); err != nil {
		return nil, err
	}
	if err := s.setValue(ctx, ConfigKeyEndpoint, endpoint); err != nil {
		return nil, err
	}
	if err := s.setValue(ctx, ConfigKeyPathPrefix, pathPrefix); err != nil {
		return nil, err
	}
	if nextSecret != current.AccountKey {
		if err := s.setValue(ctx, ConfigKeyAccountKey, nextSecret); err != nil {
			return nil, err
		}
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
