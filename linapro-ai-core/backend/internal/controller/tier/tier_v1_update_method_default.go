// This file implements the update-method-default controller method.

package tier

import (
	"context"

	"lina-plugin-linapro-ai-core/backend/api/tier/v1"
	aisvc "lina-plugin-linapro-ai-core/backend/internal/service/ai"
)

// UpdateMethodDefault updates one capability method default parameter projection.
func (c *ControllerV1) UpdateMethodDefault(ctx context.Context, req *v1.UpdateMethodDefaultReq) (res *v1.UpdateMethodDefaultRes, err error) {
	err = c.aiSvc.UpdateMethodDefault(ctx, aisvc.MethodDefaultParamSaveInput{
		CapabilityType:    req.CapabilityType,
		CapabilityMethod:  req.CapabilityMethod,
		DefaultParamsJson: req.DefaultParamsJson,
		Enabled:           req.Enabled,
	})
	if err != nil {
		return nil, err
	}
	return &v1.UpdateMethodDefaultRes{}, nil
}
