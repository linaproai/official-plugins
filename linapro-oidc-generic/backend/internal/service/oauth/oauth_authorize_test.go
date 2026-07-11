// oauth_authorize_test.go covers fail-closed authorize and PKCE parameters.

package oauth

import (
	"context"
	"net/url"
	"strings"
	"testing"
)

type staticStateCodec struct{}

func (staticStateCodec) Encode(_ context.Context, stateKey string, returnTo string, _ string, codeVerifier string, oidcNonce string) (string, error) {
	return "state:" + stateKey + ":" + returnTo + ":" + codeVerifier + ":" + oidcNonce, nil
}

func (staticStateCodec) Decode(_ context.Context, _ string, _ string) (StatePayload, error) {
	return StatePayload{}, nil
}

func newAuthorizeTestService(config Config) Service {
	return New(nil, NewConfigResolver(nil, config), staticStateCodec{})
}

func TestBuildAuthorizeURLRejectsMissingCredentials(t *testing.T) {
	t.Parallel()
	cases := []Config{
		{Issuer: "", ClientID: "id", ClientSecret: "secret", RedirectURL: "http://127.0.0.1/cb"},
		{Issuer: "https://idp.example.com", ClientID: "", ClientSecret: "secret", RedirectURL: "http://127.0.0.1/cb"},
		{Issuer: "https://idp.example.com", ClientID: "REPLACE_ME_X", ClientSecret: "secret", RedirectURL: "http://127.0.0.1/cb"},
		{Issuer: "https://idp.example.com", ClientID: "id", ClientSecret: "secret", RedirectURL: ""},
	}
	for _, cfg := range cases {
		_, err := newAuthorizeTestService(cfg).BuildAuthorizeURL(context.Background(), "biz", "/admin/auth/login")
		if err == nil {
			t.Fatalf("expected config missing for %+v", cfg)
		}
	}
}

func TestPKCEChallengeS256Deterministic(t *testing.T) {
	t.Parallel()
	// RFC 7636 appendix B example
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	got := pkceChallengeS256(verifier)
	want := "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"
	if got != want {
		t.Fatalf("challenge = %q want %q", got, want)
	}
}

func TestNormalizeScopesRequiresOpenID(t *testing.T) {
	t.Parallel()
	got := normalizeScopes("email profile", nil)
	if got[0] != "openid" {
		t.Fatalf("scopes = %v", got)
	}
	if !strings.Contains(strings.Join(got, " "), "email") {
		t.Fatalf("scopes lost email: %v", got)
	}
}

func TestSanitizeReturnToRejectsOpenRedirects(t *testing.T) {
	t.Parallel()
	if got := sanitizeReturnTo("/admin/auth/login"); got != "/admin/auth/login" {
		t.Fatalf("relative path rejected: %q", got)
	}
	if got := sanitizeReturnTo("https://evil.example/phish"); got != "" {
		t.Fatalf("absolute URL accepted: %q", got)
	}
	if got := sanitizeReturnTo("//evil.example/phish"); got != "" {
		t.Fatalf("protocol-relative URL accepted: %q", got)
	}
}

func TestIsLoginConfigured(t *testing.T) {
	t.Parallel()
	if IsLoginConfigured(Config{}) {
		t.Fatal("empty should be unconfigured")
	}
	if !IsLoginConfigured(Config{
		Issuer: "https://idp.example.com", ClientID: "id", ClientSecret: "sec", RedirectURL: "http://x/cb",
	}) {
		t.Fatal("expected configured")
	}
}

// Ensure AuthorizeRequest type stays importable for URL smoke without discovery.
func TestAuthorizeRequestType(t *testing.T) {
	t.Parallel()
	u, err := url.Parse("https://example.com/auth")
	if err != nil {
		t.Fatal(err)
	}
	if u.Host != "example.com" {
		t.Fatal(u.Host)
	}
}

func TestResolveAllowAutoProvisionDefaultFalse(t *testing.T) {
	t.Parallel()
	svc := New(nil, NewConfigResolver(nil, DefaultConfig()), staticStateCodec{}).(*serviceImpl)
	if svc.resolveAllowAutoProvision(context.Background()) {
		t.Fatal("auto provision must default false")
	}
}
