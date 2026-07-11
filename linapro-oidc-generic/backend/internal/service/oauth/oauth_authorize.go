// oauth_authorize.go builds the OIDC authorize URL with PKCE and signed state.

package oauth

import (
	"context"
	"net/url"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/bizerr"
)

// BuildAuthorizeURL prepares one IdP authorize redirect with PKCE S256.
func (s *serviceImpl) BuildAuthorizeURL(ctx context.Context, stateKey string, returnTo string) (AuthorizeRequest, error) {
	config := s.resolveConfig(ctx)
	if !IsLoginConfigured(config) {
		return AuthorizeRequest{}, bizerr.WrapCode(
			gerror.New("oidc generic: authorize configuration missing"),
			CodeConfigMissing,
		)
	}
	if s.stateCodec == nil || s.discovery == nil {
		return AuthorizeRequest{}, bizerr.WrapCode(
			gerror.New("oidc generic: authorize dependencies missing"),
			CodeStateGenerateFailed,
		)
	}
	doc, err := s.discovery.resolve(ctx, config.Issuer)
	if err != nil {
		return AuthorizeRequest{}, err
	}
	codeVerifier, err := newPKCEVerifier()
	if err != nil {
		return AuthorizeRequest{}, bizerr.WrapCode(gerror.Wrap(err, "oidc generic: pkce verifier"), CodeStateGenerateFailed)
	}
	oidcNonce, err := newOIDCNonce()
	if err != nil {
		return AuthorizeRequest{}, bizerr.WrapCode(gerror.Wrap(err, "oidc generic: nonce"), CodeStateGenerateFailed)
	}
	state, err := s.stateCodec.Encode(ctx, stateKey, returnTo, config.ClientSecret, codeVerifier, oidcNonce)
	if err != nil {
		return AuthorizeRequest{}, err
	}
	authorizeURL, err := url.Parse(doc.AuthorizationEndpoint)
	if err != nil {
		return AuthorizeRequest{}, bizerr.WrapCode(
			gerror.Wrap(err, "oidc generic: authorize URL invalid"),
			CodeConfigMissing,
		)
	}
	query := authorizeURL.Query()
	query.Set("client_id", config.ClientID)
	query.Set("redirect_uri", config.RedirectURL)
	query.Set("response_type", "code")
	query.Set("scope", strings.Join(config.Scopes, " "))
	query.Set("state", state)
	query.Set("nonce", oidcNonce)
	query.Set("code_challenge", pkceChallengeS256(codeVerifier))
	query.Set("code_challenge_method", "S256")
	authorizeURL.RawQuery = query.Encode()
	return AuthorizeRequest{
		URL:   authorizeURL.String(),
		State: state,
	}, nil
}
