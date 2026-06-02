// This file implements the list-provider controller method.

package provider

import (
	"context"

	"lina-plugin-linapro-ai-core/backend/api/provider/v1"
	aisvc "lina-plugin-linapro-ai-core/backend/internal/service/ai"
)

// List returns a paged provider list with aggregated model counts.
func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error) {
	out, err := c.aiSvc.ListProviders(ctx, aisvc.ProviderListInput{
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		Keyword:  req.Keyword,
		Enabled:  req.Enabled,
	})
	if err != nil {
		return nil, err
	}
	list := make([]*v1.ProviderItem, 0, len(out.List))
	for _, item := range out.List {
		list = append(list, toAPIProviderItem(item))
	}
	return &v1.ListRes{List: list, Total: out.Total}, nil
}
