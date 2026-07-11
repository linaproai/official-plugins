package settings

import (
	"context"

	v1 "lina-plugin-linapro-auth-ldap/backend/api/settings/v1"
	settingssvc "lina-plugin-linapro-auth-ldap/backend/internal/service/settings"
)

func (c *ControllerV1) SaveSettings(ctx context.Context, req *v1.SaveSettingsReq) (*v1.SaveSettingsRes, error) {
	p, err := c.settingsSvc.Save(ctx, settingssvc.SaveInput{
		DisplayName: req.DisplayName, Host: req.Host, Port: req.Port, TLSMode: req.TlsMode,
		BindDN: req.BindDn, BindPassword: req.BindPassword, BaseDN: req.BaseDn,
		UserFilter: req.UserFilter, UserDNTemplate: req.UserDnTemplate,
		SubjectAttr: req.SubjectAttr, EmailAttr: req.EmailAttr, DisplayNameAttr: req.DisplayNameAttr,
		AllowAutoProvision: req.AllowAutoProvision,
	})
	if err != nil {
		return nil, err
	}
	return &v1.SaveSettingsRes{Settings: project(p)}, nil
}

func project(p *settingssvc.Projection) *v1.SettingsItem {
	if p == nil {
		return &v1.SettingsItem{}
	}
	return &v1.SettingsItem{
		ConnectionKey: p.ConnectionKey, DisplayName: p.DisplayName, Host: p.Host, Port: p.Port,
		TlsMode: p.TLSMode, BindDn: p.BindDN, BindPasswordMasked: p.BindPasswordMasked,
		BindPasswordConfigured: p.BindPasswordConfigured, BaseDn: p.BaseDN, UserFilter: p.UserFilter,
		UserDnTemplate: p.UserDNTemplate, SubjectAttr: p.SubjectAttr, EmailAttr: p.EmailAttr,
		DisplayNameAttr: p.DisplayNameAttr, AllowAutoProvision: p.AllowAutoProvision,
	}
}
