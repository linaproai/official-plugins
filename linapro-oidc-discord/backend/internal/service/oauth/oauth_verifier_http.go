// oauth_verifier_http.go implements the production identity verifier: it
// exchanges the OAuth authorization code against Discord's token endpoint and
// fetches the verified identity from Discord's /users/@me endpoint. Outbound
// requests are bounded by a timeout, error bodies are truncated before
// logging so tokens never leak, and unverified email addresses are rejected
// before the identity reaches the host external-login seam.

package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
)

// httpTimeout bounds each outbound HTTP request to Discord.
const httpTimeout = 15 * time.Second

// errorBodyLimit caps how much of an error response body is embedded into
// error messages so large payloads and tokens never leak into logs.
const errorBodyLimit = 512

// SettingsSource supplies the current OAuth client configuration at request
// time so admin edits apply without a restart.
type SettingsSource interface {
	// ResolveConfig returns the effective OAuth configuration for one request.
	ResolveConfig(ctx context.Context) Config
}

// httpIdentityVerifier is the production verifier backed by Discord's token
// and user endpoints.
type httpIdentityVerifier struct {
	// httpClient performs outbound requests to Discord with a bounded timeout.
	httpClient *http.Client
	// settings supplies client credentials and endpoints per request.
	settings SettingsSource
}

// NewHTTPIdentityVerifier returns the production identity verifier that
// performs the real Discord code exchange and user fetch. The settings source
// is consulted on every Verify call so credential rotations apply without
// restarting the host.
func NewHTTPIdentityVerifier(settings SettingsSource) IdentityVerifier {
	return &httpIdentityVerifier{
		httpClient: &http.Client{Timeout: httpTimeout},
		settings:   settings,
	}
}

// Verify exchanges the authorization code for a Discord access token, fetches
// the user claims, enforces email verification, and projects the result into
// the neutral VerifiedIdentity shape.
func (v *httpIdentityVerifier) Verify(ctx context.Context, code string, redirectURL string) (*VerifiedIdentity, error) {
	trimmedCode := strings.TrimSpace(code)
	if trimmedCode == "" {
		return nil, bizerr.WrapCode(gerror.New("oidc discord: empty authorization code"), CodeIdentityVerifyFailed)
	}
	if v == nil || v.settings == nil {
		return nil, bizerr.WrapCode(gerror.New("oidc discord: verifier settings source is missing"), CodeConfigMissing)
	}
	config := v.settings.ResolveConfig(ctx)
	accessToken, err := v.exchangeCode(ctx, config, redirectURL, trimmedCode)
	if err != nil {
		return nil, err
	}
	return v.fetchUserIdentity(ctx, config, accessToken)
}

// exchangeCode trades one authorization code for a Discord access token.
func (v *httpIdentityVerifier) exchangeCode(ctx context.Context, config Config, redirectURL string, code string) (string, error) {
	clientID := strings.TrimSpace(config.ClientID)
	clientSecret := strings.TrimSpace(config.ClientSecret)
	if clientID == "" || clientSecret == "" {
		return "", bizerr.WrapCode(gerror.New("oidc discord: client credentials are not configured"), CodeConfigMissing)
	}
	resolvedRedirect := strings.TrimSpace(redirectURL)
	if resolvedRedirect == "" {
		resolvedRedirect = strings.TrimSpace(config.RedirectURL)
	}
	if resolvedRedirect == "" {
		return "", bizerr.WrapCode(gerror.New("oidc discord: redirect url is not configured"), CodeConfigMissing)
	}
	form := url.Values{}
	form.Set("code", code)
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)
	form.Set("redirect_uri", resolvedRedirect)
	form.Set("grant_type", "authorization_code")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, config.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", bizerr.WrapCode(gerror.Wrap(err, "oidc discord: build token request failed"), CodeIdentityVerifyFailed)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := v.httpClient.Do(req)
	if err != nil {
		return "", bizerr.WrapCode(gerror.Wrap(err, "oidc discord: call token endpoint failed"), CodeIdentityVerifyFailed)
	}
	defer closeResponseBody(ctx, resp, "discord oauth token endpoint")
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", bizerr.WrapCode(gerror.Wrap(err, "oidc discord: read token response failed"), CodeIdentityVerifyFailed)
	}
	if resp.StatusCode != http.StatusOK {
		return "", bizerr.WrapCode(
			gerror.Newf("oidc discord: token endpoint returned status %d: %s", resp.StatusCode, truncate(string(body), errorBodyLimit)),
			CodeIdentityVerifyFailed,
		)
	}
	var tokenResp struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
		ErrorDesc   string `json:"error_description"`
	}
	if err = json.Unmarshal(body, &tokenResp); err != nil {
		return "", bizerr.WrapCode(gerror.Wrap(err, "oidc discord: decode token response failed"), CodeIdentityVerifyFailed)
	}
	if tokenResp.Error != "" {
		return "", bizerr.WrapCode(
			gerror.Newf("oidc discord: token endpoint error: %s: %s", tokenResp.Error, tokenResp.ErrorDesc),
			CodeIdentityVerifyFailed,
		)
	}
	if strings.TrimSpace(tokenResp.AccessToken) == "" {
		return "", bizerr.WrapCode(gerror.New("oidc discord: token endpoint returned empty access token"), CodeIdentityVerifyFailed)
	}
	return tokenResp.AccessToken, nil
}

// fetchUserIdentity retrieves the Discord user claims and enforces the
// verified email requirement before projecting the identity.
func (v *httpIdentityVerifier) fetchUserIdentity(ctx context.Context, config Config, accessToken string) (*VerifiedIdentity, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, config.UserInfoURL, nil)
	if err != nil {
		return nil, bizerr.WrapCode(gerror.Wrap(err, "oidc discord: build user request failed"), CodeIdentityVerifyFailed)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := v.httpClient.Do(req)
	if err != nil {
		return nil, bizerr.WrapCode(gerror.Wrap(err, "oidc discord: call user endpoint failed"), CodeIdentityVerifyFailed)
	}
	defer closeResponseBody(ctx, resp, "discord oauth user endpoint")
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, bizerr.WrapCode(gerror.Wrap(err, "oidc discord: read user response failed"), CodeIdentityVerifyFailed)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, bizerr.WrapCode(
			gerror.Newf("oidc discord: user endpoint returned status %d: %s", resp.StatusCode, truncate(string(body), errorBodyLimit)),
			CodeIdentityVerifyFailed,
		)
	}
	var user struct {
		ID         string `json:"id"`
		Username   string `json:"username"`
		GlobalName string `json:"global_name"`
		Email      string `json:"email"`
		Verified   bool   `json:"verified"`
	}
	if err = json.Unmarshal(body, &user); err != nil {
		return nil, bizerr.WrapCode(gerror.Wrap(err, "oidc discord: decode user response failed"), CodeIdentityVerifyFailed)
	}
	if strings.TrimSpace(user.ID) == "" || strings.TrimSpace(user.Email) == "" {
		return nil, bizerr.WrapCode(gerror.New("oidc discord: user response missing id or email"), CodeIdentityVerifyFailed)
	}
	if !user.Verified {
		return nil, bizerr.WrapCode(
			gerror.New("oidc discord: account email is not verified by Discord"),
			CodeEmailNotVerified,
		)
	}
	displayName := strings.TrimSpace(user.GlobalName)
	if displayName == "" {
		displayName = user.Username
	}
	return &VerifiedIdentity{
		Subject:     user.ID,
		Email:       user.Email,
		DisplayName: displayName,
	}, nil
}

// SetHTTPClient overrides the outbound HTTP client. Intended for tests that
// need to intercept Discord traffic.
func (v *httpIdentityVerifier) SetHTTPClient(client *http.Client) {
	if client != nil {
		v.httpClient = client
	}
}

// truncate trims long strings used in error messages so secrets and large
// payloads do not leak into logs.
func truncate(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return fmt.Sprintf("%s...(truncated)", value[:limit])
}

// closeResponseBody closes resp.Body once the caller has finished reading it
// and logs any close-time error against the supplied context so error returns
// are never silently discarded.
func closeResponseBody(ctx context.Context, resp *http.Response, endpointLabel string) {
	if resp == nil || resp.Body == nil {
		return
	}
	if err := resp.Body.Close(); err != nil {
		logger.Warningf(ctx, "%s close response body failed err=%v", endpointLabel, err)
	}
}
