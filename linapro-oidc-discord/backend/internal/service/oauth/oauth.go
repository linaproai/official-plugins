// Package oauth implements the Discord OIDC login orchestration for the
// linapro-oidc-discord reference plugin. It builds the Discord authorize URL,
// consumes callback code and state values, exchanges the code for a verified
// identity, and hands the verified identity off to the host external-login
// seam so the host can complete provisioning, tenant selection, and token
// minting under its own governance.
package oauth

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	"lina-core/pkg/plugin/capability/authcap/externallogin"

	settingssvc "lina-plugin-linapro-oidc-discord/backend/internal/service/settings"
)

// Provider is the stable external-identity provider ID owned by this plugin.
// The host validates that external-login requests only pass this provider ID
// when the caller is the linapro-oidc-discord plugin.
const Provider = "discord"

// Service defines the Discord OIDC login orchestration contract. The service
// separates the authorize URL construction and callback flow so the HTTP
// controller only forwards raw request values and never talks to the host
// external-login seam directly.
type Service interface {
	// BuildAuthorizeURL prepares one authorize URL that the browser should
	// redirect to when the user clicks the "Continue with Discord" button. The
	// optional stateKey is embedded into the signed state token so the callback
	// can match it against the configured SSO delivery rules without relying
	// on browser cookies.
	BuildAuthorizeURL(ctx context.Context, stateKey string) (AuthorizeRequest, error)
	// CompleteCallback validates the signed callback state, exchanges the OAuth
	// code for a verified identity, calls the host external-login seam, and
	// returns one login outcome the controller renders as a redirect.
	CompleteCallback(ctx context.Context, in CallbackInput) (*CallbackOutput, error)
	// LoginReturnPath returns the SPA path the callback controller redirects to
	// after the host external-login exchange completes. It reflects the
	// configured LoginReturnPath value so wiring stays in one place.
	LoginReturnPath() string
}

// AuthorizeRequest describes the authorize redirect the browser should follow.
type AuthorizeRequest struct {
	// URL is the fully-qualified Discord authorize URL, including client_id,
	// redirect_uri, scope, response_type, and state parameters.
	URL string
	// State is the self-contained signed state token embedded in the URL. It
	// is returned separately for audit logging only; no cookie is required.
	State string
}

// CallbackInput carries the raw values returned by the Discord callback.
type CallbackInput struct {
	// Code is the OAuth authorization code that the plugin exchanges for a
	// verified identity.
	Code string
	// State is the signed state token Discord echoed back. It is validated by
	// signature and expiry; no stored copy is needed.
	State string
}

// CallbackOutput carries the host external-login outcome projected into the
// shape the controller needs to build the SPA redirect.
type CallbackOutput struct {
	// AccessToken is set when the resolved user has zero or one active tenant.
	AccessToken string
	// RefreshToken is set together with AccessToken for single-tenant users.
	RefreshToken string
	// PreToken is set when the resolved user has more than one active tenant
	// and the SPA must complete tenant selection through the host auth API.
	PreToken string
	// TenantCandidates lists the tenants the user may select from when
	// PreToken is set. The controller may serialize it into query parameters
	// or defer the display to the SPA.
	TenantCandidates []externallogin.TenantCandidate
	// StateKey is the business state key recovered from the signed state
	// token; the controller matches it against the SSO delivery rules.
	StateKey string
}

// Interface compliance assertion for the default OIDC login service implementation.
var _ Service = (*serviceImpl)(nil)

// serviceImpl orchestrates the Discord OIDC login for the reference plugin.
// External-login is the host runtime dependency: the host owns provisioning,
// tenant candidate resolution, and token minting once the plugin submits a
// verified identity. The settings are re-resolved on every OAuth request so
// admin edits from the settings page take effect without a restart.
type serviceImpl struct {
	externalLoginSvc externallogin.Service // externalLoginSvc completes host session issuance for verified identities.
	configResolver   *ConfigResolver       // configResolver layers persisted settings over static defaults per request.
	verifier         IdentityVerifier      // verifier exchanges an OAuth code for a verified identity.
	stateCodec       StateCodec            // stateCodec signs and validates self-contained state tokens.
}

// New creates and returns a new Discord OIDC login service instance. Each
// dependency is a narrow contract so the reference implementation can wire a
// production HTTP verifier at real deployment time while unit tests inject a
// deterministic verifier. The shared configResolver is also consumed by the
// HTTP identity verifier so both read identical request-time settings.
func New(
	externalLoginSvc externallogin.Service,
	configResolver *ConfigResolver,
	verifier IdentityVerifier,
	stateCodec StateCodec,
) Service {
	return &serviceImpl{
		externalLoginSvc: externalLoginSvc,
		configResolver:   configResolver,
		verifier:         verifier,
		stateCodec:       stateCodec,
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
	return scheme + "://" + request.Host + "/portal/linapro-oidc-discord/callback"
}
