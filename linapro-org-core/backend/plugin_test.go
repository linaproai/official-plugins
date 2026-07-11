// This file verifies the linapro-org-core provider adapter construction boundary.

package backend

import (
	"context"
	"testing"

	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/orgcap/orgspi"
	"lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/capability/usercap"
	"lina-core/pkg/statusflag"
)

// TestProvideOrgUsesTypedProviderEnv verifies provider construction only needs
// the narrow orgspi.ProviderEnv published for organization capability.
func TestProvideOrgUsesTypedProviderEnv(t *testing.T) {
	provider, err := provideOrg(context.Background(), orgspi.ProviderEnv{
		PluginID: pluginID,
		Tenant:   fakeTenantService{},
		Users:    fakeUsers{},
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

// fakeTenantService is the minimal host-published tenant service required by
// linapro-org-core provider construction tests.
type fakeTenantService struct {
	tenantcap.Service
}

// Filter returns a deterministic tenant filter service for construction-only tests.
func (fakeTenantService) Filter() tenantcap.FilterService {
	return fakeTenantFilter{}
}

// fakeTenantFilter is the minimal host-published tenant filter required by
// linapro-org-core provider construction tests.
type fakeTenantFilter struct{}

// Context returns a deterministic tenant context for construction-only tests.
func (fakeTenantFilter) Context(context.Context) tenantcap.TenantFilterContext {
	return tenantcap.TenantFilterContext{TenantID: 1}
}

// fakeUsers is the minimal usercap dependency required for provider
// construction tests; methods are not executed by these construction checks.
type fakeUsers struct{}

// Current returns no current user projection for construction-only tests.
func (fakeUsers) Current(context.Context) (*usercap.UserInfo, error) {
	return nil, nil
}

// Get returns no user projection for construction-only tests.
func (fakeUsers) Get(context.Context, usercap.UserID) (*usercap.UserInfo, error) {
	return nil, nil
}

// BatchGet returns an empty visible projection set.
func (fakeUsers) BatchGet(context.Context, []usercap.UserID) (*capmodel.BatchResult[*usercap.UserInfo, usercap.UserID], error) {
	return &capmodel.BatchResult[*usercap.UserInfo, usercap.UserID]{Items: map[usercap.UserID]*usercap.UserInfo{}}, nil
}

// BatchResolve returns an empty visible projection set.
func (fakeUsers) BatchResolve(context.Context, usercap.BatchResolveInput) (*capmodel.BatchResult[*usercap.UserInfo, usercap.ResolveKey], error) {
	return &capmodel.BatchResult[*usercap.UserInfo, usercap.ResolveKey]{Items: map[usercap.ResolveKey]*usercap.UserInfo{}}, nil
}

// List returns an empty page.
func (fakeUsers) List(context.Context, usercap.ListInput) (*capmodel.PageResult[*usercap.UserInfo], error) {
	return &capmodel.PageResult[*usercap.UserInfo]{Items: []*usercap.UserInfo{}}, nil
}

// EnsureVisible accepts all users for construction-only tests.
func (fakeUsers) EnsureVisible(context.Context, []usercap.UserID) error {
	return nil
}

// Create is unused by construction-only tests.
func (fakeUsers) Create(context.Context, usercap.CreateInput) (usercap.UserID, error) {
	return "", nil
}

func (fakeUsers) CreateFromExternal(context.Context, usercap.CreateFromExternalInput) (usercap.UserID, error) {
	return "", nil
}

// Update is unused by construction-only tests.
func (fakeUsers) Update(context.Context, usercap.UpdateInput) error {
	return nil
}

// Delete is unused by construction-only tests.
func (fakeUsers) Delete(context.Context, usercap.UserID) error {
	return nil
}

// SetStatus is unused by construction-only tests.
func (fakeUsers) SetStatus(context.Context, usercap.UserID, statusflag.Enabled) error {
	return nil
}

// ResetPassword is unused by construction-only tests.
func (fakeUsers) ResetPassword(context.Context, usercap.UserID, string) error {
	return nil
}

// Assignment returns user-role assignment operations unused by construction-only tests.
func (fakeUsers) Assignment() usercap.AssignmentService {
	return fakeUserAssignments{}
}

// fakeUserAssignments accepts unused role replacements.
type fakeUserAssignments struct{}

// ReplaceRoles is unused by construction-only tests.
func (fakeUserAssignments) ReplaceRoles(context.Context, usercap.UserID, []int) error {
	return nil
}
