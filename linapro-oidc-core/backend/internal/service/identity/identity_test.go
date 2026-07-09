// This file verifies the linapro-oidc-core external-identity provider:
// (provider, subject) resolution, idempotent provisioning, same-email conflict
// rejection, email-less deterministic anchor derivation, concurrent
// unique-index conflict absorption, and self-isolated bind/unbind/list. These
// are database-backed tests following the plugin test convention: each test
// configures the local PostgreSQL database, ensures the plugin table exists,
// creates unique fixtures, and cleans up its own rows.

package identity

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"

	"lina-core/pkg/bizerr"
	_ "lina-core/pkg/dbdriver"
	"lina-core/pkg/plugin/capability/authcap/externallogin/externalidentityspi"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/usercap"
	"lina-core/pkg/statusflag"
	"lina-plugin-linapro-oidc-core/backend/internal/dao"
	"lina-plugin-linapro-oidc-core/backend/internal/model/do"
)

// fakeUserCapability is a configurable usercap.Service double. Only
// ProvisionExternal carries behavior; the remaining methods are
// interface-conformance no-ops.
type fakeUserCapability struct {
	// provisionID is returned by ProvisionExternal.
	provisionID int
	// provisionErr fails ProvisionExternal when set.
	provisionErr error
	// provisionCalls records every provisioning request.
	provisionCalls []usercap.ProvisionExternalInput
	// beforeProvisionReturn runs after recording a provisioning call and before
	// returning, letting tests simulate a concurrent provision winning the race.
	beforeProvisionReturn func()
}

func (f *fakeUserCapability) Current(context.Context) (*usercap.UserInfo, error) { return nil, nil }
func (f *fakeUserCapability) Get(context.Context, usercap.UserID) (*usercap.UserInfo, error) {
	return nil, nil
}
func (f *fakeUserCapability) BatchGet(context.Context, []usercap.UserID) (*capmodel.BatchResult[*usercap.UserInfo, usercap.UserID], error) {
	return &capmodel.BatchResult[*usercap.UserInfo, usercap.UserID]{}, nil
}

// BatchResolve is an interface-conformance no-op: the login path carries no
// actor context, so the provider never performs data-scoped lookups.
func (f *fakeUserCapability) BatchResolve(context.Context, usercap.BatchResolveInput) (*capmodel.BatchResult[*usercap.UserInfo, usercap.ResolveKey], error) {
	return &capmodel.BatchResult[*usercap.UserInfo, usercap.ResolveKey]{}, nil
}
func (f *fakeUserCapability) List(context.Context, usercap.ListInput) (*capmodel.PageResult[*usercap.UserInfo], error) {
	return &capmodel.PageResult[*usercap.UserInfo]{}, nil
}
func (f *fakeUserCapability) EnsureVisible(context.Context, []usercap.UserID) error { return nil }
func (f *fakeUserCapability) Create(context.Context, usercap.CreateInput) (usercap.UserID, error) {
	return "", nil
}

// ProvisionExternal records the call and returns the configured outcome.
func (f *fakeUserCapability) ProvisionExternal(_ context.Context, input usercap.ProvisionExternalInput) (usercap.UserID, error) {
	f.provisionCalls = append(f.provisionCalls, input)
	if f.provisionErr != nil {
		return "", f.provisionErr
	}
	if f.beforeProvisionReturn != nil {
		f.beforeProvisionReturn()
	}
	return usercap.UserID(strconv.Itoa(f.provisionID)), nil
}
func (f *fakeUserCapability) Update(context.Context, usercap.UpdateInput) error { return nil }
func (f *fakeUserCapability) Delete(context.Context, usercap.UserID) error      { return nil }
func (f *fakeUserCapability) SetStatus(context.Context, usercap.UserID, statusflag.Enabled) error {
	return nil
}
func (f *fakeUserCapability) ResetPassword(context.Context, usercap.UserID, string) error {
	return nil
}
func (f *fakeUserCapability) Assignment() usercap.AssignmentService { return nil }

// configureIdentityTestDB points the package test at the local PostgreSQL
// database and ensures the plugin-owned table exists.
func configureIdentityTestDB(t *testing.T, ctx context.Context) {
	t.Helper()
	originalConfig := gdb.GetAllConfig()
	if err := gdb.SetConfig(gdb.Config{
		gdb.DefaultGroupName: gdb.ConfigGroup{{Link: "pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable"}},
	}); err != nil {
		t.Fatalf("configure identity test database failed: %v", err)
	}
	db := g.DB()
	if _, err := db.Exec(ctx, `
CREATE TABLE IF NOT EXISTS user_external_identity (
    "id"             BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "user_id"        INT NOT NULL,
    "provider"       VARCHAR(64) NOT NULL,
    "subject"        VARCHAR(191) NOT NULL,
    "plugin_id"      VARCHAR(128) NOT NULL DEFAULT '',
    "email_snapshot" VARCHAR(191) NOT NULL DEFAULT '',
    "created_at"     TIMESTAMPTZ,
    "updated_at"     TIMESTAMPTZ,
    "deleted_at"     TIMESTAMPTZ
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_user_external_identity_provider_subject
    ON user_external_identity ("provider", "subject") WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_user_external_identity_user
    ON user_external_identity ("user_id");
`); err != nil {
		t.Fatalf("ensure identity test table failed: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(ctx); err != nil {
			t.Errorf("close identity test database failed: %v", err)
		}
		if err := gdb.SetConfig(originalConfig); err != nil {
			t.Errorf("restore identity test database config failed: %v", err)
		}
	})
}

// cleanupIdentityRows removes every linkage row for one provider during cleanup.
func cleanupIdentityRows(t *testing.T, ctx context.Context, provider string) {
	t.Helper()
	t.Cleanup(func() {
		if _, err := dao.UserExternalIdentity.Ctx(ctx).
			Unscoped().
			Where(do.UserExternalIdentity{Provider: provider}).
			Delete(); err != nil {
			t.Errorf("cleanup identity rows failed: %v", err)
		}
	})
}

// uniqueTestProvider returns a per-test provider ID so tests stay independent
// and order-free even against a shared local database.
func uniqueTestProvider(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

// newTestService builds the identity service on the supplied user capability.
func newTestService(t *testing.T, users usercap.Service) *Service {
	t.Helper()
	svc, err := New(users)
	if err != nil {
		t.Fatalf("new identity service: %v", err)
	}
	return svc
}

// TestResolveMissingLinkageReturnsNotFound verifies resolution reports
// found=false without error for an unlinked identity, and rejects empty keys.
func TestResolveMissingLinkageReturnsNotFound(t *testing.T) {
	ctx := context.Background()
	configureIdentityTestDB(t, ctx)
	svc := newTestService(t, &fakeUserCapability{})

	if _, _, err := svc.Resolve(ctx, externalidentityspi.ResolveInput{Provider: " ", Subject: "s"}); !bizerr.Is(err, CodeIdentityInvalid) {
		t.Fatalf("blank provider: expected identity-invalid, got %v", err)
	}
	_, found, err := svc.Resolve(ctx, externalidentityspi.ResolveInput{
		Provider: uniqueTestProvider("resolve-missing"),
		Subject:  "subject-1",
	})
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if found {
		t.Fatal("expected found=false for unlinked identity")
	}
}

// TestProvisionReusesExistingLinkageIdempotently verifies a linked identity is
// reused without any account provisioning call.
func TestProvisionReusesExistingLinkageIdempotently(t *testing.T) {
	ctx := context.Background()
	configureIdentityTestDB(t, ctx)
	users := &fakeUserCapability{provisionID: 999}
	svc := newTestService(t, users)

	provider := uniqueTestProvider("prov-idem")
	cleanupIdentityRows(t, ctx, provider)
	if _, err := dao.UserExternalIdentity.Ctx(ctx).Data(do.UserExternalIdentity{
		UserId: 42, Provider: provider, Subject: "subject-1", PluginId: "test",
	}).Insert(); err != nil {
		t.Fatalf("seed linkage: %v", err)
	}

	userID, err := svc.Provision(ctx, externalidentityspi.ProvisionInput{
		Provider: provider, Subject: "subject-1", Email: "idem@example.com", AllowAutoProvision: true,
	})
	if err != nil {
		t.Fatalf("provision: %v", err)
	}
	if userID != 42 {
		t.Fatalf("expected linked user 42, got %d", userID)
	}
	if len(users.provisionCalls) != 0 {
		t.Fatalf("expected no account provisioning, got %d calls", len(users.provisionCalls))
	}
}

// TestProvisionRejectsSameEmailConflict verifies the host minting primitive's
// email-conflict sentinel is mapped into the provider's caller-visible
// conflict code and never writes a linkage. The unfiltered email-existence
// check lives in the host primitive because the login path has no actor
// context for data-scoped lookups.
func TestProvisionRejectsSameEmailConflict(t *testing.T) {
	ctx := context.Background()
	configureIdentityTestDB(t, ctx)
	email := "conflict@example.com"
	users := &fakeUserCapability{provisionErr: usercap.ErrProvisionEmailConflict}
	svc := newTestService(t, users)

	provider := uniqueTestProvider("prov-conflict")
	cleanupIdentityRows(t, ctx, provider)
	_, err := svc.Provision(ctx, externalidentityspi.ProvisionInput{
		Provider: provider, Subject: "subject-1", Email: email, AllowAutoProvision: true,
	})
	if !bizerr.Is(err, CodeProvisionEmailConflict) {
		t.Fatalf("expected email-conflict, got %v", err)
	}
	// No linkage row may exist for the conflicted identity.
	if _, found, resolveErr := svc.Resolve(ctx, externalidentityspi.ResolveInput{
		Provider: provider, Subject: "subject-1",
	}); resolveErr != nil || found {
		t.Fatalf("expected no linkage after conflict, got found=%v err=%v", found, resolveErr)
	}
}

// TestProvisionDisallowedFailsClosed verifies AllowAutoProvision=false rejects
// without provisioning.
func TestProvisionDisallowedFailsClosed(t *testing.T) {
	ctx := context.Background()
	configureIdentityTestDB(t, ctx)
	users := &fakeUserCapability{provisionID: 7}
	svc := newTestService(t, users)

	_, err := svc.Provision(ctx, externalidentityspi.ProvisionInput{
		Provider: uniqueTestProvider("prov-off"), Subject: "subject-1", Email: "off@example.com",
	})
	if !bizerr.Is(err, CodeProvisionNotAllowed) {
		t.Fatalf("expected provision-not-allowed, got %v", err)
	}
	if len(users.provisionCalls) != 0 {
		t.Fatalf("expected no account provisioning, got %d calls", len(users.provisionCalls))
	}
}

// TestProvisionEmaillessDerivesDeterministicAnchor verifies email-less
// provisioning derives a deterministic, collision-resistant username anchor and
// never triggers the email-conflict policy.
func TestProvisionEmaillessDerivesDeterministicAnchor(t *testing.T) {
	ctx := context.Background()
	configureIdentityTestDB(t, ctx)
	users := &fakeUserCapability{provisionID: 77}
	svc := newTestService(t, users)

	provider := uniqueTestProvider("prov-anchor")
	cleanupIdentityRows(t, ctx, provider)
	userID, err := svc.Provision(ctx, externalidentityspi.ProvisionInput{
		Provider: provider, Subject: "wechat-subject", AllowAutoProvision: true, PluginID: "linapro-oidc-test",
	})
	if err != nil {
		t.Fatalf("email-less provision: %v", err)
	}
	if userID != 77 {
		t.Fatalf("expected provisioned user 77, got %d", userID)
	}
	if len(users.provisionCalls) != 1 {
		t.Fatalf("expected one provisioning call, got %d", len(users.provisionCalls))
	}
	call := users.provisionCalls[0]
	if call.Email != "" {
		t.Fatalf("expected empty email, got %q", call.Email)
	}
	want := deriveUsernameAnchor(provider, "wechat-subject")
	if call.UsernameAnchor != want {
		t.Fatalf("anchor = %q, want deterministic %q", call.UsernameAnchor, want)
	}
	if again := deriveUsernameAnchor(provider, "wechat-subject"); again != want {
		t.Fatalf("anchor derivation not deterministic: %q vs %q", again, want)
	}
	if other := deriveUsernameAnchor(provider, "other-subject"); other == want {
		t.Fatal("distinct subjects must derive distinct anchors")
	}
	// The linkage row must reference the provisioned user with the stamped plugin.
	resolvedID, found, err := svc.Resolve(ctx, externalidentityspi.ResolveInput{Provider: provider, Subject: "wechat-subject"})
	if err != nil || !found || resolvedID != 77 {
		t.Fatalf("expected linkage to user 77, got id=%d found=%v err=%v", resolvedID, found, err)
	}
}

// TestProvisionAbsorbsConcurrentUniqueConflict verifies that when a concurrent
// provision wins the (provider, subject) unique index race, the losing call
// reuses the winning linkage instead of surfacing a 500.
func TestProvisionAbsorbsConcurrentUniqueConflict(t *testing.T) {
	ctx := context.Background()
	configureIdentityTestDB(t, ctx)
	provider := uniqueTestProvider("prov-race")
	cleanupIdentityRows(t, ctx, provider)

	users := &fakeUserCapability{provisionID: 88}
	// Simulate a concurrent provision winning the race after this call's
	// fast-path linkage check but before its linkage insert.
	users.beforeProvisionReturn = func() {
		if _, err := dao.UserExternalIdentity.Ctx(ctx).Data(do.UserExternalIdentity{
			UserId: 55, Provider: provider, Subject: "raced-subject", PluginId: "winner",
		}).Insert(); err != nil {
			t.Fatalf("seed winning linkage: %v", err)
		}
	}
	svc := newTestService(t, users)

	userID, err := svc.Provision(ctx, externalidentityspi.ProvisionInput{
		Provider: provider, Subject: "raced-subject", Email: "race@example.com", AllowAutoProvision: true,
	})
	if err != nil {
		t.Fatalf("expected unique conflict to be absorbed, got %v", err)
	}
	if userID != 55 {
		t.Fatalf("expected winning linkage user 55, got %d", userID)
	}
}

// TestBindUnbindListSelfIsolation verifies bind conflict and idempotency,
// unbind self-isolation, relink after unbind, list self-isolation, and
// cross-provider subject independence.
func TestBindUnbindListSelfIsolation(t *testing.T) {
	ctx := context.Background()
	configureIdentityTestDB(t, ctx)
	svc := newTestService(t, &fakeUserCapability{})

	provider := uniqueTestProvider("bind")
	otherProvider := uniqueTestProvider("bind-other")
	cleanupIdentityRows(t, ctx, provider)
	cleanupIdentityRows(t, ctx, otherProvider)
	const userA, userB = 101, 202

	// Bind for user A succeeds; re-binding the same identity is idempotent.
	bind := externalidentityspi.BindInput{UserID: userA, Provider: provider, Subject: "shared-subject", Email: "a@example.com", PluginID: "test"}
	if err := svc.Bind(ctx, bind); err != nil {
		t.Fatalf("bind: %v", err)
	}
	if err := svc.Bind(ctx, bind); err != nil {
		t.Fatalf("idempotent re-bind: %v", err)
	}
	// The same identity bound by another user is rejected as a conflict.
	conflict := bind
	conflict.UserID = userB
	if err := svc.Bind(ctx, conflict); !bizerr.Is(err, CodeBindConflict) {
		t.Fatalf("expected bind-conflict, got %v", err)
	}
	// The same subject under a DIFFERENT provider is a distinct identity.
	if err := svc.Bind(ctx, externalidentityspi.BindInput{
		UserID: userB, Provider: otherProvider, Subject: "shared-subject", PluginID: "test",
	}); err != nil {
		t.Fatalf("cross-provider bind: %v", err)
	}
	// List is self-isolated: each user only sees their own linkage.
	identitiesA, err := svc.List(ctx, userA)
	if err != nil {
		t.Fatalf("list A: %v", err)
	}
	if len(identitiesA) != 1 || identitiesA[0].Provider != provider {
		t.Fatalf("user A list = %#v, want one %s linkage", identitiesA, provider)
	}
	identitiesB, err := svc.List(ctx, userB)
	if err != nil {
		t.Fatalf("list B: %v", err)
	}
	if len(identitiesB) != 1 || identitiesB[0].Provider != otherProvider {
		t.Fatalf("user B list = %#v, want one %s linkage", identitiesB, otherProvider)
	}
	// Unbind by a non-owner reports not-found without leaking existence.
	if err = svc.Unbind(ctx, externalidentityspi.UnbindInput{
		UserID: userB, Provider: provider, Subject: "shared-subject",
	}); !bizerr.Is(err, CodeIdentityNotFound) {
		t.Fatalf("cross-user unbind: expected not-found, got %v", err)
	}
	// Owner unbind succeeds and frees the unique key for a relink.
	if err = svc.Unbind(ctx, externalidentityspi.UnbindInput{
		UserID: userA, Provider: provider, Subject: "shared-subject",
	}); err != nil {
		t.Fatalf("owner unbind: %v", err)
	}
	if err = svc.Bind(ctx, externalidentityspi.BindInput{
		UserID: userB, Provider: provider, Subject: "shared-subject", PluginID: "test",
	}); err != nil {
		t.Fatalf("relink after unbind: %v", err)
	}
}
