// identity_v1_list_identities.go implements GET list for the current session
// user's external-identity linkages via LinkageService.ListByUser.

package identity

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	v1 "lina-plugin-linapro-extid-core/backend/api/identity/v1"
)

// ListIdentities returns the current session user's bound external identities.
func (c *ControllerV1) ListIdentities(ctx context.Context, _ *v1.ListIdentitiesReq) (res *v1.ListIdentitiesRes, err error) {
	current := c.bizCtxSvc.Current(ctx)
	if current.UserID <= 0 {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized)
	}
	if c.extidSvc == nil || c.extidSvc.Linkage() == nil {
		return nil, gerror.NewCode(gcode.CodeInternalError)
	}
	identities, err := c.extidSvc.Linkage().ListByUser(ctx, current.UserID)
	if err != nil {
		return nil, err
	}
	items := make([]v1.IdentityItem, 0, len(identities))
	for _, identity := range identities {
		items = append(items, v1.IdentityItem{
			Provider:    identity.Provider,
			Subject:     identity.Subject,
			SubjectKind: string(identity.SubjectKind),
			AppContext:  identity.AppContext,
			Email:       identity.Email,
			Phone:       identity.Phone,
			DisplayName: identity.DisplayName,
			AvatarURL:   identity.AvatarURL,
		})
	}
	return &v1.ListIdentitiesRes{List: items}, nil
}
