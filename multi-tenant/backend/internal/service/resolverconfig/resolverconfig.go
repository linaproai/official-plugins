// Package resolverconfig exposes the code-owned tenant resolver policy.
package resolverconfig

import (
	"context"
)

// Service defines resolver policy operations.
type Service interface {
	// Get returns the built-in resolver policy as a detached value. It is
	// code-owned, read-only, and does not consult runtime config, cache, or i18n.
	Get(ctx context.Context) (*Config, error)
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct{}

// New creates and returns a resolver policy service.
func New() Service {
	return &serviceImpl{}
}

// Config is the API/service projection of the built-in resolver policy.
type Config struct {
	Chain              []string `json:"chain"`
	ReservedSubdomains []string `json:"reservedSubdomains"`
	RootDomain         string   `json:"rootDomain"`
	OnAmbiguous        string   `json:"onAmbiguous"`
	Version            int64    `json:"version"`
}
