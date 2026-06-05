// This file implements the invocation-log cleanup controller method.

package invocation

import (
	"context"

	"lina-plugin-linapro-ai-core/backend/api/invocation/v1"
	aisvc "lina-plugin-linapro-ai-core/backend/internal/service/ai"
)

// Clean clears AI invocation logs within one optional creation time range.
func (c *ControllerV1) Clean(ctx context.Context, req *v1.CleanReq) (res *v1.CleanRes, err error) {
	deleted, err := c.aiSvc.CleanInvocations(ctx, aisvc.InvocationCleanInput{
		StartedAt: req.StartedAt,
		EndedAt:   req.EndedAt,
	})
	if err != nil {
		return nil, err
	}
	return &v1.CleanRes{Deleted: deleted}, nil
}
