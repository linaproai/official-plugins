// This file verifies that notice creator projections go through the host
// user-domain capability instead of plugin-local host user table access.

package notice

import (
	"context"
	"reflect"
	"testing"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
	"lina-core/pkg/plugin/capability/usercap"
)

// TestResolveCreatorNameMapUsesSingleUserBatch verifies current-page creators
// are de-duplicated before calling the host user capability.
func TestResolveCreatorNameMapUsesSingleUserBatch(t *testing.T) {
	ctx := context.Background()
	userSvc := &fakeNoticeUserService{
		batchResult: &capmodel.BatchResult[*usercap.UserProjection, usercap.UserID]{
			Items: map[usercap.UserID]*usercap.UserProjection{
				"7": {ID: "7", Username: "alice"},
			},
			MissingIDs: []usercap.UserID{"9"},
		},
	}
	svc := &serviceImpl{
		bizCtxSvc: fakeNoticeBizCtxService{current: bizctxcap.CurrentContext{
			UserID:   3,
			Username: "operator",
			TenantID: 2,
		}},
		tenantFilter: fakeNoticeTenantFilter{tenantCtx: tenantspi.TenantFilterContext{
			UserID:   3,
			TenantID: 2,
		}},
		userSvc: userSvc,
	}

	names, err := svc.resolveCreatorNameMap(ctx, []*NoticeEntity{
		{CreatedBy: 7},
		{CreatedBy: 7},
		{CreatedBy: 9},
		{CreatedBy: 0},
	})
	if err != nil {
		t.Fatalf("resolveCreatorNameMap returned error: %v", err)
	}
	if want := []usercap.UserID{"7", "9"}; !reflect.DeepEqual(userSvc.batchIDs, want) {
		t.Fatalf("expected batch ids %v, got %v", want, userSvc.batchIDs)
	}
	if names[7] != "alice" {
		t.Fatalf("expected creator 7 to resolve as alice, got %q", names[7])
	}
	if _, ok := names[9]; ok {
		t.Fatal("expected missing creator 9 to stay absent from resolved names")
	}
	if userSvc.batchCapCtx.PluginID != pluginID {
		t.Fatalf("expected plugin id %q, got %q", pluginID, userSvc.batchCapCtx.PluginID)
	}
	if userSvc.batchCapCtx.TenantID != "2" {
		t.Fatalf("expected tenant id 2, got %q", userSvc.batchCapCtx.TenantID)
	}
	if userSvc.batchCapCtx.Actor.UserID != 3 {
		t.Fatalf("expected actor user 3, got %d", userSvc.batchCapCtx.Actor.UserID)
	}
}

// TestSearchCreatorUserIDsUsesBoundedUserSearch verifies creator keyword
// filtering is delegated to usercap.SearchUsers with a bounded request.
func TestSearchCreatorUserIDsUsesBoundedUserSearch(t *testing.T) {
	ctx := context.Background()
	userSvc := &fakeNoticeUserService{
		searchResult: &capmodel.PageResult[*usercap.UserProjection]{
			Items: []*usercap.UserProjection{
				{ID: "5", Username: "alice"},
				{ID: "not-a-storage-id", Username: "external"},
				{ID: "8", Username: "alex"},
			},
			Total: 3,
		},
	}
	svc := &serviceImpl{
		bizCtxSvc: fakeNoticeBizCtxService{current: bizctxcap.CurrentContext{
			UserID:   4,
			Username: "reviewer",
			TenantID: 6,
		}},
		tenantFilter: fakeNoticeTenantFilter{tenantCtx: tenantspi.TenantFilterContext{
			UserID:   4,
			TenantID: 6,
		}},
		userSvc: userSvc,
	}

	ids, err := svc.searchCreatorUserIDs(ctx, "  ali ")
	if err != nil {
		t.Fatalf("searchCreatorUserIDs returned error: %v", err)
	}
	if want := []int64{5, 8}; !reflect.DeepEqual(ids, want) {
		t.Fatalf("expected storage ids %v, got %v", want, ids)
	}
	if userSvc.searchInput.Keyword != "ali" {
		t.Fatalf("expected trimmed keyword ali, got %q", userSvc.searchInput.Keyword)
	}
	if userSvc.searchInput.Page.PageSize != noticeCreatorSearchLimit ||
		userSvc.searchInput.Page.Limit != noticeCreatorSearchLimit {
		t.Fatalf("expected bounded page %d, got %+v", noticeCreatorSearchLimit, userSvc.searchInput.Page)
	}
	if userSvc.searchCapCtx.Resource != noticeCreatorCapabilityResource {
		t.Fatalf("expected resource %q, got %q", noticeCreatorCapabilityResource, userSvc.searchCapCtx.Resource)
	}
}

type fakeNoticeUserService struct {
	batchIDs    []usercap.UserID
	batchCapCtx capmodel.CapabilityContext
	batchResult *capmodel.BatchResult[*usercap.UserProjection, usercap.UserID]

	searchInput  usercap.SearchInput
	searchCapCtx capmodel.CapabilityContext
	searchResult *capmodel.PageResult[*usercap.UserProjection]
}

func (s *fakeNoticeUserService) BatchGetUsers(
	_ context.Context,
	capCtx capmodel.CapabilityContext,
	ids []usercap.UserID,
) (*capmodel.BatchResult[*usercap.UserProjection, usercap.UserID], error) {
	s.batchCapCtx = capCtx
	s.batchIDs = append([]usercap.UserID(nil), ids...)
	return s.batchResult, nil
}

func (s *fakeNoticeUserService) SearchUsers(
	_ context.Context,
	capCtx capmodel.CapabilityContext,
	input usercap.SearchInput,
) (*capmodel.PageResult[*usercap.UserProjection], error) {
	s.searchCapCtx = capCtx
	s.searchInput = input
	return s.searchResult, nil
}

func (s *fakeNoticeUserService) EnsureUsersVisible(_ context.Context, _ capmodel.CapabilityContext, _ []usercap.UserID) error {
	return nil
}

type fakeNoticeBizCtxService struct {
	current bizctxcap.CurrentContext
}

func (s fakeNoticeBizCtxService) Current(context.Context) bizctxcap.CurrentContext {
	return s.current
}

type fakeNoticeTenantFilter struct {
	tenantCtx tenantspi.TenantFilterContext
}

func (s fakeNoticeTenantFilter) Context(context.Context) tenantspi.TenantFilterContext {
	return s.tenantCtx
}

func (s fakeNoticeTenantFilter) Apply(_ context.Context, model *gdb.Model, _ string) *gdb.Model {
	return model
}
