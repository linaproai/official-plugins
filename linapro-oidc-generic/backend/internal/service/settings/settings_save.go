// settings_save.go implements the write path of the generic OIDC settings service.

package settings

import (
	"context"
	"net/url"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/hostconfigcap"
)

const enabledFlagValue = "1"

// Save persists settings; empty or masked secret keeps the previous secret.
// AllowAutoProvision defaults to false when not enabled in the request.
func (s *serviceImpl) Save(ctx context.Context, in SaveInput) (*Projection, error) {
	if s == nil || s.sysConfigSvc == nil {
		return nil, bizerr.WrapCode(
			gerror.New("settings: host sys_config service is unavailable"),
			CodeStorageUnavailable,
		)
	}
	issuer := strings.TrimSpace(in.Issuer)
	if issuer != "" {
		if err := validateIssuer(issuer); err != nil {
			return nil, err
		}
	}
	current, err := s.Load(ctx)
	if err != nil {
		return nil, err
	}
	nextSecret := resolveNextSecret(current.ClientSecret, in.ClientSecret)
	if err := s.setValue(ctx, ConfigKeyConnectionKey, ConnectionKeyDefault); err != nil {
		return nil, err
	}
	if err := s.setValue(ctx, ConfigKeyDisplayName, strings.TrimSpace(in.DisplayName)); err != nil {
		return nil, err
	}
	if err := s.setValue(ctx, ConfigKeyIssuer, issuer); err != nil {
		return nil, err
	}
	if err := s.setValue(ctx, ConfigKeyClientID, strings.TrimSpace(in.ClientID)); err != nil {
		return nil, err
	}
	if err := s.setValue(ctx, ConfigKeyRedirectURL, strings.TrimSpace(in.RedirectURL)); err != nil {
		return nil, err
	}
	if err := s.setValue(ctx, ConfigKeyScopes, strings.TrimSpace(in.Scopes)); err != nil {
		return nil, err
	}
	if err := s.setValue(ctx, ConfigKeyDefaultBackendRedirect, strings.TrimSpace(in.DefaultBackendRedirect)); err != nil {
		return nil, err
	}
	autoProvisionFlag := ""
	if in.AllowAutoProvision {
		autoProvisionFlag = enabledFlagValue
	}
	if err := s.setValue(ctx, ConfigKeyAllowAutoProvision, autoProvisionFlag); err != nil {
		return nil, err
	}
	if nextSecret != current.ClientSecret {
		if err := s.setValue(ctx, ConfigKeyClientSecret, nextSecret); err != nil {
			return nil, err
		}
	}
	logger.Infof(ctx, "linapro-oidc-generic settings saved issuerSet=%t clientIdSet=%t secretSet=%t autoProvision=%t",
		issuer != "",
		strings.TrimSpace(in.ClientID) != "",
		nextSecret != "",
		in.AllowAutoProvision,
	)
	return projectFromSnapshot(&Snapshot{
		ConnectionKey:          ConnectionKeyDefault,
		DisplayName:            strings.TrimSpace(in.DisplayName),
		Issuer:                 issuer,
		ClientID:               strings.TrimSpace(in.ClientID),
		ClientSecret:           nextSecret,
		RedirectURL:            strings.TrimSpace(in.RedirectURL),
		Scopes:                 strings.TrimSpace(in.Scopes),
		AllowAutoProvision:     in.AllowAutoProvision,
		DefaultBackendRedirect: strings.TrimSpace(in.DefaultBackendRedirect),
	}), nil
}

func validateIssuer(issuer string) error {
	parsed, err := url.Parse(issuer)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return bizerr.WrapCode(gerror.New("settings: issuer is not a valid absolute URL"), CodeIssuerInvalid)
	}
	host := strings.ToLower(parsed.Hostname())
	if parsed.Scheme == "https" {
		return nil
	}
	if parsed.Scheme == "http" && (host == "localhost" || host == "127.0.0.1" || host == "::1") {
		return nil
	}
	return bizerr.WrapCode(gerror.New("settings: issuer must use https"), CodeIssuerInvalid)
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
