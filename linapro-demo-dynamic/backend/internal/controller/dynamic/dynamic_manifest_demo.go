// Manifest host-service demo route controller.

package dynamic

import (
	"context"

	v1 "lina-plugin-linapro-demo-dynamic/backend/api/dynamic/v1"
)

// ManifestDemo returns values read through the plugin-scoped manifest host
// service without executing the broader host-call side-effect demo.
func (c *Controller) ManifestDemo(
	ctx context.Context,
	_ *v1.ManifestDemoReq,
) (res *v1.ManifestDemoRes, err error) {
	payload, err := c.dynamicSvc.BuildManifestDemoPayload(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.ManifestDemoRes{
		ProfilePath:       payload.ProfilePath,
		ProfileFound:      payload.ProfileFound,
		ProfileName:       payload.ProfileName,
		ProfileTier:       payload.ProfileTier,
		ProfileOwner:      payload.ProfileOwner,
		ConfigPath:        payload.ConfigPath,
		ConfigFound:       payload.ConfigFound,
		ConfigBodyPreview: payload.ConfigBodyPreview,
	}, nil
}
