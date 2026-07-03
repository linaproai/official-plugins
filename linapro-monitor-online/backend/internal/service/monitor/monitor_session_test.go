// This file verifies linapro-monitor-online service operations delegate to the
// host session domain read and management capabilities.

package monitor

import (
	"context"
	"testing"

	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/sessioncap"
)

// TestListDelegatesToSessionDomain verifies online-user listing goes through
// the published session domain service with the original filter and pagination.
func TestListDelegatesToSessionDomain(t *testing.T) {
	ctx := context.Background()
	session := &sessioncap.SessionInfo{ID: "visible-token", UserID: "10", Username: "visible"}
	sessionSvc := &monitorSessionService{
		searchResult: &capmodel.PageResult[*sessioncap.SessionInfo]{
			Items: []*sessioncap.SessionInfo{session},
			Total: 1,
		},
	}
	svc := &serviceImpl{
		bizCtxSvc:  monitorBizCtxService{},
		sessionSvc: sessionSvc,
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
	if out.Total != 1 || len(out.Items) != 1 || out.Items[0] != session {
		t.Fatalf("expected session domain result, got %#v", out)
	}
}

// TestForceLogoutDelegatesToSessionService verifies online-user revocation goes
// through the published unified session capability.
func TestForceLogoutDelegatesToSessionService(t *testing.T) {
	ctx := context.Background()
	sessionSvc := &monitorSessionService{}
	svc := &serviceImpl{
		bizCtxSvc:  monitorBizCtxService{},
		sessionSvc: sessionSvc,
	}

	if err := svc.ForceLogout(ctx, "target-token"); err != nil {
		t.Fatalf("force logout online session: %v", err)
	}
	if sessionSvc.revokeCalled != 1 {
		t.Fatalf("expected one Revoke call, got %d", sessionSvc.revokeCalled)
	}
	if sessionSvc.revokedSessionID != "target-token" {
		t.Fatalf("expected token target-token, got %q", sessionSvc.revokedSessionID)
	}
}

// monitorSessionService records calls to the published host session read service.
type monitorSessionService struct {
	searchCalled int
	searchInput  sessioncap.ListInput
	searchResult *capmodel.PageResult[*sessioncap.SessionInfo]

	revokeCalled     int
	revokedSessionID sessioncap.SessionID
}

// Current returns no current session because list tests never read it.
func (s *monitorSessionService) Current(context.Context) (*sessioncap.SessionInfo, error) {
	return nil, nil
}

// Get is unused by these service tests.
func (s *monitorSessionService) Get(context.Context, sessioncap.SessionID) (*sessioncap.SessionInfo, error) {
	return nil, nil
}

// List records list arguments and returns the configured result.
func (s *monitorSessionService) List(ctx context.Context, input sessioncap.ListInput) (*capmodel.PageResult[*sessioncap.SessionInfo], error) {
	s.searchCalled++
	s.searchInput = input
	if s.searchResult != nil {
		return s.searchResult, nil
	}
	return &capmodel.PageResult[*sessioncap.SessionInfo]{Items: []*sessioncap.SessionInfo{}}, nil
}

// BatchGet is unused by these service tests.
func (s *monitorSessionService) BatchGet(context.Context, []sessioncap.SessionID) (*capmodel.BatchResult[*sessioncap.SessionInfo, sessioncap.SessionID], error) {
	return &capmodel.BatchResult[*sessioncap.SessionInfo, sessioncap.SessionID]{
		Items:      map[sessioncap.SessionID]*sessioncap.SessionInfo{},
		MissingIDs: []sessioncap.SessionID{},
	}, nil
}

// BatchGetUserOnlineStatus is unused by these service tests.
func (s *monitorSessionService) BatchGetUserOnlineStatus(_ context.Context, userIDs []string) (*capmodel.BatchResult[*sessioncap.UserOnlineStatus, string], error) {
	return &capmodel.BatchResult[*sessioncap.UserOnlineStatus, string]{
		Items:      map[string]*sessioncap.UserOnlineStatus{},
		MissingIDs: append([]string(nil), userIDs...),
	}, nil
}

// EnsureVisible accepts all session IDs because visibility is outside these tests.
func (s *monitorSessionService) EnsureVisible(context.Context, []sessioncap.SessionID) error {
	return nil
}

// Revoke records the session ID passed to the published unified session service.
func (s *monitorSessionService) Revoke(ctx context.Context, id sessioncap.SessionID) error {
	s.revokeCalled++
	s.revokedSessionID = id
	return nil
}

// RevokeMany is unused by these service tests.
func (s *monitorSessionService) RevokeMany(context.Context, []sessioncap.SessionID) error {
	return nil
}

// monitorBizCtxService returns a deterministic business context for capability calls.
type monitorBizCtxService struct{}

// Current returns a request-scoped actor and tenant projection.
func (monitorBizCtxService) Current(context.Context) bizctxcap.CurrentContext {
	return bizctxcap.CurrentContext{UserID: 7, Username: "admin", TenantID: 3}
}
