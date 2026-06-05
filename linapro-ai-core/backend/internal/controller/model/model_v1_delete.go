// This file implements the delete-model controller method.

package model

import (
	"context"

	"lina-plugin-linapro-ai-core/backend/api/model/v1"
)

// Delete deletes the provider-local same-name model group when no tier binding references it.
func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error) {
	if err = c.aiSvc.DeleteModel(ctx, req.Id); err != nil {
		return nil, err
	}
	return &v1.DeleteRes{}, nil
}
