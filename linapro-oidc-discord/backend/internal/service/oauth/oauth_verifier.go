// oauth_verifier.go defines the identity-verifier contract the Discord OIDC
// login flow depends on plus a reference stub implementation. In a real
// deployment the plugin swaps in an HTTP-backed verifier that talks to the
// Discord token and userinfo endpoints; the stub is intentionally lightweight
// so the surrounding orchestration can be exercised end-to-end without a
// live Discord application.

package oauth

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/bizerr"
)

// VerifiedIdentity is the neutral projection the plugin submits to the host
// external-login seam after a successful Discord callback exchange.
type VerifiedIdentity struct {
	// Subject is the immutable Discord-issued user identifier ("id").
	Subject string
	// Email is the Discord account email captured for audit and hook context.
	Email string
	// DisplayName is the Discord user global name or username captured for
	// audit and hook context.
	DisplayName string
}

// IdentityVerifier exchanges the OAuth authorization code for a verified
// identity. The reference stub short-circuits the exchange with a deterministic
// verified identity so the surrounding orchestration can be exercised without
// contacting Discord. The production verifier performs the token exchange and
// userinfo fetch inside its Verify method.
type IdentityVerifier interface {
	// Verify exchanges the supplied authorization code for a verified identity
	// obtained from Discord. The redirectURL parameter is the same value the
	// authorize URL was built with so token exchange can echo it back to
	// Discord. Returning bizerr.CodeIdentityVerifyFailed maps to a stable
	// client error at the controller boundary.
	Verify(ctx context.Context, code string, redirectURL string) (*VerifiedIdentity, error)
}

// stubIdentityVerifier is the deterministic identity verifier used by unit
// tests to exercise the orchestration without contacting Discord. Production
// wiring uses NewHTTPIdentityVerifier.
type stubIdentityVerifier struct{}

// NewStubIdentityVerifier returns the deterministic test verifier that echoes
// the OAuth code as a Discord subject. It must never be wired in production;
// plugin.go wires NewHTTPIdentityVerifier instead.
func NewStubIdentityVerifier() IdentityVerifier {
	return &stubIdentityVerifier{}
}

// Verify projects the OAuth code into a deterministic verified identity so
// unit tests can exercise the downstream host external-login wiring without a
// live Discord application. It rejects blank codes with a stable business
// error.
func (v *stubIdentityVerifier) Verify(_ context.Context, code string, _ string) (*VerifiedIdentity, error) {
	trimmed := strings.TrimSpace(code)
	if trimmed == "" {
		return nil, bizerr.WrapCode(gerror.New("oidc discord: empty authorization code"), CodeIdentityVerifyFailed)
	}
	return &VerifiedIdentity{
		Subject:     "discord-stub-subject-" + trimmed,
		Email:       "stub-user+" + trimmed + "@example.com",
		DisplayName: "Discord Reference User",
	}, nil
}
