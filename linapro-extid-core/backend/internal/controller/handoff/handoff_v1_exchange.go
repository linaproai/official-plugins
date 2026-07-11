// handoff_v1_exchange.go implements the public SPA handoff exchange endpoint.
package handoff

import (
	"context"

	"lina-core/pkg/bizerr"
	v1 "lina-plugin-linapro-extid-core/backend/api/handoff/v1"
	"lina-plugin-linapro-extid-core/backend/cap/extidcap"
)

// ExchangeLoginHandoff consumes a one-time external-login handoff code.
func (c *ControllerV1) ExchangeLoginHandoff(
	ctx context.Context,
	req *v1.ExchangeLoginHandoffReq,
) (res *v1.ExchangeLoginHandoffRes, err error) {
	_ = ctx
	if c == nil || c.handoffs == nil {
		return nil, bizerr.NewCode(extidcap.CodeLoginHandoffInvalid)
	}
	payload, err := c.handoffs.Exchange(req.Handoff)
	if err != nil {
		return nil, err
	}
	tenants := make([]*v1.HandoffTenantEntity, 0, len(payload.TenantCandidates))
	for _, tenant := range payload.TenantCandidates {
		tenants = append(tenants, &v1.HandoffTenantEntity{
			Id:     tenant.ID,
			Code:   tenant.Code,
			Name:   tenant.Name,
			Status: tenant.Status,
		})
	}
	return &v1.ExchangeLoginHandoffRes{
		AccessToken:  payload.AccessToken,
		RefreshToken: payload.RefreshToken,
		PreToken:     payload.PreToken,
		Tenants:      tenants,
	}, nil
}
