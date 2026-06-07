// Package tenantadapter adapts linapro-tenant-core internal services to the
// framework tenant capability provider contract.
package tenantadapter

import (
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/plugincap"
	"lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/capability/usercap"
	"lina-plugin-linapro-tenant-core/backend/internal/service/membership"
	"lina-plugin-linapro-tenant-core/backend/internal/service/provider"
	"lina-plugin-linapro-tenant-core/backend/internal/service/resolver"
	"lina-plugin-linapro-tenant-core/backend/internal/service/resolverconfig"
	"lina-plugin-linapro-tenant-core/backend/internal/service/tenantplugin"
)

// New creates the tenant framework capability provider from host-published services.
func New(
	bizCtxSvc bizctxcap.Service,
	pluginLifecycleSvc plugincap.LifecycleService,
	users usercap.Service,
	plugins plugincap.Service,
	pluginAdmin plugincap.AdminService,
) (tenantcap.Provider, error) {
	if bizCtxSvc == nil {
		return nil, gerror.New("linapro-tenant-core provider requires host bizctx service")
	}
	if users == nil {
		return nil, gerror.New("linapro-tenant-core provider requires host user capability service")
	}
	if plugins == nil || pluginAdmin == nil {
		return nil, gerror.New("linapro-tenant-core provider requires host plugin capability services")
	}
	membershipSvc := membership.New(bizCtxSvc, users)
	resolverConfigSvc := resolverconfig.New()
	tenantPluginSvc := tenantplugin.New(bizCtxSvc, pluginLifecycleSvc, plugins, pluginAdmin)
	resolverSvc := resolver.New(bizCtxSvc, membershipSvc)
	return provider.New(membershipSvc, resolverSvc, resolverConfigSvc, tenantPluginSvc)
}
