// oauth_callback.go implements the Discord callback step of the OIDC login
// flow. The controller invokes CompleteCallback with the raw callback values
// and forwards the resulting tokens or pre-token to the SPA login page. The
// state parameter is a self-contained HMAC-signed token validated by
// signature and expiry, so the flow does not depend on cookies surviving the
// cross-site round trip through Discord.

package oauth

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
	"lina-core/pkg/plugin/capability/authcap/extlogin"
	"lina-plugin-linapro-extlogin-core/backend/cap/extidcap"
)

// resolveAllowAutoProvision reads the admin-controlled auto-provision flag at
// request time so settings edits apply without a restart. A missing resolver
// or settings load failure keeps auto-provisioning disabled fail-closed.
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

// CompleteCallback validates the signed callback state, exchanges the code
// for a verified identity, and hands the identity to the host external-login
// seam. Provisioning, tenant resolution, and token minting stay host-owned;
// the plugin only forwards the returned outcome to the controller.
func (s *serviceImpl) CompleteCallback(ctx context.Context, in CallbackInput) (*CallbackOutput, error) {
	code := strings.TrimSpace(in.Code)
	state := strings.TrimSpace(in.State)
	if code == "" {
		return nil, bizerr.WrapCode(gerror.New("oidc discord: callback code is empty"), CodeCallbackCodeRequired)
	}
	if state == "" {
		return nil, bizerr.WrapCode(gerror.New("oidc discord: callback state is empty"), CodeCallbackStateMismatch)
	}
	if s.stateCodec == nil {
		return nil, bizerr.WrapCode(gerror.New("oidc discord: state codec is missing"), CodeCallbackStateMismatch)
	}
	// Re-resolve settings per request so admin edits apply without a restart.
	config := s.resolveConfig(ctx)
	statePayload, err := s.stateCodec.Decode(ctx, state, config.ClientSecret)
	if err != nil {
		return nil, err
	}
	if s.externalLoginSvc == nil {
		return nil, bizerr.WrapCode(
			gerror.New("oidc discord: external-login service is unavailable"),
			CodeExternalLoginUnavailable,
		)
	}
	identity, err := s.verifier.Verify(ctx, code, config.RedirectURL)
	if err != nil {
		return nil, err
	}
	if identity == nil || strings.TrimSpace(identity.Subject) == "" {
		return nil, bizerr.WrapCode(
			gerror.New("oidc discord: verified identity is missing subject"),
			CodeIdentityVerifyFailed,
		)
	}
	logger.Infof(
		ctx,
		"linapro-oidc-discord callback verified provider=%s subject=%s email=%s",
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
			gerror.New("oidc discord: external-login returned nil outcome"),
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
