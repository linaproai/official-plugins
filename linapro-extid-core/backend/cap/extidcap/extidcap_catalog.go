// extidcap_catalog.go publishes the provider catalog contract and a process-bound
// facade. The registry implementation lives in backend/internal/service/providercatalog
// and is bound by the owner plugin at startup. Protocol plugins may call
// RegisterProviderDescriptor from init(); descriptors registered before bind
// are buffered and flushed when the owner binds the catalog service.

package extidcap

import "sync"

// CatalogService is the owner-plugin runtime for protocol provider discovery.
type CatalogService interface {
	// Register records one protocol provider. Empty IDs are ignored.
	Register(desc ProviderDescriptor)
	// Snapshot returns a filtered copy of registered providers.
	Snapshot(filter ProviderFilter) []ProviderView
	// Lookup returns one catalog entry by provider ID.
	Lookup(providerID string) (*ProviderView, bool)
}

var (
	catalogMu      sync.Mutex
	boundCatalog   CatalogService
	pendingCatalog []ProviderDescriptor
)

// BindCatalogService wires the owner plugin catalog implementation and flushes
// any descriptors registered before bind. Passing nil clears the binding and
// keeps subsequent Register calls buffered again.
func BindCatalogService(svc CatalogService) {
	catalogMu.Lock()
	defer catalogMu.Unlock()
	boundCatalog = svc
	if svc == nil {
		return
	}
	for _, desc := range pendingCatalog {
		svc.Register(desc)
	}
	pendingCatalog = nil
}

// BoundCatalogService returns the currently bound catalog implementation.
func BoundCatalogService() CatalogService {
	catalogMu.Lock()
	defer catalogMu.Unlock()
	return boundCatalog
}

// RegisterProviderDescriptor records one protocol provider in the shared catalog.
// Safe to call from protocol plugin init() before the owner plugin binds.
func RegisterProviderDescriptor(desc ProviderDescriptor) {
	catalogMu.Lock()
	defer catalogMu.Unlock()
	if boundCatalog != nil {
		boundCatalog.Register(desc)
		return
	}
	pendingCatalog = append(pendingCatalog, desc)
}

// SnapshotProviders returns a filtered copy of registered providers.
func SnapshotProviders(filter ProviderFilter) []ProviderView {
	catalogMu.Lock()
	svc := boundCatalog
	catalogMu.Unlock()
	if svc == nil {
		return nil
	}
	return svc.Snapshot(filter)
}

// LookupProvider returns one catalog entry.
func LookupProvider(providerID string) (*ProviderView, bool) {
	catalogMu.Lock()
	svc := boundCatalog
	catalogMu.Unlock()
	if svc == nil {
		return nil, false
	}
	return svc.Lookup(providerID)
}
