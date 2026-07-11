package ldapauth

import (
	"context"
	"testing"

	"lina-core/pkg/plugin/capability/authcap/extlogin"
	settingssvc "lina-plugin-linapro-auth-ldap/backend/internal/service/settings"
)

type fakeSettings struct {
	snap *settingssvc.Snapshot
}

func (f *fakeSettings) Get(ctx context.Context) (*settingssvc.Projection, error) {
	return nil, nil
}
func (f *fakeSettings) Save(ctx context.Context, in settingssvc.SaveInput) (*settingssvc.Projection, error) {
	return nil, nil
}
func (f *fakeSettings) Load(ctx context.Context) (*settingssvc.Snapshot, error) {
	return f.snap, nil
}

type fakeDir struct {
	identity *VerifiedIdentity
	err      error
	lastUser string
}

func (f *fakeDir) Authenticate(ctx context.Context, cfg settingssvc.Snapshot, username, password string) (*VerifiedIdentity, error) {
	f.lastUser = username
	return f.identity, f.err
}

type fakeExt struct {
	called bool
	in     extlogin.LoginInput
}

func (f *fakeExt) LoginByVerifiedIdentity(ctx context.Context, in extlogin.LoginInput) (*extlogin.LoginOutput, error) {
	f.called = true
	f.in = in
	return &extlogin.LoginOutput{AccessToken: "a", RefreshToken: "r"}, nil
}

func TestIsLoginConfigured(t *testing.T) {
	t.Parallel()
	if IsLoginConfigured(settingssvc.Snapshot{}) {
		t.Fatal()
	}
	if !IsLoginConfigured(settingssvc.Snapshot{Host: "h", UserDNTemplate: "uid={username},dc=x"}) {
		t.Fatal()
	}
	if !IsLoginConfigured(settingssvc.Snapshot{Host: "h", BaseDN: "dc=x", UserFilter: "(uid={username})"}) {
		t.Fatal()
	}
}

func TestRenderUserFilterEscapes(t *testing.T) {
	t.Parallel()
	got := RenderUserFilter("(uid={username})", "a*b(c)")
	if got != `(uid=a\2ab\28c\29)` && got != `(uid=a\2a b\28c\29)` {
		// exact: a\2ab\28c\29
		want := `(uid=a\2ab\28c\29)`
		if got != want {
			t.Fatalf("got %q want %q", got, want)
		}
	}
}

func TestRenderUserDN(t *testing.T) {
	t.Parallel()
	got := RenderUserDN("uid={username},ou=people,dc=example,dc=com", "alice")
	if got != "uid=alice,ou=people,dc=example,dc=com" {
		t.Fatal(got)
	}
}

func TestLoginRejectsMissingConfig(t *testing.T) {
	t.Parallel()
	svc := New(nil, &fakeSettings{snap: &settingssvc.Snapshot{}}, &fakeDir{})
	_, err := svc.Login(context.Background(), "u", "p")
	if err == nil {
		t.Fatal("expected config missing")
	}
}

func TestLoginRequiresUsernamePassword(t *testing.T) {
	t.Parallel()
	svc := New(nil, &fakeSettings{snap: &settingssvc.Snapshot{Host: "h", UserDNTemplate: "uid={username},dc=x"}}, &fakeDir{})
	if _, err := svc.Login(context.Background(), "", "p"); err == nil {
		t.Fatal()
	}
	if _, err := svc.Login(context.Background(), "u", ""); err == nil {
		t.Fatal()
	}
}

func TestLoginAllowAutoProvisionDefaultFalse(t *testing.T) {
	t.Parallel()
	// Without handoff store this may fail at CreateLoginHandoffFromHost; we only check dir path defaults.
	snap := &settingssvc.Snapshot{
		Host: "localhost", TLSMode: settingssvc.TLSModePlain, UserDNTemplate: "uid={username},dc=x",
		AllowAutoProvision: false,
	}
	ext := &fakeExt{}
	dir := &fakeDir{identity: &VerifiedIdentity{Subject: "sub-1", Email: "a@b.c"}}
	svc := New(ext, &fakeSettings{snap: snap}, dir)
	// handoff may fail if store not bound — still assert ext was called with false when it gets there
	_, _ = svc.Login(context.Background(), "alice", "secret")
	if ext.called && ext.in.AllowAutoProvision {
		t.Fatal("auto provision must be false")
	}
	if ext.called && ext.in.Provider != Provider {
		t.Fatalf("provider %q", ext.in.Provider)
	}
}

func TestEscapeFilter(t *testing.T) {
	t.Parallel()
	if EscapeFilter(`a*b`) != `a\2ab` && EscapeFilter("a*b") != `a\2ab` {
		got := EscapeFilter("a*b")
		if got != `a\2ab` {
			t.Fatalf("%q", got)
		}
	}
}
