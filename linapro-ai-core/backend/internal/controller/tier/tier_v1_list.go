// This file implements the list-tier controller method.

package tier

import (
	"context"

	"lina-plugin-linapro-ai-core/backend/api/tier/v1"
)

// List returns the fixed AI capability tiers.
func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error) {
	items, err := c.aiSvc.ListTiers(ctx, req.CapabilityType, req.CapabilityMethod)
	if err != nil {
		return nil, err
	}
	list := make([]*v1.TierItem, 0, len(items))
	for _, item := range items {
		list = append(list, toAPITierItem(item))
	}
	return &v1.ListRes{List: list}, nil
}
