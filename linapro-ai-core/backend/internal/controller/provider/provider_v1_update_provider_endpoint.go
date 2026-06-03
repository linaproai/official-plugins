// This file implements the update-provider-endpoint controller method.

package provider

import (
	"context"

	"lina-plugin-linapro-ai-core/backend/api/provider/v1"
	aisvc "lina-plugin-linapro-ai-core/backend/internal/service/ai"
)

// UpdateProviderEndpoint updates one protocol endpoint under a provider.
func (c *ControllerV1) UpdateProviderEndpoint(ctx context.Context, req *v1.UpdateProviderEndpointReq) (res *v1.UpdateProviderEndpointRes, err error) {
	err = c.aiSvc.UpdateProviderEndpoint(ctx, aisvc.ProviderEndpointSaveInput{
		Id:           req.Id,
		ProviderId:   req.ProviderId,
		Protocol:     req.Protocol,
		BaseUrl:      req.BaseUrl,
		SecretRef:    req.SecretRef,
		Enabled:      req.Enabled,
		MetadataJson: req.MetadataJson,
	})
	if err != nil {
		return nil, err
	}
	return &v1.UpdateProviderEndpointRes{}, nil
}
