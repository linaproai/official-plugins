// This file implements the tenant-switch validation endpoint.

package auth

import (
	"context"
	"lina-core/pkg/plugin/capability/authcap/token"

	v1 "lina-plugin-linapro-tenant-core/backend/api/auth/v1"
)

// SwitchTenant validates target tenant visibility and requests a host-signed token.
func (c *ControllerV1) SwitchTenant(ctx context.Context, req *v1.SwitchTenantReq) (res *v1.SwitchTenantRes, err error) {
	tokenString, _ := token.BearerTokenFromContext(ctx)
	out, err := c.authSvc.SwitchTenant(ctx, token.SwitchTenantInput{
		BearerToken: tokenString,
		TenantID:    int(req.TenantId),
	})
	if err != nil {
		return nil, err
	}
	return &v1.SwitchTenantRes{AccessToken: out.AccessToken, RefreshToken: out.RefreshToken}, nil
}
