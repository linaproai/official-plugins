// Package ldapauth implements directory bind verification and host external login.
package ldapauth

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
	"lina-core/pkg/plugin/capability/authcap/extlogin"
	settingssvc "lina-plugin-linapro-auth-ldap/backend/internal/service/settings"
	"lina-plugin-linapro-extlogin-core/backend/cap/extidcap"
)

// Provider is the stable external-identity provider ID owned by this plugin.
const Provider = "ldap:default"

// Service defines the LDAP login orchestration contract.
type Service interface {
	// Login verifies directory credentials and returns a one-time handoff code.
	Login(ctx context.Context, username, password string) (*LoginOutput, error)
}

// LoginOutput is the successful login payload for the SPA.
type LoginOutput struct {
	Handoff string
}

// DirectoryClient verifies a username/password against LDAP and returns identity claims.
type DirectoryClient interface {
	Authenticate(ctx context.Context, cfg settingssvc.Snapshot, username, password string) (*VerifiedIdentity, error)
}

// VerifiedIdentity is projected after a successful bind.
type VerifiedIdentity struct {
	Subject     string
	Email       string
	DisplayName string
}

var _ Service = (*serviceImpl)(nil)

type serviceImpl struct {
	externalLoginSvc extlogin.Service
	settingsSvc      settingssvc.Service
	directory        DirectoryClient
}

// New creates the LDAP login service.
func New(externalLoginSvc extlogin.Service, settingsSvc settingssvc.Service, directory DirectoryClient) Service {
	if directory == nil {
		directory = NewLDAPDirectoryClient()
	}
	return &serviceImpl{
		externalLoginSvc: externalLoginSvc,
		settingsSvc:      settingsSvc,
		directory:        directory,
	}
}

// Login validates input, binds against the directory, and exchanges for a handoff.
func (s *serviceImpl) Login(ctx context.Context, username, password string) (*LoginOutput, error) {
	user := strings.TrimSpace(username)
	// Do not trim password: directories may be intentional.
	if user == "" {
		return nil, bizerr.WrapCode(gerror.New("ldap: username required"), CodeUsernameRequired)
	}
	if password == "" {
		return nil, bizerr.WrapCode(gerror.New("ldap: password required"), CodePasswordRequired)
	}
	if s == nil || s.settingsSvc == nil {
		return nil, bizerr.WrapCode(gerror.New("ldap: settings unavailable"), CodeConfigMissing)
	}
	snap, err := s.settingsSvc.Load(ctx)
	if err != nil {
		return nil, err
	}
	if snap == nil || !IsLoginConfigured(*snap) {
		return nil, bizerr.WrapCode(gerror.New("ldap: configuration missing"), CodeConfigMissing)
	}
	if err := settingssvc.ValidateTLSMode(snap.Host, snap.TLSMode); err != nil {
		return nil, err
	}
	if s.directory == nil {
		return nil, bizerr.WrapCode(gerror.New("ldap: directory client missing"), CodeDirectoryUnavailable)
	}
	identity, err := s.directory.Authenticate(ctx, *snap, user, password)
	if err != nil {
		// Never wrap underlying messages that might include DN/password details for auth failures.
		if bizerr.Is(err, CodeAuthFailed) {
			return nil, err
		}
		if bizerr.Is(err, CodeDirectoryUnavailable) || bizerr.Is(err, CodeConfigMissing) {
			return nil, err
		}
		logger.Warningf(ctx, "linapro-auth-ldap authenticate failed (details redacted)")
		return nil, bizerr.WrapCode(gerror.New("ldap: authentication failed"), CodeAuthFailed)
	}
	if identity == nil || strings.TrimSpace(identity.Subject) == "" {
		return nil, bizerr.WrapCode(gerror.New("ldap: subject missing"), CodeAuthFailed)
	}
	if s.externalLoginSvc == nil {
		return nil, bizerr.WrapCode(gerror.New("ldap: external login unavailable"), CodeExternalLoginUnavailable)
	}
	logger.Infof(ctx, "linapro-auth-ldap verified provider=%s subject=%s", Provider, identity.Subject)
	loginOut, err := s.externalLoginSvc.LoginByVerifiedIdentity(ctx, extlogin.LoginInput{
		Provider:           Provider,
		Subject:            identity.Subject,
		Email:              identity.Email,
		DisplayName:        identity.DisplayName,
		AllowAutoProvision: snap.AllowAutoProvision,
	})
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeExternalLoginFailed)
	}
	if loginOut == nil {
		return nil, bizerr.WrapCode(gerror.New("ldap: nil login outcome"), CodeExternalLoginFailed)
	}
	handoff, err := extidcap.CreateLoginHandoffFromHost(loginOut)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeExternalLoginFailed)
	}
	return &LoginOutput{Handoff: handoff}, nil
}

// IsLoginConfigured reports whether directory host and a DN resolution path exist.
func IsLoginConfigured(snap settingssvc.Snapshot) bool {
	if strings.TrimSpace(snap.Host) == "" {
		return false
	}
	if strings.TrimSpace(snap.UserDNTemplate) != "" {
		return true
	}
	return strings.TrimSpace(snap.BaseDN) != "" && strings.TrimSpace(snap.UserFilter) != ""
}

// RenderUserDN applies {username} placeholder substitution.
func RenderUserDN(template, username string) string {
	return strings.ReplaceAll(template, "{username}", username)
}

// RenderUserFilter applies {username} and escapes LDAP filter special chars in username.
func RenderUserFilter(filter, username string) string {
	escaped := EscapeFilter(username)
	return strings.ReplaceAll(filter, "{username}", escaped)
}

// EscapeFilter escapes RFC 4515 filter special characters.
func EscapeFilter(value string) string {
	replacer := strings.NewReplacer(
		`\`, `\5c`,
		`*`, `\2a`,
		`(`, `\28`,
		`)`, `\29`,
		`\x00`, `\00`,
	)
	// Also escape NUL
	var b strings.Builder
	for _, r := range value {
		switch r {
		case '\\':
			b.WriteString(`\5c`)
		case '*':
			b.WriteString(`\2a`)
		case '(':
			b.WriteString(`\28`)
		case ')':
			b.WriteString(`\29`)
		case 0:
			b.WriteString(`\00`)
		default:
			b.WriteRune(r)
		}
	}
	_ = replacer
	return b.String()
}
