package settings

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/hostconfigcap"
)

func allKeys() []hostconfigcap.SysConfigKey {
	return []hostconfigcap.SysConfigKey{
		ConfigKeyConnectionKey, ConfigKeyDisplayName, ConfigKeyHost, ConfigKeyPort, ConfigKeyTLSMode,
		ConfigKeyBindDN, ConfigKeyBindPassword, ConfigKeyBaseDN, ConfigKeyUserFilter, ConfigKeyUserDNTemplate,
		ConfigKeySubjectAttr, ConfigKeyEmailAttr, ConfigKeyDisplayNameAttr, ConfigKeyAllowAutoProvision,
	}
}

func (s *serviceImpl) Get(ctx context.Context) (*Projection, error) {
	snap, err := s.Load(ctx)
	if err != nil {
		return nil, err
	}
	return project(snap), nil
}

func (s *serviceImpl) Load(ctx context.Context) (*Snapshot, error) {
	if s == nil || s.sysConfigSvc == nil {
		return nil, bizerr.WrapCode(gerror.New("settings: sys_config unavailable"), CodeStorageUnavailable)
	}
	result, err := s.sysConfigSvc.BatchGet(ctx, allKeys())
	if err != nil {
		if bizerr.Is(err, capmodel.CodeCapabilityUnavailable) {
			return nil, bizerr.WrapCode(err, CodeStorageUnavailable)
		}
		return nil, bizerr.WrapCode(err, CodeReadFailed)
	}
	snap := &Snapshot{
		ConnectionKey:   ConnectionKeyDefault,
		TLSMode:         TLSModeLDAPS,
		Port:            "636",
		SubjectAttr:     "entryUUID",
		EmailAttr:       "mail",
		DisplayNameAttr: "cn",
		UserFilter:      "(uid={username})",
	}
	if result == nil {
		return snap, nil
	}
	get := func(k hostconfigcap.SysConfigKey) string {
		if info := result.Items[k]; info != nil {
			return info.Value
		}
		return ""
	}
	if v := get(ConfigKeyConnectionKey); v != "" {
		snap.ConnectionKey = v
	}
	snap.DisplayName = get(ConfigKeyDisplayName)
	snap.Host = get(ConfigKeyHost)
	if v := get(ConfigKeyPort); v != "" {
		snap.Port = v
	}
	if v := NormalizeTLSMode(get(ConfigKeyTLSMode)); v != "" {
		snap.TLSMode = v
	}
	snap.BindDN = get(ConfigKeyBindDN)
	snap.BindPassword = get(ConfigKeyBindPassword)
	snap.BaseDN = get(ConfigKeyBaseDN)
	if v := get(ConfigKeyUserFilter); v != "" {
		snap.UserFilter = v
	}
	snap.UserDNTemplate = get(ConfigKeyUserDNTemplate)
	if v := get(ConfigKeySubjectAttr); v != "" {
		snap.SubjectAttr = v
	}
	if v := get(ConfigKeyEmailAttr); v != "" {
		snap.EmailAttr = v
	}
	if v := get(ConfigKeyDisplayNameAttr); v != "" {
		snap.DisplayNameAttr = v
	}
	snap.AllowAutoProvision = get(ConfigKeyAllowAutoProvision) == "1"
	return snap, nil
}

func project(snap *Snapshot) *Projection {
	p := &Projection{ConnectionKey: ConnectionKeyDefault, TLSMode: TLSModeLDAPS}
	if snap == nil {
		return p
	}
	p.ConnectionKey = snap.ConnectionKey
	p.DisplayName = snap.DisplayName
	p.Host = snap.Host
	p.Port = snap.Port
	p.TLSMode = snap.TLSMode
	p.BindDN = snap.BindDN
	p.BaseDN = snap.BaseDN
	p.UserFilter = snap.UserFilter
	p.UserDNTemplate = snap.UserDNTemplate
	p.SubjectAttr = snap.SubjectAttr
	p.EmailAttr = snap.EmailAttr
	p.DisplayNameAttr = snap.DisplayNameAttr
	p.AllowAutoProvision = snap.AllowAutoProvision
	if snap.BindPassword != "" {
		p.BindPasswordConfigured = true
		p.BindPasswordMasked = SecretMask
	}
	return p
}
