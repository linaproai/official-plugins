// Package provider implements the multi-tenant plugin provider adapter.
package provider

import (
	pkgtenantcap "lina-core/pkg/tenantcap"
	"lina-plugin-multi-tenant/backend/internal/service/membership"
	"lina-plugin-multi-tenant/backend/internal/service/resolver"
	"lina-plugin-multi-tenant/backend/internal/service/resolverconfig"
)

// Provider is the plugin-owned tenant capability provider. It mirrors the
// host tenantcap contract so registration can be wired once the host seam lands.
type Provider struct {
	membershipSvc     membership.Service
	resolverSvc       resolver.Service
	resolverConfigSvc resolverconfig.Service
}

// Ensure Provider implements the host tenant capability provider contract.
var _ pkgtenantcap.Provider = (*Provider)(nil)
var _ pkgtenantcap.UserMembershipProvider = (*Provider)(nil)

// New creates and returns a Provider instance.
func New(membershipSvc membership.Service, resolverSvc resolver.Service, resolverConfigSvc resolverconfig.Service) *Provider {
	return &Provider{
		membershipSvc:     membershipSvc,
		resolverSvc:       resolverSvc,
		resolverConfigSvc: resolverConfigSvc,
	}
}
