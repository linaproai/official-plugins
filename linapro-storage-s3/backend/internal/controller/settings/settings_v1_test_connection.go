package settings

import (
	"context"
	"strings"

	"lina-core/pkg/bizerr"

	v1 "lina-plugin-linapro-storage-s3/backend/api/settings/v1"
	settingssvc "lina-plugin-linapro-storage-s3/backend/internal/service/settings"
)

// TestConnection probes connectivity without persisting form values.
func (c *ControllerV1) TestConnection(ctx context.Context, req *v1.TestConnectionReq) (*v1.TestConnectionRes, error) {
	snap, err := c.settingsSvc.Load(ctx)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.AccessKeyID) != "" {
		snap.AccessKeyID = strings.TrimSpace(req.AccessKeyID)
	}
	if secret := strings.TrimSpace(req.SecretAccessKey); secret != "" && secret != settingssvc.SecretMask {
		snap.SecretAccessKey = secret
	}
	if strings.TrimSpace(req.Region) != "" {
		snap.Region = strings.TrimSpace(req.Region)
	}
	if strings.TrimSpace(req.Bucket) != "" {
		snap.Bucket = strings.TrimSpace(req.Bucket)
	}
	if strings.TrimSpace(req.Endpoint) != "" {
		snap.Endpoint = strings.TrimSpace(req.Endpoint)
	}
	if strings.TrimSpace(req.PathPrefix) != "" {
		snap.PathPrefix = settingssvc.NormalizePathPrefix(req.PathPrefix)
	}
	snap.ForcePathStyle = req.ForcePathStyle
	if err := c.settingsSvc.ValidateReady(snap); err != nil {
		return nil, err
	}
	if c.tester == nil {
		return nil, bizerr.NewCode(settingssvc.CodeTestFailed)
	}
	if err := c.tester.TestConnection(ctx, snap); err != nil {
		return &v1.TestConnectionRes{OK: false, Message: err.Error()}, nil
	}
	return &v1.TestConnectionRes{OK: true, Message: "ok"}, nil
}
