// This file verifies platform-context enforcement for platform tenant
// management operations.

package tenant

import (
	"context"
	"testing"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-plugin-linapro-tenant-core/backend/internal/service/resolver"
	"lina-plugin-linapro-tenant-core/backend/internal/service/shared"
)

// TestEnsurePlatformTenantGovernanceRejectsTenantContext verifies tenant CRUD
// entry points fail closed when the caller is not platform bypass.
func TestEnsurePlatformTenantGovernanceRejectsTenantContext(t *testing.T) {
	svc := &serviceImpl{bizCtxSvc: tenantGuardBizCtx{current: bizctxcap.CurrentContext{
		TenantID:       1001,
		PlatformBypass: false,
	}}}

	err := svc.ensurePlatformTenantGovernance(context.Background())
	if !bizerr.Is(err, resolver.CodePlatformPermissionRequired) {
		t.Fatalf("expected platform permission error, got %v", err)
	}
}

// TestEnsurePlatformTenantGovernanceAllowsPlatformBypass verifies platform
// all-data context can manage platform tenant rows.
func TestEnsurePlatformTenantGovernanceAllowsPlatformBypass(t *testing.T) {
	svc := &serviceImpl{bizCtxSvc: tenantGuardBizCtx{current: bizctxcap.CurrentContext{
		TenantID:       0,
		PlatformBypass: true,
	}}}

	err := svc.ensurePlatformTenantGovernance(context.Background())
	if err != nil {
		t.Fatalf("expected platform bypass to allow tenant governance, got %v", err)
	}
}

// TestPlatformTenantMethodsRejectTenantContextAfterPermission verifies the
// platform tenant CRUD service still rejects tenant context after route-level
// system:tenant:* permission checks have already passed.
func TestPlatformTenantMethodsRejectTenantContextAfterPermission(t *testing.T) {
	svc := &serviceImpl{bizCtxSvc: tenantGuardBizCtx{current: bizctxcap.CurrentContext{
		TenantID:       1001,
		PlatformBypass: false,
	}}}
	name := "Blocked Tenant"
	remark := "blocked"
	cases := []struct {
		name string
		run  func() error
	}{
		{name: "list", run: func() error {
			_, err := svc.List(context.Background(), ListInput{PageNum: 1, PageSize: 10})
			return err
		}},
		{name: "get", run: func() error {
			_, err := svc.Get(context.Background(), 1)
			return err
		}},
		{name: "create", run: func() error {
			_, err := svc.Create(context.Background(), CreateInput{Code: "blocked", Name: name})
			return err
		}},
		{name: "update", run: func() error {
			return svc.Update(context.Background(), UpdateInput{Id: 1, Name: &name, Remark: &remark})
		}},
		{name: "change status", run: func() error {
			return svc.ChangeStatus(context.Background(), 1, shared.TenantStatusActive)
		}},
		{name: "delete", run: func() error {
			return svc.Delete(context.Background(), 1)
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.run(); !bizerr.Is(err, resolver.CodePlatformPermissionRequired) {
				t.Fatalf("expected platform permission error, got %v", err)
			}
		})
	}
}

// tenantGuardBizCtx returns a fixed plugin-visible business context.
type tenantGuardBizCtx struct {
	current bizctxcap.CurrentContext
}

// Current returns the fixed context configured by the test.
func (s tenantGuardBizCtx) Current(context.Context) bizctxcap.CurrentContext {
	return s.current
}
