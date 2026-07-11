// extidcap_catalog_test.go covers the provider catalog facade in
// extidcap_catalog.go, including pre-bind buffering used by protocol plugin
// init() registration.
package extidcap

import "testing"

type fakeCatalog struct {
	items map[string]ProviderDescriptor
}

func (f *fakeCatalog) Register(desc ProviderDescriptor) {
	if f.items == nil {
		f.items = make(map[string]ProviderDescriptor)
	}
	f.items[desc.ID] = desc
}

func (f *fakeCatalog) Snapshot(ProviderFilter) []ProviderView {
	out := make([]ProviderView, 0, len(f.items))
	for _, desc := range f.items {
		out = append(out, ProviderView{ProviderDescriptor: desc, Enabled: true, Configured: true})
	}
	return out
}

func (f *fakeCatalog) Lookup(providerID string) (*ProviderView, bool) {
	desc, ok := f.items[providerID]
	if !ok {
		return nil, false
	}
	view := ProviderView{ProviderDescriptor: desc, Enabled: true, Configured: true}
	return &view, true
}

func TestCatalogBuffersBeforeBind(t *testing.T) {
	BindCatalogService(nil)
	t.Cleanup(func() { BindCatalogService(nil) })

	RegisterProviderDescriptor(ProviderDescriptor{ID: "google", Order: 1})
	if got := SnapshotProviders(ProviderFilter{}); len(got) != 0 {
		t.Fatalf("expected empty snapshot before bind, got %+v", got)
	}

	fake := &fakeCatalog{}
	BindCatalogService(fake)
	if _, ok := LookupProvider("google"); !ok {
		t.Fatal("expected buffered descriptor flushed on bind")
	}
	RegisterProviderDescriptor(ProviderDescriptor{ID: "discord", Order: 2})
	if _, ok := LookupProvider("discord"); !ok {
		t.Fatal("expected live register after bind")
	}
}
