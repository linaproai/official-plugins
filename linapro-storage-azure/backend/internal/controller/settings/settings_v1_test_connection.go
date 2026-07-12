// TestConnection handler for Azure Blob connectivity probe.
package settings

import (
	"context"
	"strings"

	"lina-core/pkg/bizerr"

	v1 "lina-plugin-linapro-storage-azure/backend/api/settings/v1"
	settingssvc "lina-plugin-linapro-storage-azure/backend/internal/service/settings"
)

// TestConnection probes connectivity without persisting form values.
func (c *ControllerV1) TestConnection(ctx context.Context, req *v1.TestConnectionReq) (*v1.TestConnectionRes, error) {
	snap, err := c.settingsSvc.Load(ctx)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.AccountName) != "" {
		snap.AccountName = strings.TrimSpace(req.AccountName)
	}
	if secret := strings.TrimSpace(req.AccountKey); secret != "" && secret != settingssvc.SecretMask {
		snap.AccountKey = secret
	}
	if strings.TrimSpace(req.Container) != "" {
		snap.Container = strings.TrimSpace(req.Container)
	}
	if strings.TrimSpace(req.Endpoint) != "" {
		snap.Endpoint = strings.TrimSpace(req.Endpoint)
	}
	if strings.TrimSpace(req.PathPrefix) != "" {
		snap.PathPrefix = settingssvc.NormalizePathPrefix(req.PathPrefix)
	}
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
