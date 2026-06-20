// This file verifies the linapro-tenant-core provider adapter construction boundary.

package backend

import (
	"context"
	"testing"

	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/plugincap"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
	"lina-core/pkg/plugin/capability/usercap"
)

// TestProvideTenantUsesTypedProviderEnv verifies provider construction only
// needs the narrow tenantspi.ProviderEnv published for tenant capability.
func TestProvideTenantUsesTypedProviderEnv(t *testing.T) {
	provider, err := provideTenant(context.Background(), tenantspi.ProviderEnv{
		PluginID:    pluginID,
		BizCtx:      fakeBizCtx{},
		Users:       fakeTenantProviderUsers{},
		Plugins:     fakeTenantProviderPlugins{},
		PluginAdmin: fakeTenantProviderPluginAdmin{},
	})
	if err != nil {
		t.Fatalf("expected typed tenant provider env to construct provider: %v", err)
	}
	if provider == nil {
		t.Fatal("expected provider instance")
	}
}

// TestProvideTenantRejectsMissingBizCtx verifies provider construction does
// not silently create an adapter without the host-published bizctx service.
func TestProvideTenantRejectsMissingBizCtx(t *testing.T) {
	provider, err := provideTenant(context.Background(), tenantspi.ProviderEnv{
		PluginID: pluginID,
	})
	if err == nil {
		t.Fatal("expected missing bizctx service to fail provider construction")
	}
	if provider != nil {
		t.Fatal("expected nil provider when construction fails")
	}
}

// fakeBizCtx is the minimal host-published business context required by
// linapro-tenant-core provider construction tests.
type fakeBizCtx struct{}

// Current returns a platform context because these tests only cover construction.
func (fakeBizCtx) Current(context.Context) bizctxcap.CurrentContext {
	return bizctxcap.CurrentContext{PlatformBypass: true}
}

// fakeTenantProviderUsers is the minimal user capability required to build the provider.
type fakeTenantProviderUsers struct{}

// Current is unused by provider construction tests.
func (fakeTenantProviderUsers) Current(context.Context, capmodel.CapabilityContext) (*usercap.UserProjection, error) {
	return nil, nil
}

// BatchGet is unused by provider construction tests.
func (fakeTenantProviderUsers) BatchGet(context.Context, capmodel.CapabilityContext, []usercap.UserID) (*capmodel.BatchResult[*usercap.UserProjection, usercap.UserID], error) {
	return &capmodel.BatchResult[*usercap.UserProjection, usercap.UserID]{Items: map[usercap.UserID]*usercap.UserProjection{}}, nil
}

// BatchResolve is unused by provider construction tests.
func (fakeTenantProviderUsers) BatchResolve(context.Context, capmodel.CapabilityContext, usercap.BatchResolveInput) (*capmodel.BatchResult[*usercap.UserProjection, usercap.ResolveKey], error) {
	return &capmodel.BatchResult[*usercap.UserProjection, usercap.ResolveKey]{Items: map[usercap.ResolveKey]*usercap.UserProjection{}}, nil
}

// Search is unused by provider construction tests.
func (fakeTenantProviderUsers) Search(context.Context, capmodel.CapabilityContext, usercap.SearchInput) (*capmodel.PageResult[*usercap.UserProjection], error) {
	return &capmodel.PageResult[*usercap.UserProjection]{}, nil
}

// EnsureVisible is unused by provider construction tests.
func (fakeTenantProviderUsers) EnsureVisible(context.Context, capmodel.CapabilityContext, []usercap.UserID) error {
	return nil
}

// fakeTenantProviderPlugins is the minimal plugin capability required to build the provider.
type fakeTenantProviderPlugins struct{}

// Current is unused by provider construction tests.
func (fakeTenantProviderPlugins) Current(context.Context, capmodel.CapabilityContext) (*plugincap.Projection, error) {
	return nil, nil
}

// BatchGet is unused by provider construction tests.
func (fakeTenantProviderPlugins) BatchGet(context.Context, capmodel.CapabilityContext, []plugincap.PluginID) (*capmodel.BatchResult[*plugincap.Projection, plugincap.PluginID], error) {
	return &capmodel.BatchResult[*plugincap.Projection, plugincap.PluginID]{Items: map[plugincap.PluginID]*plugincap.Projection{}}, nil
}

// Search is unused by provider construction tests.
func (fakeTenantProviderPlugins) Search(context.Context, capmodel.CapabilityContext, plugincap.SearchInput) (*capmodel.PageResult[*plugincap.Projection], error) {
	return &capmodel.PageResult[*plugincap.Projection]{Items: []*plugincap.Projection{}}, nil
}

// Config is unused by provider construction tests.
func (fakeTenantProviderPlugins) Config() plugincap.ConfigService {
	return nil
}

// State is unused by provider construction tests.
func (fakeTenantProviderPlugins) State() plugincap.StateService {
	return nil
}

// Lifecycle is unused by provider construction tests.
func (fakeTenantProviderPlugins) Lifecycle() plugincap.LifecycleService {
	return nil
}

// Registry returns the fake registry service for provider construction tests.
func (s fakeTenantProviderPlugins) Registry() plugincap.RegistryService {
	return s
}

// ListTenantPlugins is unused by provider construction tests.
func (fakeTenantProviderPlugins) ListTenantPlugins(context.Context, capmodel.CapabilityContext, plugincap.TenantListInput) (*capmodel.PageResult[*plugincap.TenantProjection], error) {
	return &capmodel.PageResult[*plugincap.TenantProjection]{Items: []*plugincap.TenantProjection{}}, nil
}

// BatchGetCapabilityStatus is unused by provider construction tests.
func (fakeTenantProviderPlugins) BatchGetCapabilityStatus(_ context.Context, _ capmodel.CapabilityContext, keys []plugincap.CapabilityKey) (*capmodel.BatchResult[*capmodel.CapabilityStatus, plugincap.CapabilityKey], error) {
	result := &capmodel.BatchResult[*capmodel.CapabilityStatus, plugincap.CapabilityKey]{
		Items: map[plugincap.CapabilityKey]*capmodel.CapabilityStatus{},
	}
	for _, key := range keys {
		result.Items[key] = &capmodel.CapabilityStatus{Available: false, Reason: "test_no_provider"}
	}
	return result, nil
}

// fakeTenantProviderPluginAdmin is the minimal plugin admin capability required to build the provider.
type fakeTenantProviderPluginAdmin struct{}

// SetEnabled is unused by provider construction tests.
func (fakeTenantProviderPluginAdmin) SetEnabled(context.Context, capmodel.CapabilityContext, plugincap.PluginID, bool) error {
	return nil
}

// ProvisionTenantDefaults is unused by provider construction tests.
func (fakeTenantProviderPluginAdmin) ProvisionTenantDefaults(context.Context, capmodel.CapabilityContext, capmodel.DomainID) error {
	return nil
}
