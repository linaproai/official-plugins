// This file verifies the linapro-ops-demo-guard lifecycle gate that prevents
// accidental management-page installation while allowing startup configuration.

package backend

import (
	"context"
	"testing"

	"lina-core/pkg/pluginhost"
	middlewaresvc "lina-plugin-linapro-ops-demo-guard/backend/internal/service/middleware"
)

// TestBeforeInstallRejectsManualInstall verifies management-page installs are vetoed.
func TestBeforeInstallRejectsManualInstall(t *testing.T) {
	input := pluginhost.NewSourcePluginLifecycleInput(pluginID, pluginhost.LifecycleHookBeforeInstall.String())
	ok, reason, err := beforeInstall(context.Background(), input)
	if err != nil {
		t.Fatalf("expected manual install veto without technical error, got %v", err)
	}
	if ok {
		t.Fatal("expected manual install to be vetoed")
	}
	expectedReason := middlewaresvc.CodeDemoControlInstallManualDenied.MessageKey()
	if reason != expectedReason {
		t.Fatalf("expected veto reason %q, got %q", expectedReason, reason)
	}
}

// TestBeforeInstallAllowsStartupAutoEnable verifies plugin.autoEnable startup
// bootstrap can install demo protection intentionally.
func TestBeforeInstallAllowsStartupAutoEnable(t *testing.T) {
	input := pluginhost.NewSourcePluginLifecycleInputWithPolicy(
		pluginID,
		pluginhost.LifecycleHookBeforeInstall.String(),
		pluginhost.SourcePluginLifecyclePolicy{StartupAutoEnable: true},
	)
	ok, reason, err := beforeInstall(context.Background(), input)
	if err != nil {
		t.Fatalf("expected startup install to avoid technical error, got %v", err)
	}
	if !ok {
		t.Fatalf("expected startup install to be allowed, got reason %q", reason)
	}
	if reason != "" {
		t.Fatalf("expected empty allow reason, got %q", reason)
	}
}
