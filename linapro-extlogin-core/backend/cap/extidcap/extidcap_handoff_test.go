// extidcap_handoff_test.go covers the process-bound handoff facade in
// extidcap_handoff.go, including the unbound failure path used when the owner
// plugin has not bound a store yet.
package extidcap

import (
	"testing"

	"lina-core/pkg/plugin/capability/authcap/extlogin"
)

type fakeHandoff struct {
	created int
	last    LoginHandoffPayload
}

func (f *fakeHandoff) Create(payload LoginHandoffPayload) (string, error) {
	f.created++
	f.last = payload
	return "code-1", nil
}

func (f *fakeHandoff) CreateFromHost(out *extlogin.LoginOutput) (string, error) {
	if out == nil {
		return "", errHandoffInvalid()
	}
	return f.Create(LoginHandoffPayload{AccessToken: out.AccessToken, PreToken: out.PreToken})
}

func (f *fakeHandoff) Exchange(code string) (*LoginHandoffPayload, error) {
	if code == "" {
		return nil, errHandoffInvalid()
	}
	return &LoginHandoffPayload{AccessToken: "a"}, nil
}

func TestHandoffFacadeRequiresBind(t *testing.T) {
	BindHandoffService(nil)
	t.Cleanup(func() { BindHandoffService(nil) })

	if _, err := CreateLoginHandoff(LoginHandoffPayload{AccessToken: "x"}); err == nil {
		t.Fatal("expected unbound create to fail")
	}

	fake := &fakeHandoff{}
	BindHandoffService(fake)
	code, err := CreateLoginHandoff(LoginHandoffPayload{AccessToken: "x"})
	if err != nil || code != "code-1" || fake.created != 1 {
		t.Fatalf("bound create failed: code=%s err=%v created=%d", code, err, fake.created)
	}
}
