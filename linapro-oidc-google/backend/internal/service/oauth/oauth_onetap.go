// oauth_onetap.go implements the Google One Tap login step. The embeddable
// GSI snippet form-POSTs an ID Token credential to the plugin's public
// /onetap route; this service validates the token locally (JWKS signature,
// issuer, audience, expiry, verified email) and exchanges the verified
// identity through the same host external-login seam as the authorize-code
// flow, so auto-provisioning and SSO delivery behave identically across both
// entry points.

package oauth

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
	"lina-core/pkg/plugin/capability/authcap/extlogin"
	"lina-plugin-linapro-extid-core/backend/cap/extidcap"
)

// CompleteOneTap validates the One Tap credential and returns the host login
// outcome plus the echoed SSO state key. See the Service interface contract.
func (s *serviceImpl) CompleteOneTap(ctx context.Context, credential string, stateKey string) (*CallbackOutput, error) {
	if s == nil || s.idTokenVerifier == nil {
		return nil, bizerr.WrapCode(
			gerror.New("oidc google: one tap verifier is not wired"),
			CodeExternalLoginUnavailable,
		)
	}
	if !s.resolveEnableOneTap(ctx) {
		return nil, bizerr.WrapCode(
			gerror.New("oidc google: one tap login is disabled"),
			CodeOneTapDisabled,
		)
	}
	if s.externalLoginSvc == nil {
		return nil, bizerr.WrapCode(
			gerror.New("oidc google: external-login service is unavailable"),
			CodeExternalLoginUnavailable,
		)
	}
	identity, err := s.idTokenVerifier.VerifyIDToken(ctx, credential)
	if err != nil {
		return nil, err
	}
	logger.Infof(
		ctx,
		"linapro-oidc-google one tap verified provider=%s subject=%s email=%s",
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
			gerror.New("oidc google: external-login returned nil outcome"),
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
		StateKey:         strings.TrimSpace(stateKey),
	}, nil
}

// resolveEnableOneTap reads the admin-controlled One Tap flag at request time
// so settings edits apply without a restart. Missing wiring or a settings
// load failure keeps One Tap disabled fail-closed.
func (s *serviceImpl) resolveEnableOneTap(ctx context.Context) bool {
	if s == nil || s.configResolver == nil || s.configResolver.settingsSvc == nil {
		return false
	}
	snapshot, err := s.configResolver.settingsSvc.Load(ctx)
	if err != nil || snapshot == nil {
		return false
	}
	return snapshot.EnableOneTap
}
