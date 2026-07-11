// identity_v1_bind_identity.go implements POST bind: consume a verified ticket
// and link the external identity to the current session user via LinkageService.

package identity

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	v1 "lina-plugin-linapro-extlogin-core/backend/api/identity/v1"
)

// BindIdentity consumes a verified-identity ticket and links it to the current user.
func (c *ControllerV1) BindIdentity(ctx context.Context, req *v1.BindIdentityReq) (res *v1.BindIdentityRes, err error) {
	current := c.bizCtxSvc.Current(ctx)
	if current.UserID <= 0 {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized)
	}
	if c.extidSvc == nil || c.extidSvc.Linkage() == nil {
		return nil, gerror.NewCode(gcode.CodeInternalError)
	}
	if err = c.extidSvc.Linkage().BindByTicket(ctx, current.UserID, req.TicketID); err != nil {
		return nil, err
	}
	return &v1.BindIdentityRes{}, nil
}
