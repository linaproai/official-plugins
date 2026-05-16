// This file verifies lifecycle precondition tenant-existence checks.

package lifecycleprecondition

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"

	_ "lina-core/pkg/dbdriver"
	"lina-core/pkg/pluginhost"
	"lina-plugin-multi-tenant/backend/internal/service/shared"
	"lina-plugin-multi-tenant/backend/internal/service/tenant"
)

// lifecyclePreconditionTestTenantData is a typed insert payload for precondition tests.
type lifecyclePreconditionTestTenantData struct {
	Code   string `orm:"code"`
	Name   string `orm:"name"`
	Status string `orm:"status"`
}

// TestPreconditionRejectsSuspendedTenantBeforePluginRemoval verifies non-deleted
// suspended tenants still block disabling or uninstalling the multi-tenant
// plugin after the archive lifecycle is removed.
func TestPreconditionRejectsSuspendedTenantBeforePluginRemoval(t *testing.T) {
	ctx := context.Background()
	configureLifecyclePreconditionTestDB(t, ctx)

	tenantID, err := shared.Model(ctx, shared.TableTenant).Data(lifecyclePreconditionTestTenantData{
		Code:   "lifecycle-precondition-suspended-test",
		Name:   "Lifecycle Precondition Suspended Test",
		Status: string(shared.TenantStatusSuspended),
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert suspended tenant failed: %v", err)
	}
	t.Cleanup(func() {
		if _, err := shared.Model(ctx, shared.TableTenant).Unscoped().Where("id", tenantID).Delete(); err != nil {
			t.Errorf("cleanup suspended tenant failed: %v", err)
		}
	})

	checker, err := New(tenant.ExistingCounter{})
	if err != nil {
		t.Fatalf("create lifecycle precondition checker failed: %v", err)
	}
	input := pluginhost.NewSourcePluginLifecycleInput("multi-tenant", pluginhost.LifecycleHookBeforeUninstall.String())
	if ok, reason, err := checker.BeforeUninstall(ctx, input); err != nil || ok || reason != ReasonUninstallTenantsExist {
		t.Fatalf("expected suspended tenant to block uninstall, ok=%v reason=%q err=%v", ok, reason, err)
	}
	input = pluginhost.NewSourcePluginLifecycleInput("multi-tenant", pluginhost.LifecycleHookBeforeDisable.String())
	if ok, reason, err := checker.BeforeDisable(ctx, input); err != nil || ok || reason != ReasonDisableTenantsExist {
		t.Fatalf("expected suspended tenant to block disable, ok=%v reason=%q err=%v", ok, reason, err)
	}

	if _, err := shared.Model(ctx, shared.TableTenant).Where("id", tenantID).Delete(); err != nil {
		t.Fatalf("soft delete suspended tenant failed: %v", err)
	}
	input = pluginhost.NewSourcePluginLifecycleInput("multi-tenant", pluginhost.LifecycleHookBeforeUninstall.String())
	if ok, reason, err := checker.BeforeUninstall(ctx, input); err != nil || !ok || reason != "" {
		t.Fatalf("expected soft-deleted tenant not to block uninstall, ok=%v reason=%q err=%v", ok, reason, err)
	}
}

// TestNewRequiresTenantCounter verifies dependency construction fails fast.
func TestNewRequiresTenantCounter(t *testing.T) {
	if _, err := New(nil); err == nil {
		t.Fatal("expected missing tenant counter to fail")
	}
}

// configureLifecyclePreconditionTestDB points the package test at the local
// PostgreSQL database initialized by the repository test workflow.
func configureLifecyclePreconditionTestDB(t *testing.T, ctx context.Context) {
	t.Helper()

	originalConfig := gdb.GetAllConfig()
	if err := gdb.SetConfig(gdb.Config{
		gdb.DefaultGroupName: gdb.ConfigGroup{{Link: "pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable"}},
	}); err != nil {
		t.Fatalf("configure lifecycle precondition test database failed: %v", err)
	}
	db := g.DB()
	ensureLifecyclePreconditionTestTables(t, ctx)
	t.Cleanup(func() {
		if err := db.Close(ctx); err != nil {
			t.Errorf("close lifecycle precondition test database failed: %v", err)
		}
		if err := gdb.SetConfig(originalConfig); err != nil {
			t.Errorf("restore lifecycle precondition test database config failed: %v", err)
		}
	})
}

// ensureLifecyclePreconditionTestTables creates the minimal tenant table
// required by precondition tests when the local database has not installed the plugin.
func ensureLifecyclePreconditionTestTables(t *testing.T, ctx context.Context) {
	t.Helper()

	statement := `CREATE TABLE IF NOT EXISTS plugin_multi_tenant_tenant (
		"id" BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
		"code" VARCHAR(64) NOT NULL,
		"name" VARCHAR(128) NOT NULL,
		"status" VARCHAR(32) NOT NULL DEFAULT 'active',
		"remark" VARCHAR(512) NOT NULL DEFAULT '',
		"created_by" BIGINT NOT NULL DEFAULT 0,
		"updated_by" BIGINT NOT NULL DEFAULT 0,
		"created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		"updated_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		"deleted_at" TIMESTAMP,
		CONSTRAINT uk_plugin_multi_tenant_tenant_code UNIQUE ("code")
	)`
	if _, err := g.DB().Exec(ctx, statement); err != nil {
		t.Fatalf("ensure lifecycle precondition test table failed: %v", err)
	}
}
