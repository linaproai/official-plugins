package identity

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/capability/authcap/externallogin/externalidentityspi"
	v1 "lina-plugin-linapro-oidc-core/backend/api/identity/v1"
)

// apiBindPluginID stamps API-driven binds with the plugin that owns the linkage
// storage; OAuth-flow binds are stamped with the verifying plugin by the host.
const apiBindPluginID = "linapro-oidc-core"

// BindIdentity links one verified external identity to the current session user.
func (c *ControllerV1) BindIdentity(ctx context.Context, req *v1.BindIdentityReq) (res *v1.BindIdentityRes, err error) {
	current := c.bizCtxSvc.Current(ctx)
	if current.UserID <= 0 {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized)
	}
	if err = c.identitySvc.Bind(ctx, externalidentityspi.BindInput{
		UserID:   current.UserID,
		Provider: req.Provider,
		Subject:  req.Subject,
		Email:    req.Email,
		PluginID: apiBindPluginID,
	}); err != nil {
		return nil, err
	}
	return &v1.BindIdentityRes{}, nil
}
