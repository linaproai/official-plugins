// lifecycleprecondition_impl.go implements plugin lifecycle guards for tenant
// disable and tenant delete operations. It checks tenant-plugin bindings before
// allowing destructive lifecycle steps so plugin state cannot be orphaned.

package lifecycleprecondition

import (
	"context"

	"lina-core/pkg/pluginhost"
)

// BeforeUninstall rejects uninstall while tenants exist.
func (c *Checker) BeforeUninstall(
	ctx context.Context,
	input pluginhost.SourcePluginLifecycleInput,
) (bool, string, error) {
	count, err := c.tenantCounter.CountExisting(ctx)
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
	count, err := c.tenantCounter.CountExisting(ctx)
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
