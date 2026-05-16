// Package lifecycleprecondition implements multi-tenant plugin lifecycle preconditions.
package lifecycleprecondition

import (
	"context"

	"lina-core/pkg/pluginhost"
	"lina-plugin-multi-tenant/backend/internal/service/tenant"
)

const (
	// ReasonUninstallTenantsExist is returned when existing tenants block uninstall.
	ReasonUninstallTenantsExist = "plugin.multi-tenant.uninstall_blocked.tenants_exist"
	// ReasonDisableTenantsExist is returned when existing tenants block disable.
	ReasonDisableTenantsExist = "plugin.multi-tenant.disable_blocked.tenants_exist"
)

// Checker implements plugin-owned lifecycle precondition checks.
type Checker struct {
	tenantSvc tenant.Service
}

// New creates and returns a lifecycle precondition checker.
func New(tenantSvc tenant.Service) *Checker {
	return &Checker{tenantSvc: tenantSvc}
}

// BeforeUninstall rejects uninstall while tenants exist.
func (c *Checker) BeforeUninstall(
	ctx context.Context,
	input pluginhost.SourcePluginLifecycleInput,
) (bool, string, error) {
	count, err := c.tenantSvc.CountExisting(ctx)
	if err != nil {
		return false, ReasonUninstallTenantsExist, err
	}
	if count > 0 {
		return false, ReasonUninstallTenantsExist, nil
	}
	return true, "", nil
}

// BeforeDisable rejects global disable while tenants exist.
func (c *Checker) BeforeDisable(
	ctx context.Context,
	input pluginhost.SourcePluginLifecycleInput,
) (bool, string, error) {
	count, err := c.tenantSvc.CountExisting(ctx)
	if err != nil {
		return false, ReasonDisableTenantsExist, err
	}
	if count > 0 {
		return false, ReasonDisableTenantsExist, nil
	}
	return true, "", nil
}

// BeforeTenantDelete reserves the cross-plugin tenant delete precondition surface.
func (c *Checker) BeforeTenantDelete(
	ctx context.Context,
	input pluginhost.SourcePluginTenantLifecycleInput,
) (bool, string, error) {
	return true, "", nil
}
