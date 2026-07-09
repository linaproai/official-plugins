package identity

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	v1 "lina-plugin-linapro-oidc-core/backend/api/identity/v1"
)

// ListIdentities returns the current session user's bound external identities.
func (c *ControllerV1) ListIdentities(ctx context.Context, _ *v1.ListIdentitiesReq) (res *v1.ListIdentitiesRes, err error) {
	current := c.bizCtxSvc.Current(ctx)
	if current.UserID <= 0 {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized)
	}
	identities, err := c.identitySvc.List(ctx, current.UserID)
	if err != nil {
		return nil, err
	}
	items := make([]v1.BoundIdentityItem, 0, len(identities))
	for _, identity := range identities {
		items = append(items, v1.BoundIdentityItem{
			Provider: identity.Provider,
			Subject:  identity.Subject,
			Email:    identity.Email,
		})
	}
	return &v1.ListIdentitiesRes{Items: items}, nil
}
