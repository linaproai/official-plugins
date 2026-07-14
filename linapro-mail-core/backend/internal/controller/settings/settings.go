// This file implements platform mail settings request handlers.

package settings

import (
	"context"

	"lina-core/pkg/bizerr"
	v1 "lina-plugin-linapro-mail-core/backend/api/settings/v1"
	mailsvc "lina-plugin-linapro-mail-core/backend/internal/service/mail"
)

// GetSettings returns the masked single-account projection.
func (c *ControllerV1) GetSettings(ctx context.Context, _ *v1.GetSettingsReq) (*v1.GetSettingsRes, error) {
	out, err := c.mailSvc.GetPlatformSettings(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.GetSettingsRes{Settings: projectSettings(out)}, nil
}

// SaveSettings upserts the single platform account and connections.
func (c *ControllerV1) SaveSettings(ctx context.Context, req *v1.SaveSettingsReq) (*v1.SaveSettingsRes, error) {
	out, err := c.mailSvc.SavePlatformSettings(ctx, mailsvc.PlatformSettingsInput{
		Name:            req.Name,
		FromAddress:     req.FromAddress,
		SmtpHost:        req.SmtpHost,
		SmtpPort:        req.SmtpPort,
		SmtpUsername:    req.SmtpUsername,
		SmtpPassword:    req.SmtpPassword,
		SmtpTlsMode:     req.SmtpTlsMode,
		InboundKind:     req.InboundKind,
		InboundHost:     req.InboundHost,
		InboundPort:     req.InboundPort,
		InboundUsername: req.InboundUsername,
		InboundPassword: req.InboundPassword,
		InboundTlsMode:  req.InboundTlsMode,
	})
	if err != nil {
		return nil, err
	}
	return &v1.SaveSettingsRes{Settings: projectSettings(out)}, nil
}

// TestSettings probes SMTP and optional inbound without persisting.
// Diagnostic failures stay soft (HTTP success, ok=false) so the page modal can
// show the localized reason without the global API error toast path.
func (c *ControllerV1) TestSettings(ctx context.Context, req *v1.TestSettingsReq) (*v1.TestSettingsRes, error) {
	out, err := c.mailSvc.TestPlatformSettings(ctx, mailsvc.PlatformSettingsInput{
		Name:            req.Name,
		FromAddress:     req.FromAddress,
		SmtpHost:        req.SmtpHost,
		SmtpPort:        req.SmtpPort,
		SmtpUsername:    req.SmtpUsername,
		SmtpPassword:    req.SmtpPassword,
		SmtpTlsMode:     req.SmtpTlsMode,
		InboundKind:     req.InboundKind,
		InboundHost:     req.InboundHost,
		InboundPort:     req.InboundPort,
		InboundUsername: req.InboundUsername,
		InboundPassword: req.InboundPassword,
		InboundTlsMode:  req.InboundTlsMode,
	})
	if err != nil {
		return &v1.TestSettingsRes{OK: false, Message: c.localizeDiagnosticError(ctx, err)}, nil
	}
	return &v1.TestSettingsRes{OK: out.OK, Message: out.Message}, nil
}

// SendTestMail delivers one diagnostic message through the form SMTP values.
// Diagnostic failures stay soft (HTTP success, ok=false) with a localized Message.
func (c *ControllerV1) SendTestMail(ctx context.Context, req *v1.SendTestMailReq) (*v1.SendTestMailRes, error) {
	out, err := c.mailSvc.SendPlatformTestMail(ctx, mailsvc.PlatformSettingsInput{
		Name:            req.Name,
		FromAddress:     req.FromAddress,
		SmtpHost:        req.SmtpHost,
		SmtpPort:        req.SmtpPort,
		SmtpUsername:    req.SmtpUsername,
		SmtpPassword:    req.SmtpPassword,
		SmtpTlsMode:     req.SmtpTlsMode,
		InboundKind:     req.InboundKind,
		InboundHost:     req.InboundHost,
		InboundPort:     req.InboundPort,
		InboundUsername: req.InboundUsername,
		InboundPassword: req.InboundPassword,
		InboundTlsMode:  req.InboundTlsMode,
	}, req.To, req.Subject, req.Body)
	if err != nil {
		return &v1.SendTestMailRes{OK: false, Message: c.localizeDiagnosticError(ctx, err)}, nil
	}
	return &v1.SendTestMailRes{OK: out.OK, Message: out.Message}, nil
}

// ReceiveTestMail probes inbound receive capability through the form IMAP/POP3 values.
// Diagnostic failures stay soft (HTTP success, ok=false) with a localized Message.
func (c *ControllerV1) ReceiveTestMail(ctx context.Context, req *v1.ReceiveTestMailReq) (*v1.ReceiveTestMailRes, error) {
	out, err := c.mailSvc.TestPlatformReceive(ctx, mailsvc.PlatformSettingsInput{
		Name:            req.Name,
		FromAddress:     req.FromAddress,
		SmtpHost:        req.SmtpHost,
		SmtpPort:        req.SmtpPort,
		SmtpUsername:    req.SmtpUsername,
		SmtpPassword:    req.SmtpPassword,
		SmtpTlsMode:     req.SmtpTlsMode,
		InboundKind:     req.InboundKind,
		InboundHost:     req.InboundHost,
		InboundPort:     req.InboundPort,
		InboundUsername: req.InboundUsername,
		InboundPassword: req.InboundPassword,
		InboundTlsMode:  req.InboundTlsMode,
	})
	if err != nil {
		return &v1.ReceiveTestMailRes{OK: false, Message: c.localizeDiagnosticError(ctx, err)}, nil
	}
	return &v1.ReceiveTestMailRes{OK: out.OK, Message: out.Message}, nil
}

// localizeDiagnosticError renders one structured bizerr into the request locale.
// Soft probe/send results must not call err.Error() alone, which only exposes the
// English fallback and skips plugin error.json catalogs.
func (c *ControllerV1) localizeDiagnosticError(ctx context.Context, err error) string {
	if err == nil {
		return ""
	}
	messageErr, ok := bizerr.As(err)
	if !ok {
		return err.Error()
	}
	template := messageErr.Fallback()
	if c != nil && c.i18nSvc != nil {
		template = c.i18nSvc.Translate(ctx, messageErr.MessageKey(), messageErr.Fallback())
	}
	if template == "" {
		template = messageErr.Fallback()
	}
	return bizerr.Format(template, messageErr.Params())
}

func projectSettings(p *mailsvc.PlatformSettingsProjection) *v1.SettingsItem {
	if p == nil {
		return &v1.SettingsItem{}
	}
	return &v1.SettingsItem{
		Name:                      p.Name,
		FromAddress:               p.FromAddress,
		SmtpHost:                  p.SmtpHost,
		SmtpPort:                  p.SmtpPort,
		SmtpUsername:              p.SmtpUsername,
		SmtpPasswordConfigured:    p.SmtpPasswordConfigured,
		SmtpTlsMode:               p.SmtpTlsMode,
		InboundKind:               p.InboundKind,
		InboundHost:               p.InboundHost,
		InboundPort:               p.InboundPort,
		InboundUsername:           p.InboundUsername,
		InboundPasswordConfigured: p.InboundPasswordConfigured,
		InboundTlsMode:            p.InboundTlsMode,
		Configured:                p.Configured,
	}
}
