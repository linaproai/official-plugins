// providercatalog_test.go covers providercatalog.go registration, lookup,
// filtering, and ordering.
package providercatalog

import (
	"testing"

	"lina-plugin-linapro-extlogin-core/backend/cap/extidcap"
)

func TestRegisterLookupAndOverwrite(t *testing.T) {
	t.Parallel()
	svc := New()
	svc.Register(extidcap.ProviderDescriptor{ID: " google ", PluginID: " p1 ", Order: 2})
	view, ok := svc.Lookup("google")
	if !ok || view == nil {
		t.Fatal("expected lookup hit")
	}
	if view.ID != "google" || view.PluginID != "p1" || !view.Enabled || !view.Configured {
		t.Fatalf("unexpected view: %+v", view)
	}
	svc.Register(extidcap.ProviderDescriptor{ID: "google", PluginID: "p2", Order: 1})
	view, ok = svc.Lookup("google")
	if !ok || view.PluginID != "p2" || view.Order != 1 {
		t.Fatalf("expected overwrite: %+v ok=%v", view, ok)
	}
}

func TestSnapshotFilterAndOrder(t *testing.T) {
	t.Parallel()
	svc := New()
	svc.Register(extidcap.ProviderDescriptor{
		ID: "b", Order: 20,
		Capabilities: extidcap.ProviderCapabilities{Login: true, Bind: false},
	})
	svc.Register(extidcap.ProviderDescriptor{
		ID: "a", Order: 10,
		Capabilities: extidcap.ProviderCapabilities{Login: true, Bind: true},
	})
	svc.Register(extidcap.ProviderDescriptor{
		ID: "c", Order: 10,
		Capabilities: extidcap.ProviderCapabilities{Login: false, Bind: true},
	})

	all := svc.Snapshot(extidcap.ProviderFilter{})
	if len(all) != 3 || all[0].ID != "a" || all[1].ID != "c" || all[2].ID != "b" {
		t.Fatalf("unexpected order: %+v", all)
	}

	bindOnly := svc.Snapshot(extidcap.ProviderFilter{SupportsBind: true})
	if len(bindOnly) != 2 || bindOnly[0].ID != "a" || bindOnly[1].ID != "c" {
		t.Fatalf("unexpected bind filter: %+v", bindOnly)
	}

	loginOnly := svc.Snapshot(extidcap.ProviderFilter{SupportsLogin: true})
	if len(loginOnly) != 2 || loginOnly[0].ID != "a" || loginOnly[1].ID != "b" {
		t.Fatalf("unexpected login filter: %+v", loginOnly)
	}
}

func TestRegisterIgnoresEmptyID(t *testing.T) {
	t.Parallel()
	svc := New()
	svc.Register(extidcap.ProviderDescriptor{ID: "  "})
	if got := svc.Snapshot(extidcap.ProviderFilter{}); len(got) != 0 {
		t.Fatalf("expected empty catalog, got %+v", got)
	}
}
