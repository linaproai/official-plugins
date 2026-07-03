// Package monitor implements the online-user governance service for the
// linapro-monitor-online source plugin. It consumes the published host session seam
// so the host continues to own authentication and session truth.
package monitor

import (
	"context"

	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/sessioncap"
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
	bizCtxSvc  bizctxcap.Service  // Business context bridge
	sessionSvc sessioncap.Service // Online-session capability
}

// New creates and returns a new linapro-monitor-online service instance.
func New(
	bizCtxSvc bizctxcap.Service,
	sessionSvc sessioncap.Service,
) Service {
	return &serviceImpl{
		bizCtxSvc:  bizCtxSvc,
		sessionSvc: sessionSvc,
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
	Items []*sessioncap.SessionInfo
	Total int
}
