// This file verifies that notice creator info goes through the host
// user-domain capability instead of plugin-local host user table access.

package notice

import (
	"context"
	"reflect"
	"testing"

	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/capability/usercap"
	"lina-core/pkg/statusflag"
)

// TestResolveCreatorNameMapUsesSingleUserBatch verifies current-page creators
// are de-duplicated before calling the host user capability.
func TestResolveCreatorNameMapUsesSingleUserBatch(t *testing.T) {
	ctx := context.Background()
	userSvc := &fakeNoticeUserService{
		batchResult: &capmodel.BatchResult[*usercap.UserInfo, usercap.UserID]{
			Items: map[usercap.UserID]*usercap.UserInfo{
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
		tenantSvc: fakeNoticeTenantService{filter: fakeNoticeTenantFilter{
			tenantCtx: tenantcap.TenantFilterContext{
				UserID:   3,
				TenantID: 2,
			},
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
}

// TestSearchCreatorUserIDsUsesBoundedUserSearch verifies creator keyword
// filtering is delegated to usercap.List with a bounded request.
func TestSearchCreatorUserIDsUsesBoundedUserSearch(t *testing.T) {
	ctx := context.Background()
	userSvc := &fakeNoticeUserService{
		searchResult: &capmodel.PageResult[*usercap.UserInfo]{
			Items: []*usercap.UserInfo{
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
		tenantSvc: fakeNoticeTenantService{filter: fakeNoticeTenantFilter{
			tenantCtx: tenantcap.TenantFilterContext{
				UserID:   4,
				TenantID: 6,
			},
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
}

type fakeNoticeUserService struct {
	batchIDs    []usercap.UserID
	batchResult *capmodel.BatchResult[*usercap.UserInfo, usercap.UserID]

	searchInput  usercap.ListInput
	searchResult *capmodel.PageResult[*usercap.UserInfo]
}

func (s *fakeNoticeUserService) Current(
	context.Context,
) (*usercap.UserInfo, error) {
	return nil, nil
}

func (s *fakeNoticeUserService) Get(
	context.Context,
	usercap.UserID,
) (*usercap.UserInfo, error) {
	return nil, nil
}

func (s *fakeNoticeUserService) BatchGet(
	ctx context.Context,
	ids []usercap.UserID,
) (*capmodel.BatchResult[*usercap.UserInfo, usercap.UserID], error) {
	s.batchIDs = append([]usercap.UserID(nil), ids...)
	return s.batchResult, nil
}

func (s *fakeNoticeUserService) BatchResolve(
	ctx context.Context,
	input usercap.BatchResolveInput,
) (*capmodel.BatchResult[*usercap.UserInfo, usercap.ResolveKey], error) {
	s.batchIDs = append([]usercap.UserID(nil), input.IDs...)
	result := &capmodel.BatchResult[*usercap.UserInfo, usercap.ResolveKey]{
		Items:      map[usercap.ResolveKey]*usercap.UserInfo{},
		MissingIDs: []usercap.ResolveKey{},
	}
	if s.batchResult == nil {
		for _, id := range input.IDs {
			result.MissingIDs = append(result.MissingIDs, usercap.ResolveKey(id))
		}
		return result, nil
	}
	for _, id := range input.IDs {
		key := usercap.ResolveKey(id)
		item, ok := s.batchResult.Items[id]
		if !ok {
			result.MissingIDs = append(result.MissingIDs, key)
			continue
		}
		result.Items[key] = item
	}
	return result, nil
}

func (s *fakeNoticeUserService) List(
	ctx context.Context,
	input usercap.ListInput,
) (*capmodel.PageResult[*usercap.UserInfo], error) {
	s.searchInput = input
	return s.searchResult, nil
}

func (s *fakeNoticeUserService) EnsureVisible(_ context.Context, _ []usercap.UserID) error {
	return nil
}

func (s *fakeNoticeUserService) Create(context.Context, usercap.CreateInput) (usercap.UserID, error) {
	return "", nil
}

func (s *fakeNoticeUserService) Update(context.Context, usercap.UpdateInput) error {
	return nil
}

func (s *fakeNoticeUserService) Delete(context.Context, usercap.UserID) error {
	return nil
}

func (s *fakeNoticeUserService) SetStatus(_ context.Context, _ usercap.UserID, _ statusflag.Enabled) error {
	return nil
}

func (s *fakeNoticeUserService) ResetPassword(context.Context, usercap.UserID, string) error {
	return nil
}

// Assignment returns user-role assignment operations unused by notice tests.
func (s *fakeNoticeUserService) Assignment() usercap.AssignmentService {
	return fakeNoticeUserAssignments{}
}

// fakeNoticeUserAssignments accepts unused role replacements.
type fakeNoticeUserAssignments struct{}

// ReplaceRoles is unused by notice tests.
func (fakeNoticeUserAssignments) ReplaceRoles(context.Context, usercap.UserID, []int) error {
	return nil
}

type fakeNoticeBizCtxService struct {
	current bizctxcap.CurrentContext
}

func (s fakeNoticeBizCtxService) Current(context.Context) bizctxcap.CurrentContext {
	return s.current
}

type fakeNoticeTenantFilter struct {
	tenantCtx tenantcap.TenantFilterContext
}

func (s fakeNoticeTenantFilter) Context(context.Context) tenantcap.TenantFilterContext {
	return s.tenantCtx
}

type fakeNoticeTenantService struct {
	filter tenantcap.FilterService
}

func (s fakeNoticeTenantService) Available(context.Context) bool {
	return true
}

func (s fakeNoticeTenantService) Status(context.Context) capmodel.CapabilityStatus {
	return capmodel.CapabilityStatus{Available: true}
}

func (s fakeNoticeTenantService) Context() tenantcap.ContextService {
	return nil
}

func (s fakeNoticeTenantService) Directory() tenantcap.DirectoryService {
	return nil
}

func (s fakeNoticeTenantService) Membership() tenantcap.MembershipService {
	return nil
}

func (s fakeNoticeTenantService) Plugins() tenantcap.PluginService {
	return nil
}

func (s fakeNoticeTenantService) Filter() tenantcap.FilterService {
	return s.filter
}
