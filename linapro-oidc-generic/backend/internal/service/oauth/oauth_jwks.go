// oauth_jwks.go implements a cached JWKS client for OIDC ID token verification.

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

const jwksCacheTTL = time.Hour

type jwksClient struct {
	httpClient *http.Client
	mu         sync.Mutex
	// keysByURL caches RSA keys per JWKS URL.
	keysByURL map[string]cachedJWKS
}

type cachedJWKS struct {
	keys      map[string]*rsa.PublicKey
	fetchedAt time.Time
}

func newJWKSClient() *jwksClient {
	return &jwksClient{
		httpClient: &http.Client{Timeout: httpTimeout},
		keysByURL:  map[string]cachedJWKS{},
	}
}

func (c *jwksClient) keyForKid(ctx context.Context, jwksURL string, kid string) (*rsa.PublicKey, error) {
	jwksURL = strings.TrimSpace(jwksURL)
	if jwksURL == "" {
		return nil, bizerr.WrapCode(gerror.New("oidc generic: jwks url is empty"), CodeIdentityVerifyFailed)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.keysByURL[jwksURL]
	if ok {
		if key, found := entry.keys[kid]; found && time.Since(entry.fetchedAt) < jwksCacheTTL {
			return key, nil
		}
	}
	keys, err := c.refreshLocked(ctx, jwksURL)
	if err != nil {
		return nil, err
	}
	c.keysByURL[jwksURL] = cachedJWKS{keys: keys, fetchedAt: time.Now()}
	if key, found := keys[kid]; found {
		return key, nil
	}
	return nil, bizerr.WrapCode(
		gerror.Newf("oidc generic: no JWKS key for kid %q", kid),
		CodeIdentityVerifyFailed,
	)
}

func (c *jwksClient) refreshLocked(ctx context.Context, jwksURL string) (map[string]*rsa.PublicKey, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, jwksURL, nil)
	if err != nil {
		return nil, bizerr.WrapCode(gerror.Wrap(err, "oidc generic: build JWKS request failed"), CodeIdentityVerifyFailed)
	}
	req.Header.Set("Accept", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, bizerr.WrapCode(gerror.Wrap(err, "oidc generic: fetch JWKS failed"), CodeIdentityVerifyFailed)
	}
	defer closeResponseBody(ctx, resp, "oidc jwks")
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, bizerr.WrapCode(gerror.Wrap(err, "oidc generic: read JWKS failed"), CodeIdentityVerifyFailed)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, bizerr.WrapCode(
			gerror.Newf("oidc generic: JWKS status %d: %s", resp.StatusCode, truncate(string(body), errorBodyLimit)),
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
		return nil, bizerr.WrapCode(gerror.Wrap(err, "oidc generic: decode JWKS failed"), CodeIdentityVerifyFailed)
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
		return nil, bizerr.WrapCode(gerror.New("oidc generic: JWKS contained no usable RSA keys"), CodeIdentityVerifyFailed)
	}
	return keys, nil
}
