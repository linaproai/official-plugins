// This file implements the update-tier controller method.

package tier

import (
	"context"

	"lina-plugin-linapro-ai-core/backend/api/tier/v1"
	aisvc "lina-plugin-linapro-ai-core/backend/internal/service/ai"
)

// Update updates one fixed AI capability tier.
func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error) {
	err = c.aiSvc.UpdateTier(ctx, aisvc.TierUpdateInput{
		CapabilityType:   req.CapabilityType,
		CapabilityMethod: req.CapabilityMethod,
		Code:             req.Code,
		ProviderId:       req.ProviderId,
		ModelId:          req.ModelId,
		DefaultEffort:    req.DefaultEffort,
		Enabled:          req.Enabled,
	})
	if err != nil {
		return nil, err
	}
	return &v1.UpdateRes{}, nil
}
