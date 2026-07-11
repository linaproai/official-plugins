// oauth_discovery.go fetches and caches OIDC discovery documents.

package oauth

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/bizerr"
)

const discoveryCacheTTL = 15 * time.Minute

// discoveryDocument is the subset of OpenID Provider Metadata we need.
type discoveryDocument struct {
	Issuer                string `json:"issuer"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	JWKSURI               string `json:"jwks_uri"`
	UserInfoEndpoint      string `json:"userinfo_endpoint"`
}

type discoveryCache struct {
	httpClient *http.Client
	mu         sync.Mutex
	byIssuer   map[string]cachedDiscovery
}

type cachedDiscovery struct {
	doc       discoveryDocument
	fetchedAt time.Time
}

func newDiscoveryCache() *discoveryCache {
	return &discoveryCache{
		httpClient: &http.Client{Timeout: httpTimeout},
		byIssuer:   map[string]cachedDiscovery{},
	}
}

func (c *discoveryCache) resolve(ctx context.Context, issuer string) (discoveryDocument, error) {
	normalized := normalizeIssuer(issuer)
	if normalized == "" {
		return discoveryDocument{}, bizerr.WrapCode(gerror.New("oidc generic: issuer is empty"), CodeConfigMissing)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if entry, ok := c.byIssuer[normalized]; ok && time.Since(entry.fetchedAt) < discoveryCacheTTL {
		return entry.doc, nil
	}
	doc, err := c.fetchLocked(ctx, normalized)
	if err != nil {
		return discoveryDocument{}, err
	}
	c.byIssuer[normalized] = cachedDiscovery{doc: doc, fetchedAt: time.Now()}
	return doc, nil
}

func (c *discoveryCache) fetchLocked(ctx context.Context, issuer string) (discoveryDocument, error) {
	wellKnown := issuer + "/.well-known/openid-configuration"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, wellKnown, nil)
	if err != nil {
		return discoveryDocument{}, bizerr.WrapCode(gerror.Wrap(err, "oidc generic: build discovery request failed"), CodeDiscoveryFailed)
	}
	req.Header.Set("Accept", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return discoveryDocument{}, bizerr.WrapCode(gerror.Wrap(err, "oidc generic: discovery request failed"), CodeDiscoveryFailed)
	}
	defer closeResponseBody(ctx, resp, "oidc discovery")
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return discoveryDocument{}, bizerr.WrapCode(gerror.Wrap(err, "oidc generic: read discovery body failed"), CodeDiscoveryFailed)
	}
	if resp.StatusCode != http.StatusOK {
		return discoveryDocument{}, bizerr.WrapCode(
			gerror.Newf("oidc generic: discovery status %d: %s", resp.StatusCode, truncate(string(body), errorBodyLimit)),
			CodeDiscoveryFailed,
		)
	}
	var doc discoveryDocument
	if err = json.Unmarshal(body, &doc); err != nil {
		return discoveryDocument{}, bizerr.WrapCode(gerror.Wrap(err, "oidc generic: decode discovery failed"), CodeDiscoveryFailed)
	}
	if strings.TrimSpace(doc.AuthorizationEndpoint) == "" || strings.TrimSpace(doc.TokenEndpoint) == "" || strings.TrimSpace(doc.JWKSURI) == "" {
		return discoveryDocument{}, bizerr.WrapCode(gerror.New("oidc generic: discovery missing required endpoints"), CodeDiscoveryFailed)
	}
	// Prefer discovery issuer when present; callers still validate id_token iss.
	if strings.TrimSpace(doc.Issuer) == "" {
		doc.Issuer = issuer
	}
	return doc, nil
}
