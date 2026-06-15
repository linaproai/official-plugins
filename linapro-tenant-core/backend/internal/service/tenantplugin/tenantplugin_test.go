// This file verifies tenant plugin governance delegates host-owned plugin state
// reads and writes to plugincap instead of touching host plugin tables.

package tenantplugin

import (
	"context"
	"errors"
	"testing"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/plugincap"
)

// TestListUsesPlugincapProjection verifies tenant plugin list projection is
// assembled from the host plugincap service.
func TestListUsesPlugincapProjection(t *testing.T) {
	plugins := &fakePlugincap{
		list: []*plugincap.TenantProjection{
			{
				ID:            "demo-plugin",
				Name:          "Demo Plugin",
				Version:       "v1.0.0",
				Type:          "source",
				Description:   "Demo",
				Installed:     true,
				Enabled:       true,
				ScopeNature:   "tenant_aware",
				InstallMode:   "tenant_scoped",
				TenantEnabled: true,
			},
		},
	}
	svc := New(testBizContextService{tenantID: 1001}, nil, plugins, plugins)

	out, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("list tenant plugins failed: %v", err)
	}
	if out.Total != 1 || len(out.List) != 1 {
		t.Fatalf("unexpected list output: %#v", out)
	}
	item := out.List[0]
	if item.Id != "demo-plugin" || item.Installed != 1 || item.Enabled != 1 || item.TenantEnabled != 1 {
		t.Fatalf("unexpected plugin projection: %#v", item)
	}
	if plugins.lastListTenantID != capmodel.DomainID("1001") {
		t.Fatalf("expected tenant context 1001, got %q", plugins.lastListTenantID)
	}
}

// TestSetEnabledRunsLifecycleBeforePlugincap verifies disable preconditions run
// before tenant plugin state mutation.
func TestSetEnabledRunsLifecycleBeforePlugincap(t *testing.T) {
	vetoErr := errors.New("disable vetoed")
	plugins := &fakePlugincap{}
	lifecycle := &fakeLifecycle{disableErr: vetoErr}
	svc := New(testBizContextService{tenantID: 1002}, lifecycle, plugins, plugins)

	err := svc.SetEnabled(context.Background(), "demo-plugin", false)
	if !errors.Is(err, vetoErr) {
		t.Fatalf("expected lifecycle veto, got %v", err)
	}
	if plugins.setCalls != 0 {
		t.Fatalf("expected plugincap not called after veto, got %d", plugins.setCalls)
	}
}

// TestProvisionForTenantDelegatesToPlugincap verifies startup provisioning only
// delegates tenant IDs to the host plugin-governance owner.
func TestProvisionForTenantDelegatesToPlugincap(t *testing.T) {
	plugins := &fakePlugincap{}
	svc := New(testBizContextService{}, nil, plugins, plugins)

	if err := svc.ProvisionForTenant(context.Background(), 1003); err != nil {
		t.Fatalf("provision tenant plugins failed: %v", err)
	}
	if plugins.provisionTenantID != capmodel.DomainID("1003") {
		t.Fatalf("expected provision tenant 1003, got %q", plugins.provisionTenantID)
	}
}

// TestRequireTenantRejectsPlatform verifies tenant plugin governance requires a tenant context.
func TestRequireTenantRejectsPlatform(t *testing.T) {
	plugins := &fakePlugincap{}
	svc := New(testBizContextService{}, nil, plugins, plugins)

	_, err := svc.List(context.Background())
	if !bizerr.Is(err, CodeTenantRequired) {
		t.Fatalf("expected tenant required error, got %v", err)
	}
}

type testBizContextService struct {
	tenantID int
	userID   int
}

func (s testBizContextService) Current(context.Context) bizctxcap.CurrentContext {
	return bizctxcap.CurrentContext{TenantID: s.tenantID, UserID: s.userID}
}

type fakePlugincap struct {
	list              []*plugincap.TenantProjection
	lastListTenantID  capmodel.DomainID
	setCalls          int
	setPluginID       plugincap.PluginID
	setEnabled        bool
	provisionTenantID capmodel.DomainID
}

func (s *fakePlugincap) BatchGet(context.Context, capmodel.CapabilityContext, []plugincap.PluginID) (*capmodel.BatchResult[*plugincap.Projection, plugincap.PluginID], error) {
	return &capmodel.BatchResult[*plugincap.Projection, plugincap.PluginID]{Items: map[plugincap.PluginID]*plugincap.Projection{}}, nil
}

func (s *fakePlugincap) Config() plugincap.ConfigService {
	return nil
}

func (s *fakePlugincap) State() plugincap.StateService {
	return nil
}

func (s *fakePlugincap) Lifecycle() plugincap.LifecycleService {
	return nil
}

func (s *fakePlugincap) Registry() plugincap.RegistryService {
	return s
}

func (s *fakePlugincap) ListTenantPlugins(_ context.Context, capCtx capmodel.CapabilityContext) (*capmodel.PageResult[*plugincap.TenantProjection], error) {
	s.lastListTenantID = capCtx.TenantID
	return &capmodel.PageResult[*plugincap.TenantProjection]{Items: s.list, Total: len(s.list)}, nil
}

func (s *fakePlugincap) SetEnabled(_ context.Context, _ capmodel.CapabilityContext, id plugincap.PluginID, enabled bool) error {
	s.setCalls++
	s.setPluginID = id
	s.setEnabled = enabled
	return nil
}

func (s *fakePlugincap) ProvisionTenantDefaults(_ context.Context, _ capmodel.CapabilityContext, tenantID capmodel.DomainID) error {
	s.provisionTenantID = tenantID
	return nil
}

type fakeLifecycle struct {
	disableErr error
}

func (s *fakeLifecycle) EnsureTenantPluginDisableAllowed(context.Context, string, int) error {
	return s.disableErr
}

func (s *fakeLifecycle) NotifyTenantPluginDisabled(context.Context, string, int) {}

func (s *fakeLifecycle) EnsureTenantDeleteAllowed(context.Context, int) error { return nil }

func (s *fakeLifecycle) NotifyTenantDeleted(context.Context, int) {}
