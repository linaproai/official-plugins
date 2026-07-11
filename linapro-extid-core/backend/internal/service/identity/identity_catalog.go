// identity_catalog.go adapts the bound extidcap provider catalog facade for the
// identity service. Protocol plugins register through
// extidcap.RegisterProviderDescriptor; this file only reads the owner-bound
// catalog snapshot.
package identity

import (
	"context"

	"lina-plugin-linapro-extid-core/backend/cap/extidcap"
)

// ListProviders returns the process-local protocol catalog.
func (s *Service) listProvidersFromCap(_ context.Context, filter extidcap.ProviderFilter) []extidcap.ProviderView {
	return extidcap.SnapshotProviders(filter)
}

// getProviderFromCap returns one catalog entry.
func (s *Service) getProviderFromCap(_ context.Context, providerID string) (*extidcap.ProviderView, bool) {
	return extidcap.LookupProvider(providerID)
}

// RegisterProviderDescriptor is retained as a thin alias for protocol plugins
// that historically imported the identity package; new code should call
// extidcap.RegisterProviderDescriptor directly.
func RegisterProviderDescriptor(desc extidcap.ProviderDescriptor) {
	extidcap.RegisterProviderDescriptor(desc)
}
