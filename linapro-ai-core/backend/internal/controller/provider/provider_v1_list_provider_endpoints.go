// This file implements the list-provider-endpoints controller method.

package provider

import (
	"context"

	"lina-plugin-linapro-ai-core/backend/api/provider/v1"
	aisvc "lina-plugin-linapro-ai-core/backend/internal/service/ai"
)

// ListProviderEndpoints returns protocol endpoints belonging to one provider.
func (c *ControllerV1) ListProviderEndpoints(ctx context.Context, req *v1.ListProviderEndpointsReq) (res *v1.ListProviderEndpointsRes, err error) {
	items, err := c.aiSvc.ListProviderEndpoints(ctx, aisvc.ProviderEndpointListInput{
		ProviderId: req.ProviderId,
		Protocol:   req.Protocol,
		Enabled:    req.Enabled,
	})
	if err != nil {
		return nil, err
	}
	list := make([]*v1.ProviderEndpointItem, 0, len(items))
	for _, item := range items {
		list = append(list, toAPIProviderEndpointItem(item))
	}
	return &v1.ListProviderEndpointsRes{List: list}, nil
}
