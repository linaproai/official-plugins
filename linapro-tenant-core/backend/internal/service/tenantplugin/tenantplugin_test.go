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
	"lina-core/pkg/plugin/capability/tenantcap"
)

// TestListUsesPlugincapProjection verifies tenant plugin list projection is
// assembled from the host plugincap service.
func TestListUsesPlugincapProjection(t *testing.T) {
	plugins := &fakePlugincap{
		list: []*plugincap.TenantPluginInfo{
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
	governance := &fakePluginGovernance{}
	svc := New(testBizContextService{tenantID: 1001}, fakeTenantDomain{plugins: governance}, plugins)

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
	if plugins.listCalled != 1 {
		t.Fatalf("expected one tenant plugin list call, got %d", plugins.listCalled)
	}
}

// TestSetEnabledRunsLifecycleBeforePlugincap verifies disable preconditions run
// before tenant plugin state mutation.
func TestSetEnabledRunsLifecycleBeforePlugincap(t *testing.T) {
	var (
		vetoErr    = errors.New("disable vetoed")
		governance = &fakePluginGovernance{}
		lifecycle  = &fakeLifecycle{disableErr: vetoErr}
		plugins    = &fakePlugincap{lifecycle: lifecycle}
		svc        = New(testBizContextService{tenantID: 1002}, fakeTenantDomain{plugins: governance}, plugins)
	)

	err := svc.SetEnabled(context.Background(), "demo-plugin", false)
	if !errors.Is(err, vetoErr) {
		t.Fatalf("expected lifecycle veto, got %v", err)
	}
	if governance.setCalls != 0 {
		t.Fatalf("expected plugin governance not called after veto, got %d", governance.setCalls)
	}
}

// TestProvisionForTenantDelegatesToPlugincap verifies startup provisioning only
// delegates tenant IDs to the host plugin-governance owner.
func TestProvisionForTenantDelegatesToPlugincap(t *testing.T) {
	var (
		plugins    = &fakePlugincap{}
		governance = &fakePluginGovernance{}
		svc        = New(testBizContextService{}, fakeTenantDomain{plugins: governance}, plugins)
	)

	if err := svc.ProvisionForTenant(context.Background(), 1003); err != nil {
		t.Fatalf("provision tenant plugins failed: %v", err)
	}
	if governance.provisionTenantID != capmodel.DomainID("1003") {
		t.Fatalf("expected provision tenant 1003, got %q", governance.provisionTenantID)
	}
}

// TestRequireTenantRejectsPlatform verifies tenant plugin governance requires a tenant context.
func TestRequireTenantRejectsPlatform(t *testing.T) {
	var (
		plugins    = &fakePlugincap{}
		governance = &fakePluginGovernance{}
		svc        = New(testBizContextService{}, fakeTenantDomain{plugins: governance}, plugins)
	)

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

type fakeTenantDomain struct {
	tenantcap.Service
	plugins tenantcap.PluginService
}

func (s fakeTenantDomain) Plugins() tenantcap.PluginService {
	return s.plugins
}

type fakePlugincap struct {
	list       []*plugincap.TenantPluginInfo
	listCalled int
	lifecycle  plugincap.LifecycleService
}

func (s *fakePlugincap) BatchGet(context.Context, []plugincap.PluginID) (*capmodel.BatchResult[*plugincap.PluginInfo, plugincap.PluginID], error) {
	return &capmodel.BatchResult[*plugincap.PluginInfo, plugincap.PluginID]{Items: map[plugincap.PluginID]*plugincap.PluginInfo{}}, nil
}

func (s *fakePlugincap) Current(context.Context) (*plugincap.PluginInfo, error) {
	return nil, nil
}

func (s *fakePlugincap) Get(context.Context, plugincap.PluginID) (*plugincap.PluginInfo, error) {
	return nil, nil
}

func (s *fakePlugincap) List(context.Context, plugincap.ListInput) (*capmodel.PageResult[*plugincap.PluginInfo], error) {
	return &capmodel.PageResult[*plugincap.PluginInfo]{Items: []*plugincap.PluginInfo{}}, nil
}

func (s *fakePlugincap) Config() plugincap.ConfigService {
	return nil
}

func (s *fakePlugincap) Registry() plugincap.RegistryService {
	return s
}

func (s *fakePlugincap) State() plugincap.StateService {
	return s
}

func (s *fakePlugincap) Lifecycle() plugincap.LifecycleService {
	return s.lifecycle
}

func (s *fakePlugincap) ListTenantPlugins(context.Context, plugincap.TenantListInput) (*capmodel.PageResult[*plugincap.TenantPluginInfo], error) {
	s.listCalled++
	return &capmodel.PageResult[*plugincap.TenantPluginInfo]{Items: s.list, Total: len(s.list)}, nil
}

func (s *fakePlugincap) IsEnabled(context.Context, plugincap.PluginID) (bool, error) {
	return false, nil
}

func (s *fakePlugincap) IsProviderEnabled(context.Context, plugincap.PluginID) (bool, error) {
	return false, nil
}

func (s *fakePlugincap) IsEnabledAuthoritative(context.Context, plugincap.PluginID) (bool, error) {
	return false, nil
}

type fakePluginGovernance struct {
	setCalls          int
	setPluginID       plugincap.PluginID
	setEnabled        bool
	provisionTenantID capmodel.DomainID
}

func (s *fakePluginGovernance) SetTenantPluginEnabled(_ context.Context, id plugincap.PluginID, enabled bool) error {
	s.setCalls++
	s.setPluginID = id
	s.setEnabled = enabled
	return nil
}

func (s *fakePluginGovernance) ProvisionTenantPluginDefaults(_ context.Context, tenantID capmodel.DomainID) error {
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
