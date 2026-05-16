// Package lifecycleprecondition implements multi-tenant plugin lifecycle preconditions.
package lifecycleprecondition

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
)

const (
	// ReasonUninstallTenantsExist is returned when existing tenants block uninstall.
	ReasonUninstallTenantsExist = "plugin.multi-tenant.uninstall_blocked.tenants_exist"
	// ReasonDisableTenantsExist is returned when existing tenants block disable.
	ReasonDisableTenantsExist = "plugin.multi-tenant.disable_blocked.tenants_exist"
)

// TenantCounter counts tenants relevant to lifecycle precondition decisions.
type TenantCounter interface {
	// CountExisting returns the number of non-deleted tenants that make uninstall,
	// disable, or tenant-delete operations unsafe. It returns storage errors from
	// the authoritative tenant table.
	CountExisting(ctx context.Context) (int, error)
}

// Checker implements plugin-owned lifecycle precondition checks without changing
// tenant data, cache state, i18n resources, or plugin registration.
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
