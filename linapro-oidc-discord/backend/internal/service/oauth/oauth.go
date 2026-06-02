// Package oauth implements the Discord OAuth2 authorization-code flow used by
// the linapro-oidc-discord source plugin. It owns authorization URL building,
// HMAC-signed state encoding, token exchange against Discord's token endpoint,
// and user-info retrieval against Discord's REST API. The service is stateless
// and only depends on OAuth client settings; provider enablement is enforced by
// the controller through the host PluginState capability before this service is
// called.
package oauth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/logger"
)

// Discord OAuth2 endpoint constants used by the authorization-code flow.
const (
	// authorizeEndpoint is the Discord OAuth2 authorization endpoint.
	authorizeEndpoint = "https://discord.com/oauth2/authorize"
	// tokenEndpoint is the Discord OAuth2 token endpoint used for code exchange.
	tokenEndpoint = "https://discord.com/api/oauth2/token"
	// userInfoEndpoint is the Discord REST API endpoint for the current user.
	userInfoEndpoint = "https://discord.com/api/users/@me"
	// requestedScopes are the OAuth2 scopes requested at authorization time.
	requestedScopes = "identify email"
	// stateTTL bounds the OAuth state lifetime to mitigate replay attacks.
	stateTTL = 10 * time.Minute
	// httpTimeout bounds each outbound HTTP request to Discord.
	httpTimeout = 15 * time.Second
)

// ProviderID is the stable Discord provider identifier published to host auth.
const ProviderID = "discord"

// PluginID is the owning source-plugin identifier.
const PluginID = "linapro-oidc-discord"

// Settings captures the OAuth runtime settings required to start the flow.
// Callers fetch settings from the plugin config service before each request
// so secret rotations take effect without restarting the host.
type Settings struct {
	// ClientID is the Discord OAuth2 client identifier.
	ClientID string
	// ClientSecret is the Discord OAuth2 client secret.
	ClientSecret string
	// RedirectURI is the registered OAuth2 redirect URI. When empty the
	// caller must supply a default based on the current request.
	RedirectURI string
}

// StatePayload holds the decoded contents of an OAuth state token.
type StatePayload struct {
	// StateKey carries the user-supplied state key used by frontend redirect rules.
	StateKey string `json:"stateKey"`
	// Nonce is a random value bound to one authorization round-trip.
	Nonce string `json:"nonce"`
	// ExpiresAt is the absolute deadline beyond which the state is rejected.
	ExpiresAt int64 `json:"expiresAt"`
}

// UserIdentity is the verified identity returned by Discord's user endpoint.
type UserIdentity struct {
	// Subject is Discord's stable user identifier.
	Subject string
	// Email is the verified email address. Discord omits the field when the
	// requesting application is not authorized for the email scope.
	Email string
	// EmailVerified reports whether Discord has verified the email address.
	EmailVerified bool
	// Name is the human-readable display name (global_name when present,
	// otherwise the legacy username).
	Name string
}

// Service implements the Discord OAuth2 authorization-code flow.
type Service struct {
	// httpClient performs outbound requests to Discord. Tests can substitute
	// a custom client through SetHTTPClient.
	httpClient *http.Client
}

// New creates a Discord OAuth Service with a bounded outbound HTTP client.
func New() *Service {
	return &Service{
		httpClient: &http.Client{Timeout: httpTimeout},
	}
}

// SetHTTPClient overrides the outbound HTTP client. Intended for tests that
// need to intercept Discord traffic.
func (s *Service) SetHTTPClient(client *http.Client) {
	if client == nil {
		return
	}
	s.httpClient = client
}

// BuildAuthorizeURL produces the redirect URL the browser must follow to start
// Discord's authorization flow. It also returns the signed state token so the
// caller can include it in audit logs.
func (s *Service) BuildAuthorizeURL(settings Settings, redirectURI string, stateKey string) (string, string, error) {
	clientID := strings.TrimSpace(settings.ClientID)
	if clientID == "" {
		return "", "", gerror.New("discord client id is not configured")
	}
	resolvedRedirectURI := strings.TrimSpace(redirectURI)
	if resolvedRedirectURI == "" {
		resolvedRedirectURI = strings.TrimSpace(settings.RedirectURI)
	}
	if resolvedRedirectURI == "" {
		return "", "", gerror.New("discord redirect uri is not configured")
	}
	nonce, err := randomToken(16)
	if err != nil {
		return "", "", gerror.Wrap(err, "generate oauth state nonce failed")
	}
	state, err := encodeState(StatePayload{
		StateKey:  strings.TrimSpace(stateKey),
		Nonce:     nonce,
		ExpiresAt: time.Now().Add(stateTTL).Unix(),
	}, settings.ClientSecret)
	if err != nil {
		return "", "", gerror.Wrap(err, "encode oauth state failed")
	}
	params := url.Values{}
	params.Set("client_id", clientID)
	params.Set("redirect_uri", resolvedRedirectURI)
	params.Set("response_type", "code")
	params.Set("scope", requestedScopes)
	params.Set("state", state)
	params.Set("prompt", "consent")
	return authorizeEndpoint + "?" + params.Encode(), state, nil
}

// DecodeState verifies the HMAC signature, expiration, and JSON payload of a
// state token previously produced by BuildAuthorizeURL.
func (s *Service) DecodeState(state string, clientSecret string) (StatePayload, error) {
	payload, err := decodeState(state, clientSecret)
	if err != nil {
		return StatePayload{}, err
	}
	if payload.ExpiresAt > 0 && time.Now().Unix() > payload.ExpiresAt {
		return StatePayload{}, gerror.New("oauth state has expired")
	}
	return payload, nil
}

// ExchangeCode trades a Discord authorization code for a Discord access token.
func (s *Service) ExchangeCode(ctx context.Context, settings Settings, redirectURI string, code string) (string, error) {
	clientID := strings.TrimSpace(settings.ClientID)
	clientSecret := strings.TrimSpace(settings.ClientSecret)
	if clientID == "" || clientSecret == "" {
		return "", gerror.New("discord client credentials are not configured")
	}
	resolvedRedirectURI := strings.TrimSpace(redirectURI)
	if resolvedRedirectURI == "" {
		resolvedRedirectURI = strings.TrimSpace(settings.RedirectURI)
	}
	if resolvedRedirectURI == "" {
		return "", gerror.New("discord redirect uri is not configured")
	}
	form := url.Values{}
	form.Set("code", code)
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)
	form.Set("redirect_uri", resolvedRedirectURI)
	form.Set("grant_type", "authorization_code")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return "", gerror.Wrap(err, "build discord token request failed")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", gerror.Wrap(err, "call discord token endpoint failed")
	}
	defer closeResponseBody(ctx, resp, "discord oauth token endpoint")
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", gerror.Wrap(err, "read discord token response failed")
	}
	if resp.StatusCode != http.StatusOK {
		return "", gerror.Newf("discord token endpoint returned status %d: %s", resp.StatusCode, truncate(string(body), 512))
	}
	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Error       string `json:"error"`
		ErrorDesc   string `json:"error_description"`
	}
	if err = json.Unmarshal(body, &tokenResp); err != nil {
		return "", gerror.Wrap(err, "decode discord token response failed")
	}
	if tokenResp.Error != "" {
		return "", gerror.Newf("discord token endpoint error: %s: %s", tokenResp.Error, tokenResp.ErrorDesc)
	}
	if strings.TrimSpace(tokenResp.AccessToken) == "" {
		return "", gerror.New("discord token endpoint returned empty access token")
	}
	return tokenResp.AccessToken, nil
}

// FetchUserIdentity calls Discord's REST API and returns the verified identity.
func (s *Service) FetchUserIdentity(ctx context.Context, accessToken string) (*UserIdentity, error) {
	trimmed := strings.TrimSpace(accessToken)
	if trimmed == "" {
		return nil, gerror.New("discord access token is empty")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, userInfoEndpoint, nil)
	if err != nil {
		return nil, gerror.Wrap(err, "build discord userinfo request failed")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+trimmed)
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, gerror.Wrap(err, "call discord userinfo endpoint failed")
	}
	defer closeResponseBody(ctx, resp, "discord oauth userinfo endpoint")
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, gerror.Wrap(err, "read discord userinfo response failed")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, gerror.Newf("discord userinfo endpoint returned status %d: %s", resp.StatusCode, truncate(string(body), 512))
	}
	var userInfo struct {
		ID         string `json:"id"`
		Email      string `json:"email"`
		Verified   bool   `json:"verified"`
		Username   string `json:"username"`
		GlobalName string `json:"global_name"`
	}
	if err = json.Unmarshal(body, &userInfo); err != nil {
		return nil, gerror.Wrap(err, "decode discord userinfo response failed")
	}
	if strings.TrimSpace(userInfo.ID) == "" || strings.TrimSpace(userInfo.Email) == "" {
		return nil, gerror.New("discord userinfo response missing id or email")
	}
	displayName := strings.TrimSpace(userInfo.GlobalName)
	if displayName == "" {
		displayName = strings.TrimSpace(userInfo.Username)
	}
	return &UserIdentity{
		Subject:       userInfo.ID,
		Email:         userInfo.Email,
		EmailVerified: userInfo.Verified,
		Name:          displayName,
	}, nil
}

// encodeState serializes one state payload and signs it with HMAC-SHA256
// using a key derived from the client secret.
func encodeState(payload StatePayload, clientSecret string) (string, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	encoded := base64.RawURLEncoding.EncodeToString(raw)
	mac := computeStateMAC(encoded, clientSecret)
	return encoded + "." + mac, nil
}

// decodeState verifies the HMAC signature on a state token and returns the
// decoded payload.
func decodeState(state string, clientSecret string) (StatePayload, error) {
	parts := strings.SplitN(state, ".", 2)
	if len(parts) != 2 {
		return StatePayload{}, gerror.New("oauth state token is malformed")
	}
	expected := computeStateMAC(parts[0], clientSecret)
	if !hmac.Equal([]byte(parts[1]), []byte(expected)) {
		return StatePayload{}, gerror.New("oauth state signature mismatch")
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return StatePayload{}, gerror.Wrap(err, "decode oauth state payload failed")
	}
	var payload StatePayload
	if err = json.Unmarshal(raw, &payload); err != nil {
		return StatePayload{}, gerror.Wrap(err, "parse oauth state payload failed")
	}
	return payload, nil
}

// computeStateMAC derives a stable HMAC for one state payload using a key
// scoped to the plugin so the signature cannot be reused outside this plugin.
func computeStateMAC(encodedPayload string, clientSecret string) string {
	key := []byte("linapro-oidc-discord::" + clientSecret)
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(encodedPayload))
	return hex.EncodeToString(mac.Sum(nil))
}

// randomToken returns a URL-safe random string of the requested byte length.
func randomToken(byteLen int) (string, error) {
	buffer := make([]byte, byteLen)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

// truncate trims long strings used in error messages so secrets and large
// payloads do not leak into logs.
func truncate(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return fmt.Sprintf("%s...(truncated)", value[:limit])
}

// closeResponseBody closes resp.Body once the caller has finished reading
// it and logs any close-time error against the supplied context. It exists
// so the call sites stay compatible with the project rule that forbids
// discarding error returns via `_ = err`. The endpointLabel parameter
// disambiguates token vs userinfo log lines without leaking secret values.
func closeResponseBody(ctx context.Context, resp *http.Response, endpointLabel string) {
	if resp == nil || resp.Body == nil {
		return
	}
	if err := resp.Body.Close(); err != nil {
		logger.Warningf(ctx, "%s close response body failed err=%v", endpointLabel, err)
	}
}
