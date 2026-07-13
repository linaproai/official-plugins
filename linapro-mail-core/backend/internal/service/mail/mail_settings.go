// This file implements the single platform mail-account settings surface.
// The admin page only manages one default account with SMTP and optional IMAP/POP3.

package mail

import (
	"context"
	"strings"
	"time"

	"lina-core/pkg/bizerr"
	"lina-plugin-linapro-mail-core/backend/cap/mailcap"
	"lina-plugin-linapro-mail-core/backend/cap/mailcap/spi"
	"lina-plugin-linapro-mail-core/backend/internal/dao"
	"lina-plugin-linapro-mail-core/backend/internal/model/do"
	"lina-plugin-linapro-mail-core/backend/internal/model/entity"
)

// InboundKindNone disables inbound for the platform account.
const InboundKindNone = "none"

// PlatformSettingsInput is the form payload for save and test.
type PlatformSettingsInput struct {
	Name            string
	FromAddress     string
	SmtpHost        string
	SmtpPort        int
	SmtpUsername    string
	SmtpPassword    string
	SmtpTlsMode     string
	InboundKind     string
	InboundHost     string
	InboundPort     int
	InboundUsername string
	InboundPassword string
	InboundTlsMode  string
}

// PlatformSettingsProjection is the masked GET projection for the admin page.
type PlatformSettingsProjection struct {
	Name                      string
	FromAddress               string
	SmtpHost                  string
	SmtpPort                  int
	SmtpUsername              string
	SmtpPasswordConfigured    bool
	SmtpTlsMode               string
	InboundKind               string
	InboundHost               string
	InboundPort               int
	InboundUsername           string
	InboundPasswordConfigured bool
	InboundTlsMode            string
	Configured                bool
}

// PlatformSettingsTestResult is the connectivity probe outcome.
type PlatformSettingsTestResult struct {
	OK      bool
	Message string
}

// GetPlatformSettings loads the singleton platform account and bound connections.
func (s *serviceImpl) GetPlatformSettings(ctx context.Context) (*PlatformSettingsProjection, error) {
	account, err := s.findSingletonAccount(ctx)
	if err != nil {
		return nil, err
	}
	out := &PlatformSettingsProjection{
		SmtpTlsMode:    TLSModeStartTLS,
		InboundKind:    InboundKindNone,
		InboundTlsMode: TLSModeStartTLS,
	}
	if account == nil {
		return out, nil
	}
	out.Configured = true
	out.Name = account.Name
	out.FromAddress = account.FromAddress
	if account.OutboundConnectionId > 0 {
		conn, getErr := s.GetConnection(ctx, account.OutboundConnectionId)
		if getErr == nil && conn != nil {
			out.SmtpHost = conn.Host
			out.SmtpPort = conn.Port
			out.SmtpUsername = conn.Username
			out.SmtpPasswordConfigured = strings.TrimSpace(conn.SecretRef) != ""
			out.SmtpTlsMode = normalizeTLSMode(conn.TlsMode)
		}
	}
	if account.InboundConnectionId > 0 {
		conn, getErr := s.GetConnection(ctx, account.InboundConnectionId)
		if getErr == nil && conn != nil {
			out.InboundKind = strings.TrimSpace(conn.Kind)
			out.InboundHost = conn.Host
			out.InboundPort = conn.Port
			out.InboundUsername = conn.Username
			out.InboundPasswordConfigured = strings.TrimSpace(conn.SecretRef) != ""
			out.InboundTlsMode = normalizeTLSMode(conn.TlsMode)
		}
	}
	return out, nil
}

// SavePlatformSettings upserts the singleton account with SMTP and optional inbound.
// Empty passwords keep previously stored secrets. Extra accounts are soft-deleted.
// Empty FromAddress defaults to the mailbox account username; account name is derived
// when omitted so the admin page no longer requires a separate display-name field.
func (s *serviceImpl) SavePlatformSettings(ctx context.Context, in PlatformSettingsInput) (*PlatformSettingsProjection, error) {
	in = normalizePlatformSettingsInput(in)
	if err := validatePlatformSettingsSave(in); err != nil {
		return nil, err
	}
	account, err := s.findSingletonAccount(ctx)
	if err != nil {
		return nil, err
	}
	var prevOutbound, prevInbound *entity.Connection
	if account != nil {
		if account.OutboundConnectionId > 0 {
			prevOutbound, _ = s.GetConnection(ctx, account.OutboundConnectionId)
		}
		if account.InboundConnectionId > 0 {
			prevInbound, _ = s.GetConnection(ctx, account.InboundConnectionId)
		}
	}

	smtpSecret := resolveSecret(in.SmtpPassword, prevOutbound)
	if smtpSecret == "" {
		return nil, bizerr.NewCode(CodeSettingsPasswordRequired, bizerr.P("field", "smtp"))
	}
	accountName := resolvePlatformAccountName(in)
	outboundID, err := s.upsertConnection(ctx, prevOutbound, CreateConnectionInput{
		Name:      platformConnectionName("smtp", accountName),
		Kind:      string(mailcap.KindSMTP),
		Host:      in.SmtpHost,
		Port:      in.SmtpPort,
		Username:  in.SmtpUsername,
		SecretRef: smtpSecret,
		TLSMode:   in.SmtpTlsMode,
		AuthMode:  AuthModePassword,
		Status:    StatusEnabled,
		Remark:    "platform-singleton",
	})
	if err != nil {
		return nil, err
	}

	inboundKind := normalizeInboundKind(in.InboundKind)
	var inboundID int64
	if inboundKind != InboundKindNone {
		if strings.TrimSpace(in.InboundHost) == "" {
			return nil, bizerr.NewCode(CodeConnectionHostRequired)
		}
		if in.InboundPort <= 0 || in.InboundPort > 65535 {
			return nil, bizerr.NewCode(CodeConnectionPortInvalid)
		}
		// Reuse previous inbound only when kind matches.
		var prevForKind *entity.Connection
		if prevInbound != nil && strings.EqualFold(prevInbound.Kind, inboundKind) {
			prevForKind = prevInbound
		}
		inboundSecret := resolveSecret(in.InboundPassword, prevForKind)
		if inboundSecret == "" {
			// Fall back to SMTP password when inbound password is empty and SMTP just provided.
			inboundSecret = resolveSecret(in.SmtpPassword, prevOutbound)
		}
		if inboundSecret == "" {
			return nil, bizerr.NewCode(CodeSettingsPasswordRequired, bizerr.P("field", "inbound"))
		}
		inboundID, err = s.upsertConnection(ctx, prevForKind, CreateConnectionInput{
			Name:      platformConnectionName(inboundKind, accountName),
			Kind:      inboundKind,
			Host:      in.InboundHost,
			Port:      in.InboundPort,
			Username:  firstNonEmpty(in.InboundUsername, in.SmtpUsername),
			SecretRef: inboundSecret,
			TLSMode:   in.InboundTlsMode,
			AuthMode:  AuthModePassword,
			Status:    StatusEnabled,
			Remark:    "platform-singleton",
		})
		if err != nil {
			return nil, err
		}
		// Drop previous inbound connection when protocol kind changed.
		if prevInbound != nil && prevForKind == nil {
			_ = s.DeleteConnections(ctx, []int64{prevInbound.Id})
		}
	} else if prevInbound != nil {
		_ = s.DeleteConnections(ctx, []int64{prevInbound.Id})
	}

	if account == nil {
		_, err = s.CreateAccount(ctx, CreateAccountInput{
			Name:                 accountName,
			FromAddress:          strings.TrimSpace(in.FromAddress),
			OutboundConnectionID: outboundID,
			InboundConnectionID:  inboundID,
			IsDefault:            true,
			Status:               StatusEnabled,
			Remark:               "platform-singleton",
		})
	} else {
		err = s.UpdateAccount(ctx, UpdateAccountInput{
			ID:                   account.Id,
			Name:                 accountName,
			FromAddress:          strings.TrimSpace(in.FromAddress),
			OutboundConnectionID: outboundID,
			InboundConnectionID:  inboundID,
			IsDefault:            true,
			Status:               StatusEnabled,
			Remark:               "platform-singleton",
		})
	}
	if err != nil {
		return nil, err
	}
	if err = s.deleteExtraAccounts(ctx); err != nil {
		return nil, err
	}
	return s.GetPlatformSettings(ctx)
}

// TestPlatformReceive probes inbound receive capability using the form IMAP/POP3 values.
// Secrets may be resolved from previously stored inbound (or SMTP fallback) when the password
// field is empty. The probe uses an ephemeral endpoint built from the form — not by resolving
// the platform default Account's inbound connection ID alone.
func (s *serviceImpl) TestPlatformReceive(
	ctx context.Context,
	in PlatformSettingsInput,
) (*PlatformSettingsTestResult, error) {
	in = normalizePlatformSettingsInput(in)
	inboundKind := normalizeInboundKind(in.InboundKind)
	if inboundKind == InboundKindNone {
		return nil, bizerr.NewCode(CodeSettingsInboundRequired)
	}
	if inboundKind != string(mailcap.KindIMAP) && inboundKind != string(mailcap.KindPOP3) {
		return nil, bizerr.NewCode(CodeConnectionKindInvalid)
	}
	if strings.TrimSpace(in.InboundHost) == "" {
		return nil, bizerr.NewCode(CodeConnectionHostRequired)
	}
	if in.InboundPort <= 0 || in.InboundPort > 65535 {
		return nil, bizerr.NewCode(CodeConnectionPortInvalid)
	}
	username := firstNonEmpty(in.InboundUsername, in.SmtpUsername)
	if username == "" {
		return nil, bizerr.NewCode(CodeSettingsUsernameRequired)
	}

	account, err := s.findSingletonAccount(ctx)
	if err != nil {
		return nil, err
	}
	var prevOutbound, prevInbound *entity.Connection
	if account != nil {
		if account.OutboundConnectionId > 0 {
			prevOutbound, _ = s.GetConnection(ctx, account.OutboundConnectionId)
		}
		if account.InboundConnectionId > 0 {
			prevInbound, _ = s.GetConnection(ctx, account.InboundConnectionId)
		}
	}
	var prevForKind *entity.Connection
	if prevInbound != nil && strings.EqualFold(prevInbound.Kind, inboundKind) {
		prevForKind = prevInbound
	}
	inboundSecret := resolveSecret(in.InboundPassword, prevForKind)
	if inboundSecret == "" {
		// Shared mailbox password: reuse SMTP secret when inbound field is empty.
		inboundSecret = resolveSecret(in.SmtpPassword, prevOutbound)
	}
	if inboundSecret == "" {
		return nil, bizerr.NewCode(CodeSettingsPasswordRequired, bizerr.P("field", "inbound"))
	}
	if err = s.probeEndpoint(ctx, mailcap.Kind(inboundKind), mailcap.ConnectionEndpoint{
		Kind:     mailcap.Kind(inboundKind),
		Host:     strings.TrimSpace(in.InboundHost),
		Port:     in.InboundPort,
		Username: username,
		Secret:   inboundSecret,
		TLSMode:  normalizeTLSMode(in.InboundTlsMode),
		AuthMode: AuthModePassword,
	}); err != nil {
		// Propagate structured bizerr so callers see errorCode/messageKey/localized message.
		return nil, err
	}
	return &PlatformSettingsTestResult{OK: true, Message: "ok"}, nil
}

// SendPlatformTestMail sends one diagnostic email using the form SMTP values.
// Secrets may be resolved from previously stored outbound when the password field is empty.
// The message is delivered through the SMTP transport SPI with an ephemeral endpoint built
// from the form — not by resolving the platform default Account's outbound connection ID alone.
func (s *serviceImpl) SendPlatformTestMail(
	ctx context.Context,
	in PlatformSettingsInput,
	to, subject, body string,
) (*PlatformSettingsTestResult, error) {
	in = normalizePlatformSettingsInput(in)
	to = strings.TrimSpace(to)
	subject = strings.TrimSpace(subject)
	body = strings.TrimSpace(body)
	if strings.TrimSpace(in.SmtpUsername) == "" {
		return nil, bizerr.NewCode(CodeSettingsUsernameRequired)
	}
	if strings.TrimSpace(in.SmtpHost) == "" {
		return nil, bizerr.NewCode(CodeConnectionHostRequired)
	}
	if in.SmtpPort <= 0 || in.SmtpPort > 65535 {
		return nil, bizerr.NewCode(CodeConnectionPortInvalid)
	}
	if to == "" {
		return nil, bizerr.NewCode(CodeSettingsRecipientRequired)
	}
	if body == "" {
		return nil, bizerr.NewCode(CodeSettingsBodyRequired)
	}
	if subject == "" {
		subject = "LinaPro mail configuration test"
	}

	account, err := s.findSingletonAccount(ctx)
	if err != nil {
		return nil, err
	}
	var prevOutbound *entity.Connection
	if account != nil && account.OutboundConnectionId > 0 {
		prevOutbound, _ = s.GetConnection(ctx, account.OutboundConnectionId)
	}
	smtpSecret := resolveSecret(in.SmtpPassword, prevOutbound)
	if smtpSecret == "" {
		// Return bizerr so the host Response middleware localizes via messageKey.
		return nil, bizerr.NewCode(CodeSettingsPasswordRequired, bizerr.P("field", "smtp"))
	}

	endpoint := mailcap.ConnectionEndpoint{
		Kind:     mailcap.KindSMTP,
		Host:     strings.TrimSpace(in.SmtpHost),
		Port:     in.SmtpPort,
		Username: strings.TrimSpace(in.SmtpUsername),
		Secret:   smtpSecret,
		TLSMode:  normalizeTLSMode(in.SmtpTlsMode),
		AuthMode: AuthModePassword,
	}
	_, transport, err := spi.ResolveOutbound(ctx, mailcap.KindSMTP, s.enablement())
	if err != nil {
		// Propagate structured bizerr (e.g. transport unavailable) for locale rendering.
		return nil, err
	}
	from := strings.TrimSpace(in.FromAddress)
	if from == "" {
		from = strings.TrimSpace(in.SmtpUsername)
	}
	if _, err = transport.Send(ctx, endpoint, mailcap.MailMessage{
		From:     from,
		To:       []string{to},
		Subject:  subject,
		TextBody: body,
	}); err != nil {
		// Transport send failures may already be bizerr; keep the chain intact.
		return nil, err
	}
	return &PlatformSettingsTestResult{OK: true, Message: "ok"}, nil
}

// TestPlatformSettings probes SMTP and optional inbound without writing.
func (s *serviceImpl) TestPlatformSettings(ctx context.Context, in PlatformSettingsInput) (*PlatformSettingsTestResult, error) {
	in = normalizePlatformSettingsInput(in)
	if strings.TrimSpace(in.SmtpUsername) == "" {
		return nil, bizerr.NewCode(CodeSettingsUsernameRequired)
	}
	if strings.TrimSpace(in.SmtpHost) == "" {
		return nil, bizerr.NewCode(CodeConnectionHostRequired)
	}
	if in.SmtpPort <= 0 || in.SmtpPort > 65535 {
		return nil, bizerr.NewCode(CodeConnectionPortInvalid)
	}
	account, err := s.findSingletonAccount(ctx)
	if err != nil {
		return nil, err
	}
	var prevOutbound, prevInbound *entity.Connection
	if account != nil {
		if account.OutboundConnectionId > 0 {
			prevOutbound, _ = s.GetConnection(ctx, account.OutboundConnectionId)
		}
		if account.InboundConnectionId > 0 {
			prevInbound, _ = s.GetConnection(ctx, account.InboundConnectionId)
		}
	}
	smtpSecret := resolveSecret(in.SmtpPassword, prevOutbound)
	if smtpSecret == "" {
		// Return bizerr so the host Response middleware localizes via messageKey.
		return nil, bizerr.NewCode(CodeSettingsPasswordRequired, bizerr.P("field", "smtp"))
	}
	if err = s.probeEndpoint(ctx, mailcap.KindSMTP, mailcap.ConnectionEndpoint{
		Kind:     mailcap.KindSMTP,
		Host:     strings.TrimSpace(in.SmtpHost),
		Port:     in.SmtpPort,
		Username: strings.TrimSpace(in.SmtpUsername),
		Secret:   smtpSecret,
		TLSMode:  normalizeTLSMode(in.SmtpTlsMode),
		AuthMode: AuthModePassword,
	}); err != nil {
		// Propagate structured bizerr so callers see errorCode/messageKey/localized message.
		return nil, err
	}

	inboundKind := normalizeInboundKind(in.InboundKind)
	if inboundKind != InboundKindNone {
		if strings.TrimSpace(in.InboundHost) == "" {
			return nil, bizerr.NewCode(CodeConnectionHostRequired)
		}
		if in.InboundPort <= 0 || in.InboundPort > 65535 {
			return nil, bizerr.NewCode(CodeConnectionPortInvalid)
		}
		var prevForKind *entity.Connection
		if prevInbound != nil && strings.EqualFold(prevInbound.Kind, inboundKind) {
			prevForKind = prevInbound
		}
		inboundSecret := resolveSecret(in.InboundPassword, prevForKind)
		if inboundSecret == "" {
			inboundSecret = smtpSecret
		}
		if err = s.probeEndpoint(ctx, mailcap.Kind(inboundKind), mailcap.ConnectionEndpoint{
			Kind:     mailcap.Kind(inboundKind),
			Host:     strings.TrimSpace(in.InboundHost),
			Port:     in.InboundPort,
			Username: firstNonEmpty(in.InboundUsername, in.SmtpUsername),
			Secret:   inboundSecret,
			TLSMode:  normalizeTLSMode(in.InboundTlsMode),
			AuthMode: AuthModePassword,
		}); err != nil {
			// Propagate structured bizerr for inbound transport probe failures.
			return nil, err
		}
	}
	return &PlatformSettingsTestResult{OK: true, Message: "ok"}, nil
}

func (s *serviceImpl) probeEndpoint(ctx context.Context, kind mailcap.Kind, endpoint mailcap.ConnectionEndpoint) error {
	switch kind {
	case mailcap.KindSMTP:
		_, transport, err := spi.ResolveOutbound(ctx, kind, s.enablement())
		if err != nil {
			return err
		}
		return transport.Probe(ctx, endpoint)
	case mailcap.KindIMAP, mailcap.KindPOP3:
		_, transport, err := spi.ResolveInbound(ctx, kind, s.enablement())
		if err != nil {
			return err
		}
		return transport.Probe(ctx, endpoint)
	default:
		return bizerr.NewCode(CodeConnectionKindInvalid)
	}
}

func (s *serviceImpl) findSingletonAccount(ctx context.Context) (*entity.Account, error) {
	var preferred entity.Account
	err := dao.Account.Ctx(ctx).
		Where(dao.Account.Columns().IsDefault, FlagDefault).
		Where(dao.Account.Columns().Status, StatusEnabled).
		OrderAsc(dao.Account.Columns().Id).
		Limit(1).
		Scan(&preferred)
	if err != nil && !isNoRows(err) {
		return nil, err
	}
	if preferred.Id > 0 {
		return &preferred, nil
	}
	var first entity.Account
	err = dao.Account.Ctx(ctx).OrderAsc(dao.Account.Columns().Id).Limit(1).Scan(&first)
	if err != nil && !isNoRows(err) {
		return nil, err
	}
	if first.Id == 0 {
		return nil, nil
	}
	return &first, nil
}

// isNoRows reports empty Scan results from GoFrame/database/sql.
func isNoRows(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "no rows in result set") ||
		strings.Contains(msg, "sql: no rows")
}

func (s *serviceImpl) upsertConnection(ctx context.Context, prev *entity.Connection, in CreateConnectionInput) (int64, error) {
	if prev != nil && prev.Id > 0 {
		err := s.UpdateConnection(ctx, UpdateConnectionInput{
			ID:        prev.Id,
			Name:      in.Name,
			Kind:      in.Kind,
			Host:      in.Host,
			Port:      in.Port,
			Username:  in.Username,
			SecretRef: in.SecretRef,
			TLSMode:   in.TLSMode,
			AuthMode:  in.AuthMode,
			ExtraJSON: in.ExtraJSON,
			Status:    in.Status,
			Remark:    in.Remark,
		})
		if err != nil {
			return 0, err
		}
		return prev.Id, nil
	}
	return s.CreateConnection(ctx, in)
}

func (s *serviceImpl) deleteExtraAccounts(ctx context.Context) error {
	singleton, err := s.findSingletonAccount(ctx)
	if err != nil || singleton == nil {
		return err
	}
	// Ensure the singleton is default.
	if singleton.IsDefault != FlagDefault {
		if _, err = dao.Account.Ctx(ctx).Where(dao.Account.Columns().Id, singleton.Id).
			Data(do.Account{IsDefault: FlagDefault}).Update(); err != nil {
			return err
		}
	}
	var extras []*entity.Account
	if err = dao.Account.Ctx(ctx).WhereNot(dao.Account.Columns().Id, singleton.Id).Scan(&extras); err != nil {
		return err
	}
	if len(extras) == 0 {
		return nil
	}
	ids := make([]int64, 0, len(extras))
	for _, row := range extras {
		if row != nil && row.Id > 0 {
			ids = append(ids, row.Id)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return s.DeleteAccounts(ctx, ids)
}

func validatePlatformSettingsSave(in PlatformSettingsInput) error {
	// Callers must run normalizePlatformSettingsInput first so From defaults to account.
	if strings.TrimSpace(in.SmtpUsername) == "" {
		return bizerr.NewCode(CodeSettingsUsernameRequired)
	}
	if strings.TrimSpace(in.FromAddress) == "" {
		return bizerr.NewCode(CodeSettingsUsernameRequired)
	}
	if strings.TrimSpace(in.SmtpHost) == "" {
		return bizerr.NewCode(CodeConnectionHostRequired)
	}
	if in.SmtpPort <= 0 || in.SmtpPort > 65535 {
		return bizerr.NewCode(CodeConnectionPortInvalid)
	}
	kind := normalizeInboundKind(in.InboundKind)
	if kind != InboundKindNone && kind != string(mailcap.KindIMAP) && kind != string(mailcap.KindPOP3) {
		return bizerr.NewCode(CodeConnectionKindInvalid)
	}
	return nil
}

// normalizePlatformSettingsInput trims credentials and defaults From to the mailbox account.
func normalizePlatformSettingsInput(in PlatformSettingsInput) PlatformSettingsInput {
	in.Name = strings.TrimSpace(in.Name)
	in.SmtpUsername = strings.TrimSpace(in.SmtpUsername)
	in.InboundUsername = strings.TrimSpace(in.InboundUsername)
	in.FromAddress = strings.TrimSpace(in.FromAddress)
	in.SmtpHost = strings.TrimSpace(in.SmtpHost)
	in.InboundHost = strings.TrimSpace(in.InboundHost)
	if in.FromAddress == "" {
		in.FromAddress = in.SmtpUsername
	}
	if in.InboundUsername == "" {
		in.InboundUsername = in.SmtpUsername
	}
	return in
}

// resolvePlatformAccountName picks a stable internal label when the admin page omits name.
func resolvePlatformAccountName(in PlatformSettingsInput) string {
	if name := strings.TrimSpace(in.Name); name != "" {
		return name
	}
	return firstNonEmpty(in.FromAddress, in.SmtpUsername, "platform")
}

func normalizeInboundKind(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", InboundKindNone, "0", "false":
		return InboundKindNone
	case string(mailcap.KindIMAP):
		return string(mailcap.KindIMAP)
	case string(mailcap.KindPOP3):
		return string(mailcap.KindPOP3)
	default:
		return strings.ToLower(strings.TrimSpace(value))
	}
}

func resolveSecret(submitted string, prev *entity.Connection) string {
	if secret := strings.TrimSpace(submitted); secret != "" {
		return secret
	}
	if prev != nil {
		return strings.TrimSpace(prev.SecretRef)
	}
	return ""
}

func platformConnectionName(kind, accountName string) string {
	base := strings.TrimSpace(accountName)
	if base == "" {
		base = "platform"
	}
	return base + "-" + kind + "-" + time.Now().UTC().Format("20060102")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
