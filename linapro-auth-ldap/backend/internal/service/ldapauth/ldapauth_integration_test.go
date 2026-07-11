//go:build integration

// ldapauth_integration_test.go exercises real OpenLDAP bind authentication and
// the plugin Login path against a live directory started by CI.
//
// Required environment (defaults match hack/ci/ldap-mock):
//
//	LINA_CI_LDAP_HOST=127.0.0.1
//	LINA_CI_LDAP_PORT=1389
//	LINA_CI_LDAP_USER=alice
//	LINA_CI_LDAP_PASSWORD=alice-secret
package ldapauth

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"lina-core/pkg/plugin/capability/authcap/extlogin"
	settingssvc "lina-plugin-linapro-auth-ldap/backend/internal/service/settings"
	"lina-plugin-linapro-extlogin-core/backend/cap/extidcap"
)

type integrationSettings struct {
	snap *settingssvc.Snapshot
}

func (s integrationSettings) Get(context.Context) (*settingssvc.Projection, error) {
	return nil, nil
}

func (s integrationSettings) Save(context.Context, settingssvc.SaveInput) (*settingssvc.Projection, error) {
	return nil, nil
}

func (s integrationSettings) Load(context.Context) (*settingssvc.Snapshot, error) {
	return s.snap, nil
}

type integrationExtLogin struct {
	last extlogin.LoginInput
}

func (f *integrationExtLogin) LoginByVerifiedIdentity(
	_ context.Context,
	in extlogin.LoginInput,
) (*extlogin.LoginOutput, error) {
	f.last = in
	return &extlogin.LoginOutput{
		AccessToken:  "ci-access-token",
		RefreshToken: "ci-refresh-token",
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

func TestIntegrationLDAPDirectoryAuthenticateAndLogin(t *testing.T) {
	host := envOr("LINA_CI_LDAP_HOST", "127.0.0.1")
	port := envOr("LINA_CI_LDAP_PORT", "1389")
	user := envOr("LINA_CI_LDAP_USER", "alice")
	password := envOr("LINA_CI_LDAP_PASSWORD", "alice-secret")

	// ldap-mock (and optional bitnami OpenLDAP) seed users as cn=<user>,ou=users,<root>.
	// ldap-mock（以及可选的 bitnami OpenLDAP）将用户种子为 cn=<user>,ou=users,<root>。
	snap := &settingssvc.Snapshot{
		Host:               host,
		Port:               port,
		TLSMode:            settingssvc.TLSModePlain,
		UserDNTemplate:     "cn={username},ou=users,dc=example,dc=com",
		SubjectAttr:        "cn",
		EmailAttr:          "mail",
		DisplayNameAttr:    "cn",
		AllowAutoProvision: true,
	}

	waitForLDAP(t, *snap, user, password)

	client := NewLDAPDirectoryClient()
	identity, err := client.Authenticate(context.Background(), *snap, user, password)
	if err != nil {
		t.Fatalf("real directory authenticate failed: %v", err)
	}
	if identity == nil || identity.Subject == "" {
		t.Fatalf("expected verified subject from directory, got %+v", identity)
	}
	if !strings.EqualFold(identity.Subject, user) {
		t.Fatalf("subject = %q want %q", identity.Subject, user)
	}

	// Wrong password must fail closed against the live directory.
	// 错误密码必须在真实目录上 fail-closed。
	if _, err := client.Authenticate(context.Background(), *snap, user, "wrong-password"); err == nil {
		t.Fatal("expected wrong password to fail against live LDAP")
	}

	handoffSvc := newMemoryHandoff()
	extidcap.BindHandoffService(handoffSvc)
	t.Cleanup(func() { extidcap.BindHandoffService(nil) })

	ext := &integrationExtLogin{}
	loginSvc := New(ext, integrationSettings{snap: snap}, client)
	out, err := loginSvc.Login(context.Background(), user, password)
	if err != nil {
		t.Fatalf("plugin Login against live LDAP failed: %v", err)
	}
	if out == nil || strings.TrimSpace(out.Handoff) == "" {
		t.Fatalf("expected non-empty handoff, got %+v", out)
	}
	if ext.last.Provider != Provider {
		t.Fatalf("external login provider = %q want %q", ext.last.Provider, Provider)
	}
	if ext.last.Subject != identity.Subject {
		t.Fatalf("external login subject = %q want %q", ext.last.Subject, identity.Subject)
	}
	if !ext.last.AllowAutoProvision {
		t.Fatal("expected AllowAutoProvision true from settings")
	}

	payload, err := handoffSvc.Exchange(out.Handoff)
	if err != nil {
		t.Fatalf("exchange handoff: %v", err)
	}
	if payload == nil || payload.AccessToken != "ci-access-token" {
		t.Fatalf("unexpected handoff payload: %+v", payload)
	}
}

func waitForLDAP(t *testing.T, snap settingssvc.Snapshot, user, password string) {
	t.Helper()
	client := NewLDAPDirectoryClient()
	deadline := time.Now().Add(60 * time.Second)
	var lastErr error
	for time.Now().Before(deadline) {
		_, lastErr = client.Authenticate(context.Background(), snap, user, password)
		if lastErr == nil {
			return
		}
		time.Sleep(1 * time.Second)
	}
	t.Fatalf("LDAP not ready within timeout (last error: %v)", lastErr)
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
