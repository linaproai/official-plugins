// settings_read.go implements read paths of the generic OIDC settings service.

package settings

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/hostconfigcap"
)

func allSettingsKeys() []hostconfigcap.SysConfigKey {
	return []hostconfigcap.SysConfigKey{
		ConfigKeyConnectionKey,
		ConfigKeyDisplayName,
		ConfigKeyIssuer,
		ConfigKeyClientID,
		ConfigKeyClientSecret,
		ConfigKeyRedirectURL,
		ConfigKeyScopes,
		ConfigKeyAllowAutoProvision,
		ConfigKeyDefaultBackendRedirect,
	}
}

// Get returns the masked settings projection for the admin page.
func (s *serviceImpl) Get(ctx context.Context) (*Projection, error) {
	snapshot, err := s.Load(ctx)
	if err != nil {
		return nil, err
	}
	return projectFromSnapshot(snapshot), nil
}

// Load returns the raw settings snapshot for OAuth orchestration.
func (s *serviceImpl) Load(ctx context.Context) (*Snapshot, error) {
	if s == nil || s.sysConfigSvc == nil {
		return nil, bizerr.WrapCode(
			gerror.New("settings: host sys_config service is unavailable"),
			CodeStorageUnavailable,
		)
	}
	result, err := s.sysConfigSvc.BatchGet(ctx, allSettingsKeys())
	if err != nil {
		if bizerr.Is(err, capmodel.CodeCapabilityUnavailable) {
			return nil, bizerr.WrapCode(err, CodeStorageUnavailable)
		}
		return nil, bizerr.WrapCode(err, CodeReadFailed)
	}
	snapshot := &Snapshot{ConnectionKey: ConnectionKeyDefault}
	if result != nil {
		if info := result.Items[ConfigKeyConnectionKey]; info != nil && info.Value != "" {
			snapshot.ConnectionKey = info.Value
		}
		if info := result.Items[ConfigKeyDisplayName]; info != nil {
			snapshot.DisplayName = info.Value
		}
		if info := result.Items[ConfigKeyIssuer]; info != nil {
			snapshot.Issuer = info.Value
		}
		if info := result.Items[ConfigKeyClientID]; info != nil {
			snapshot.ClientID = info.Value
		}
		if info := result.Items[ConfigKeyClientSecret]; info != nil {
			snapshot.ClientSecret = info.Value
		}
		if info := result.Items[ConfigKeyRedirectURL]; info != nil {
			snapshot.RedirectURL = info.Value
		}
		if info := result.Items[ConfigKeyScopes]; info != nil {
			snapshot.Scopes = info.Value
		}
		if info := result.Items[ConfigKeyAllowAutoProvision]; info != nil {
			snapshot.AllowAutoProvision = info.Value == enabledFlagValue
		}
		if info := result.Items[ConfigKeyDefaultBackendRedirect]; info != nil {
			snapshot.DefaultBackendRedirect = info.Value
		}
	}
	if snapshot.ConnectionKey == "" {
		snapshot.ConnectionKey = ConnectionKeyDefault
	}
	return snapshot, nil
}

func projectFromSnapshot(snapshot *Snapshot) *Projection {
	projection := &Projection{ConnectionKey: ConnectionKeyDefault}
	if snapshot == nil {
		return projection
	}
	projection.ConnectionKey = snapshot.ConnectionKey
	if projection.ConnectionKey == "" {
		projection.ConnectionKey = ConnectionKeyDefault
	}
	projection.DisplayName = snapshot.DisplayName
	projection.Issuer = snapshot.Issuer
	projection.ClientID = snapshot.ClientID
	projection.RedirectURL = snapshot.RedirectURL
	projection.Scopes = snapshot.Scopes
	projection.AllowAutoProvision = snapshot.AllowAutoProvision
	projection.DefaultBackendRedirect = snapshot.DefaultBackendRedirect
	if snapshot.ClientSecret != "" {
		projection.ClientSecretConfigured = true
		projection.ClientSecretMasked = SecretMask
	}
	return projection
}
