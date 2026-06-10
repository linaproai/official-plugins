// Jobs registration controller.

package dynamic

import (
	"context"

	v1 "lina-plugin-linapro-demo-dynamic/backend/api/dynamic/v1"

	"lina-core/pkg/plugin/pluginbridge"
)

// RegisterJobs publishes the dynamic sample plugin's built-in Jobs declarations
// for host-side discovery.
func (c *Controller) RegisterJobs(
	_ context.Context,
	_ *v1.RegisterJobsReq,
) (*v1.RegisterJobsRes, error) {
	if err := c.dynamicSvc.RegisterJobs(pluginbridge.NewDeclarations()); err != nil {
		return nil, err
	}
	return &v1.RegisterJobsRes{}, nil
}
