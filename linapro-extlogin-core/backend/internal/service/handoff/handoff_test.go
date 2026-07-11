// handoff_test.go covers handoff.go single-use exchange, invalid inputs, and
// host mapping for the process-local login handoff store.
package handoff

import (
	"testing"

	"lina-core/pkg/plugin/capability/authcap/extlogin"
	"lina-plugin-linapro-extlogin-core/backend/cap/extidcap"
)

func TestCreateAndExchangeSingleUse(t *testing.T) {
	t.Parallel()
	svc := New()
	code, err := svc.Create(extidcap.LoginHandoffPayload{
		AccessToken:  "access-1",
		RefreshToken: "refresh-1",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	first, err := svc.Exchange(code)
	if err != nil {
		t.Fatalf("first exchange: %v", err)
	}
	if first.AccessToken != "access-1" || first.RefreshToken != "refresh-1" {
		t.Fatalf("unexpected payload: %+v", first)
	}
	if _, err = svc.Exchange(code); err == nil {
		t.Fatal("expected second exchange to fail")
	}
}

func TestCreateRejectsEmptyTokens(t *testing.T) {
	t.Parallel()
	svc := New()
	if _, err := svc.Create(extidcap.LoginHandoffPayload{}); err == nil {
		t.Fatal("expected empty payload to fail")
	}
}

func TestCreateFromHost(t *testing.T) {
	t.Parallel()
	svc := New()
	if _, err := svc.CreateFromHost(nil); err == nil {
		t.Fatal("expected nil host output to fail")
	}
	code, err := svc.CreateFromHost(&extlogin.LoginOutput{
		PreToken: "pre-1",
		TenantCandidates: []extlogin.TenantCandidate{
			{ID: 7, Code: "acme", Name: "Acme", Status: "enabled"},
		},
	})
	if err != nil {
		t.Fatalf("create from host: %v", err)
	}
	payload, err := svc.Exchange(code)
	if err != nil {
		t.Fatalf("exchange: %v", err)
	}
	if payload.PreToken != "pre-1" || len(payload.TenantCandidates) != 1 || payload.TenantCandidates[0].ID != 7 {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestExchangeRejectsBlankCode(t *testing.T) {
	t.Parallel()
	svc := New()
	if _, err := svc.Exchange("  "); err == nil {
		t.Fatal("expected blank code to fail")
	}
}
