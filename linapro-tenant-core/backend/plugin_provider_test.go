// This file verifies the linapro-tenant-core provider adapter construction boundary.

package backend

import (
	"context"
	"testing"

	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/plugincap"
	"lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
	"lina-core/pkg/plugin/capability/usercap"
	"lina-core/pkg/statusflag"
)

// TestProvideTenantUsesTypedProviderEnv verifies provider construction only
// needs the typed tenantspi.ProviderEnv published for tenant capability.
func TestProvideTenantUsesTypedProviderEnv(t *testing.T) {
	provider, err := provideTenant(context.Background(), tenantspi.ProviderEnv{
		PluginID: pluginID,
		BizCtx:   fakeBizCtx{},
		Tenant:   fakeTenantProviderTenant{plugins: fakeTenantProviderPluginGovernance{}},
		Users:    fakeTenantProviderUsers{},
		Plugins:  fakeTenantProviderPlugins{},
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
func (fakeTenantProviderUsers) Current(context.Context) (*usercap.UserInfo, error) {
	return nil, nil
}

// Get is unused by provider construction tests.
func (fakeTenantProviderUsers) Get(context.Context, usercap.UserID) (*usercap.UserInfo, error) {
	return nil, nil
}

// BatchGet is unused by provider construction tests.
func (fakeTenantProviderUsers) BatchGet(context.Context, []usercap.UserID) (*capmodel.BatchResult[*usercap.UserInfo, usercap.UserID], error) {
	return &capmodel.BatchResult[*usercap.UserInfo, usercap.UserID]{Items: map[usercap.UserID]*usercap.UserInfo{}}, nil
}

// BatchResolve is unused by provider construction tests.
func (fakeTenantProviderUsers) BatchResolve(context.Context, usercap.BatchResolveInput) (*capmodel.BatchResult[*usercap.UserInfo, usercap.ResolveKey], error) {
	return &capmodel.BatchResult[*usercap.UserInfo, usercap.ResolveKey]{Items: map[usercap.ResolveKey]*usercap.UserInfo{}}, nil
}

// List is unused by provider construction tests.
func (fakeTenantProviderUsers) List(context.Context, usercap.ListInput) (*capmodel.PageResult[*usercap.UserInfo], error) {
	return &capmodel.PageResult[*usercap.UserInfo]{}, nil
}

// EnsureVisible is unused by provider construction tests.
func (fakeTenantProviderUsers) EnsureVisible(context.Context, []usercap.UserID) error {
	return nil
}

// Create is unused by provider construction tests.
func (fakeTenantProviderUsers) Create(context.Context, usercap.CreateInput) (usercap.UserID, error) {
	return "", nil
}

// Update is unused by provider construction tests.
func (fakeTenantProviderUsers) Update(context.Context, usercap.UpdateInput) error {
	return nil
}

// Delete is unused by provider construction tests.
func (fakeTenantProviderUsers) Delete(context.Context, usercap.UserID) error {
	return nil
}

// SetStatus is unused by provider construction tests.
func (fakeTenantProviderUsers) SetStatus(context.Context, usercap.UserID, statusflag.Enabled) error {
	return nil
}

// ResetPassword is unused by provider construction tests.
func (fakeTenantProviderUsers) ResetPassword(context.Context, usercap.UserID, string) error {
	return nil
}

// Assignment returns user-role assignment operations unused by provider construction tests.
func (fakeTenantProviderUsers) Assignment() usercap.AssignmentService {
	return fakeTenantProviderUserAssignments{}
}

// fakeTenantProviderUserAssignments accepts unused role replacements.
type fakeTenantProviderUserAssignments struct{}

// ReplaceRoles is unused by provider construction tests.
func (fakeTenantProviderUserAssignments) ReplaceRoles(context.Context, usercap.UserID, []int) error {
	return nil
}

// fakeTenantProviderPlugins is the minimal plugin capability required to build the provider.
type fakeTenantProviderPlugins struct{}

// Current is unused by provider construction tests.
func (fakeTenantProviderPlugins) Current(context.Context) (*plugincap.PluginInfo, error) {
	return nil, nil
}

// Get is unused by provider construction tests.
func (fakeTenantProviderPlugins) Get(context.Context, plugincap.PluginID) (*plugincap.PluginInfo, error) {
	return nil, nil
}

// BatchGet is unused by provider construction tests.
func (fakeTenantProviderPlugins) BatchGet(context.Context, []plugincap.PluginID) (*capmodel.BatchResult[*plugincap.PluginInfo, plugincap.PluginID], error) {
	return &capmodel.BatchResult[*plugincap.PluginInfo, plugincap.PluginID]{Items: map[plugincap.PluginID]*plugincap.PluginInfo{}}, nil
}

// List is unused by provider construction tests.
func (fakeTenantProviderPlugins) List(context.Context, plugincap.ListInput) (*capmodel.PageResult[*plugincap.PluginInfo], error) {
	return &capmodel.PageResult[*plugincap.PluginInfo]{Items: []*plugincap.PluginInfo{}}, nil
}

// Config is unused by provider construction tests.
func (fakeTenantProviderPlugins) Config() plugincap.ConfigService {
	return nil
}

// Registry returns the fake registry service for provider construction tests.
func (s fakeTenantProviderPlugins) Registry() plugincap.RegistryService {
	return s
}

// State returns the fake plugin state service for provider construction tests.
func (s fakeTenantProviderPlugins) State() plugincap.StateService {
	return s
}

// Lifecycle is unused by provider construction tests.
func (fakeTenantProviderPlugins) Lifecycle() plugincap.LifecycleService {
	return nil
}

// ListTenantPlugins is unused by provider construction tests.
func (fakeTenantProviderPlugins) ListTenantPlugins(context.Context, plugincap.TenantListInput) (*capmodel.PageResult[*plugincap.TenantPluginInfo], error) {
	return &capmodel.PageResult[*plugincap.TenantPluginInfo]{Items: []*plugincap.TenantPluginInfo{}}, nil
}

// IsEnabled is unused by provider construction tests.
func (fakeTenantProviderPlugins) IsEnabled(context.Context, plugincap.PluginID) (bool, error) {
	return false, nil
}

// IsProviderEnabled is unused by provider construction tests.
func (fakeTenantProviderPlugins) IsProviderEnabled(context.Context, plugincap.PluginID) (bool, error) {
	return false, nil
}

// IsEnabledAuthoritative is unused by provider construction tests.
func (fakeTenantProviderPlugins) IsEnabledAuthoritative(context.Context, plugincap.PluginID) (bool, error) {
	return false, nil
}

// fakeTenantProviderTenant is the minimal tenant capability required by provider construction tests.
type fakeTenantProviderTenant struct {
	tenantcap.Service
	plugins tenantcap.PluginService
}

// Plugins returns tenant plugin-governance operations for provider construction tests.
func (s fakeTenantProviderTenant) Plugins() tenantcap.PluginService {
	return s.plugins
}

// fakeTenantProviderPluginGovernance is the minimal tenant plugin-governance
// writer required to build the provider.
type fakeTenantProviderPluginGovernance struct{}

// SetTenantPluginEnabled is unused by provider construction tests.
func (fakeTenantProviderPluginGovernance) SetTenantPluginEnabled(context.Context, plugincap.PluginID, bool) error {
	return nil
}

// ProvisionTenantPluginDefaults is unused by provider construction tests.
func (fakeTenantProviderPluginGovernance) ProvisionTenantPluginDefaults(context.Context, capmodel.DomainID) error {
	return nil
}
