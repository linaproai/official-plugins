// Package oauth implements generic OIDC login orchestration for
// linapro-oidc-generic: discovery, authorize (PKCE), callback id_token verify,
// and host external-login exchange with handoff.
package oauth

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/frame/g"

	"lina-core/pkg/plugin/capability/authcap/extlogin"

	settingssvc "lina-plugin-linapro-oidc-generic/backend/internal/service/settings"
)

// Provider is the stable external-identity provider ID owned by this plugin.
// v1 uses a fixed connection key "default".
const Provider = "oidc:default"

// ConnectionKey is the v1 fixed connection key embedded in Provider.
const ConnectionKey = settingssvc.ConnectionKeyDefault

// Service defines the generic OIDC login orchestration contract.
type Service interface {
	BuildAuthorizeURL(ctx context.Context, stateKey string, returnTo string) (AuthorizeRequest, error)
	CompleteCallback(ctx context.Context, in CallbackInput) (*CallbackOutput, error)
	LoginReturnPath() string
}

// AuthorizeRequest describes the authorize redirect the browser should follow.
type AuthorizeRequest struct {
	URL   string
	State string
}

// CallbackInput carries raw callback values from the IdP.
type CallbackInput struct {
	Code  string
	State string
}

// CallbackOutput carries the host login outcome for SPA redirect.
type CallbackOutput struct {
	Handoff          string
	AccessToken      string
	RefreshToken     string
	PreToken         string
	TenantCandidates []extlogin.TenantCandidate
	StateKey         string
	ReturnTo         string
}

var _ Service = (*serviceImpl)(nil)

type serviceImpl struct {
	externalLoginSvc extlogin.Service
	configResolver   *ConfigResolver
	stateCodec       StateCodec
	discovery        *discoveryCache
	jwks             *jwksClient
}

// New creates the generic OIDC login service.
func New(
	externalLoginSvc extlogin.Service,
	configResolver *ConfigResolver,
	stateCodec StateCodec,
) Service {
	return &serviceImpl{
		externalLoginSvc: externalLoginSvc,
		configResolver:   configResolver,
		stateCodec:       stateCodec,
		discovery:        newDiscoveryCache(),
		jwks:             newJWKSClient(),
	}
}

func (s *serviceImpl) resolveConfig(ctx context.Context) Config {
	if s == nil || s.configResolver == nil {
		return Config{}
	}
	return s.configResolver.ResolveConfig(ctx)
}

// ConfigResolver layers persisted settings over static defaults per request.
type ConfigResolver struct {
	settingsSvc settingssvc.Service
	config      Config
}

// NewConfigResolver creates the shared request-time configuration resolver.
func NewConfigResolver(settingsSvc settingssvc.Service, config Config) *ConfigResolver {
	return &ConfigResolver{settingsSvc: settingsSvc, config: config}
}

// ResolveConfig layers persisted settings over static defaults.
func (r *ConfigResolver) ResolveConfig(ctx context.Context) Config {
	if r == nil {
		return Config{}
	}
	resolved := r.config
	if r.settingsSvc != nil {
		snapshot, err := r.settingsSvc.Load(ctx)
		if err == nil && snapshot != nil {
			if snapshot.Issuer != "" {
				resolved.Issuer = normalizeIssuer(snapshot.Issuer)
			}
			if snapshot.ClientID != "" {
				resolved.ClientID = snapshot.ClientID
			}
			if snapshot.ClientSecret != "" {
				resolved.ClientSecret = snapshot.ClientSecret
			}
			if snapshot.RedirectURL != "" {
				resolved.RedirectURL = snapshot.RedirectURL
			}
			if snapshot.DisplayName != "" {
				resolved.DisplayName = snapshot.DisplayName
			}
			resolved.Scopes = normalizeScopes(snapshot.Scopes, resolved.Scopes)
		}
	}
	if len(resolved.Scopes) == 0 {
		resolved.Scopes = normalizeScopes("", defaultConfig().Scopes)
	}
	if resolved.RedirectURL == "" {
		resolved.RedirectURL = deriveCallbackURL(ctx)
	}
	return resolved
}

func deriveCallbackURL(ctx context.Context) string {
	request := g.RequestFromCtx(ctx)
	if request == nil || request.Host == "" {
		return ""
	}
	scheme := "http"
	if request.TLS != nil || request.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	return scheme + "://" + request.Host + "/portal/linapro-oidc-generic/callback"
}

// IsLoginConfigured reports whether issuer and credentials are usable.
func IsLoginConfigured(cfg Config) bool {
	return normalizeIssuer(cfg.Issuer) != "" &&
		isConfiguredCredential(cfg.ClientID) &&
		isConfiguredCredential(cfg.ClientSecret) &&
		strings.TrimSpace(cfg.RedirectURL) != ""
}
