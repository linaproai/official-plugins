// settings_read.go implements the read paths of the linapro-oidc-google
// settings service. Reads project stored client secrets into a masked
// indicator so plaintext material never reaches the HTTP boundary.

package settings

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/hostconfigcap"
)

// allSettingsKeys returns the ordered list of sys_config keys the plugin owns.
// The order is preserved on batch reads so the projection is deterministic.
func allSettingsKeys() []hostconfigcap.SysConfigKey {
	return []hostconfigcap.SysConfigKey{
		ConfigKeyClientID,
		ConfigKeyClientSecret,
		ConfigKeyRedirectURL,
		ConfigKeyEnableBackendRedirect,
		ConfigKeyDefaultBackendRedirect,
		ConfigKeyBackendRedirects,
	}
}

// Get returns the masked settings projection consumed by the admin settings page.
func (s *serviceImpl) Get(ctx context.Context) (*Projection, error) {
	snapshot, err := s.Load(ctx)
	if err != nil {
		return nil, err
	}
	return projectFromSnapshot(snapshot), nil
}

// Load returns the raw settings snapshot consumed by the OAuth login flow.
// Missing sys_config rows are reported through the host BatchGet MissingIDs
// contract and treated as empty values so the OAuth service can layer its
// own defaults on top.
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
	snapshot := &Snapshot{}
	if result != nil {
		if info := result.Items[ConfigKeyClientID]; info != nil {
			snapshot.ClientID = info.Value
		}
		if info := result.Items[ConfigKeyClientSecret]; info != nil {
			snapshot.ClientSecret = info.Value
		}
		if info := result.Items[ConfigKeyRedirectURL]; info != nil {
			snapshot.RedirectURL = info.Value
		}
		if info := result.Items[ConfigKeyEnableBackendRedirect]; info != nil {
			snapshot.EnableBackendRedirect = info.Value == enabledFlagValue
		}
		if info := result.Items[ConfigKeyDefaultBackendRedirect]; info != nil {
			snapshot.DefaultBackendRedirect = info.Value
		}
		if info := result.Items[ConfigKeyBackendRedirects]; info != nil {
			snapshot.BackendRedirects = info.Value
		}
	}
	return snapshot, nil
}

// projectFromSnapshot masks the client secret before returning the projection.
func projectFromSnapshot(snapshot *Snapshot) *Projection {
	projection := &Projection{}
	if snapshot == nil {
		return projection
	}
	projection.ClientID = snapshot.ClientID
	projection.RedirectURL = snapshot.RedirectURL
	projection.EnableBackendRedirect = snapshot.EnableBackendRedirect
	projection.DefaultBackendRedirect = snapshot.DefaultBackendRedirect
	projection.BackendRedirects = snapshot.BackendRedirects
	if snapshot.ClientSecret != "" {
		projection.ClientSecretConfigured = true
		projection.ClientSecretMasked = SecretMask
	}
	return projection
}
