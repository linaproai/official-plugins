// This file implements the list-provider-operations controller method.

package invocation

import (
	"context"

	"lina-plugin-linapro-ai-core/backend/api/invocation/v1"

	aisvc "lina-plugin-linapro-ai-core/backend/internal/service/ai"
)

// ListProviderOperations returns masked provider operation projections.
func (c *ControllerV1) ListProviderOperations(ctx context.Context, req *v1.ListProviderOperationsReq) (res *v1.ListProviderOperationsRes, err error) {
	out, err := c.aiSvc.ListProviderOperations(ctx, aisvc.ProviderOperationListInput{
		PageNum:          req.PageNum,
		PageSize:         req.PageSize,
		CapabilityType:   req.CapabilityType,
		CapabilityMethod: req.CapabilityMethod,
		Purpose:          req.Purpose,
		Status:           req.Status,
		ProviderId:       req.ProviderId,
		ModelId:          req.ModelId,
		SourcePluginId:   req.SourcePluginId,
		StartedAt:        req.StartedAt,
		EndedAt:          req.EndedAt,
	})
	if err != nil {
		return nil, err
	}
	list := make([]*v1.ProviderOperationItem, 0, len(out.List))
	for _, item := range out.List {
		list = append(list, toAPIProviderOperationItem(item))
	}
	return &v1.ListProviderOperationsRes{List: list, Total: out.Total}, nil
}
