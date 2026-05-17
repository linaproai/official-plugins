// This file verifies lifecycle precondition tenant-existence checks.

package lifecycleprecondition

import (
	"context"
	"testing"

	"lina-core/pkg/pluginhost"
)

// TestPreconditionRejectsSuspendedTenantBeforePluginRemoval verifies non-deleted
// suspended tenants still block disabling or uninstalling the multi-tenant
// plugin after the archive lifecycle is removed.
func TestPreconditionRejectsSuspendedTenantBeforePluginRemoval(t *testing.T) {
	ctx := context.Background()
	counter := &lifecyclePreconditionTestCounter{count: 1}
	checker, err := New(counter)
	if err != nil {
		t.Fatalf("create lifecycle precondition checker failed: %v", err)
	}
	input := pluginhost.NewSourcePluginLifecycleInput("multi-tenant", pluginhost.LifecycleHookBeforeUninstall.String())
	if ok, reason, err := checker.BeforeUninstall(ctx, input); err != nil || ok || reason != ReasonUninstallTenantsExist {
		t.Fatalf("expected suspended tenant to block uninstall, ok=%v reason=%q err=%v", ok, reason, err)
	}
	input = pluginhost.NewSourcePluginLifecycleInputWithUninstallPolicy(
		"multi-tenant",
		pluginhost.LifecycleHookBeforeUninstall.String(),
		true,
	)
	if ok, reason, err := checker.BeforeUninstall(ctx, input); err != nil || !ok || reason != "" {
		t.Fatalf("expected data-purging uninstall to allow tenant cleanup, ok=%v reason=%q err=%v", ok, reason, err)
	}
	input = pluginhost.NewSourcePluginLifecycleInput("multi-tenant", pluginhost.LifecycleHookBeforeDisable.String())
	if ok, reason, err := checker.BeforeDisable(ctx, input); err != nil || ok || reason != ReasonDisableTenantsExist {
		t.Fatalf("expected suspended tenant to block disable, ok=%v reason=%q err=%v", ok, reason, err)
	}

	counter.count = 0
	input = pluginhost.NewSourcePluginLifecycleInput("multi-tenant", pluginhost.LifecycleHookBeforeUninstall.String())
	if ok, reason, err := checker.BeforeUninstall(ctx, input); err != nil || !ok || reason != "" {
		t.Fatalf("expected no tenant count not to block uninstall, ok=%v reason=%q err=%v", ok, reason, err)
	}
}

// TestNewRequiresTenantCounter verifies dependency construction fails fast.
func TestNewRequiresTenantCounter(t *testing.T) {
	if _, err := New(nil); err == nil {
		t.Fatal("expected missing tenant counter to fail")
	}
}

// lifecyclePreconditionTestCounter is a deterministic tenant counter for
// lifecycle precondition tests.
type lifecyclePreconditionTestCounter struct {
	count int
}

// CountExisting returns the configured non-deleted tenant count.
func (c *lifecyclePreconditionTestCounter) CountExisting(ctx context.Context) (int, error) {
	return c.count, nil
}
