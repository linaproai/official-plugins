// This file verifies the linapro-org-core provider adapter construction boundary.

package backend

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/orgcap/orgspi"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
	"lina-core/pkg/plugin/capability/usercap"
)

// TestProvideOrgUsesTypedProviderEnv verifies provider construction only needs
// the narrow orgspi.ProviderEnv published for organization capability.
func TestProvideOrgUsesTypedProviderEnv(t *testing.T) {
	provider, err := provideOrg(context.Background(), orgspi.ProviderEnv{
		PluginID:     pluginID,
		TenantFilter: fakeTenantFilter{},
		Users:        fakeUsers{},
	})
	if err != nil {
		t.Fatalf("expected typed org provider env to construct provider: %v", err)
	}
	if provider == nil {
		t.Fatal("expected provider instance")
	}
}

// TestProvideOrgRejectsMissingTenantFilter verifies provider construction does
// not silently create an adapter without the host-published tenant filter.
func TestProvideOrgRejectsMissingTenantFilter(t *testing.T) {
	provider, err := provideOrg(context.Background(), orgspi.ProviderEnv{
		PluginID: pluginID,
	})
	if err == nil {
		t.Fatal("expected missing tenant filter to fail provider construction")
	}
	if provider != nil {
		t.Fatal("expected nil provider when construction fails")
	}
}

// fakeTenantFilter is the minimal host-published tenant filter required by
// linapro-org-core provider construction tests.
type fakeTenantFilter struct{}

// Context returns a deterministic tenant context for construction-only tests.
func (fakeTenantFilter) Context(context.Context) tenantspi.TenantFilterContext {
	return tenantspi.TenantFilterContext{TenantID: 1}
}

// Apply returns the input model unchanged because these tests do not execute queries.
func (fakeTenantFilter) Apply(_ context.Context, model *gdb.Model, _ string) *gdb.Model {
	return model
}

// fakeUsers is the minimal usercap dependency required for provider
// construction tests; methods are not executed by these construction checks.
type fakeUsers struct{}

// Current returns no current user projection for construction-only tests.
func (fakeUsers) Current(context.Context, capmodel.CapabilityContext) (*usercap.UserProjection, error) {
	return nil, nil
}

// BatchGet returns an empty visible projection set.
func (fakeUsers) BatchGet(context.Context, capmodel.CapabilityContext, []usercap.UserID) (*capmodel.BatchResult[*usercap.UserProjection, usercap.UserID], error) {
	return &capmodel.BatchResult[*usercap.UserProjection, usercap.UserID]{Items: map[usercap.UserID]*usercap.UserProjection{}}, nil
}

// BatchResolve returns an empty visible projection set.
func (fakeUsers) BatchResolve(context.Context, capmodel.CapabilityContext, usercap.BatchResolveInput) (*capmodel.BatchResult[*usercap.UserProjection, usercap.ResolveKey], error) {
	return &capmodel.BatchResult[*usercap.UserProjection, usercap.ResolveKey]{Items: map[usercap.ResolveKey]*usercap.UserProjection{}}, nil
}

// Search returns an empty page.
func (fakeUsers) Search(context.Context, capmodel.CapabilityContext, usercap.SearchInput) (*capmodel.PageResult[*usercap.UserProjection], error) {
	return &capmodel.PageResult[*usercap.UserProjection]{Items: []*usercap.UserProjection{}}, nil
}

// EnsureVisible accepts all users for construction-only tests.
func (fakeUsers) EnsureVisible(context.Context, capmodel.CapabilityContext, []usercap.UserID) error {
	return nil
}
