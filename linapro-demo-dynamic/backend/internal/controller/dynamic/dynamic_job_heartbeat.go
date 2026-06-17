// Declared Jobs heartbeat controller.

package dynamic

import (
	"context"

	v1 "lina-plugin-linapro-demo-dynamic/backend/api/dynamic/v1"
)

// JobHeartbeat executes the declared Jobs heartbeat task for the dynamic
// sample plugin.
func (c *Controller) JobHeartbeat(
	_ context.Context,
	_ *v1.JobHeartbeatReq,
) (*v1.JobHeartbeatRes, error) {
	payload, err := c.dynamicSvc.BuildJobHeartbeatPayload()
	if err != nil {
		return nil, err
	}
	return &v1.JobHeartbeatRes{
		Count:   payload.Count,
		Message: payload.Message,
	}, nil
}
