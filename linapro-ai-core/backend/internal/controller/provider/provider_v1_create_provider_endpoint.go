// This file implements the create-provider-endpoint controller method.

package provider

import (
	"context"

	"lina-plugin-linapro-ai-core/backend/api/provider/v1"
	aisvc "lina-plugin-linapro-ai-core/backend/internal/service/ai"
)

// CreateProviderEndpoint creates one protocol endpoint under a provider.
func (c *ControllerV1) CreateProviderEndpoint(ctx context.Context, req *v1.CreateProviderEndpointReq) (res *v1.CreateProviderEndpointRes, err error) {
	id, err := c.aiSvc.CreateProviderEndpoint(ctx, aisvc.ProviderEndpointSaveInput{
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
	return &v1.CreateProviderEndpointRes{Id: id}, nil
}
