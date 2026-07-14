// ldapauth_client.go implements LDAP bind authentication with go-ldap.

package ldapauth

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"strings"
	"time"

	ldap "github.com/go-ldap/ldap/v3"
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/bizerr"
	settingssvc "lina-plugin-linapro-auth-ldap/backend/internal/service/settings"
)

const ldapTimeout = 10 * time.Second

// LDAPDirectoryClient is the production DirectoryClient.
type LDAPDirectoryClient struct {
	// DialTimeout bounds dial and operations.
	DialTimeout time.Duration
}

// NewLDAPDirectoryClient returns the production client.
func NewLDAPDirectoryClient() *LDAPDirectoryClient {
	return &LDAPDirectoryClient{DialTimeout: ldapTimeout}
}

// Authenticate connects, resolves the user DN, binds with the password, and reads attributes.
func (c *LDAPDirectoryClient) Authenticate(ctx context.Context, cfg settingssvc.Snapshot, username, password string) (*VerifiedIdentity, error) {
	_ = ctx
	timeout := c.DialTimeout
	if timeout <= 0 {
		timeout = ldapTimeout
	}
	conn, err := c.dial(cfg, timeout)
	if err != nil {
		return nil, bizerr.WrapCode(gerror.Wrap(err, "ldap dial failed"), CodeDirectoryUnavailable)
	}
	defer conn.Close()
	conn.SetTimeout(timeout)

	userDN, entry, err := c.resolveUser(conn, cfg, username)
	if err != nil {
		return nil, err
	}
	if err := conn.Bind(userDN, password); err != nil {
		return nil, bizerr.WrapCode(gerror.New("ldap bind failed"), CodeAuthFailed)
	}
	// Re-read attributes after user bind when search entry missing.
	if entry == nil {
		entry, err = c.readEntry(conn, userDN, cfg)
		if err != nil {
			return nil, err
		}
	}
	identity := identityFromEntry(entry, cfg, username)
	if identity.Subject == "" {
		return nil, bizerr.WrapCode(gerror.New("ldap subject attribute empty"), CodeAuthFailed)
	}
	return identity, nil
}

func (c *LDAPDirectoryClient) dial(cfg settingssvc.Snapshot, timeout time.Duration) (*ldap.Conn, error) {
	var (
		host = strings.TrimSpace(cfg.Host)
		port = strings.TrimSpace(cfg.Port)
		mode = settingssvc.NormalizeTLSMode(cfg.TLSMode)
	)
	if port == "" {
		if mode == settingssvc.TLSModeLDAPS {
			port = "636"
		} else {
			port = "389"
		}
	}
	addr := net.JoinHostPort(host, port)
	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12, ServerName: host}

	switch mode {
	case settingssvc.TLSModeLDAPS:
		return ldap.DialURL("ldaps://"+addr, ldap.DialWithTLSConfig(tlsConfig), ldap.DialWithDialer(&net.Dialer{Timeout: timeout}))
	case settingssvc.TLSModeStartTLS:
		conn, err := ldap.DialURL("ldap://"+addr, ldap.DialWithDialer(&net.Dialer{Timeout: timeout}))
		if err != nil {
			return nil, err
		}
		if err := conn.StartTLS(tlsConfig); err != nil {
			conn.Close()
			return nil, err
		}
		return conn, nil
	case settingssvc.TLSModePlain:
		if !settingssvc.IsLocalHost(host) {
			return nil, gerror.New("plain ldap not allowed")
		}
		return ldap.DialURL("ldap://"+addr, ldap.DialWithDialer(&net.Dialer{Timeout: timeout}))
	default:
		return nil, gerror.New("invalid tls mode")
	}
}

func (c *LDAPDirectoryClient) resolveUser(conn *ldap.Conn, cfg settingssvc.Snapshot, username string) (string, *ldap.Entry, error) {
	template := strings.TrimSpace(cfg.UserDNTemplate)
	if template != "" {
		return RenderUserDN(template, username), nil, nil
	}
	// Search path requires optional service bind.
	if strings.TrimSpace(cfg.BindDN) != "" {
		if err := conn.Bind(cfg.BindDN, cfg.BindPassword); err != nil {
			return "", nil, bizerr.WrapCode(gerror.New("ldap service bind failed"), CodeDirectoryUnavailable)
		}
	}
	filter := RenderUserFilter(cfg.UserFilter, username)
	if strings.TrimSpace(filter) == "" {
		return "", nil, bizerr.WrapCode(gerror.New("ldap user filter empty"), CodeConfigMissing)
	}
	attrs := uniqueAttrs(cfg.SubjectAttr, cfg.EmailAttr, cfg.DisplayNameAttr, "dn")
	req := ldap.NewSearchRequest(
		cfg.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 2, int(ldapTimeout.Seconds()), false,
		filter,
		attrs,
		nil,
	)
	res, err := conn.Search(req)
	if err != nil {
		return "", nil, bizerr.WrapCode(gerror.New("ldap search failed"), CodeDirectoryUnavailable)
	}
	if res == nil || len(res.Entries) == 0 {
		return "", nil, bizerr.WrapCode(gerror.New("ldap user not found"), CodeAuthFailed)
	}
	if len(res.Entries) > 1 {
		return "", nil, bizerr.WrapCode(gerror.New("ldap user not unique"), CodeAuthFailed)
	}
	entry := res.Entries[0]
	return entry.DN, entry, nil
}

func (c *LDAPDirectoryClient) readEntry(conn *ldap.Conn, dn string, cfg settingssvc.Snapshot) (*ldap.Entry, error) {
	attrs := uniqueAttrs(cfg.SubjectAttr, cfg.EmailAttr, cfg.DisplayNameAttr)
	req := ldap.NewSearchRequest(
		dn,
		ldap.ScopeBaseObject, ldap.NeverDerefAliases, 1, int(ldapTimeout.Seconds()), false,
		"(objectClass=*)",
		attrs,
		nil,
	)
	res, err := conn.Search(req)
	if err != nil || res == nil || len(res.Entries) == 0 {
		// Fallback: synthetic entry with DN only; subject may come from DN uid.
		return &ldap.Entry{DN: dn}, nil
	}
	return res.Entries[0], nil
}

func identityFromEntry(entry *ldap.Entry, cfg settingssvc.Snapshot, username string) *VerifiedIdentity {
	if entry == nil {
		return &VerifiedIdentity{}
	}
	subjectAttr := strings.TrimSpace(cfg.SubjectAttr)
	if subjectAttr == "" {
		subjectAttr = "entryUUID"
	}
	subject := firstAttr(entry, subjectAttr)
	// Binary attributes (e.g. AD objectGUID) via raw bytes.
	if subject == "" {
		if raw := entry.GetRawAttributeValue(subjectAttr); len(raw) > 0 {
			subject = base64.RawURLEncoding.EncodeToString(raw)
		}
	}
	if subject == "" && strings.EqualFold(subjectAttr, "uid") {
		subject = username
	}
	if subject == "" {
		// last resort: use username only if subject attr configured as uid-like
		if strings.EqualFold(subjectAttr, "sAMAccountName") || strings.EqualFold(subjectAttr, "userPrincipalName") {
			subject = username
		}
	}
	emailAttr := cfg.EmailAttr
	if emailAttr == "" {
		emailAttr = "mail"
	}
	nameAttr := cfg.DisplayNameAttr
	if nameAttr == "" {
		nameAttr = "cn"
	}
	return &VerifiedIdentity{
		Subject:     strings.TrimSpace(subject),
		Email:       strings.TrimSpace(firstAttr(entry, emailAttr)),
		DisplayName: strings.TrimSpace(firstAttr(entry, nameAttr)),
	}
}

func firstAttr(entry *ldap.Entry, name string) string {
	if entry == nil || name == "" {
		return ""
	}
	vals := entry.GetAttributeValues(name)
	if len(vals) == 0 {
		// case-insensitive scan
		for _, a := range entry.Attributes {
			if strings.EqualFold(a.Name, name) && len(a.Values) > 0 {
				return a.Values[0]
			}
		}
		return ""
	}
	return vals[0]
}

func uniqueAttrs(names ...string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, n := range names {
		n = strings.TrimSpace(n)
		if n == "" {
			continue
		}
		key := strings.ToLower(n)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, n)
	}
	if len(out) == 0 {
		return []string{"dn"}
	}
	return out
}

// FormatLDAPURL is a test helper for address construction.
func FormatLDAPURL(host, port, mode string) string {
	if port == "" {
		port = "389"
	}
	addr := net.JoinHostPort(host, port)
	if settingssvc.NormalizeTLSMode(mode) == settingssvc.TLSModeLDAPS {
		return fmt.Sprintf("ldaps://%s", addr)
	}
	return fmt.Sprintf("ldap://%s", addr)
}
