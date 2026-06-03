// This file implements the delete-provider-endpoint controller method.

package provider

import (
	"context"

	"lina-plugin-linapro-ai-core/backend/api/provider/v1"
)

// DeleteProviderEndpoint deletes one provider endpoint after reference checks.
func (c *ControllerV1) DeleteProviderEndpoint(ctx context.Context, req *v1.DeleteProviderEndpointReq) (res *v1.DeleteProviderEndpointRes, err error) {
	if err = c.aiSvc.DeleteProviderEndpoint(ctx, req.ProviderId, req.Id); err != nil {
		return nil, err
	}
	return &v1.DeleteProviderEndpointRes{}, nil
}
