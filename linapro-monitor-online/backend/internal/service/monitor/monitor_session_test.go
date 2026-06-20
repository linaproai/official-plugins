// This file verifies linapro-monitor-online service operations delegate to the
// host session domain read and management capabilities.

package monitor

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/sessioncap"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
)

// TestListDelegatesToSessionDomain verifies online-user listing goes through
// the published session domain service with the original filter and pagination.
func TestListDelegatesToSessionDomain(t *testing.T) {
	ctx := context.Background()
	session := &sessioncap.Projection{ID: "visible-token", UserID: "10", Username: "visible"}
	sessionSvc := &monitorSessionService{
		searchResult: &capmodel.PageResult[*sessioncap.Projection]{
			Items: []*sessioncap.Projection{session},
			Total: 1,
		},
	}
	svc := &serviceImpl{
		bizCtxSvc:    monitorBizCtxService{},
		tenantFilter: monitorTenantFilterService{},
		sessionSvc:   sessionSvc,
	}

	out, err := svc.List(ctx, ListInput{
		PageNum:  2,
		PageSize: 25,
		Username: "visible",
		Ip:       "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("list online sessions: %v", err)
	}
	if sessionSvc.searchCalled != 1 {
		t.Fatalf("expected one Search call, got %d", sessionSvc.searchCalled)
	}
	if sessionSvc.searchInput.Page.PageNum != 2 || sessionSvc.searchInput.Page.PageSize != 25 {
		t.Fatalf("expected page 2 size 25, got page %d size %d", sessionSvc.searchInput.Page.PageNum, sessionSvc.searchInput.Page.PageSize)
	}
	if sessionSvc.searchInput.Username != "visible" || sessionSvc.searchInput.IP != "127.0.0.1" {
		t.Fatalf("expected forwarded filter, got %#v", sessionSvc.searchInput)
	}
	if sessionSvc.searchCapCtx.PluginID != pluginID || sessionSvc.searchCapCtx.Resource != monitorOnlineSessionResource {
		t.Fatalf("expected audited capability context, got %#v", sessionSvc.searchCapCtx)
	}
	if out.Total != 1 || len(out.Items) != 1 || out.Items[0] != session {
		t.Fatalf("expected session domain result, got %#v", out)
	}
}

// TestForceLogoutDelegatesToSessionAdmin verifies online-user revocation goes
// through the published session management capability.
func TestForceLogoutDelegatesToSessionAdmin(t *testing.T) {
	ctx := context.Background()
	sessionAdminSvc := &monitorSessionAdminService{}
	svc := &serviceImpl{
		bizCtxSvc:       monitorBizCtxService{},
		tenantFilter:    monitorTenantFilterService{},
		sessionAdminSvc: sessionAdminSvc,
	}

	if err := svc.ForceLogout(ctx, "target-token"); err != nil {
		t.Fatalf("force logout online session: %v", err)
	}
	if sessionAdminSvc.revokeCalled != 1 {
		t.Fatalf("expected one Revoke call, got %d", sessionAdminSvc.revokeCalled)
	}
	if sessionAdminSvc.revokedSessionID != "target-token" {
		t.Fatalf("expected token target-token, got %q", sessionAdminSvc.revokedSessionID)
	}
	if sessionAdminSvc.revokeCapCtx.PluginID != pluginID || sessionAdminSvc.revokeCapCtx.Resource != monitorOnlineSessionResource {
		t.Fatalf("expected audited capability context, got %#v", sessionAdminSvc.revokeCapCtx)
	}
}

// monitorSessionService records calls to the published host session read service.
type monitorSessionService struct {
	searchCalled int
	searchCapCtx capmodel.CapabilityContext
	searchInput  sessioncap.SearchInput
	searchResult *capmodel.PageResult[*sessioncap.Projection]
}

// Current returns no current session because list tests never read it.
func (s *monitorSessionService) Current(context.Context, capmodel.CapabilityContext) (*sessioncap.Projection, error) {
	return nil, nil
}

// Search records search arguments and returns the configured result.
func (s *monitorSessionService) Search(_ context.Context, capCtx capmodel.CapabilityContext, input sessioncap.SearchInput) (*capmodel.PageResult[*sessioncap.Projection], error) {
	s.searchCalled++
	s.searchCapCtx = capCtx
	s.searchInput = input
	if s.searchResult != nil {
		return s.searchResult, nil
	}
	return &capmodel.PageResult[*sessioncap.Projection]{Items: []*sessioncap.Projection{}}, nil
}

// BatchGet is unused by these service tests.
func (s *monitorSessionService) BatchGet(context.Context, capmodel.CapabilityContext, []sessioncap.SessionID) (*capmodel.BatchResult[*sessioncap.Projection, sessioncap.SessionID], error) {
	return &capmodel.BatchResult[*sessioncap.Projection, sessioncap.SessionID]{
		Items:      map[sessioncap.SessionID]*sessioncap.Projection{},
		MissingIDs: []sessioncap.SessionID{},
	}, nil
}

// BatchGetUserOnlineStatus is unused by these service tests.
func (s *monitorSessionService) BatchGetUserOnlineStatus(_ context.Context, _ capmodel.CapabilityContext, userIDs []string) (*capmodel.BatchResult[*sessioncap.UserOnlineStatusProjection, string], error) {
	return &capmodel.BatchResult[*sessioncap.UserOnlineStatusProjection, string]{
		Items:      map[string]*sessioncap.UserOnlineStatusProjection{},
		MissingIDs: append([]string(nil), userIDs...),
	}, nil
}

// EnsureVisible accepts all session IDs because visibility is outside these tests.
func (s *monitorSessionService) EnsureVisible(context.Context, capmodel.CapabilityContext, []sessioncap.SessionID) error {
	return nil
}

// monitorSessionAdminService records calls to the published host session admin service.
type monitorSessionAdminService struct {
	revokeCalled     int
	revokeCapCtx     capmodel.CapabilityContext
	revokedSessionID sessioncap.SessionID
}

// Revoke records the session ID passed to the published session admin service.
func (s *monitorSessionAdminService) Revoke(_ context.Context, capCtx capmodel.CapabilityContext, id sessioncap.SessionID) error {
	s.revokeCalled++
	s.revokeCapCtx = capCtx
	s.revokedSessionID = id
	return nil
}

// monitorBizCtxService returns a deterministic business context for capability calls.
type monitorBizCtxService struct{}

// Current returns a request-scoped actor and tenant projection.
func (monitorBizCtxService) Current(context.Context) bizctxcap.CurrentContext {
	return bizctxcap.CurrentContext{UserID: 7, Username: "admin", TenantID: 3}
}

// monitorTenantFilterService returns a deterministic tenant filter context.
type monitorTenantFilterService struct{}

// Context returns a request-scoped tenant and actor projection.
func (monitorTenantFilterService) Context(context.Context) tenantspi.TenantFilterContext {
	return tenantspi.TenantFilterContext{UserID: 7, TenantID: 3}
}

// Apply is unused by these service tests.
func (monitorTenantFilterService) Apply(_ context.Context, model *gdb.Model, _ string) *gdb.Model {
	return model
}
