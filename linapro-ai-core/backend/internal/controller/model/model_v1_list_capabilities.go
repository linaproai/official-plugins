// This file implements the list-model-capabilities controller method.

package model

import (
	"context"

	"lina-plugin-linapro-ai-core/backend/api/model/v1"
)

// ListCapabilities returns explicit capability methods for one model.
func (c *ControllerV1) ListCapabilities(ctx context.Context, req *v1.ListCapabilitiesReq) (res *v1.ListCapabilitiesRes, err error) {
	items, err := c.aiSvc.ListModelCapabilities(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	list := make([]*v1.ModelCapabilityItem, 0, len(items))
	for _, item := range items {
		list = append(list, toAPIModelCapabilityItem(item))
	}
	return &v1.ListCapabilitiesRes{List: list}, nil
}
