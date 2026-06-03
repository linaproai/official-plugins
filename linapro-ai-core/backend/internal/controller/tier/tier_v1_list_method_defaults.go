// This file implements the list-method-defaults controller method.

package tier

import (
	"context"

	"lina-plugin-linapro-ai-core/backend/api/tier/v1"
)

// ListMethodDefaults returns all governed capability method default parameters.
func (c *ControllerV1) ListMethodDefaults(ctx context.Context, req *v1.ListMethodDefaultsReq) (res *v1.ListMethodDefaultsRes, err error) {
	items, err := c.aiSvc.ListMethodDefaults(ctx)
	if err != nil {
		return nil, err
	}
	list := make([]*v1.MethodDefaultParamItem, 0, len(items))
	for _, item := range items {
		list = append(list, toAPIMethodDefaultParamItem(item))
	}
	return &v1.ListMethodDefaultsRes{List: list}, nil
}
