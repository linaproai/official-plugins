package settings

import (
	"context"

	v1 "lina-plugin-linapro-storage-cos/backend/api/settings/v1"
	settingssvc "lina-plugin-linapro-storage-cos/backend/internal/service/settings"
)

// SaveSettings persists settings.
func (c *ControllerV1) SaveSettings(ctx context.Context, req *v1.SaveSettingsReq) (*v1.SaveSettingsRes, error) {
	projection, err := c.settingsSvc.Save(ctx, settingssvc.SaveInput{
		AccessKeyID:     req.AccessKeyID,
		SecretAccessKey: req.SecretAccessKey,
		Region:          req.Region,
		Bucket:          req.Bucket,
		Endpoint:        req.Endpoint,
		PathPrefix:      req.PathPrefix,
	})
	if err != nil {
		return nil, err
	}
	return &v1.SaveSettingsRes{Settings: projectItem(projection)}, nil
}
