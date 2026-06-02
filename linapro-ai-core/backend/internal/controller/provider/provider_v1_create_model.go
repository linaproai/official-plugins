// This file implements the create-provider-model controller method.

package provider

import (
	"context"

	"lina-plugin-linapro-ai-core/backend/api/provider/v1"
	aisvc "lina-plugin-linapro-ai-core/backend/internal/service/ai"
)

// CreateModel creates one model under a provider.
func (c *ControllerV1) CreateModel(ctx context.Context, req *v1.CreateModelReq) (res *v1.CreateModelRes, err error) {
	id, err := c.aiSvc.CreateModel(ctx, aisvc.ModelSaveInput{
		ProviderId:       req.ProviderId,
		CapabilityType:   req.CapabilityType,
		ModelName:        req.ModelName,
		Protocol:         req.Protocol,
		Source:           aisvc.ModelSourceManual,
		SupportsThinking: req.SupportsThinking,
		SupportedEfforts: req.SupportedEfforts,
		MaxInputTokens:   req.MaxInputTokens,
		MaxOutputTokens:  req.MaxOutputTokens,
		Enabled:          req.Enabled,
	})
	if err != nil {
		return nil, err
	}
	return &v1.CreateModelRes{Id: id}, nil
}
