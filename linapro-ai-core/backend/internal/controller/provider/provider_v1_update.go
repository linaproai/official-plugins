// This file implements the update-provider controller method.

package provider

import (
	"context"

	"lina-plugin-linapro-ai-core/backend/api/provider/v1"
	aisvc "lina-plugin-linapro-ai-core/backend/internal/service/ai"
)

// Update updates one AI provider.
func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error) {
	err = c.aiSvc.UpdateProvider(ctx, aisvc.ProviderSaveInput{
		Id:         req.Id,
		Name:       req.Name,
		WebsiteUrl: req.WebsiteUrl,
		Remark:     req.Remark,
		Enabled:    req.Enabled,
	})
	if err != nil {
		return nil, err
	}
	return &v1.UpdateRes{}, nil
}
