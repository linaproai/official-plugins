// This file implements the list-invocation controller method.

package invocation

import (
	"context"

	"lina-plugin-linapro-ai-core/backend/api/invocation/v1"
	aisvc "lina-plugin-linapro-ai-core/backend/internal/service/ai"
)

// List returns masked AI invocation logs with database-side filters.
func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error) {
	out, err := c.aiSvc.ListInvocations(ctx, aisvc.InvocationListInput{
		PageNum:        req.PageNum,
		PageSize:       req.PageSize,
		CapabilityType: req.CapabilityType,
		Purpose:        req.Purpose,
		TierCode:       req.TierCode,
		Status:         req.Status,
		ProviderId:     req.ProviderId,
		ModelId:        req.ModelId,
		SourcePluginId: req.SourcePluginId,
		StartedAt:      req.StartedAt,
		EndedAt:        req.EndedAt,
	})
	if err != nil {
		return nil, err
	}
	list := make([]*v1.InvocationItem, 0, len(out.List))
	for _, item := range out.List {
		list = append(list, toAPIInvocationItem(item))
	}
	return &v1.ListRes{List: list, Total: out.Total}, nil
}
