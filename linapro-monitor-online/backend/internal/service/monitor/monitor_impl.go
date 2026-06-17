// monitor_impl.go implements online-session listing and forced logout for the
// linapro-monitor-online plugin. It delegates to the host-published session service so
// plugin pages observe the same tenant, data-scope, and session-store state as
// the core host.

package monitor

import (
	"context"

	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/sessioncap"
)

// List returns one paginated online-user list.
func (s *serviceImpl) List(ctx context.Context, in ListInput) (*ListOutput, error) {
	out, err := s.sessionSvc.Search(ctx, s.capabilityContext(ctx, monitorOnlineSessionResource), sessioncap.SearchInput{
		Username: in.Username,
		IP:       in.Ip,
		Page: capmodel.PageRequest{
			PageNum:  in.PageNum,
			PageSize: in.PageSize,
		},
	})
	if err != nil {
		return nil, err
	}
	if out == nil {
		return &ListOutput{Items: []*sessioncap.Projection{}}, nil
	}
	return &ListOutput{Items: out.Items, Total: out.Total}, nil
}

// ForceLogout invalidates one online-user session by token ID.
func (s *serviceImpl) ForceLogout(ctx context.Context, tokenID string) error {
	return s.sessionAdminSvc.Revoke(ctx, s.capabilityContext(ctx, monitorOnlineSessionResource), sessioncap.SessionID(tokenID))
}
