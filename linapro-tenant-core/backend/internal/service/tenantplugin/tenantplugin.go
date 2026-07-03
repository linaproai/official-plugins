// Package tenantplugin implements tenant-scoped plugin enablement governance.
package tenantplugin

import (
	"context"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/plugincap"
	"lina-core/pkg/plugin/capability/tenantcap"
)

// Service defines tenant plugin-governance operations and cache revision updates.
type Service interface {
	// List returns tenant-controllable plugins with current tenant enablement for
	// ctx's tenant. It is read-only and returns database errors.
	List(ctx context.Context) (*ListOutput, error)
	// SetEnabled updates one tenant plugin enablement row for ctx's tenant, runs
	// lifecycle preconditions, and bumps the shared plugin-runtime cache revision.
	SetEnabled(ctx context.Context, pluginID string, enabled bool) error
	// ProvisionForTenant provisions missing default tenant plugin enablement for
	// one tenant and bumps runtime cache revision through the shared revision
	// table when it writes new rows. Existing tenant-owned enablement rows are
	// preserved so startup reconciliation cannot override explicit choices.
	ProvisionForTenant(ctx context.Context, tenantID int64) error
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	bizCtxSvc bizctxcap.Service
	tenantSvc tenantcap.Service
	plugins   plugincap.Service
}

// New creates and returns a tenant plugin governance service.
func New(
	bizCtxSvc bizctxcap.Service,
	tenantSvc tenantcap.Service,
	plugins plugincap.Service,
) Service {
	return &serviceImpl{
		bizCtxSvc: bizCtxSvc,
		tenantSvc: tenantSvc,
		plugins:   plugins,
	}
}

// Entity is the tenant plugin-governance projection.
type Entity struct {
	Id            string
	Name          string
	Version       string
	Type          string
	Description   string
	Installed     int
	Enabled       int
	ScopeNature   string
	InstallMode   string
	TenantEnabled int
}

// ListOutput defines tenant plugin list output.
type ListOutput struct {
	List  []*Entity
	Total int
}

// requirePlugincap verifies tenant plugin governance dependencies.
func (s *serviceImpl) requirePlugincap() error {
	if s == nil || s.plugins == nil || s.plugins.Registry() == nil {
		return bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", "plugin"))
	}
	return nil
}

// requirePluginGovernance verifies tenant plugin governance is reachable from the tenant domain.
func (s *serviceImpl) requirePluginGovernance() error {
	if s == nil || s.tenantSvc == nil || s.tenantSvc.Plugins() == nil {
		return bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", "tenant-plugin-governance"))
	}
	return nil
}
