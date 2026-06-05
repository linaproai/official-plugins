// This file implements the list-provider-models controller method.

package provider

import (
	"context"

	"lina-plugin-linapro-ai-core/backend/api/provider/v1"
	aisvc "lina-plugin-linapro-ai-core/backend/internal/service/ai"
)

// ListModels returns models belonging to one provider.
func (c *ControllerV1) ListModels(ctx context.Context, req *v1.ListModelsReq) (res *v1.ListModelsRes, err error) {
	out, err := c.aiSvc.ListModels(ctx, aisvc.ModelListInput{
		ProviderId: req.ProviderId,
		PageNum:    req.PageNum,
		PageSize:   req.PageSize,
		Enabled:    req.Enabled,
	})
	if err != nil {
		return nil, err
	}
	list := make([]*v1.ModelItem, 0, len(out.List))
	for _, item := range out.List {
		list = append(list, toAPIModelItem(item))
	}
	return &v1.ListModelsRes{List: list, Total: out.Total}, nil
}
