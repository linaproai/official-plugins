// This file verifies impersonation token metadata and error contracts.

package impersonate

import (
	"context"
	"testing"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/authcap/authz"
	"lina-core/pkg/plugin/capability/authcap/token"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/usercap"
	"lina-core/pkg/statusflag"
	"lina-plugin-linapro-tenant-core/backend/internal/service/shared"
	tenantsvc "lina-plugin-linapro-tenant-core/backend/internal/service/tenant"
)

// TestImpersonationBusinessErrorMetadata verifies impersonation errors expose stable metadata.
func TestImpersonationBusinessErrorMetadata(t *testing.T) {
	testCases := []struct {
		name        string
		code        *bizerr.Code
		runtimeCode string
		messageKey  string
	}{
		{name: "permission denied", code: CodeImpersonationPermissionDenied, runtimeCode: "MULTI_TENANT_IMPERSONATION_PERMISSION_DENIED", messageKey: "error.multi.tenant.impersonation.permission.denied"},
		{name: "tenant unavailable", code: CodeImpersonationTenantUnavailable, runtimeCode: "MULTI_TENANT_IMPERSONATION_TENANT_UNAVAILABLE", messageKey: "error.multi.tenant.impersonation.tenant.unavailable"},
		{name: "token invalid", code: CodeImpersonationTokenInvalid, runtimeCode: "MULTI_TENANT_IMPERSONATION_TOKEN_INVALID", messageKey: "error.multi.tenant.impersonation.token.invalid"},
		{name: "token unavailable", code: CodeImpersonationTokenUnavailable, runtimeCode: "MULTI_TENANT_IMPERSONATION_TOKEN_UNAVAILABLE", messageKey: "error.multi.tenant.impersonation.token.unavailable"},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			err := bizerr.NewCode(testCase.code)
			messageErr, ok := bizerr.As(err)
			if !ok {
				t.Fatalf("expected structured business error, got %T", err)
			}
			if messageErr.RuntimeCode() != testCase.runtimeCode {
				t.Fatalf("expected runtime code %q, got %q", testCase.runtimeCode, messageErr.RuntimeCode())
			}
			if messageErr.MessageKey() != testCase.messageKey {
				t.Fatalf("expected message key %q, got %q", testCase.messageKey, messageErr.MessageKey())
			}
		})
	}
}

// TestStartDelegatesTokenIssuanceToHostAuth verifies plugin impersonation does
// not read host JWT config and instead delegates token/session ownership to the
// host auth service.
func TestStartDelegatesTokenIssuanceToHostAuth(t *testing.T) {
	ctx := context.Background()

	userID := int64(42)
	authSvc := &fakeImpersonationAuthService{}
	svc := &serviceImpl{
		authSvc:   authSvc,
		authzSvc:  fakeImpersonationAuthz{platformAdmin: true},
		bizCtxSvc: impersonateGuardBizCtx{current: bizctxcap.CurrentContext{UserID: int(userID), PlatformBypass: true}},
		tenantSvc: fakeImpersonationTenantService{tenant: &tenantsvc.Entity{Id: 42, Status: string(shared.TenantStatusActive)}},
		users: fakeImpersonationUsers{users: map[usercap.UserID]*usercap.UserInfo{
			usercap.UserID("42"): &usercap.UserInfo{ID: usercap.UserID("42"), Username: "platform-admin"},
		}},
	}

	out, err := svc.Start(ctx, StartInput{TenantID: 42, Reason: "unit test"})
	if err != nil {
		t.Fatalf("start impersonation: %v", err)
	}
	if out.Token != "host-impersonation-token" || out.TenantID != 42 || out.ActingUserID != userID || !out.IsImpersonated {
		t.Fatalf("unexpected impersonation output: %#v", out)
	}
	if authSvc.issuedActingUserID != int(userID) || authSvc.issuedTenantID != 42 {
		t.Fatalf("expected host auth issue call, got %#v", authSvc)
	}
}

// TestStopDelegatesTokenRevocationToHostAuth verifies impersonation stop keeps
// token parsing and revocation inside the host auth service.
func TestStopDelegatesTokenRevocationToHostAuth(t *testing.T) {
	authSvc := &fakeImpersonationAuthService{}
	svc := &serviceImpl{authSvc: authSvc}

	if err := svc.Stop(context.Background(), StopInput{TenantID: 42, Token: "Bearer host-impersonation-token"}); err != nil {
		t.Fatalf("stop impersonation: %v", err)
	}
	if authSvc.revokedBearer != "host-impersonation-token" || authSvc.revokedTenantID != 42 {
		t.Fatalf("expected host auth revoke call, got %#v", authSvc)
	}
}

// fakeImpersonationAuthService records host auth calls made by impersonation tests.
type fakeImpersonationAuthService struct {
	issuedActingUserID int
	issuedTenantID     int
	revokedBearer      string
	revokedTenantID    int
}

// SelectTenant is unused by impersonation tests.
func (s *fakeImpersonationAuthService) SelectTenant(
	context.Context,
	token.SelectTenantInput,
) (*token.TenantTokenOutput, error) {
	return &token.TenantTokenOutput{}, nil
}

// SwitchTenant is unused by impersonation tests.
func (s *fakeImpersonationAuthService) SwitchTenant(
	context.Context,
	token.SwitchTenantInput,
) (*token.TenantTokenOutput, error) {
	return &token.TenantTokenOutput{}, nil
}

// IssueImpersonationToken records the host auth impersonation request.
func (s *fakeImpersonationAuthService) IssueImpersonationToken(
	_ context.Context,
	in token.ImpersonationTokenIssueInput,
) (*token.ImpersonationTokenOutput, error) {
	s.issuedActingUserID = in.ActingUserID
	s.issuedTenantID = in.TenantID
	return &token.ImpersonationTokenOutput{
		AccessToken:  "host-impersonation-token",
		TokenID:      "host-token-id",
		TenantID:     in.TenantID,
		ActingUserID: in.ActingUserID,
	}, nil
}

// RevokeImpersonationToken records the host auth impersonation revoke request.
func (s *fakeImpersonationAuthService) RevokeImpersonationToken(
	_ context.Context,
	in token.ImpersonationTokenRevokeInput,
) error {
	s.revokedBearer = in.BearerToken
	s.revokedTenantID = in.TenantID
	return nil
}

// fakeImpersonationTenantService returns one configured tenant.
type fakeImpersonationTenantService struct {
	tenant *tenantsvc.Entity
}

// List is unused by impersonation tests.
func (s fakeImpersonationTenantService) List(context.Context, tenantsvc.ListInput) (*tenantsvc.ListOutput, error) {
	return &tenantsvc.ListOutput{}, nil
}

// Get returns the configured tenant.
func (s fakeImpersonationTenantService) Get(context.Context, int64) (*tenantsvc.Entity, error) {
	return s.tenant, nil
}

// Create is unused by impersonation tests.
func (s fakeImpersonationTenantService) Create(context.Context, tenantsvc.CreateInput) (int64, error) {
	return 0, nil
}

// Update is unused by impersonation tests.
func (s fakeImpersonationTenantService) Update(context.Context, tenantsvc.UpdateInput) error {
	return nil
}

// ChangeStatus is unused by impersonation tests.
func (s fakeImpersonationTenantService) ChangeStatus(context.Context, int64, shared.TenantStatus) error {
	return nil
}

// Delete is unused by impersonation tests.
func (s fakeImpersonationTenantService) Delete(context.Context, int64) error {
	return nil
}

// fakeImpersonationAuthz returns a configured platform-admin decision.
type fakeImpersonationAuthz struct {
	platformAdmin bool
}

// BatchGetPermissions is unused by impersonation tests.
func (s fakeImpersonationAuthz) BatchGetPermissions(context.Context, []authz.PermissionKey) (*capmodel.BatchResult[*authz.PermissionInfo, authz.PermissionKey], error) {
	return &capmodel.BatchResult[*authz.PermissionInfo, authz.PermissionKey]{Items: map[authz.PermissionKey]*authz.PermissionInfo{}}, nil
}

// BatchHasPermissions is unused by impersonation tests.
func (s fakeImpersonationAuthz) BatchHasPermissions(_ context.Context, keys []authz.PermissionKey) (map[authz.PermissionKey]bool, error) {
	result := make(map[authz.PermissionKey]bool, len(keys))
	for _, key := range keys {
		result[key] = false
	}
	return result, nil
}

// HasPermission is unused by impersonation tests.
func (s fakeImpersonationAuthz) HasPermission(context.Context, authz.PermissionKey) (bool, error) {
	return false, nil
}

// IsPlatformAdmin returns the configured decision.
func (s fakeImpersonationAuthz) IsPlatformAdmin(context.Context, authz.UserID) (bool, error) {
	return s.platformAdmin, nil
}

// ReplaceRolePermissions is unused by impersonation tests.
func (s fakeImpersonationAuthz) ReplaceRolePermissions(context.Context, authz.RoleID, []authz.PermissionKey) error {
	return nil
}

// fakeImpersonationUsers returns configured user projections.
type fakeImpersonationUsers struct {
	users map[usercap.UserID]*usercap.UserInfo
}

// Current returns no current user projection because these tests resolve explicit users.
func (s fakeImpersonationUsers) Current(context.Context) (*usercap.UserInfo, error) {
	return nil, nil
}

// Get is unused by impersonation tests.
func (s fakeImpersonationUsers) Get(context.Context, usercap.UserID) (*usercap.UserInfo, error) {
	return nil, nil
}

// BatchGet returns configured user projections and opaque missing IDs.
func (s fakeImpersonationUsers) BatchGet(_ context.Context, ids []usercap.UserID) (*capmodel.BatchResult[*usercap.UserInfo, usercap.UserID], error) {
	out := &capmodel.BatchResult[*usercap.UserInfo, usercap.UserID]{Items: map[usercap.UserID]*usercap.UserInfo{}}
	for _, id := range ids {
		if item := s.users[id]; item != nil {
			out.Items[id] = item
			continue
		}
		out.MissingIDs = append(out.MissingIDs, id)
	}
	return out, nil
}

// BatchResolve resolves configured users by ID for impersonation tests.
func (s fakeImpersonationUsers) BatchResolve(_ context.Context, input usercap.BatchResolveInput) (*capmodel.BatchResult[*usercap.UserInfo, usercap.ResolveKey], error) {
	out := &capmodel.BatchResult[*usercap.UserInfo, usercap.ResolveKey]{Items: map[usercap.ResolveKey]*usercap.UserInfo{}}
	for _, id := range input.IDs {
		key := usercap.ResolveKey(id)
		if item := s.users[id]; item != nil {
			out.Items[key] = item
			continue
		}
		out.MissingIDs = append(out.MissingIDs, key)
	}
	return out, nil
}

// List is unused by impersonation tests.
func (s fakeImpersonationUsers) List(context.Context, usercap.ListInput) (*capmodel.PageResult[*usercap.UserInfo], error) {
	return &capmodel.PageResult[*usercap.UserInfo]{Items: []*usercap.UserInfo{}}, nil
}

// EnsureVisible is unused by impersonation tests.
func (s fakeImpersonationUsers) EnsureVisible(context.Context, []usercap.UserID) error {
	return nil
}

// Create is unused by impersonation tests.
func (s fakeImpersonationUsers) Create(context.Context, usercap.CreateInput) (usercap.UserID, error) {
	return "", nil
}

func (s fakeImpersonationUsers) CreateFromExternal(context.Context, usercap.CreateFromExternalInput) (usercap.UserID, error) {
	return "", nil
}

// Update is unused by impersonation tests.
func (s fakeImpersonationUsers) Update(context.Context, usercap.UpdateInput) error {
	return nil
}

// Delete is unused by impersonation tests.
func (s fakeImpersonationUsers) Delete(context.Context, usercap.UserID) error {
	return nil
}

// SetStatus is unused by impersonation tests.
func (s fakeImpersonationUsers) SetStatus(context.Context, usercap.UserID, statusflag.Enabled) error {
	return nil
}

// ResetPassword is unused by impersonation tests.
func (s fakeImpersonationUsers) ResetPassword(context.Context, usercap.UserID, string) error {
	return nil
}

// Assignment returns user-role assignment operations unused by impersonation tests.
func (s fakeImpersonationUsers) Assignment() usercap.AssignmentService {
	return fakeImpersonationUserAssignments{}
}

// fakeImpersonationUserAssignments accepts unused role replacements.
type fakeImpersonationUserAssignments struct{}

// ReplaceRoles is unused by impersonation tests.
func (fakeImpersonationUserAssignments) ReplaceRoles(context.Context, usercap.UserID, []int) error {
	return nil
}
