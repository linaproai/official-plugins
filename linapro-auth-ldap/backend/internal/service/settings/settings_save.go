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

func (s *serviceImpl) Save(ctx context.Context, in SaveInput) (*Projection, error) {
	if s == nil || s.sysConfigSvc == nil {
		return nil, bizerr.WrapCode(gerror.New("settings: sys_config unavailable"), CodeStorageUnavailable)
	}
	tlsMode := NormalizeTLSMode(in.TLSMode)
	if tlsMode == "" {
		tlsMode = TLSModeLDAPS
	}
	host := strings.TrimSpace(in.Host)
	if err := ValidateTLSMode(host, tlsMode); err != nil {
		return nil, err
	}
	if host != "" && strings.TrimSpace(in.UserDNTemplate) == "" && strings.TrimSpace(in.BaseDN) == "" {
		return nil, bizerr.WrapCode(gerror.New("settings: base DN or user DN template required"), CodeConfigInvalid)
	}
	current, err := s.Load(ctx)
	if err != nil {
		return nil, err
	}
	nextSecret := resolveSecret(current.BindPassword, in.BindPassword)
	port := strings.TrimSpace(in.Port)
	if port == "" {
		if tlsMode == TLSModeLDAPS {
			port = "636"
		} else {
			port = "389"
		}
	}
	writes := []struct {
		key hostconfigcap.SysConfigKey
		val string
	}{
		{ConfigKeyConnectionKey, ConnectionKeyDefault},
		{ConfigKeyDisplayName, strings.TrimSpace(in.DisplayName)},
		{ConfigKeyHost, host},
		{ConfigKeyPort, port},
		{ConfigKeyTLSMode, tlsMode},
		{ConfigKeyBindDN, strings.TrimSpace(in.BindDN)},
		{ConfigKeyBaseDN, strings.TrimSpace(in.BaseDN)},
		{ConfigKeyUserFilter, strings.TrimSpace(in.UserFilter)},
		{ConfigKeyUserDNTemplate, strings.TrimSpace(in.UserDNTemplate)},
		{ConfigKeySubjectAttr, strings.TrimSpace(in.SubjectAttr)},
		{ConfigKeyEmailAttr, strings.TrimSpace(in.EmailAttr)},
		{ConfigKeyDisplayNameAttr, strings.TrimSpace(in.DisplayNameAttr)},
		{ConfigKeyAllowAutoProvision, boolFlag(in.AllowAutoProvision)},
	}
	for _, item := range writes {
		if err := s.set(ctx, item.key, item.val); err != nil {
			return nil, err
		}
	}
	if nextSecret != current.BindPassword {
		if err := s.set(ctx, ConfigKeyBindPassword, nextSecret); err != nil {
			return nil, err
		}
	}
	logger.Infof(ctx, "linapro-auth-ldap settings saved hostSet=%t tls=%s autoProvision=%t",
		host != "", tlsMode, in.AllowAutoProvision)
	return project(&Snapshot{
		ConnectionKey: ConnectionKeyDefault, DisplayName: strings.TrimSpace(in.DisplayName),
		Host: host, Port: port, TLSMode: tlsMode, BindDN: strings.TrimSpace(in.BindDN),
		BindPassword: nextSecret, BaseDN: strings.TrimSpace(in.BaseDN),
		UserFilter: strings.TrimSpace(in.UserFilter), UserDNTemplate: strings.TrimSpace(in.UserDNTemplate),
		SubjectAttr: strings.TrimSpace(in.SubjectAttr), EmailAttr: strings.TrimSpace(in.EmailAttr),
		DisplayNameAttr: strings.TrimSpace(in.DisplayNameAttr), AllowAutoProvision: in.AllowAutoProvision,
	}), nil
}

// ValidateTLSMode enforces plain-only-on-localhost.
func ValidateTLSMode(host, tlsMode string) error {
	mode := NormalizeTLSMode(tlsMode)
	if mode == "" {
		return bizerr.WrapCode(gerror.New("settings: invalid tls mode"), CodeTLSInvalid)
	}
	if mode == TLSModePlain && host != "" && !IsLocalHost(host) {
		return bizerr.WrapCode(gerror.New("settings: plain LDAP only for localhost"), CodeTLSInvalid)
	}
	return nil
}

func (s *serviceImpl) set(ctx context.Context, key hostconfigcap.SysConfigKey, value string) error {
	if err := s.sysConfigSvc.SetValue(ctx, key, value); err != nil {
		if bizerr.Is(err, capmodel.CodeCapabilityUnavailable) {
			return bizerr.WrapCode(err, CodeStorageUnavailable)
		}
		return bizerr.WrapCode(err, CodeSaveFailed)
	}
	return nil
}

func resolveSecret(current, submitted string) string {
	t := strings.TrimSpace(submitted)
	if t == "" || t == SecretMask {
		return current
	}
	return t
}

func boolFlag(v bool) string {
	if v {
		return "1"
	}
	return ""
}
