// Package monitor implements the online-user governance service for the
// monitor-online source plugin. It consumes the published host session seam
// so the host continues to own authentication and session truth.
package monitor

import (
	"context"

	sessionsvc "lina-core/pkg/pluginservice/contract"
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
	sessionSvc sessionsvc.SessionService // published host session service
}

// New creates and returns a new monitor-online service instance.
func New(sessionSvc sessionsvc.SessionService) Service {
	return &serviceImpl{sessionSvc: sessionSvc}
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
	Items []*sessionsvc.Session
	Total int
}
