// resolverconfig_impl.go implements resolver configuration validation,
// persistence, ordering, and cache-friendly lookup for the multi-tenant plugin.
// It keeps strategy validation aligned with resolver behavior so invalid
// tenant-detection rules are rejected before runtime use.

package resolverconfig

import (
	"context"

	"lina-plugin-multi-tenant/backend/internal/service/resolver"
	"lina-plugin-multi-tenant/backend/internal/service/shared"
)

// Get returns the code-owned resolver policy.
func (s *serviceImpl) Get(ctx context.Context) (*Config, error) {
	return defaultConfig(), nil
}

// ToResolverConfig returns the resolver package config from built-in policy.
func ToResolverConfig(config *Config) resolver.Config {
	// The first multi-tenant iteration intentionally keeps resolver behavior
	// code-owned. Even if an internal caller passes an edited Config value, the
	// active chain remains override -> jwt -> session -> header -> subdomain ->
	// default, subdomain resolution remains disabled by the empty root domain,
	// and ambiguous requests keep prompting for explicit tenant selection.
	defaults := defaultConfig()
	return resolver.Config{
		Chain:              cloneStrings(defaults.Chain),
		ReservedSubdomains: cloneStrings(defaults.ReservedSubdomains),
		RootDomain:         shared.DefaultRootDomain,
		OnAmbiguous:        shared.OnAmbiguousPrompt,
	}
}

// defaultConfig returns the built-in resolver configuration.
func defaultConfig() *Config {
	return &Config{
		Chain:              shared.DefaultResolverChain(),
		ReservedSubdomains: shared.DefaultReservedSubdomains(),
		RootDomain:         shared.DefaultRootDomain,
		OnAmbiguous:        shared.OnAmbiguousPrompt,
		Version:            1,
	}
}

// cloneStrings returns a detached copy of string slices stored in the policy.
func cloneStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	cloned := make([]string, len(values))
	copy(cloned, values)
	return cloned
}
