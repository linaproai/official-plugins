// oauth_config.go defines the plugin-owned Google OIDC client configuration
// used by the reference login flow. The configuration is intentionally kept
// as a plain value struct so it can be sourced from the host plugin config
// section, a plugin manifest file, or a startup override without pulling in
// runtime dependencies here.

package oauth

// Config carries the plugin-owned Google OIDC client credentials and the
// OAuth endpoints the plugin talks to. Real deployments load this from the
// plugin configuration source chain; the reference implementation ships with
// placeholder values so the surrounding integration flow can be exercised
// end-to-end without a live Google project.
type Config struct {
	// ClientID is the Google OAuth 2.0 client ID issued by Google Cloud.
	ClientID string
	// ClientSecret is the Google OAuth 2.0 client secret paired with ClientID.
	ClientSecret string
	// RedirectURL is the callback URL registered with Google. It must resolve
	// to the plugin callback route so Google routes the browser back through
	// the plugin's own controller.
	RedirectURL string
	// AuthorizeURL is the Google OAuth authorize endpoint.
	AuthorizeURL string
	// TokenURL is the Google OAuth token endpoint.
	TokenURL string
	// UserInfoURL is the Google OIDC userinfo endpoint used to project the
	// verified identity into a stable subject, email, and display name.
	UserInfoURL string
	// Scopes are the OAuth scopes the plugin requests. The reference flow
	// requests openid, email, and profile so it can build the verified
	// identity DTO the host expects.
	Scopes []string
	// LoginReturnPath is the SPA path the callback controller redirects to
	// after the host external-login exchange completes. The host issues the
	// tokens or pre-token that must reach the SPA.
	LoginReturnPath string
}

// defaultConfig returns the reference-implementation configuration. The
// endpoints are the well-known Google OAuth 2.0 URLs; the client credentials
// are intentionally placeholder values that a real deployment must override
// through the plugin configuration source chain.
func defaultConfig() Config {
	return Config{
		ClientID:     "REPLACE_ME_GOOGLE_CLIENT_ID",
		ClientSecret: "REPLACE_ME_GOOGLE_CLIENT_SECRET",
		// RedirectURL is intentionally empty: the OAuth service derives the
		// callback URL from the live request host unless the admin overrides
		// it through the settings page.
		RedirectURL:  "",
		AuthorizeURL: "https://accounts.google.com/o/oauth2/v2/auth",
		TokenURL:     "https://oauth2.googleapis.com/token",
		UserInfoURL:  "https://openidconnect.googleapis.com/v1/userinfo",
		Scopes:       []string{"openid", "email", "profile"},
		// The default workspace serves the SPA under /admin/ with a hash
		// router, so the callback outcome query must live inside the hash
		// fragment for vue-router to expose it through route.query.
		LoginReturnPath: "/admin/#/auth/login",
	}
}

// DefaultConfig returns one exported copy of the reference-implementation
// configuration so wiring code can start from the reference values and
// override individual fields as needed.
func DefaultConfig() Config {
	return defaultConfig()
}
