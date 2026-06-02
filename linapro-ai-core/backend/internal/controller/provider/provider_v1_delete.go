// This file implements the delete-provider controller method.

package provider

import (
	"context"

	"lina-plugin-linapro-ai-core/backend/api/provider/v1"
)

// Delete deletes one AI provider when no tier binding references it.
func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error) {
	if err = c.aiSvc.DeleteProvider(ctx, req.Id); err != nil {
		return nil, err
	}
	return &v1.DeleteRes{}, nil
}
