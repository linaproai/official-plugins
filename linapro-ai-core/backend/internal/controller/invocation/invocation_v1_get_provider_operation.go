// This file implements the get-provider-operation controller method.

package invocation

import (
	"context"

	"lina-plugin-linapro-ai-core/backend/api/invocation/v1"
)

// GetProviderOperation returns one masked provider operation projection.
func (c *ControllerV1) GetProviderOperation(ctx context.Context, req *v1.GetProviderOperationReq) (res *v1.GetProviderOperationRes, err error) {
	item, err := c.aiSvc.GetProviderOperation(ctx, req.OperationRef)
	if err != nil {
		return nil, err
	}
	dto := toAPIProviderOperationItem(item)
	if dto == nil {
		return &v1.GetProviderOperationRes{}, nil
	}
	return &v1.GetProviderOperationRes{ProviderOperationItem: *dto}, nil
}
