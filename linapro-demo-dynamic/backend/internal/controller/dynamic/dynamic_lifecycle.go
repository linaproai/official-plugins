// This file implements typed lifecycle callback handlers for the dynamic sample plugin.

package dynamic

import (
	"context"
	"strings"

	v1 "lina-plugin-linapro-demo-dynamic/backend/api/dynamic/v1"
	dynamicservice "lina-plugin-linapro-demo-dynamic/backend/internal/service/dynamic"

	bridgeplugin "lina-core/pkg/plugin/pluginbridge"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// beforeInstallTimeoutMs demonstrates a code-owned lifecycle timeout override
// for the dynamic plugin builder.
const beforeInstallTimeoutMs = 120000

// BeforeInstall logs the dynamic plugin install precondition.
func (c *Controller) BeforeInstall(ctx context.Context, req *v1.BeforeInstallReq) (*v1.LifecycleDecisionRes, error) {
	return c.runLifecycleDebugHook(ctx, req)
}

// AfterInstall logs the dynamic plugin post-install notification.
func (c *Controller) AfterInstall(ctx context.Context, req *v1.AfterInstallReq) (*v1.LifecycleDecisionRes, error) {
	return c.runLifecycleDebugHook(ctx, req)
}

// BeforeUpgrade logs the dynamic plugin upgrade precondition.
func (c *Controller) BeforeUpgrade(ctx context.Context, req *v1.BeforeUpgradeReq) (*v1.LifecycleDecisionRes, error) {
	return c.runLifecycleDebugHook(ctx, req)
}

// Upgrade logs the dynamic plugin upgrade execution callback.
func (c *Controller) Upgrade(ctx context.Context, req *v1.UpgradeReq) (*v1.LifecycleDecisionRes, error) {
	return c.runLifecycleDebugHook(ctx, req)
}

// AfterUpgrade logs the dynamic plugin post-upgrade notification.
func (c *Controller) AfterUpgrade(ctx context.Context, req *v1.AfterUpgradeReq) (*v1.LifecycleDecisionRes, error) {
	return c.runLifecycleDebugHook(ctx, req)
}

// BeforeDisable logs the dynamic plugin disable precondition.
func (c *Controller) BeforeDisable(ctx context.Context, req *v1.BeforeDisableReq) (*v1.LifecycleDecisionRes, error) {
	return c.runLifecycleDebugHook(ctx, req)
}

// AfterDisable logs the dynamic plugin post-disable notification.
func (c *Controller) AfterDisable(ctx context.Context, req *v1.AfterDisableReq) (*v1.LifecycleDecisionRes, error) {
	return c.runLifecycleDebugHook(ctx, req)
}

// BeforeUninstall logs the dynamic plugin uninstall precondition.
func (c *Controller) BeforeUninstall(ctx context.Context, req *v1.BeforeUninstallReq) (*v1.LifecycleDecisionRes, error) {
	return c.runLifecycleDebugHook(ctx, req)
}

// Uninstall logs the dynamic plugin uninstall cleanup callback.
func (c *Controller) Uninstall(ctx context.Context, req *v1.UninstallReq) (*v1.LifecycleDecisionRes, error) {
	return c.runLifecycleDebugHook(ctx, req)
}

// AfterUninstall logs the dynamic plugin post-uninstall notification.
func (c *Controller) AfterUninstall(ctx context.Context, req *v1.AfterUninstallReq) (*v1.LifecycleDecisionRes, error) {
	return c.runLifecycleDebugHook(ctx, req)
}

// BeforeTenantDisable logs the dynamic plugin tenant-disable precondition.
func (c *Controller) BeforeTenantDisable(ctx context.Context, req *v1.BeforeTenantDisableReq) (*v1.LifecycleDecisionRes, error) {
	return c.runLifecycleDebugHook(ctx, req)
}

// AfterTenantDisable logs the dynamic plugin post-tenant-disable notification.
func (c *Controller) AfterTenantDisable(ctx context.Context, req *v1.AfterTenantDisableReq) (*v1.LifecycleDecisionRes, error) {
	return c.runLifecycleDebugHook(ctx, req)
}

// BeforeTenantDelete logs the dynamic plugin tenant-delete precondition.
func (c *Controller) BeforeTenantDelete(ctx context.Context, req *v1.BeforeTenantDeleteReq) (*v1.LifecycleDecisionRes, error) {
	return c.runLifecycleDebugHook(ctx, req)
}

// AfterTenantDelete logs the dynamic plugin post-tenant-delete notification.
func (c *Controller) AfterTenantDelete(ctx context.Context, req *v1.AfterTenantDeleteReq) (*v1.LifecycleDecisionRes, error) {
	return c.runLifecycleDebugHook(ctx, req)
}

// BeforeInstallModeChange logs the dynamic plugin install-mode change precondition.
func (c *Controller) BeforeInstallModeChange(ctx context.Context, req *v1.BeforeInstallModeChangeReq) (*v1.LifecycleDecisionRes, error) {
	return c.runLifecycleDebugHook(ctx, req)
}

// AfterInstallModeChange logs the dynamic plugin post-install-mode-change notification.
func (c *Controller) AfterInstallModeChange(ctx context.Context, req *v1.AfterInstallModeChangeReq) (*v1.LifecycleDecisionRes, error) {
	return c.runLifecycleDebugHook(ctx, req)
}

// runLifecycleDebugHook logs one dynamic lifecycle request and allows the host
// lifecycle operation to continue.
func (c *Controller) runLifecycleDebugHook(
	ctx context.Context,
	req *protocol.LifecycleRequest,
) (*v1.LifecycleDecisionRes, error) {
	input := buildLifecycleDebugInput(ctx, req)
	if err := c.dynamicSvc.RunLifecycleDebugHook(input); err != nil {
		return nil, err
	}
	return &v1.LifecycleDecisionRes{OK: true}, nil
}

// buildLifecycleDebugInput converts a typed lifecycle request into the service input.
func buildLifecycleDebugInput(ctx context.Context, req *protocol.LifecycleRequest) *dynamicservice.LifecycleDebugInput {
	input := &dynamicservice.LifecycleDebugInput{}
	if envelope := bridgeplugin.RequestEnvelopeFromContext(ctx); envelope != nil {
		input.PluginID = strings.TrimSpace(envelope.PluginID)
	}
	if req == nil {
		return input
	}
	if strings.TrimSpace(req.PluginID) != "" {
		input.PluginID = strings.TrimSpace(req.PluginID)
	}
	input.Operation = strings.TrimSpace(req.Operation)
	input.FromVersion = strings.TrimSpace(req.FromVersion)
	input.ToVersion = strings.TrimSpace(req.ToVersion)
	input.TenantID = req.TenantID
	input.FromMode = strings.TrimSpace(req.FromMode)
	input.ToMode = strings.TrimSpace(req.ToMode)
	input.PurgeStorageData = req.PurgeStorageData
	return input
}
