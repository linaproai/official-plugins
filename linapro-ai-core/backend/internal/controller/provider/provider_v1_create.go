// This file implements the create-provider controller method.

package provider

import (
	"context"

	"lina-plugin-linapro-ai-core/backend/api/provider/v1"
	aisvc "lina-plugin-linapro-ai-core/backend/internal/service/ai"
)

// Create creates one AI provider.
func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error) {
	id, err := c.aiSvc.CreateProvider(ctx, aisvc.ProviderSaveInput{
		Name:       req.Name,
		WebsiteUrl: req.WebsiteUrl,
		Remark:     req.Remark,
		Enabled:    req.Enabled,
	})
	if err != nil {
		return nil, err
	}
	return &v1.CreateRes{Id: id}, nil
}
