// Demo-record delete route controller.

package dynamic

import (
	"context"

	v1 "lina-plugin-linapro-demo-dynamic/backend/api/dynamic/v1"
)

// DeleteDemoRecord deletes one plugin-owned demo record.
func (c *Controller) DeleteDemoRecord(
	ctx context.Context,
	req *v1.DeleteDemoRecordReq,
) (res *v1.DeleteDemoRecordRes, err error) {
	payload, err := c.dynamicSvc.DeleteDemoRecordPayload(ctx, req.Id)
	if err != nil {
		return nil, wrapDynamicError(err)
	}
	return &v1.DeleteDemoRecordRes{
		Id:      payload.Id,
		Deleted: payload.Deleted,
	}, nil
}
