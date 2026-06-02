// This file implements the get-provider controller method.

package provider

import (
	"context"

	"lina-plugin-linapro-ai-core/backend/api/provider/v1"
)

// Get returns one AI provider detail.
func (c *ControllerV1) Get(ctx context.Context, req *v1.GetReq) (res *v1.GetRes, err error) {
	item, err := c.aiSvc.GetProvider(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	dto := toAPIProviderItem(item)
	if dto == nil {
		return &v1.GetRes{}, nil
	}
	return &v1.GetRes{ProviderItem: *dto}, nil
}
