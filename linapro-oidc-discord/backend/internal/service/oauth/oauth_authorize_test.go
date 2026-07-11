// oauth_authorize_test.go verifies that login-start fails closed when Discord
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
	return New(nil, NewConfigResolver(nil, config), nil, staticStateCodec{})
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
				RedirectURL:  "http://127.0.0.1:9120/portal/linapro-oidc-discord/callback",
				AuthorizeURL: "https://discord.com/oauth2/authorize",
				Scopes:       []string{"identify"},
			},
		},
		{
			name: "placeholder credentials",
			config: Config{
				ClientID:     "REPLACE_ME_DISCORD_CLIENT_ID",
				ClientSecret: "REPLACE_ME_DISCORD_CLIENT_SECRET",
				RedirectURL:  "http://127.0.0.1:9120/portal/linapro-oidc-discord/callback",
				AuthorizeURL: "https://discord.com/oauth2/authorize",
				Scopes:       []string{"identify"},
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
			if !strings.Contains(err.Error(), "missing") {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestBuildAuthorizeURLSucceedsWithConfiguredCredentials(t *testing.T) {
	t.Parallel()

	svc := newAuthorizeTestService(Config{
		ClientID:     "real-discord-client-id",
		ClientSecret: "real-discord-client-secret",
		RedirectURL:  "http://127.0.0.1:9120/portal/linapro-oidc-discord/callback",
		AuthorizeURL: "https://discord.com/oauth2/authorize",
		Scopes:       []string{"identify", "email"},
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
	if query.Get("client_id") != "real-discord-client-id" {
		t.Fatalf("client_id = %q", query.Get("client_id"))
	}
	if query.Get("state") != "state:biz:/admin/auth/login" {
		t.Fatalf("state = %q", query.Get("state"))
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
