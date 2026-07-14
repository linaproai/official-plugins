// settings_save.go implements the write path of the linapro-oidc-google
// settings service. An empty or masked client secret keeps the previously
// stored value so admins can freely edit other fields without re-typing the
// secret.

package settings

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/hostconfigcap"
)

// enabledFlagValue is the persisted representation of an enabled boolean flag.
const enabledFlagValue = "1"

// Save persists the caller-supplied settings values through the host
// sys_config seam and returns the fresh masked projection so the caller can
// render the updated form state without a second read.
func (s *serviceImpl) Save(ctx context.Context, in SaveInput) (*Projection, error) {
	if s == nil || s.sysConfigSvc == nil {
		return nil, bizerr.WrapCode(
			gerror.New("settings: host sys_config service is unavailable"),
			CodeStorageUnavailable,
		)
	}
	rules := strings.TrimSpace(in.BackendRedirects)
	if err := validateBackendRedirects(rules); err != nil {
		return nil, err
	}
	current, err := s.Load(ctx)
	if err != nil {
		return nil, err
	}
	nextSecret := resolveNextSecret(current.ClientSecret, in.ClientSecret)
	enableFlag := ""
	if in.EnableBackendRedirect {
		enableFlag = enabledFlagValue
	}
	autoProvisionFlag := ""
	if in.AllowAutoProvision {
		autoProvisionFlag = enabledFlagValue
	}
	oneTapFlag := ""
	if in.EnableOneTap {
		oneTapFlag = enabledFlagValue
	}
	items := []hostconfigcap.SetSysConfigValueItem{
		{Key: ConfigKeyClientID, Value: strings.TrimSpace(in.ClientID)},
		{Key: ConfigKeyRedirectURL, Value: strings.TrimSpace(in.RedirectURL)},
		{Key: ConfigKeyEnableBackendRedirect, Value: enableFlag},
		{Key: ConfigKeyDefaultBackendRedirect, Value: strings.TrimSpace(in.DefaultBackendRedirect)},
		{Key: ConfigKeyBackendRedirects, Value: rules},
		{Key: ConfigKeyAllowAutoProvision, Value: autoProvisionFlag},
		{Key: ConfigKeyEnableOneTap, Value: oneTapFlag},
	}
	if nextSecret != current.ClientSecret {
		items = append(items, hostconfigcap.SetSysConfigValueItem{Key: ConfigKeyClientSecret, Value: nextSecret})
	}
	if err := s.setValues(ctx, items); err != nil {
		return nil, err
	}
	logger.Infof(ctx, "linapro-oidc-google settings saved clientIdSet=%t redirectUrlSet=%t secretSet=%t ssoEnabled=%t ruleSet=%t",
		strings.TrimSpace(in.ClientID) != "",
		strings.TrimSpace(in.RedirectURL) != "",
		nextSecret != "",
		in.EnableBackendRedirect,
		rules != "",
	)
	return projectFromSnapshot(&Snapshot{
		ClientID:               strings.TrimSpace(in.ClientID),
		ClientSecret:           nextSecret,
		RedirectURL:            strings.TrimSpace(in.RedirectURL),
		EnableBackendRedirect:  in.EnableBackendRedirect,
		DefaultBackendRedirect: strings.TrimSpace(in.DefaultBackendRedirect),
		BackendRedirects:       rules,
		AllowAutoProvision:     in.AllowAutoProvision,
		EnableOneTap:           in.EnableOneTap,
	}), nil
}

// validateBackendRedirects rejects rule payloads that are not a JSON object of
// string receiver URLs, so a malformed dictionary is surfaced at save time
// instead of silently breaking SSO delivery at login time.
func validateBackendRedirects(rules string) error {
	if rules == "" {
		return nil
	}
	var parsed map[string]string
	if err := json.Unmarshal([]byte(rules), &parsed); err != nil {
		return bizerr.WrapCode(err, CodeRulesInvalid)
	}
	return nil
}

// setValues writes plugin-scoped sys_config values in one batch and wraps host
// errors into the plugin's stable business error surface.
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

// resolveNextSecret honors the "empty or masked secret keeps existing" rule
// so admins can freely edit other fields without re-typing the client secret.
func resolveNextSecret(current string, submitted string) string {
	trimmed := strings.TrimSpace(submitted)
	if trimmed == "" || trimmed == SecretMask {
		return current
	}
	return trimmed
}
