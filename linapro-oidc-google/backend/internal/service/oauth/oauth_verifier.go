// oauth_verifier.go defines the identity-verifier contract the Google OIDC
// login flow depends on plus a deterministic stub implementation used by unit
// tests. Production wiring uses the HTTP-backed verifier in
// oauth_verifier_http.go, which performs the real token exchange and userinfo
// fetch against Google.

package oauth

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/bizerr"
)

// VerifiedIdentity is the neutral projection the plugin submits to the host
// external-login seam after a successful Google callback exchange.
type VerifiedIdentity struct {
	// Subject is the immutable Google-issued subject identifier ("sub").
	Subject string
	// Email is the Google account email captured for audit and hook context.
	Email string
	// DisplayName is the Google profile display name captured for audit and
	// hook context.
	DisplayName string
}

// IdentityVerifier exchanges the OAuth authorization code for a verified
// identity. The reference stub short-circuits the exchange with a deterministic
// verified identity so the surrounding orchestration can be exercised without
// contacting Google. The production verifier performs the token exchange and
// userinfo fetch inside its Verify method.
type IdentityVerifier interface {
	// Verify exchanges the supplied authorization code for a verified identity
	// obtained from Google. The redirectURL parameter is the same value the
	// authorize URL was built with so token exchange can echo it back to
	// Google. Returning bizerr.CodeIdentityVerifyFailed maps to a stable client
	// error at the controller boundary.
	Verify(ctx context.Context, code string, redirectURL string) (*VerifiedIdentity, error)
}

// stubIdentityVerifier is the deterministic identity verifier used by unit
// tests to exercise the orchestration without contacting Google. Production
// wiring uses NewHTTPIdentityVerifier.
type stubIdentityVerifier struct{}

// NewStubIdentityVerifier returns the deterministic test verifier that echoes
// the OAuth code as a Google subject. It must never be wired in production;
// plugin.go wires NewHTTPIdentityVerifier instead.
func NewStubIdentityVerifier() IdentityVerifier {
	return &stubIdentityVerifier{}
}

// Verify projects the OAuth code into a deterministic verified identity so
// unit tests can exercise the downstream host external-login wiring without a
// live Google project. It rejects blank codes with a stable business error.
func (v *stubIdentityVerifier) Verify(_ context.Context, code string, _ string) (*VerifiedIdentity, error) {
	trimmed := strings.TrimSpace(code)
	if trimmed == "" {
		return nil, bizerr.WrapCode(gerror.New("oidc google: empty authorization code"), CodeIdentityVerifyFailed)
	}
	return &VerifiedIdentity{
		Subject:     "google-stub-subject-" + trimmed,
		Email:       "stub-user+" + trimmed + "@example.com",
		DisplayName: "Google Reference User",
	}, nil
}
