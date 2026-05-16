// This file implements lifecycle callback handlers for the dynamic sample plugin.

package dynamic

import (
	"encoding/json"
	"strings"

	"lina-core/pkg/pluginbridge"
	dynamicservice "lina-plugin-demo-dynamic/backend/internal/service/dynamic"
)

// BeforeInstall logs the dynamic plugin install precondition.
func (c *Controller) BeforeInstall(request *pluginbridge.BridgeRequestEnvelopeV1) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	return c.runLifecycleDebugHook(request)
}

// AfterInstall logs the dynamic plugin post-install notification.
func (c *Controller) AfterInstall(request *pluginbridge.BridgeRequestEnvelopeV1) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	return c.runLifecycleDebugHook(request)
}

// BeforeUpgrade logs the dynamic plugin upgrade precondition.
func (c *Controller) BeforeUpgrade(request *pluginbridge.BridgeRequestEnvelopeV1) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	return c.runLifecycleDebugHook(request)
}

// Upgrade logs the dynamic plugin upgrade execution callback.
func (c *Controller) Upgrade(request *pluginbridge.BridgeRequestEnvelopeV1) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	return c.runLifecycleDebugHook(request)
}

// AfterUpgrade logs the dynamic plugin post-upgrade notification.
func (c *Controller) AfterUpgrade(request *pluginbridge.BridgeRequestEnvelopeV1) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	return c.runLifecycleDebugHook(request)
}

// BeforeDisable logs the dynamic plugin disable precondition.
func (c *Controller) BeforeDisable(request *pluginbridge.BridgeRequestEnvelopeV1) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	return c.runLifecycleDebugHook(request)
}

// AfterDisable logs the dynamic plugin post-disable notification.
func (c *Controller) AfterDisable(request *pluginbridge.BridgeRequestEnvelopeV1) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	return c.runLifecycleDebugHook(request)
}

// BeforeUninstall logs the dynamic plugin uninstall precondition.
func (c *Controller) BeforeUninstall(request *pluginbridge.BridgeRequestEnvelopeV1) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	return c.runLifecycleDebugHook(request)
}

// Uninstall logs the dynamic plugin uninstall cleanup callback.
func (c *Controller) Uninstall(request *pluginbridge.BridgeRequestEnvelopeV1) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	return c.runLifecycleDebugHook(request)
}

// AfterUninstall logs the dynamic plugin post-uninstall notification.
func (c *Controller) AfterUninstall(request *pluginbridge.BridgeRequestEnvelopeV1) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	return c.runLifecycleDebugHook(request)
}

// BeforeTenantDisable logs the dynamic plugin tenant-disable precondition.
func (c *Controller) BeforeTenantDisable(request *pluginbridge.BridgeRequestEnvelopeV1) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	return c.runLifecycleDebugHook(request)
}

// AfterTenantDisable logs the dynamic plugin post-tenant-disable notification.
func (c *Controller) AfterTenantDisable(request *pluginbridge.BridgeRequestEnvelopeV1) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	return c.runLifecycleDebugHook(request)
}

// BeforeTenantDelete logs the dynamic plugin tenant-delete precondition.
func (c *Controller) BeforeTenantDelete(request *pluginbridge.BridgeRequestEnvelopeV1) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	return c.runLifecycleDebugHook(request)
}

// AfterTenantDelete logs the dynamic plugin post-tenant-delete notification.
func (c *Controller) AfterTenantDelete(request *pluginbridge.BridgeRequestEnvelopeV1) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	return c.runLifecycleDebugHook(request)
}

// BeforeInstallModeChange logs the dynamic plugin install-mode change precondition.
func (c *Controller) BeforeInstallModeChange(request *pluginbridge.BridgeRequestEnvelopeV1) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	return c.runLifecycleDebugHook(request)
}

// AfterInstallModeChange logs the dynamic plugin post-install-mode-change notification.
func (c *Controller) AfterInstallModeChange(request *pluginbridge.BridgeRequestEnvelopeV1) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	return c.runLifecycleDebugHook(request)
}

// runLifecycleDebugHook logs one dynamic lifecycle request and allows the host
// lifecycle operation to continue.
func (c *Controller) runLifecycleDebugHook(
	request *pluginbridge.BridgeRequestEnvelopeV1,
) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	input, err := buildLifecycleDebugInput(request)
	if err != nil {
		return nil, err
	}
	if err = c.dynamicSvc.RunLifecycleDebugHook(input); err != nil {
		return nil, err
	}
	return pluginbridge.WriteJSON(200, &pluginbridge.LifecycleDecision{OK: true})
}

// buildLifecycleDebugInput converts a bridge lifecycle request into the service input.
func buildLifecycleDebugInput(request *pluginbridge.BridgeRequestEnvelopeV1) (*dynamicservice.LifecycleDebugInput, error) {
	input := &dynamicservice.LifecycleDebugInput{}
	if request == nil {
		return input, nil
	}
	input.PluginID = strings.TrimSpace(request.PluginID)
	if request.Request == nil || len(request.Request.Body) == 0 {
		return input, nil
	}
	body := &pluginbridge.LifecycleRequest{}
	if err := json.Unmarshal(request.Request.Body, body); err != nil {
		return nil, err
	}
	if strings.TrimSpace(body.PluginID) != "" {
		input.PluginID = strings.TrimSpace(body.PluginID)
	}
	input.Operation = strings.TrimSpace(body.Operation)
	input.FromVersion = strings.TrimSpace(body.FromVersion)
	input.ToVersion = strings.TrimSpace(body.ToVersion)
	input.TenantID = body.TenantID
	input.FromMode = strings.TrimSpace(body.FromMode)
	input.ToMode = strings.TrimSpace(body.ToMode)
	input.PurgeStorageData = body.PurgeStorageData
	return input, nil
}
