// This file implements the sync-provider-models controller method.

package provider

import (
	"context"

	"lina-plugin-linapro-ai-core/backend/api/provider/v1"
	aisvc "lina-plugin-linapro-ai-core/backend/internal/service/ai"
)

// SyncModels imports public model metadata from the selected provider protocol.
func (c *ControllerV1) SyncModels(ctx context.Context, req *v1.SyncModelsReq) (res *v1.SyncModelsRes, err error) {
	out, err := c.aiSvc.SyncModels(ctx, aisvc.ModelSyncInput{
		ProviderId: req.ProviderId,
		Protocol:   req.Protocol,
	})
	if err != nil {
		return nil, err
	}
	return &v1.SyncModelsRes{Created: out.Created, Kept: out.Kept}, nil
}
