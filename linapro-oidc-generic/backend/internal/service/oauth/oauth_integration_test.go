//go:build integration

// oauth_integration_test.go exercises a real OIDC code+PKCE+id_token login path
// against the CI oidc-mock provider (hack/ci/oidc-mock).
//
// Required environment (defaults match CI):
//
//	LINA_CI_OIDC_ISSUER=http://127.0.0.1:18080
//	LINA_CI_OIDC_CLIENT_ID=linapro-ci
//	LINA_CI_OIDC_CLIENT_SECRET=linapro-ci-secret
//	LINA_CI_OIDC_REDIRECT_URL=http://127.0.0.1:9999/portal/linapro-oidc-generic/callback
package oauth

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"lina-core/pkg/plugin/capability/authcap/extlogin"
	"lina-plugin-linapro-extlogin-core/backend/cap/extidcap"
)

type integrationExtLogin struct {
	last extlogin.LoginInput
}

func (f *integrationExtLogin) LoginByVerifiedIdentity(
	_ context.Context,
	in extlogin.LoginInput,
) (*extlogin.LoginOutput, error) {
	f.last = in
	return &extlogin.LoginOutput{
		AccessToken:  "ci-oidc-access",
		RefreshToken: "ci-oidc-refresh",
	}, nil
}

// memoryHandoff is a process-local handoff store for integration tests only.
// memoryHandoff 仅用于集成测试的进程内 handoff 存储。
type memoryHandoff struct {
	mu    sync.Mutex
	items map[string]extidcap.LoginHandoffPayload
	seq   int
}

func newMemoryHandoff() *memoryHandoff {
	return &memoryHandoff{items: map[string]extidcap.LoginHandoffPayload{}}
}

func (m *memoryHandoff) Create(payload extidcap.LoginHandoffPayload) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.seq++
	code := fmt.Sprintf("ci_handoff_%d", m.seq)
	m.items[code] = payload
	return code, nil
}

func (m *memoryHandoff) CreateFromHost(out *extlogin.LoginOutput) (string, error) {
	if out == nil {
		return "", fmt.Errorf("nil host login output")
	}
	return m.Create(extidcap.LoginHandoffPayload{
		AccessToken:      out.AccessToken,
		RefreshToken:     out.RefreshToken,
		PreToken:         out.PreToken,
		TenantCandidates: out.TenantCandidates,
	})
}

func (m *memoryHandoff) Exchange(code string) (*extidcap.LoginHandoffPayload, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	payload, ok := m.items[code]
	if !ok {
		return nil, fmt.Errorf("unknown handoff")
	}
	delete(m.items, code)
	return &payload, nil
}

func TestIntegrationOIDCAuthorizeCallbackLogin(t *testing.T) {
	issuer := strings.TrimRight(envOr("LINA_CI_OIDC_ISSUER", "http://127.0.0.1:18080"), "/")
	clientID := envOr("LINA_CI_OIDC_CLIENT_ID", "linapro-ci")
	clientSecret := envOr("LINA_CI_OIDC_CLIENT_SECRET", "linapro-ci-secret")
	redirectURL := envOr(
		"LINA_CI_OIDC_REDIRECT_URL",
		"http://127.0.0.1:9999/portal/linapro-oidc-generic/callback",
	)

	waitForOIDC(t, issuer)

	handoffSvc := newMemoryHandoff()
	extidcap.BindHandoffService(handoffSvc)
	t.Cleanup(func() { extidcap.BindHandoffService(nil) })

	ext := &integrationExtLogin{}
	cfg := Config{
		Issuer:          issuer,
		ClientID:        clientID,
		ClientSecret:    clientSecret,
		RedirectURL:     redirectURL,
		Scopes:          []string{"openid", "email", "profile"},
		LoginReturnPath: "/admin/auth/login",
	}
	svc := New(ext, NewConfigResolver(nil, cfg), NewHMACStateCodec())

	// 1) Plugin builds authorize URL with real discovery + PKCE.
	// 1) 插件通过真实 discovery + PKCE 构造 authorize URL。
	authReq, err := svc.BuildAuthorizeURL(context.Background(), "ci-state-key", "/admin/dashboard")
	if err != nil {
		t.Fatalf("BuildAuthorizeURL against live IdP failed: %v", err)
	}
	if !strings.HasPrefix(authReq.URL, issuer+"/authorize") {
		t.Fatalf("authorize URL = %q, want prefix %s/authorize", authReq.URL, issuer)
	}
	if strings.TrimSpace(authReq.State) == "" {
		t.Fatal("expected non-empty signed state")
	}

	// 2) Follow authorize (oidc-mock auto-login) and capture callback code.
	// 2) 跟随 authorize（oidc-mock 自动登录）并捕获 callback code。
	code, state := followAuthorizeForCode(t, authReq.URL, redirectURL)
	if code == "" || state == "" {
		t.Fatalf("missing code/state from IdP callback: code=%q state=%q", code, state)
	}
	if state != authReq.State {
		t.Fatalf("callback state mismatch: got %q want %q", state, authReq.State)
	}

	// 3) Plugin completes callback: token exchange, id_token verify, handoff.
	// 3) 插件完成 callback：换 token、校验 id_token、生成 handoff。
	out, err := svc.CompleteCallback(context.Background(), CallbackInput{
		Code:  code,
		State: state,
	})
	if err != nil {
		t.Fatalf("CompleteCallback against live IdP failed: %v", err)
	}
	if out == nil || strings.TrimSpace(out.Handoff) == "" {
		t.Fatalf("expected non-empty handoff, got %+v", out)
	}
	if out.ReturnTo != "/admin/dashboard" {
		t.Fatalf("returnTo = %q want /admin/dashboard", out.ReturnTo)
	}
	if ext.last.Provider != Provider {
		t.Fatalf("external login provider = %q want %q", ext.last.Provider, Provider)
	}
	if strings.TrimSpace(ext.last.Subject) == "" {
		t.Fatal("expected verified subject from id_token")
	}
	if !strings.Contains(ext.last.Email, "@") {
		t.Fatalf("expected email claim, got %q", ext.last.Email)
	}

	payload, err := handoffSvc.Exchange(out.Handoff)
	if err != nil {
		t.Fatalf("exchange handoff: %v", err)
	}
	if payload == nil || payload.AccessToken != "ci-oidc-access" {
		t.Fatalf("unexpected handoff payload: %+v", payload)
	}

	// Replayed code must fail (single-use auth code).
	// 重放 code 必须失败（授权码一次性）。
	if _, err := svc.CompleteCallback(context.Background(), CallbackInput{
		Code:  code,
		State: state,
	}); err == nil {
		t.Fatal("expected replayed authorization code to fail")
	}
}

func followAuthorizeForCode(t *testing.T, authorizeURL, redirectURL string) (code, state string) {
	t.Helper()
	client := &http.Client{
		Timeout: 15 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Stop when the IdP redirects back to the plugin callback.
			// 当 IdP 重定向回插件 callback 时停止跟随。
			if strings.HasPrefix(req.URL.String(), redirectURL) {
				return http.ErrUseLastResponse
			}
			if len(via) >= 5 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}
	resp, err := client.Get(authorizeURL)
	if err != nil {
		t.Fatalf("GET authorize: %v", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	location := resp.Header.Get("Location")
	if location == "" {
		t.Fatalf("authorize response missing Location (status=%d)", resp.StatusCode)
	}
	u, err := url.Parse(location)
	if err != nil {
		t.Fatalf("parse callback location: %v", err)
	}
	return u.Query().Get("code"), u.Query().Get("state")
}

func waitForOIDC(t *testing.T, issuer string) {
	t.Helper()
	client := &http.Client{Timeout: 2 * time.Second}
	discovery := issuer + "/.well-known/openid-configuration"
	deadline := time.Now().Add(60 * time.Second)
	var lastErr error
	for time.Now().Before(deadline) {
		resp, err := client.Get(discovery)
		if err == nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
			lastErr = errStatus(resp.StatusCode)
		} else {
			lastErr = err
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatalf("OIDC discovery not ready within timeout (last error: %v)", lastErr)
}

type statusError int

func (e statusError) Error() string { return "status " + http.StatusText(int(e)) }

func errStatus(code int) error { return statusError(code) }

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
