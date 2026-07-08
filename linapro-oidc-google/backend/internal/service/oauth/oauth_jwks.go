// oauth_jwks.go implements a cached Google JWKS (JSON Web Key Set) client
// used to verify Google-issued ID Tokens locally. Keys are fetched from
// Google's certs endpoint through the bounded HTTP client and cached for a
// fixed TTL so One Tap logins do not hit Google on every verification.

package oauth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/bizerr"
)

// googleJWKSEndpoint serves Google's current ID Token signing keys.
const googleJWKSEndpoint = "https://www.googleapis.com/oauth2/v3/certs"

// jwksCacheTTL bounds how long fetched keys are reused before a refresh.
// Google rotates keys on the order of days; one hour keeps rotation lag
// negligible while avoiding per-login fetches.
const jwksCacheTTL = time.Hour

// jwksClient fetches and caches Google's RSA public keys indexed by key ID.
type jwksClient struct {
	httpClient *http.Client
	mu         sync.Mutex
	keys       map[string]*rsa.PublicKey
	fetchedAt  time.Time
}

// newJWKSClient creates a JWKS client with a bounded outbound HTTP client.
func newJWKSClient() *jwksClient {
	return &jwksClient{
		httpClient: &http.Client{Timeout: httpTimeout},
		keys:       map[string]*rsa.PublicKey{},
	}
}

// keyForKid returns the RSA public key for one JWT key ID, refreshing the
// cached key set when the kid is unknown or the cache expired.
func (c *jwksClient) keyForKid(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if key, ok := c.keys[kid]; ok && time.Since(c.fetchedAt) < jwksCacheTTL {
		return key, nil
	}
	if err := c.refreshLocked(ctx); err != nil {
		return nil, err
	}
	if key, ok := c.keys[kid]; ok {
		return key, nil
	}
	return nil, bizerr.WrapCode(
		gerror.Newf("oidc google: no JWKS key for kid %q", kid),
		CodeIdentityVerifyFailed,
	)
}

// refreshLocked fetches the current key set. Callers must hold c.mu.
func (c *jwksClient) refreshLocked(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, googleJWKSEndpoint, nil)
	if err != nil {
		return bizerr.WrapCode(gerror.Wrap(err, "oidc google: build JWKS request failed"), CodeIdentityVerifyFailed)
	}
	req.Header.Set("Accept", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return bizerr.WrapCode(gerror.Wrap(err, "oidc google: fetch JWKS failed"), CodeIdentityVerifyFailed)
	}
	defer closeResponseBody(ctx, resp, "google jwks endpoint")
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return bizerr.WrapCode(gerror.Wrap(err, "oidc google: read JWKS response failed"), CodeIdentityVerifyFailed)
	}
	if resp.StatusCode != http.StatusOK {
		return bizerr.WrapCode(
			gerror.Newf("oidc google: JWKS endpoint returned status %d: %s", resp.StatusCode, truncate(string(body), errorBodyLimit)),
			CodeIdentityVerifyFailed,
		)
	}
	var payload struct {
		Keys []struct {
			Kid string `json:"kid"`
			Kty string `json:"kty"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err = json.Unmarshal(body, &payload); err != nil {
		return bizerr.WrapCode(gerror.Wrap(err, "oidc google: decode JWKS response failed"), CodeIdentityVerifyFailed)
	}
	keys := make(map[string]*rsa.PublicKey, len(payload.Keys))
	for _, jwk := range payload.Keys {
		if !strings.EqualFold(jwk.Kty, "RSA") || jwk.Kid == "" {
			continue
		}
		nBytes, decodeErr := base64.RawURLEncoding.DecodeString(jwk.N)
		if decodeErr != nil {
			continue
		}
		eBytes, decodeErr := base64.RawURLEncoding.DecodeString(jwk.E)
		if decodeErr != nil {
			continue
		}
		keys[jwk.Kid] = &rsa.PublicKey{
			N: new(big.Int).SetBytes(nBytes),
			E: int(new(big.Int).SetBytes(eBytes).Int64()),
		}
	}
	if len(keys) == 0 {
		return bizerr.WrapCode(gerror.New("oidc google: JWKS response contained no usable RSA keys"), CodeIdentityVerifyFailed)
	}
	c.keys = keys
	c.fetchedAt = time.Now()
	return nil
}
