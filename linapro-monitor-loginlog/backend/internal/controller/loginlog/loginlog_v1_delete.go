package loginlog

import (
	"context"

	v1 "lina-plugin-linapro-monitor-loginlog/backend/api/loginlog/v1"
)

// Delete deletes login logs.
func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error) {
	deleted, err := c.loginLogSvc.DeleteByIds(ctx, req.Ids)
	if err != nil {
		return nil, err
	}
	return &v1.DeleteRes{Deleted: deleted}, nil
}
