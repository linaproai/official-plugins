package platform

import (
	"context"

	"lina-plugin-linapro-tenant-core/backend/api/platform/v1"

	"github.com/gogf/gf/v2/frame/g"

	"lina-plugin-linapro-tenant-core/backend/internal/service/impersonate"
)

// TenantEndImpersonate ends platform impersonation for the current token.
func (c *ControllerV1) TenantEndImpersonate(ctx context.Context, req *v1.TenantEndImpersonateReq) (res *v1.TenantEndImpersonateRes, err error) {
	token := ""
	if request := g.RequestFromCtx(ctx); request != nil {
		token = request.GetHeader("Authorization")
	}
	if err = c.impersonateSvc.Stop(ctx, impersonate.StopInput{TenantID: req.Id, Token: token}); err != nil {
		return nil, err
	}
	return &v1.TenantEndImpersonateRes{}, nil
}
