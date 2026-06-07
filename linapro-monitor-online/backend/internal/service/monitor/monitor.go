// Package monitor implements the online-user governance service for the
// linapro-monitor-online source plugin. It consumes the published host session seam
// so the host continues to own authentication and session truth.
package monitor

import (
	"context"
	"strconv"
	"time"

	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/sessioncap"
	"lina-core/pkg/plugin/capability/tenantcap"
)

const (
	pluginID                       = "linapro-monitor-online"
	monitorOnlineSessionResource   = "monitor.online.session"
	monitorOnlineRevokeAuditReason = "monitor.online.force_logout"
)

// Service defines online-session read and revocation operations backed by the host session seam.
type Service interface {
	// List returns one paginated online-user list filtered by username and IP.
	// The host session service remains the source of truth for tenant/data visibility.
	List(ctx context.Context, in ListInput) (*ListOutput, error)
	// ForceLogout invalidates one online-user session by token ID through the
	// host session service and returns its authorization or revocation error.
	ForceLogout(ctx context.Context, tokenID string) error
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	bizCtxSvc       bizctxcap.Service                  // Business context bridge
	tenantFilter    tenantcap.PluginTableFilterService // Tenant query filter bridge
	sessionSvc      sessioncap.Service                 // Online-session read capability
	sessionAdminSvc sessioncap.AdminService            // Online-session management capability
}

// New creates and returns a new linapro-monitor-online service instance.
func New(
	bizCtxSvc bizctxcap.Service,
	tenantFilter tenantcap.PluginTableFilterService,
	sessionSvc sessioncap.Service,
	sessionAdminSvc sessioncap.AdminService,
) Service {
	return &serviceImpl{
		bizCtxSvc:       bizCtxSvc,
		tenantFilter:    tenantFilter,
		sessionSvc:      sessionSvc,
		sessionAdminSvc: sessionAdminSvc,
	}
}

// ListInput defines the online-user list filter input.
type ListInput struct {
	PageNum  int
	PageSize int
	Username string
	Ip       string
}

// ListOutput defines the online-user list result.
type ListOutput struct {
	Items []*sessioncap.Projection
	Total int
}

// capabilityContext builds the audited domain-call context used by host
// session capabilities without exposing host-private request or storage objects.
func (s *serviceImpl) capabilityContext(ctx context.Context, resource string) capmodel.CapabilityContext {
	var current bizctxcap.CurrentContext
	if s.bizCtxSvc != nil {
		current = s.bizCtxSvc.Current(ctx)
	}
	tenantCtx := tenantcap.TenantFilterContext{TenantID: current.TenantID}
	if s.tenantFilter != nil {
		tenantCtx = s.tenantFilter.Context(ctx)
	}
	actorUserID := tenantCtx.ActingUserID
	if actorUserID == 0 {
		actorUserID = current.ActingUserID
	}
	if actorUserID == 0 {
		actorUserID = tenantCtx.UserID
	}
	if actorUserID == 0 {
		actorUserID = current.UserID
	}
	tenantID := tenantCtx.TenantID
	if tenantID == 0 {
		tenantID = current.TenantID
	}
	return capmodel.CapabilityContext{
		PluginID: pluginID,
		Actor: capmodel.CapabilityActor{
			Type:   capmodel.ActorTypeUser,
			UserID: int64(actorUserID),
			Name:   current.Username,
		},
		TenantID:    capmodel.DomainID(strconv.Itoa(tenantID)),
		Source:      capmodel.CapabilitySourceHTTP,
		Resource:    resource,
		RequestedAt: time.Now(),
	}
}
