// oauth_idtoken.go verifies OIDC id_token JWTs using provider JWKS.

package oauth

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/golang-jwt/jwt/v5"

	"lina-core/pkg/bizerr"
)

// VerifiedIdentity is the neutral identity projected after OIDC verification.
type VerifiedIdentity struct {
	Subject     string
	Email       string
	DisplayName string
}

type idTokenClaims struct {
	Email         string `json:"email"`
	EmailVerified any    `json:"email_verified"`
	Name          string `json:"name"`
	Nonce         string `json:"nonce"`
	jwt.RegisteredClaims
}

func (c *idTokenClaims) emailVerifiedBool() bool {
	switch v := c.EmailVerified.(type) {
	case bool:
		return v
	case string:
		return strings.EqualFold(v, "true")
	default:
		return false
	}
}

func verifyIDToken(ctx context.Context, jwks *jwksClient, rawToken string, jwksURL string, expectedIssuer string, clientID string, expectedNonce string) (*VerifiedIdentity, error) {
	trimmed := strings.TrimSpace(rawToken)
	if trimmed == "" {
		return nil, bizerr.WrapCode(gerror.New("oidc generic: id_token is empty"), CodeIdentityVerifyFailed)
	}
	if jwks == nil {
		return nil, bizerr.WrapCode(gerror.New("oidc generic: jwks client missing"), CodeIdentityVerifyFailed)
	}
	claims := &idTokenClaims{}
	token, err := jwt.ParseWithClaims(trimmed, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, gerror.Newf("unexpected signing method %v", token.Header["alg"])
		}
		kid, _ := token.Header["kid"].(string)
		return jwks.keyForKid(ctx, jwksURL, kid)
	}, jwt.WithExpirationRequired(), jwt.WithAudience(clientID))
	if err != nil || token == nil || !token.Valid {
		return nil, bizerr.WrapCode(gerror.Wrap(err, "oidc generic: id_token validation failed"), CodeIdentityVerifyFailed)
	}
	issuer := normalizeIssuer(claims.Issuer)
	wantIssuer := normalizeIssuer(expectedIssuer)
	if issuer == "" || (wantIssuer != "" && issuer != wantIssuer) {
		return nil, bizerr.WrapCode(gerror.Newf("oidc generic: unexpected issuer %q", claims.Issuer), CodeIdentityVerifyFailed)
	}
	if expectedNonce != "" && strings.TrimSpace(claims.Nonce) != expectedNonce {
		return nil, bizerr.WrapCode(gerror.New("oidc generic: id_token nonce mismatch"), CodeIdentityVerifyFailed)
	}
	subject := strings.TrimSpace(claims.Subject)
	if subject == "" {
		return nil, bizerr.WrapCode(gerror.New("oidc generic: id_token missing sub"), CodeIdentityVerifyFailed)
	}
	email := strings.TrimSpace(claims.Email)
	// Email is optional for enterprise IdPs; when present and marked unverified, drop it
	// rather than rejecting login so subject-based linking still works.
	if email != "" && !claims.emailVerifiedBool() {
		// Some IdPs omit email_verified; keep email when claim absent (nil).
		if claims.EmailVerified != nil {
			email = ""
		}
	}
	return &VerifiedIdentity{
		Subject:     subject,
		Email:       email,
		DisplayName: strings.TrimSpace(claims.Name),
	}, nil
}
