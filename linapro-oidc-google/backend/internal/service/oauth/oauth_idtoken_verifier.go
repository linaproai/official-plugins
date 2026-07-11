// oauth_idtoken_verifier.go verifies Google-issued ID Tokens delivered by the
// Google One Tap embed (GSI login_uri form POST). Verification is fully local
// once the JWKS keys are cached: RSA signature, issuer, audience (the
// configured client ID), expiry, and the verified-email claim are all
// enforced before the identity reaches the host external-login seam.

package oauth

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/golang-jwt/jwt/v5"

	"lina-core/pkg/bizerr"
)

// googleIssuer and googleIssuerHTTPS are the accepted `iss` claim values for
// Google ID Tokens per Google's OpenID Connect documentation.
const (
	googleIssuer      = "accounts.google.com"
	googleIssuerHTTPS = "https://accounts.google.com"
)

// IDTokenVerifier validates one Google ID Token and projects the verified
// identity. It is the One Tap counterpart of IdentityVerifier: both produce
// the same neutral VerifiedIdentity consumed by the login orchestration.
type IDTokenVerifier interface {
	// VerifyIDToken validates the raw JWT credential and returns the identity.
	VerifyIDToken(ctx context.Context, rawToken string) (*VerifiedIdentity, error)
}

// googleIDTokenClaims models the Google ID Token claims the plugin consumes.
type googleIDTokenClaims struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	jwt.RegisteredClaims
}

// jwksIDTokenVerifier is the production ID Token verifier backed by Google's
// JWKS keys and the shared request-time settings resolver.
type jwksIDTokenVerifier struct {
	jwks     *jwksClient
	settings SettingsSource
}

// NewJWKSIDTokenVerifier returns the production One Tap ID Token verifier.
// The settings source supplies the expected audience (client ID) at request
// time so credential rotations apply without a restart.
func NewJWKSIDTokenVerifier(settings SettingsSource) IDTokenVerifier {
	return &jwksIDTokenVerifier{jwks: newJWKSClient(), settings: settings}
}

// VerifyIDToken validates signature, issuer, audience, expiry, and the
// verified-email claim, then projects the neutral identity.
func (v *jwksIDTokenVerifier) VerifyIDToken(ctx context.Context, rawToken string) (*VerifiedIdentity, error) {
	trimmed := strings.TrimSpace(rawToken)
	if trimmed == "" {
		return nil, bizerr.WrapCode(gerror.New("oidc google: one tap credential is empty"), CodeIdentityVerifyFailed)
	}
	if v == nil || v.settings == nil || v.jwks == nil {
		return nil, bizerr.WrapCode(gerror.New("oidc google: one tap verifier is not wired"), CodeConfigMissing)
	}
	clientID := strings.TrimSpace(v.settings.ResolveConfig(ctx).ClientID)
	if !isConfiguredCredential(clientID) {
		return nil, bizerr.WrapCode(gerror.New("oidc google: client id is not configured"), CodeConfigMissing)
	}
	claims := &googleIDTokenClaims{}
	token, err := jwt.ParseWithClaims(trimmed, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, gerror.Newf("unexpected signing method %v", token.Header["alg"])
		}
		kid, _ := token.Header["kid"].(string)
		return v.jwks.keyForKid(ctx, kid)
	}, jwt.WithExpirationRequired(), jwt.WithAudience(clientID))
	if err != nil || !token.Valid {
		return nil, bizerr.WrapCode(gerror.Wrap(err, "oidc google: one tap credential validation failed"), CodeIdentityVerifyFailed)
	}
	issuer := strings.TrimSpace(claims.Issuer)
	if issuer != googleIssuer && issuer != googleIssuerHTTPS {
		return nil, bizerr.WrapCode(gerror.Newf("oidc google: unexpected issuer %q", issuer), CodeIdentityVerifyFailed)
	}
	subject := strings.TrimSpace(claims.Subject)
	email := strings.TrimSpace(claims.Email)
	if subject == "" || email == "" {
		return nil, bizerr.WrapCode(gerror.New("oidc google: one tap claims missing sub or email"), CodeIdentityVerifyFailed)
	}
	if !claims.EmailVerified {
		return nil, bizerr.WrapCode(gerror.New("oidc google: account email is not verified by Google"), CodeEmailNotVerified)
	}
	return &VerifiedIdentity{
		Subject:     subject,
		Email:       email,
		DisplayName: claims.Name,
	}, nil
}
