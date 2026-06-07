// This file implements the list-model controller method.

package model

import (
	"context"

	"lina-plugin-linapro-ai-core/backend/api/model/v1"

	aisvc "lina-plugin-linapro-ai-core/backend/internal/service/ai"
)

// List returns a paged AI model list with provider and endpoint projections.
func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error) {
	out, err := c.aiSvc.ListAllModels(ctx, aisvc.ModelGlobalListInput{
		PageNum:    req.PageNum,
		PageSize:   req.PageSize,
		Keyword:    req.Keyword,
		ProviderId: req.ProviderId,
		Enabled:    req.Enabled,
	})
	if err != nil {
		return nil, err
	}
	list := make([]*v1.ModelItem, 0, len(out.List))
	for _, item := range out.List {
		list = append(list, toAPIModelItem(item))
	}
	return &v1.ListRes{List: list, Total: out.Total}, nil
}
