// oauth_authorize.go implements the Google authorize URL construction step
// of the Google OIDC login flow. The controller invokes BuildAuthorizeURL and
// then issues an HTTP 302 redirect using the returned URL. The anti-CSRF
// state is a self-contained HMAC-signed token, so no cookie is required to
// survive the round trip through Google.

package oauth

import (
	"context"
	"net/url"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/bizerr"
)

// BuildAuthorizeURL prepares one Google authorize redirect. The optional
// stateKey is embedded into the signed state token so the callback can match
// it against the configured SSO delivery rules. returnTo is also embedded so
// the callback can restore the originating SPA login page.
func (s *serviceImpl) BuildAuthorizeURL(ctx context.Context, stateKey string, returnTo string) (AuthorizeRequest, error) {
	// Re-resolve settings per request so admin edits apply without a restart.
	config := s.resolveConfig(ctx)
	if !isConfiguredCredential(config.ClientID) || !isConfiguredCredential(config.ClientSecret) ||
		strings.TrimSpace(config.RedirectURL) == "" || strings.TrimSpace(config.AuthorizeURL) == "" {
		return AuthorizeRequest{}, bizerr.WrapCode(
			gerror.New("oidc google: authorize configuration missing"),
			CodeConfigMissing,
		)
	}
	if s.stateCodec == nil {
		return AuthorizeRequest{}, bizerr.WrapCode(
			gerror.New("oidc google: state codec is missing"),
			CodeStateGenerateFailed,
		)
	}
	state, err := s.stateCodec.Encode(ctx, stateKey, returnTo, config.ClientSecret)
	if err != nil {
		return AuthorizeRequest{}, err
	}
	authorizeURL, err := url.Parse(config.AuthorizeURL)
	if err != nil {
		return AuthorizeRequest{}, bizerr.WrapCode(
			gerror.Wrap(err, "oidc google: authorize URL is not a valid URL"),
			CodeConfigMissing,
		)
	}
	query := authorizeURL.Query()
	query.Set("client_id", config.ClientID)
	query.Set("redirect_uri", config.RedirectURL)
	query.Set("response_type", "code")
	query.Set("scope", strings.Join(config.Scopes, " "))
	query.Set("state", state)
	query.Set("access_type", "online")
	query.Set("prompt", "select_account")
	authorizeURL.RawQuery = query.Encode()
	return AuthorizeRequest{
		URL:   authorizeURL.String(),
		State: state,
	}, nil
}
