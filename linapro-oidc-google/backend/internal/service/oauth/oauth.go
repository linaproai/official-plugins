// Package oauth implements the Google OIDC login orchestration for the
// linapro-oidc-google reference plugin. It builds the Google authorize URL,
// consumes callback code and state values, exchanges the code for a verified
// identity, and hands the verified identity off to the host external-login
// seam so the host can complete provisioning, tenant selection, and token
// minting under its own governance.
package oauth

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	"lina-core/pkg/plugin/capability/authcap/extlogin"

	settingssvc "lina-plugin-linapro-oidc-google/backend/internal/service/settings"
)

// Provider is the stable external-identity provider ID owned by this plugin.
// The host validates that external-login requests only pass this provider ID
// when the caller is the linapro-oidc-google plugin.
const Provider = "google"

// Service defines the Google OIDC login orchestration contract. The service
// separates the authorize URL construction and callback flow so the HTTP
// controller only forwards raw request values and never talks to the host
// external-login seam directly.
type Service interface {
	// BuildAuthorizeURL prepares one authorize URL that the browser should
	// redirect to when the user clicks the "Continue with Google" button. The
	// optional stateKey is embedded into the signed state token so the callback
	// can match it against the configured SSO delivery rules without relying
	// on browser cookies. returnTo is the sanitized SPA path echoed back after
	// the round trip so history-mode and hash-mode login pages both recover.
	BuildAuthorizeURL(ctx context.Context, stateKey string, returnTo string) (AuthorizeRequest, error)
	// CompleteCallback validates the signed callback state, exchanges the OAuth
	// code for a verified identity, calls the host external-login seam, and
	// returns one login outcome the controller renders as a redirect.
	CompleteCallback(ctx context.Context, in CallbackInput) (*CallbackOutput, error)
	// LoginReturnPath returns the SPA path the callback controller redirects to
	// after the host external-login exchange completes. It reflects the
	// configured LoginReturnPath value so wiring stays in one place.
	LoginReturnPath() string
	// CompleteOneTap validates one Google One Tap ID Token credential and
	// exchanges the verified identity for a host login outcome. The stateKey
	// comes from the embed's login_uri query and selects the SSO delivery
	// rule; the ID Token's own signature, audience, and expiry provide the
	// integrity guarantees that the authorize-code flow gets from the signed
	// state token.
	CompleteOneTap(ctx context.Context, credential string, stateKey string) (*CallbackOutput, error)
}

// AuthorizeRequest describes the authorize redirect the browser should follow.
type AuthorizeRequest struct {
	// URL is the fully-qualified Google authorize URL, including client_id,
	// redirect_uri, scope, response_type, and state parameters.
	URL string
	// State is the self-contained signed state token embedded in the URL. It
	// is returned separately for audit logging only; no cookie is required.
	State string
}

// CallbackInput carries the raw values returned by the Google callback.
type CallbackInput struct {
	// Code is the OAuth authorization code that the plugin exchanges for a
	// verified identity.
	Code string
	// State is the signed state token Google echoed back. It is validated by
	// signature and expiry; no stored copy is needed.
	State string
}

// CallbackOutput carries the host external-login outcome projected into the
// shape the controller needs to build the SPA redirect.
type CallbackOutput struct {
	// Handoff is the single-use host code the SPA exchanges for tokens. SPA
	// redirects MUST use this field instead of putting JWTs in the URL.
	Handoff string
	// AccessToken is retained only for optional admin-configured SSO receiver
	// delivery; it MUST NOT be placed on the SPA login redirect.
	AccessToken string
	// RefreshToken pairs with AccessToken for SSO receiver delivery only.
	RefreshToken string
	// PreToken is retained for multi-tenant SSO edge cases; SPA redirects use Handoff.
	PreToken string
	// TenantCandidates lists the tenants the user may select from when
	// multi-tenant selection is required (also embedded in handoff payload).
	TenantCandidates []extlogin.TenantCandidate
	// StateKey is the business state key recovered from the signed state
	// token; the controller matches it against the SSO delivery rules.
	StateKey string
	// ReturnTo is the sanitized SPA path recovered from the signed state
	// token so the callback can redirect back to the page that started login.
	ReturnTo string
}

// Interface compliance assertion for the default OIDC login service implementation.
var _ Service = (*serviceImpl)(nil)

// serviceImpl orchestrates the Google OIDC login for the reference plugin.
// External-login is the host runtime dependency: the host owns provisioning,
// tenant candidate resolution, and token minting once the plugin submits a
// verified identity. The settings service is re-read on every OAuth request
// so admin edits from the settings page take effect without a restart.
type serviceImpl struct {
	externalLoginSvc extlogin.Service // externalLoginSvc completes host session issuance for verified identities.
	configResolver   *ConfigResolver       // configResolver layers persisted settings over static defaults per request.
	verifier         IdentityVerifier      // verifier exchanges an OAuth code for a verified identity.
	stateCodec       StateCodec            // stateCodec signs and validates self-contained state tokens.
	idTokenVerifier  IDTokenVerifier       // idTokenVerifier validates One Tap ID Token credentials; nil disables One Tap.
}

// New creates and returns a new Google OIDC login service instance. Each
// dependency is a narrow contract so the reference implementation can wire a
// production HTTP verifier at real deployment time while unit tests inject a
// deterministic verifier. The shared configResolver is also consumed by the
// HTTP identity verifier so both read identical request-time settings.
// idTokenVerifier may be nil, which disables the One Tap endpoint fail-closed.
func New(
	externalLoginSvc extlogin.Service,
	configResolver *ConfigResolver,
	verifier IdentityVerifier,
	stateCodec StateCodec,
	idTokenVerifier IDTokenVerifier,
) Service {
	return &serviceImpl{
		externalLoginSvc: externalLoginSvc,
		configResolver:   configResolver,
		verifier:         verifier,
		stateCodec:       stateCodec,
		idTokenVerifier:  idTokenVerifier,
	}
}

// resolveConfig returns the effective request-time configuration through the
// shared resolver, degrading to the zero config when wiring is incomplete.
func (s *serviceImpl) resolveConfig(ctx context.Context) Config {
	if s == nil || s.configResolver == nil {
		return Config{}
	}
	return s.configResolver.ResolveConfig(ctx)
}

// ConfigResolver layers the persisted admin settings on top of the static
// defaults per request so settings edits take effect immediately. It is
// shared between the login orchestration and the HTTP identity verifier so
// both read identical request-time settings.
type ConfigResolver struct {
	settingsSvc settingssvc.Service // settingsSvc supplies persisted admin settings; nil keeps static config only.
	config      Config              // config carries the static defaults layered under persisted settings.
}

// NewConfigResolver creates the shared request-time configuration resolver.
func NewConfigResolver(settingsSvc settingssvc.Service, config Config) *ConfigResolver {
	return &ConfigResolver{settingsSvc: settingsSvc, config: config}
}

// ResolveConfig layers persisted settings over the static defaults; persisted
// empty fields keep the static default, and an unset redirect URL is derived
// from the live request host.
func (r *ConfigResolver) ResolveConfig(ctx context.Context) Config {
	if r == nil {
		return Config{}
	}
	resolved := r.config
	if r.settingsSvc != nil {
		snapshot, err := r.settingsSvc.Load(ctx)
		if err == nil && snapshot != nil {
			if snapshot.ClientID != "" {
				resolved.ClientID = snapshot.ClientID
			}
			if snapshot.ClientSecret != "" {
				resolved.ClientSecret = snapshot.ClientSecret
			}
			if snapshot.RedirectURL != "" {
				resolved.RedirectURL = snapshot.RedirectURL
			}
		}
		// A load failure keeps the static defaults so a temporarily
		// unavailable settings store does not take the login entry down; the
		// settings service already logs the underlying failure.
	}
	if resolved.RedirectURL == "" {
		resolved.RedirectURL = deriveCallbackURL(ctx)
	}
	return resolved
}

// deriveCallbackURL builds the plugin callback URL from the current request
// host so deployments do not need to configure the redirect URL manually. The
// callback path is fixed by the plugin's portal route registration.
func deriveCallbackURL(ctx context.Context) string {
	request := g.RequestFromCtx(ctx)
	if request == nil || request.Host == "" {
		return ""
	}
	scheme := "http"
	if request.TLS != nil || request.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	return scheme + "://" + request.Host + "/portal/linapro-oidc-google/callback"
}
