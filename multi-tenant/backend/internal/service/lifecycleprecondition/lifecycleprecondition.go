// Package lifecycleprecondition implements multi-tenant plugin lifecycle preconditions.
package lifecycleprecondition

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/pluginhost"
)

const (
	// ReasonUninstallTenantsExist is returned when existing tenants block uninstall.
	ReasonUninstallTenantsExist = "plugin.multi-tenant.uninstall_blocked.tenants_exist"
	// ReasonDisableTenantsExist is returned when existing tenants block disable.
	ReasonDisableTenantsExist = "plugin.multi-tenant.disable_blocked.tenants_exist"
)

// TenantCounter counts tenants relevant to lifecycle precondition decisions.
type TenantCounter interface {
	// CountExisting returns the number of non-deleted tenants.
	CountExisting(ctx context.Context) (int, error)
}

// Checker implements plugin-owned lifecycle precondition checks.
type Checker struct {
	tenantCounter TenantCounter
}

// New creates and returns a lifecycle precondition checker.
func New(tenantCounter TenantCounter) (*Checker, error) {
	if tenantCounter == nil {
		return nil, gerror.New("multi-tenant lifecycle precondition requires tenant counter")
	}
	return &Checker{tenantCounter: tenantCounter}, nil
}

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
