// oauth_authorize_test.go verifies that login-start fails closed when Google
// client credentials are missing or still set to historical REPLACE_ME
// placeholders, and that a configured client produces a valid authorize URL.

package oauth

import (
	"context"
	"net/url"
	"strings"
	"testing"
)

type staticStateCodec struct{}

func (staticStateCodec) Encode(_ context.Context, stateKey string, returnTo string, _ string) (string, error) {
	return "state:" + stateKey + ":" + returnTo, nil
}

func (staticStateCodec) Decode(_ context.Context, _ string, _ string) (StatePayload, error) {
	return StatePayload{}, nil
}

func newAuthorizeTestService(config Config) Service {
	return New(nil, NewConfigResolver(nil, config), nil, staticStateCodec{}, nil)
}

func TestBuildAuthorizeURLRejectsMissingCredentials(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		config Config
	}{
		{
			name: "empty credentials",
			config: Config{
				ClientID:     "",
				ClientSecret: "",
				RedirectURL:  "http://127.0.0.1:9120/portal/linapro-oidc-google/callback",
				AuthorizeURL: "https://accounts.google.com/o/oauth2/v2/auth",
				Scopes:       []string{"openid"},
			},
		},
		{
			name: "placeholder credentials",
			config: Config{
				ClientID:     "REPLACE_ME_GOOGLE_CLIENT_ID",
				ClientSecret: "REPLACE_ME_GOOGLE_CLIENT_SECRET",
				RedirectURL:  "http://127.0.0.1:9120/portal/linapro-oidc-google/callback",
				AuthorizeURL: "https://accounts.google.com/o/oauth2/v2/auth",
				Scopes:       []string{"openid"},
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := newAuthorizeTestService(tc.config).BuildAuthorizeURL(context.Background(), "biz", "/admin/auth/login")
			if err == nil {
				t.Fatal("expected configuration missing error")
			}
			if !strings.Contains(err.Error(), "configuration missing") && !strings.Contains(err.Error(), "CONFIG_MISSING") {
				// bizerr wraps the message; accept either human text or runtime code surface.
				if !strings.Contains(err.Error(), "missing") {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestBuildAuthorizeURLSucceedsWithConfiguredCredentials(t *testing.T) {
	t.Parallel()

	svc := newAuthorizeTestService(Config{
		ClientID:     "real-client-id",
		ClientSecret: "real-client-secret",
		RedirectURL:  "http://127.0.0.1:9120/portal/linapro-oidc-google/callback",
		AuthorizeURL: "https://accounts.google.com/o/oauth2/v2/auth",
		Scopes:       []string{"openid", "email"},
	})
	authorize, err := svc.BuildAuthorizeURL(context.Background(), "biz", "/admin/auth/login")
	if err != nil {
		t.Fatalf("BuildAuthorizeURL: %v", err)
	}
	parsed, err := url.Parse(authorize.URL)
	if err != nil {
		t.Fatalf("parse authorize URL: %v", err)
	}
	query := parsed.Query()
	if query.Get("client_id") != "real-client-id" {
		t.Fatalf("client_id = %q", query.Get("client_id"))
	}
	if query.Get("state") != "state:biz:/admin/auth/login" {
		t.Fatalf("state = %q", query.Get("state"))
	}
	if strings.Contains(query.Get("client_id"), "REPLACE_ME") {
		t.Fatal("authorize URL must not use placeholder client id")
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
