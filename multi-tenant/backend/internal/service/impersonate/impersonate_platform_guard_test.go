// This file verifies tenant impersonation can only start from a strict
// platform-bypass context.

package impersonate

import (
	"context"
	"testing"

	"lina-core/pkg/bizerr"
	plugincontract "lina-core/pkg/pluginservice/contract"
)

// TestStartRejectsNonPlatformBypassBeforeRoleLookup verifies the impersonation
// entry point fails closed for tenant or delegated contexts even if later role
// checks might find platform-like grants.
func TestStartRejectsNonPlatformBypassBeforeRoleLookup(t *testing.T) {
	svc := &serviceImpl{bizCtxSvc: impersonateGuardBizCtx{current: plugincontract.CurrentContext{
		UserID:         1,
		TenantID:       1001,
		PlatformBypass: false,
	}}}

	_, err := svc.Start(context.Background(), StartInput{TenantID: 1001})
	if !bizerr.Is(err, CodeImpersonationPermissionDenied) {
		t.Fatalf("expected impersonation permission denied, got %v", err)
	}
}

// impersonateGuardBizCtx returns a fixed plugin-visible business context.
type impersonateGuardBizCtx struct {
	current plugincontract.CurrentContext
}

// Current returns the fixed context configured by the test.
func (s impersonateGuardBizCtx) Current(context.Context) plugincontract.CurrentContext {
	return s.current
}
