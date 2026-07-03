// Package provider implements the linapro-tenant-core plugin provider adapter.
package provider

import (
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
	"lina-plugin-linapro-tenant-core/backend/internal/service/membership"
	"lina-plugin-linapro-tenant-core/backend/internal/service/resolver"
	"lina-plugin-linapro-tenant-core/backend/internal/service/resolverconfig"
	"lina-plugin-linapro-tenant-core/backend/internal/service/tenantplugin"
)

// Provider is the plugin-owned tenant capability provider. It mirrors the
// framework tenant capability contract while keeping storage access internal.
type Provider struct {
	membershipSvc     membership.Service
	resolverSvc       resolver.Service
	resolverConfigSvc resolverconfig.Service
	tenantPluginSvc   tenantplugin.Service
}

// Ensure Provider implements the host tenant capability provider contract.
var (
	_ tenantspi.Provider               = (*Provider)(nil)
	_ tenantspi.UserMembershipProvider = (*Provider)(nil)
	_ tenantspi.DirectoryProvider      = (*Provider)(nil)
)

// New creates and returns a Provider instance.
func New(
	membershipSvc membership.Service,
	resolverSvc resolver.Service,
	resolverConfigSvc resolverconfig.Service,
	tenantPluginSvc tenantplugin.Service,
) (*Provider, error) {
	if membershipSvc == nil {
		return nil, gerror.New("linapro-tenant-core provider requires membership service")
	}
	if resolverSvc == nil {
		return nil, gerror.New("linapro-tenant-core provider requires resolver service")
	}
	if resolverConfigSvc == nil {
		return nil, gerror.New("linapro-tenant-core provider requires resolver config service")
	}
	if tenantPluginSvc == nil {
		return nil, gerror.New("linapro-tenant-core provider requires tenant plugin service")
	}
	return &Provider{
		membershipSvc:     membershipSvc,
		resolverSvc:       resolverSvc,
		resolverConfigSvc: resolverConfigSvc,
		tenantPluginSvc:   tenantPluginSvc,
	}, nil
}
