// This file implements the upsert-model-capabilities controller method.

package model

import (
	"context"

	"lina-plugin-linapro-ai-core/backend/api/model/v1"
)

// UpsertCapabilities replaces or inserts explicit capability methods for one model.
func (c *ControllerV1) UpsertCapabilities(ctx context.Context, req *v1.UpsertCapabilitiesReq) (res *v1.UpsertCapabilitiesRes, err error) {
	if err = c.aiSvc.UpsertModelCapabilities(ctx, req.Id, toServiceModelCapabilityInputs(req.Items)); err != nil {
		return nil, err
	}
	return &v1.UpsertCapabilitiesRes{}, nil
}
