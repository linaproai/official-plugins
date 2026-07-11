// oauth_config.go defines plugin-owned generic OIDC configuration values.

package oauth

import "strings"

// Config carries OIDC client credentials and resolved endpoints.
type Config struct {
	Issuer          string
	ClientID        string
	ClientSecret    string
	RedirectURL     string
	AuthorizeURL    string
	TokenURL        string
	JWKSURL         string
	UserInfoURL     string
	Scopes          []string
	LoginReturnPath string
	DisplayName     string
}

func defaultConfig() Config {
	return Config{
		Scopes:          []string{"openid", "email", "profile"},
		LoginReturnPath: "/admin/auth/login",
	}
}

// DefaultConfig returns a copy of the static defaults.
func DefaultConfig() Config {
	return defaultConfig()
}

func isConfiguredCredential(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return false
	}
	upper := strings.ToUpper(trimmed)
	return !strings.HasPrefix(upper, "REPLACE_ME")
}

func sanitizeReturnTo(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	if !strings.HasPrefix(trimmed, "/") || strings.HasPrefix(trimmed, "//") {
		return ""
	}
	if strings.Contains(trimmed, "://") || strings.ContainsAny(trimmed, "\r\n") {
		return ""
	}
	return trimmed
}

// SanitizeReturnTo is the exported form for HTTP controllers.
func SanitizeReturnTo(raw string) string {
	return sanitizeReturnTo(raw)
}

func normalizeScopes(raw string, fallback []string) []string {
	fields := strings.Fields(strings.TrimSpace(raw))
	if len(fields) == 0 {
		fields = append([]string{}, fallback...)
	}
	hasOpenID := false
	for _, s := range fields {
		if s == "openid" {
			hasOpenID = true
			break
		}
	}
	if !hasOpenID {
		fields = append([]string{"openid"}, fields...)
	}
	return fields
}

func normalizeIssuer(issuer string) string {
	return strings.TrimRight(strings.TrimSpace(issuer), "/")
}
