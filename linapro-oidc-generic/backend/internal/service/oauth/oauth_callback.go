// oauth_callback.go completes the OIDC code flow and host external login.

package oauth

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
	"lina-core/pkg/plugin/capability/authcap/extlogin"
	"lina-plugin-linapro-extid-core/backend/cap/extidcap"
)

func (s *serviceImpl) resolveAllowAutoProvision(ctx context.Context) bool {
	if s == nil || s.configResolver == nil || s.configResolver.settingsSvc == nil {
		return false
	}
	snapshot, err := s.configResolver.settingsSvc.Load(ctx)
	if err != nil || snapshot == nil {
		return false
	}
	return snapshot.AllowAutoProvision
}

// CompleteCallback validates state, exchanges code, verifies id_token, and
// hands the identity to the host external-login seam.
func (s *serviceImpl) CompleteCallback(ctx context.Context, in CallbackInput) (*CallbackOutput, error) {
	code := strings.TrimSpace(in.Code)
	state := strings.TrimSpace(in.State)
	if code == "" {
		return nil, bizerr.WrapCode(gerror.New("oidc generic: callback code is empty"), CodeCallbackCodeRequired)
	}
	if state == "" {
		return nil, bizerr.WrapCode(gerror.New("oidc generic: callback state is empty"), CodeCallbackStateMismatch)
	}
	if s.stateCodec == nil || s.discovery == nil || s.jwks == nil {
		return nil, bizerr.WrapCode(gerror.New("oidc generic: callback dependencies missing"), CodeIdentityVerifyFailed)
	}
	config := s.resolveConfig(ctx)
	if !IsLoginConfigured(config) {
		return nil, bizerr.WrapCode(gerror.New("oidc generic: configuration missing"), CodeConfigMissing)
	}
	statePayload, err := s.stateCodec.Decode(ctx, state, config.ClientSecret)
	if err != nil {
		return nil, err
	}
	if s.externalLoginSvc == nil {
		return nil, bizerr.WrapCode(
			gerror.New("oidc generic: external-login service is unavailable"),
			CodeExternalLoginUnavailable,
		)
	}
	doc, err := s.discovery.resolve(ctx, config.Issuer)
	if err != nil {
		return nil, err
	}
	identity, err := s.exchangeAndVerify(ctx, config, doc, code, statePayload.CodeVerifier, statePayload.OIDCNonce)
	if err != nil {
		return nil, err
	}
	if identity == nil || strings.TrimSpace(identity.Subject) == "" {
		return nil, bizerr.WrapCode(
			gerror.New("oidc generic: verified identity is missing subject"),
			CodeIdentityVerifyFailed,
		)
	}
	logger.Infof(
		ctx,
		"linapro-oidc-generic callback verified provider=%s subject=%s email=%s",
		Provider,
		identity.Subject,
		identity.Email,
	)
	loginOut, err := s.externalLoginSvc.LoginByVerifiedIdentity(ctx, extlogin.LoginInput{
		Provider:           Provider,
		Subject:            identity.Subject,
		Email:              identity.Email,
		DisplayName:        identity.DisplayName,
		AllowAutoProvision: s.resolveAllowAutoProvision(ctx),
	})
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeExternalLoginFailed)
	}
	if loginOut == nil {
		return nil, bizerr.WrapCode(
			gerror.New("oidc generic: external-login returned nil outcome"),
			CodeExternalLoginFailed,
		)
	}
	handoff, err := extidcap.CreateLoginHandoffFromHost(loginOut)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeExternalLoginFailed)
	}
	return &CallbackOutput{
		Handoff:          handoff,
		AccessToken:      loginOut.AccessToken,
		RefreshToken:     loginOut.RefreshToken,
		PreToken:         loginOut.PreToken,
		TenantCandidates: loginOut.TenantCandidates,
		StateKey:         statePayload.StateKey,
		ReturnTo:         sanitizeReturnTo(statePayload.ReturnTo),
	}, nil
}

func (s *serviceImpl) exchangeAndVerify(
	ctx context.Context,
	config Config,
	doc discoveryDocument,
	code string,
	codeVerifier string,
	oidcNonce string,
) (*VerifiedIdentity, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", config.RedirectURL)
	form.Set("client_id", config.ClientID)
	form.Set("client_secret", config.ClientSecret)
	if strings.TrimSpace(codeVerifier) != "" {
		form.Set("code_verifier", codeVerifier)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, doc.TokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, bizerr.WrapCode(gerror.Wrap(err, "oidc generic: build token request failed"), CodeIdentityVerifyFailed)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, bizerr.WrapCode(gerror.Wrap(err, "oidc generic: token request failed"), CodeIdentityVerifyFailed)
	}
	defer closeResponseBody(ctx, resp, "oidc token")
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, bizerr.WrapCode(gerror.Wrap(err, "oidc generic: read token response failed"), CodeIdentityVerifyFailed)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, bizerr.WrapCode(
			gerror.Newf("oidc generic: token status %d: %s", resp.StatusCode, truncate(string(body), errorBodyLimit)),
			CodeIdentityVerifyFailed,
		)
	}
	var tokenResp struct {
		IDToken     string `json:"id_token"`
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
		ErrorDesc   string `json:"error_description"`
	}
	if err = json.Unmarshal(body, &tokenResp); err != nil {
		return nil, bizerr.WrapCode(gerror.Wrap(err, "oidc generic: decode token response failed"), CodeIdentityVerifyFailed)
	}
	if tokenResp.Error != "" {
		return nil, bizerr.WrapCode(
			gerror.Newf("oidc generic: token error: %s: %s", tokenResp.Error, tokenResp.ErrorDesc),
			CodeIdentityVerifyFailed,
		)
	}
	if strings.TrimSpace(tokenResp.IDToken) == "" {
		return nil, bizerr.WrapCode(gerror.New("oidc generic: token response missing id_token"), CodeIdentityVerifyFailed)
	}
	expectedIssuer := normalizeIssuer(doc.Issuer)
	if expectedIssuer == "" {
		expectedIssuer = normalizeIssuer(config.Issuer)
	}
	identity, err := verifyIDToken(ctx, s.jwks, tokenResp.IDToken, doc.JWKSURI, expectedIssuer, config.ClientID, oidcNonce)
	if err != nil {
		return nil, err
	}
	// Optionally enrich display name from userinfo when empty.
	if identity.DisplayName == "" && tokenResp.AccessToken != "" && strings.TrimSpace(doc.UserInfoEndpoint) != "" {
		if name := fetchUserInfoName(ctx, doc.UserInfoEndpoint, tokenResp.AccessToken); name != "" {
			identity.DisplayName = name
		}
	}
	return identity, nil
}

func fetchUserInfoName(ctx context.Context, userInfoURL string, accessToken string) string {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, userInfoURL, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer closeResponseBody(ctx, resp, "oidc userinfo")
	if resp.StatusCode != http.StatusOK {
		return ""
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	var payload struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return ""
	}
	return strings.TrimSpace(payload.Name)
}
