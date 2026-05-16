// monitor_impl.go implements online-session listing and forced logout for the
// monitor-online plugin. It delegates to the host-published session service so
// plugin pages observe the same tenant, data-scope, and session-store state as
// the core host.

package monitor

import (
	"context"

	sessionsvc "lina-core/pkg/pluginservice/contract"
)

// List returns one paginated online-user list.
func (s *serviceImpl) List(ctx context.Context, in ListInput) (*ListOutput, error) {
	out, err := s.sessionSvc.ListPage(ctx, &sessionsvc.ListFilter{
		Username: in.Username,
		Ip:       in.Ip,
	}, in.PageNum, in.PageSize)
	if err != nil {
		return nil, err
	}
	return &ListOutput{Items: out.Items, Total: out.Total}, nil
}

// ForceLogout invalidates one online-user session by token ID.
func (s *serviceImpl) ForceLogout(ctx context.Context, tokenID string) error {
	return s.sessionSvc.Revoke(ctx, tokenID)
}
