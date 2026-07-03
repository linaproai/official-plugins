// Package lifecycleprecondition implements linapro-tenant-core plugin lifecycle preconditions.
package lifecycleprecondition

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
)

const (
	// ReasonUninstallTenantsExist is returned when existing tenants block uninstall.
	ReasonUninstallTenantsExist = "plugin.linapro-tenant-core.uninstall_blocked.tenants_exist"
	// ReasonDisableTenantsExist is returned when existing tenants block disable.
	ReasonDisableTenantsExist = "plugin.linapro-tenant-core.disable_blocked.tenants_exist"
)

// tenantCounter counts tenants relevant to lifecycle precondition decisions.
type tenantCounter interface {
	// CountExisting returns the number of non-deleted tenants that make uninstall,
	// disable, or tenant-delete operations unsafe. It returns storage errors from
	// the authoritative tenant table.
	CountExisting(ctx context.Context) (int, error)
}

// Checker implements plugin-owned lifecycle precondition checks without changing
// tenant data, cache state, i18n resources, or plugin registration.
type Checker struct {
	tenantCounter tenantCounter
}

// New creates and returns a lifecycle precondition checker.
func New(tenantCounter tenantCounter) (*Checker, error) {
	if tenantCounter == nil {
		return nil, gerror.New("linapro-tenant-core lifecycle precondition requires tenant counter")
	}
	return &Checker{tenantCounter: tenantCounter}, nil
}
