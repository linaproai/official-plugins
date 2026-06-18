// This file verifies startup tenant-plugin provisioning exposed through the
// host-facing linapro-tenant-core provider.

package provider

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"

	_ "lina-core/pkg/dbdriver"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/usercap"
	"lina-plugin-linapro-tenant-core/backend/internal/dao"
	"lina-plugin-linapro-tenant-core/backend/internal/model/do"
	"lina-plugin-linapro-tenant-core/backend/internal/service/membership"
	"lina-plugin-linapro-tenant-core/backend/internal/service/resolver"
	"lina-plugin-linapro-tenant-core/backend/internal/service/resolverconfig"
	"lina-plugin-linapro-tenant-core/backend/internal/service/shared"
	"lina-plugin-linapro-tenant-core/backend/internal/service/tenantplugin"
)

// TestProvisionAutoEnabledTenantPluginsAppliesExistingActiveTenants verifies
// startup reconciliation provisions tenant-scoped default plugins for active
// tenants that existed before the host image started.
func TestProvisionAutoEnabledTenantPluginsAppliesExistingActiveTenants(t *testing.T) {
	ctx := context.Background()
	configureProviderProvisioningTestDB(t, ctx)

	const (
		activeTenant    = "provider-active-tenant"
		suspendedTenant = "provider-suspended-tenant"
	)
	activeTenantID := insertProviderProvisioningTenant(t, ctx, activeTenant, shared.TenantStatusActive)
	suspendedTenantID := insertProviderProvisioningTenant(t, ctx, suspendedTenant, shared.TenantStatusSuspended)
	t.Cleanup(func() {
		cleanupProviderProvisioningRows(t, ctx, activeTenantID, suspendedTenantID)
	})

	bizCtxSvc := providerProvisioningBizCtxService{}
	membershipSvc := membership.New(bizCtxSvc, providerProvisioningUsers{})
	resolverConfigSvc := resolverconfig.New()
	tenantPluginSvc := &providerProvisioningTenantPlugins{}
	resolverSvc := resolver.New(bizCtxSvc, membershipSvc)
	providerSvc, err := New(membershipSvc, resolverSvc, resolverConfigSvc, tenantPluginSvc)
	if err != nil {
		t.Fatalf("create provider service failed: %v", err)
	}
	if err := providerSvc.ProvisionAutoEnabledTenantPlugins(ctx); err != nil {
		t.Fatalf("provision existing tenants failed: %v", err)
	}

	if !containsProvisionedTenantID(tenantPluginSvc.provisionedTenantIDs, activeTenantID) {
		t.Fatalf("expected active tenant %d to be provisioned, got %v", activeTenantID, tenantPluginSvc.provisionedTenantIDs)
	}
	if containsProvisionedTenantID(tenantPluginSvc.provisionedTenantIDs, suspendedTenantID) {
		t.Fatalf("expected suspended tenant %d to be skipped, got %v", suspendedTenantID, tenantPluginSvc.provisionedTenantIDs)
	}
}

// providerProvisioningBizCtxService is unused by startup provisioning but
// satisfies the tenantplugin constructor contract.
type providerProvisioningBizCtxService struct{}

// Current returns an empty plugin-visible business context.
func (providerProvisioningBizCtxService) Current(context.Context) bizctxcap.CurrentContext {
	return bizctxcap.CurrentContext{}
}

// providerProvisioningUsers is unused by the provisioning path.
type providerProvisioningUsers struct{}

// Current returns nil because startup provisioning never reads the current user.
func (providerProvisioningUsers) Current(context.Context, capmodel.CapabilityContext) (*usercap.UserProjection, error) {
	return nil, nil
}

// BatchGet returns an empty projection map for provisioning-only tests.
func (providerProvisioningUsers) BatchGet(context.Context, capmodel.CapabilityContext, []usercap.UserID) (*capmodel.BatchResult[*usercap.UserProjection, usercap.UserID], error) {
	return &capmodel.BatchResult[*usercap.UserProjection, usercap.UserID]{Items: map[usercap.UserID]*usercap.UserProjection{}}, nil
}

// BatchResolve returns an empty projection map for provisioning-only tests.
func (providerProvisioningUsers) BatchResolve(context.Context, capmodel.CapabilityContext, usercap.BatchResolveInput) (*capmodel.BatchResult[*usercap.UserProjection, usercap.ResolveKey], error) {
	return &capmodel.BatchResult[*usercap.UserProjection, usercap.ResolveKey]{Items: map[usercap.ResolveKey]*usercap.UserProjection{}}, nil
}

// Search returns an empty page because startup provisioning never searches users.
func (providerProvisioningUsers) Search(context.Context, capmodel.CapabilityContext, usercap.SearchInput) (*capmodel.PageResult[*usercap.UserProjection], error) {
	return &capmodel.PageResult[*usercap.UserProjection]{Items: []*usercap.UserProjection{}}, nil
}

// EnsureVisible accepts all inputs because this fixture is construction-only.
func (providerProvisioningUsers) EnsureVisible(context.Context, capmodel.CapabilityContext, []usercap.UserID) error {
	return nil
}

// providerProvisioningTenantPlugins records provisioning calls.
type providerProvisioningTenantPlugins struct {
	provisionedTenantIDs []int64
}

// List returns no tenant plugin rows because this test only verifies provisioning calls.
func (s *providerProvisioningTenantPlugins) List(context.Context) (*tenantplugin.ListOutput, error) {
	return &tenantplugin.ListOutput{}, nil
}

// SetEnabled accepts updates without mutating state because it is outside this test path.
func (s *providerProvisioningTenantPlugins) SetEnabled(context.Context, string, bool) error {
	return nil
}

// ProvisionForTenant records the tenant targeted by startup provisioning.
func (s *providerProvisioningTenantPlugins) ProvisionForTenant(_ context.Context, tenantID int64) error {
	s.provisionedTenantIDs = append(s.provisionedTenantIDs, tenantID)
	return nil
}

// configureProviderProvisioningTestDB points the provider package test at local PostgreSQL.
func configureProviderProvisioningTestDB(t *testing.T, ctx context.Context) {
	t.Helper()

	originalConfig := gdb.GetAllConfig()
	if err := gdb.SetConfig(gdb.Config{
		gdb.DefaultGroupName: gdb.ConfigGroup{{Link: "pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable"}},
	}); err != nil {
		t.Fatalf("configure provider provisioning test database failed: %v", err)
	}
	db := g.DB()
	t.Cleanup(func() {
		if err := db.Close(ctx); err != nil {
			t.Errorf("close provider provisioning test database failed: %v", err)
		}
		if err := gdb.SetConfig(originalConfig); err != nil {
			t.Errorf("restore provider provisioning test database config failed: %v", err)
		}
	})
}

// insertProviderProvisioningTenant creates one tenant row for provisioning tests.
func insertProviderProvisioningTenant(
	t *testing.T,
	ctx context.Context,
	code string,
	status shared.TenantStatus,
) int64 {
	t.Helper()
	if _, err := dao.Tenant.Ctx(ctx).Unscoped().Where(do.Tenant{Code: code}).Delete(); err != nil {
		t.Fatalf("cleanup stale tenant %s failed: %v", code, err)
	}
	id, err := dao.Tenant.Ctx(ctx).Data(do.Tenant{
		Code:   code,
		Name:   code,
		Status: string(status),
		Remark: "",
	}).OmitEmptyData().InsertAndGetId()
	if err != nil {
		t.Fatalf("insert tenant %s failed: %v", code, err)
	}
	return id
}

// cleanupProviderProvisioningRows removes rows created by provider provisioning tests.
func cleanupProviderProvisioningRows(t *testing.T, ctx context.Context, tenantIDs ...int64) {
	t.Helper()
	if len(tenantIDs) > 0 {
		if _, err := dao.Tenant.Ctx(ctx).Unscoped().WhereIn(dao.Tenant.Columns().Id, tenantIDs).Delete(); err != nil {
			t.Errorf("cleanup tenant rows failed: %v", err)
		}
	}
}

// containsProvisionedTenantID reports whether a provisioning call targeted tenantID.
func containsProvisionedTenantID(values []int64, tenantID int64) bool {
	for _, value := range values {
		if value == tenantID {
			return true
		}
	}
	return false
}
