// oauth_verifier_http.go implements the production identity verifier: it
// exchanges the OAuth authorization code against Google's token endpoint and
// fetches the verified identity from Google's OpenID Connect userinfo
// endpoint. Outbound requests are bounded by a timeout, error bodies are
// truncated before logging so tokens never leak, and unverified email
// addresses are rejected before the identity reaches the host external-login
// seam.

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

// httpTimeout bounds each outbound HTTP request to Google.
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

// httpIdentityVerifier is the production verifier backed by Google's token
// and userinfo endpoints.
type httpIdentityVerifier struct {
	// httpClient performs outbound requests to Google with a bounded timeout.
	httpClient *http.Client
	// settings supplies client credentials and endpoints per request.
	settings SettingsSource
}

// NewHTTPIdentityVerifier returns the production identity verifier that
// performs the real Google code exchange and userinfo fetch. The settings
// source is consulted on every Verify call so credential rotations apply
// without restarting the host.
func NewHTTPIdentityVerifier(settings SettingsSource) IdentityVerifier {
	return &httpIdentityVerifier{
		httpClient: &http.Client{Timeout: httpTimeout},
		settings:   settings,
	}
}

// Verify exchanges the authorization code for a Google access token, fetches
// the userinfo claims, enforces email verification, and projects the result
// into the neutral VerifiedIdentity shape.
func (v *httpIdentityVerifier) Verify(ctx context.Context, code string, redirectURL string) (*VerifiedIdentity, error) {
	trimmedCode := strings.TrimSpace(code)
	if trimmedCode == "" {
		return nil, bizerr.WrapCode(gerror.New("oidc google: empty authorization code"), CodeIdentityVerifyFailed)
	}
	if v == nil || v.settings == nil {
		return nil, bizerr.WrapCode(gerror.New("oidc google: verifier settings source is missing"), CodeConfigMissing)
	}
	config := v.settings.ResolveConfig(ctx)
	accessToken, err := v.exchangeCode(ctx, config, redirectURL, trimmedCode)
	if err != nil {
		return nil, err
	}
	return v.fetchUserIdentity(ctx, config, accessToken)
}

// exchangeCode trades one authorization code for a Google access token.
func (v *httpIdentityVerifier) exchangeCode(ctx context.Context, config Config, redirectURL string, code string) (string, error) {
	clientID := strings.TrimSpace(config.ClientID)
	clientSecret := strings.TrimSpace(config.ClientSecret)
	if !isConfiguredCredential(clientID) || !isConfiguredCredential(clientSecret) {
		return "", bizerr.WrapCode(gerror.New("oidc google: client credentials are not configured"), CodeConfigMissing)
	}
	resolvedRedirect := strings.TrimSpace(redirectURL)
	if resolvedRedirect == "" {
		resolvedRedirect = strings.TrimSpace(config.RedirectURL)
	}
	if resolvedRedirect == "" {
		return "", bizerr.WrapCode(gerror.New("oidc google: redirect url is not configured"), CodeConfigMissing)
	}
	form := url.Values{}
	form.Set("code", code)
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)
	form.Set("redirect_uri", resolvedRedirect)
	form.Set("grant_type", "authorization_code")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, config.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", bizerr.WrapCode(gerror.Wrap(err, "oidc google: build token request failed"), CodeIdentityVerifyFailed)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := v.httpClient.Do(req)
	if err != nil {
		return "", bizerr.WrapCode(gerror.Wrap(err, "oidc google: call token endpoint failed"), CodeIdentityVerifyFailed)
	}
	defer closeResponseBody(ctx, resp, "google oauth token endpoint")
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", bizerr.WrapCode(gerror.Wrap(err, "oidc google: read token response failed"), CodeIdentityVerifyFailed)
	}
	if resp.StatusCode != http.StatusOK {
		return "", bizerr.WrapCode(
			gerror.Newf("oidc google: token endpoint returned status %d: %s", resp.StatusCode, truncate(string(body), errorBodyLimit)),
			CodeIdentityVerifyFailed,
		)
	}
	var tokenResp struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
		ErrorDesc   string `json:"error_description"`
	}
	if err = json.Unmarshal(body, &tokenResp); err != nil {
		return "", bizerr.WrapCode(gerror.Wrap(err, "oidc google: decode token response failed"), CodeIdentityVerifyFailed)
	}
	if tokenResp.Error != "" {
		return "", bizerr.WrapCode(
			gerror.Newf("oidc google: token endpoint error: %s: %s", tokenResp.Error, tokenResp.ErrorDesc),
			CodeIdentityVerifyFailed,
		)
	}
	if strings.TrimSpace(tokenResp.AccessToken) == "" {
		return "", bizerr.WrapCode(gerror.New("oidc google: token endpoint returned empty access token"), CodeIdentityVerifyFailed)
	}
	return tokenResp.AccessToken, nil
}

// fetchUserIdentity retrieves the userinfo claims and enforces the verified
// email requirement before projecting the identity.
func (v *httpIdentityVerifier) fetchUserIdentity(ctx context.Context, config Config, accessToken string) (*VerifiedIdentity, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, config.UserInfoURL, nil)
	if err != nil {
		return nil, bizerr.WrapCode(gerror.Wrap(err, "oidc google: build userinfo request failed"), CodeIdentityVerifyFailed)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := v.httpClient.Do(req)
	if err != nil {
		return nil, bizerr.WrapCode(gerror.Wrap(err, "oidc google: call userinfo endpoint failed"), CodeIdentityVerifyFailed)
	}
	defer closeResponseBody(ctx, resp, "google oauth userinfo endpoint")
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, bizerr.WrapCode(gerror.Wrap(err, "oidc google: read userinfo response failed"), CodeIdentityVerifyFailed)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, bizerr.WrapCode(
			gerror.Newf("oidc google: userinfo endpoint returned status %d: %s", resp.StatusCode, truncate(string(body), errorBodyLimit)),
			CodeIdentityVerifyFailed,
		)
	}
	var userInfo struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
	}
	if err = json.Unmarshal(body, &userInfo); err != nil {
		return nil, bizerr.WrapCode(gerror.Wrap(err, "oidc google: decode userinfo response failed"), CodeIdentityVerifyFailed)
	}
	if strings.TrimSpace(userInfo.Sub) == "" || strings.TrimSpace(userInfo.Email) == "" {
		return nil, bizerr.WrapCode(gerror.New("oidc google: userinfo response missing sub or email"), CodeIdentityVerifyFailed)
	}
	if !userInfo.EmailVerified {
		return nil, bizerr.WrapCode(
			gerror.New("oidc google: account email is not verified by Google"),
			CodeEmailNotVerified,
		)
	}
	return &VerifiedIdentity{
		Subject:     userInfo.Sub,
		Email:       userInfo.Email,
		DisplayName: userInfo.Name,
	}, nil
}

// SetHTTPClient overrides the outbound HTTP client. Intended for tests that
// need to intercept Google traffic.
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
