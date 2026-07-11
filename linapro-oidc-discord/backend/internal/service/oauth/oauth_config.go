// oauth_config.go defines the plugin-owned Discord OIDC client configuration
// used by the reference login flow. The configuration is intentionally kept
// as a plain value struct so it can be sourced from the host plugin config
// section, a plugin manifest file, or a startup override without pulling in
// runtime dependencies here.

package oauth

import "strings"

// Config carries the plugin-owned Discord OIDC client credentials and the
// OAuth endpoints the plugin talks to. Real deployments load this from the
// plugin configuration source chain; client credentials stay empty until an
// administrator configures them so login-start fails closed by default.
type Config struct {
	// ClientID is the Discord OAuth 2.0 client ID issued by the Discord
	// Developer Portal.
	ClientID string
	// ClientSecret is the Discord OAuth 2.0 client secret paired with ClientID.
	ClientSecret string
	// RedirectURL is the callback URL registered with Discord. It must resolve
	// to the plugin callback route so Discord routes the browser back through
	// the plugin's own controller.
	RedirectURL string
	// AuthorizeURL is the Discord OAuth authorize endpoint.
	AuthorizeURL string
	// TokenURL is the Discord OAuth token endpoint.
	TokenURL string
	// UserInfoURL is the Discord OIDC userinfo endpoint used to project the
	// verified identity into a stable subject, email, and display name.
	UserInfoURL string
	// Scopes are the OAuth scopes the plugin requests. The reference flow
	// requests identify and email so it can build the verified identity DTO
	// the host expects.
	Scopes []string
	// LoginReturnPath is the SPA path the callback controller redirects to
	// after the host external-login exchange completes. The host issues the
	// tokens or pre-token that must reach the SPA.
	LoginReturnPath string
}

// defaultConfig returns the reference-implementation configuration. The
// endpoints are the well-known Discord OAuth 2.0 URLs. Client credentials stay
// empty until an admin configures them on the settings page so the login-start
// route fails closed instead of redirecting browsers to Discord with a
// placeholder client_id.
func defaultConfig() Config {
	return Config{
		ClientID:     "",
		ClientSecret: "",
		// RedirectURL is intentionally empty: the OAuth service derives the
		// callback URL from the live request host unless the admin overrides
		// it through the settings page.
		RedirectURL:  "",
		AuthorizeURL: "https://discord.com/oauth2/authorize",
		TokenURL:     "https://discord.com/api/oauth2/token",
		UserInfoURL:  "https://discord.com/api/users/@me",
		Scopes:       []string{"identify", "email"},
		// Prefer the history-mode SPA path used by the default workspace base
		// (/admin/). Hash-mode deployments can still pass returnTo from the
		// login entry button so error/success redirects land on the live page.
		LoginReturnPath: "/admin/auth/login",
	}
}

// DefaultConfig returns one exported copy of the reference-implementation
// configuration so wiring code can start from the reference values and
// override individual fields as needed.
func DefaultConfig() Config {
	return defaultConfig()
}

// isConfiguredCredential reports whether a client id or secret is usable.
// Empty values and the historical REPLACE_ME_* placeholders are treated as
// unconfigured so authorize redirects never leave the product with a fake
// client_id.
func isConfiguredCredential(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return false
	}
	upper := strings.ToUpper(trimmed)
	return !strings.HasPrefix(upper, "REPLACE_ME")
}

// sanitizeReturnTo accepts only same-origin relative SPA paths that the login
// entry button may echo back for error/success redirects. Absolute URLs and
// protocol-relative paths are rejected to prevent open redirects.
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

// SanitizeReturnTo is the exported form of sanitizeReturnTo for HTTP controllers.
func SanitizeReturnTo(raw string) string {
	return sanitizeReturnTo(raw)
}
