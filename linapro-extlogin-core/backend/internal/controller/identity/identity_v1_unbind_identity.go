// identity_v1_unbind_identity.go implements DELETE/unbind for one of the
// current session user's external-identity linkages via LinkageService.Unbind.

package identity

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/capability/authcap/extlogin/extidspi"
	v1 "lina-plugin-linapro-extlogin-core/backend/api/identity/v1"
)

// UnbindIdentity removes one of the current session user's external identities.
func (c *ControllerV1) UnbindIdentity(ctx context.Context, req *v1.UnbindIdentityReq) (res *v1.UnbindIdentityRes, err error) {
	current := c.bizCtxSvc.Current(ctx)
	if current.UserID <= 0 {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized)
	}
	if c.extidSvc == nil || c.extidSvc.Linkage() == nil {
		return nil, gerror.NewCode(gcode.CodeInternalError)
	}
	if err = c.extidSvc.Linkage().Unbind(ctx, extidspi.UnbindInput{
		UserID:   current.UserID,
		Provider: req.Provider,
		Subject:  req.Subject,
	}); err != nil {
		return nil, err
	}
	return &v1.UnbindIdentityRes{}, nil
}
