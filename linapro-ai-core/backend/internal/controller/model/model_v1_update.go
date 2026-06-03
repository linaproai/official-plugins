// This file implements the update-model controller method.

package model

import (
	"context"

	"lina-plugin-linapro-ai-core/backend/api/model/v1"
	aisvc "lina-plugin-linapro-ai-core/backend/internal/service/ai"
)

// Update updates one AI model.
func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error) {
	err = c.aiSvc.UpdateModel(ctx, aisvc.ModelSaveInput{
		Id:         req.Id,
		EndpointId: req.EndpointId,
		ModelName:  req.ModelName,
		Protocol:   req.Protocol,
		Enabled:    req.Enabled,
	})
	if err != nil {
		return nil, err
	}
	return &v1.UpdateRes{}, nil
}
