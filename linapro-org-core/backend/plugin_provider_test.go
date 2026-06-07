// This file verifies the linapro-org-core provider adapter construction boundary.

package backend

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/orgcap"
	"lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/capability/usercap"
)

// TestProvideOrgUsesTypedProviderEnv verifies provider construction only needs
// the narrow orgcap.ProviderEnv published for organization capability.
func TestProvideOrgUsesTypedProviderEnv(t *testing.T) {
	provider, err := provideOrg(context.Background(), orgcap.ProviderEnv{
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
	provider, err := provideOrg(context.Background(), orgcap.ProviderEnv{
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
func (fakeTenantFilter) Context(context.Context) tenantcap.TenantFilterContext {
	return tenantcap.TenantFilterContext{TenantID: 1}
}

// Apply returns the input model unchanged because these tests do not execute queries.
func (fakeTenantFilter) Apply(_ context.Context, model *gdb.Model, _ string) *gdb.Model {
	return model
}

// fakeUsers is the minimal usercap dependency required for provider
// construction tests; methods are not executed by these construction checks.
type fakeUsers struct{}

// BatchGetUsers returns an empty visible projection set.
func (fakeUsers) BatchGetUsers(context.Context, capmodel.CapabilityContext, []usercap.UserID) (*capmodel.BatchResult[*usercap.UserProjection, usercap.UserID], error) {
	return &capmodel.BatchResult[*usercap.UserProjection, usercap.UserID]{Items: map[usercap.UserID]*usercap.UserProjection{}}, nil
}

// SearchUsers returns an empty page.
func (fakeUsers) SearchUsers(context.Context, capmodel.CapabilityContext, usercap.SearchInput) (*capmodel.PageResult[*usercap.UserProjection], error) {
	return &capmodel.PageResult[*usercap.UserProjection]{Items: []*usercap.UserProjection{}}, nil
}

// EnsureUsersVisible accepts all users for construction-only tests.
func (fakeUsers) EnsureUsersVisible(context.Context, capmodel.CapabilityContext, []usercap.UserID) error {
	return nil
}
