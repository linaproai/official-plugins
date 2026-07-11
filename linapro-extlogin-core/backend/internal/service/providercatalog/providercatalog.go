// Package providercatalog implements the process-local protocol provider
// catalog owned by linapro-extlogin-core. Protocol plugins register descriptors
// through the extidcap public facade; the identity service reads filtered
// snapshots without importing other plugins' internals.
//
// Runtime boundary: entries live in the current host process only. Without a
// persistent admin configuration store, every registered descriptor is treated
// as both Enabled and Configured (registration presence). Admin enablement and
// credential-configured flags are a later hardening step.
package providercatalog

import (
	"sort"
	"strings"
	"sync"

	"lina-plugin-linapro-extlogin-core/backend/cap/extidcap"
)

// Ensure Service implements the public catalog contract.
var _ extidcap.CatalogService = (*Service)(nil)

// Service is the process-local provider catalog registry.
type Service struct {
	mu    sync.RWMutex
	items map[string]extidcap.ProviderDescriptor
}

// New creates an empty provider catalog.
func New() *Service {
	return &Service{items: make(map[string]extidcap.ProviderDescriptor)}
}

// Register records one protocol provider. Empty IDs are ignored.
func (s *Service) Register(desc extidcap.ProviderDescriptor) {
	if s == nil {
		return
	}
	id := strings.TrimSpace(desc.ID)
	if id == "" {
		return
	}
	desc.ID = id
	desc.PluginID = strings.TrimSpace(desc.PluginID)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[id] = desc
}

// Snapshot returns a filtered copy of registered providers ordered by Order, then ID.
func (s *Service) Snapshot(filter extidcap.ProviderFilter) []extidcap.ProviderView {
	if s == nil {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]extidcap.ProviderView, 0, len(s.items))
	for _, desc := range s.items {
		// Without persistent config, registration implies enabled+configured.
		view := extidcap.ProviderView{ProviderDescriptor: desc, Enabled: true, Configured: true}
		if filter.EnabledOnly && !view.Enabled {
			continue
		}
		if filter.ConfiguredOnly && !view.Configured {
			continue
		}
		if filter.SupportsBind && !desc.Capabilities.Bind {
			continue
		}
		if filter.SupportsLogin && !desc.Capabilities.Login {
			continue
		}
		out = append(out, view)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Order == out[j].Order {
			return out[i].ID < out[j].ID
		}
		return out[i].Order < out[j].Order
	})
	return out
}

// Lookup returns one catalog entry by provider ID.
func (s *Service) Lookup(providerID string) (*extidcap.ProviderView, bool) {
	if s == nil {
		return nil, false
	}
	providerID = strings.TrimSpace(providerID)
	s.mu.RLock()
	defer s.mu.RUnlock()
	desc, ok := s.items[providerID]
	if !ok {
		return nil, false
	}
	view := extidcap.ProviderView{ProviderDescriptor: desc, Enabled: true, Configured: true}
	return &view, true
}
