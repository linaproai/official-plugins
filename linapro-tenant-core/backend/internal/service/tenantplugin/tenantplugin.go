// Package tenantplugin implements tenant-scoped plugin enablement governance.
package tenantplugin

import (
	"context"
	"strconv"
	"time"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/plugincap"
)

const (
	// tenantPluginCapabilityPluginID is the caller identifier used in plugincap contexts.
	tenantPluginCapabilityPluginID = "linapro-tenant-core"
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
	bizCtxSvc          bizctxcap.Service
	pluginLifecycleSvc plugincap.LifecycleService
	plugins            plugincap.Service
	pluginAdmin        plugincap.AdminService
}

// New creates and returns a tenant plugin governance service.
func New(
	bizCtxSvc bizctxcap.Service,
	pluginLifecycleSvc plugincap.LifecycleService,
	plugins plugincap.Service,
	pluginAdmin plugincap.AdminService,
) Service {
	return &serviceImpl{
		bizCtxSvc:          bizCtxSvc,
		pluginLifecycleSvc: pluginLifecycleSvc,
		plugins:            plugins,
		pluginAdmin:        pluginAdmin,
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

// capabilityContext builds plugin-visible metadata for tenant plugin
// governance calls into host-owned plugincap.
func (s *serviceImpl) capabilityContext(ctx context.Context, tenantID int64, resource string) capmodel.CapabilityContext {
	current := bizctxcap.CurrentContext{}
	if s != nil && s.bizCtxSvc != nil {
		current = s.bizCtxSvc.Current(ctx)
	}
	if tenantID <= 0 {
		tenantID = int64(current.TenantID)
	}
	actorID := current.ActingUserID
	if actorID == 0 {
		actorID = current.UserID
	}
	actor := capmodel.CapabilityActor{
		Type:   capmodel.ActorTypeUser,
		UserID: int64(actorID),
		Name:   current.Username,
	}
	if actorID == 0 {
		actor = capmodel.CapabilityActor{
			Type:         capmodel.ActorTypeSystem,
			Name:         tenantPluginCapabilityPluginID,
			SystemReason: "tenant plugin governance",
		}
	}
	return capmodel.CapabilityContext{
		PluginID:    tenantPluginCapabilityPluginID,
		Actor:       actor,
		TenantID:    capmodel.DomainID(strconv.FormatInt(tenantID, 10)),
		Source:      capmodel.CapabilitySourceHTTP,
		SystemCall:  actor.Type == capmodel.ActorTypeSystem,
		Resource:    resource,
		RequestedAt: time.Now(),
	}
}

// requirePlugincap verifies tenant plugin governance dependencies.
func (s *serviceImpl) requirePlugincap() error {
	if s == nil || s.plugins == nil || s.pluginAdmin == nil {
		return bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", "plugin"))
	}
	return nil
}
