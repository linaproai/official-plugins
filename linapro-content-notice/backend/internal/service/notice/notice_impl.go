// notice_impl.go implements notice CRUD, publish state transitions, and
// tenant-scoped query/export behavior for the linapro-content-notice plugin. It keeps
// host capability and tenant dependencies injected so plugin-owned notice rows
// remain isolated from host-internal services.

package notice

import (
	"context"
	"strconv"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	usermsgv1 "lina-core/api/usermsg/v1"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
	"lina-core/pkg/plugin/capability/usercap"
	"lina-plugin-linapro-content-notice/backend/internal/dao"
	"lina-plugin-linapro-content-notice/backend/internal/model/do"
)

const (
	pluginID                        = "linapro-content-notice"
	noticeCreatorCapabilityResource = "notice.creator"
	noticeCreatorSearchLimit        = 200
)

// List queries notice list with pagination and filters.
func (s *serviceImpl) List(ctx context.Context, in ListInput) (*ListOutput, error) {
	noticeColumns := dao.Notice.Columns()

	m := tenantspi.ApplyPluginTableFilter(ctx, s.pluginTableFilter(), dao.Notice.Ctx(ctx), "")

	// Apply filters
	if in.Title != "" {
		m = m.WhereLike(noticeColumns.Title, "%"+in.Title+"%")
	}
	if in.Type > 0 {
		m = m.Where(noticeColumns.Type, in.Type)
	}
	if createdBy := strings.TrimSpace(in.CreatedBy); createdBy != "" {
		creatorIDs, err := s.searchCreatorUserIDs(ctx, createdBy)
		if err != nil {
			return nil, err
		}
		if len(creatorIDs) == 0 {
			return &ListOutput{List: []*ListItem{}, Total: 0}, nil
		}
		m = m.WhereIn(noticeColumns.CreatedBy, creatorIDs)
	}

	// Get total count
	total, err := m.Count()
	if err != nil {
		return nil, err
	}

	// Query with pagination
	list := make([]*NoticeEntity, 0)
	err = m.Page(in.PageNum, in.PageSize).
		OrderDesc(noticeColumns.Id).
		Scan(&list)
	if err != nil {
		return nil, err
	}

	userNameMap, err := s.resolveCreatorNameMap(ctx, list)
	if err != nil {
		return nil, err
	}

	// Build result
	items := make([]*ListItem, 0, len(list))
	for _, n := range list {
		items = append(items, &ListItem{
			NoticeEntity:  n,
			CreatedByName: userNameMap[n.CreatedBy],
		})
	}

	return &ListOutput{
		List:  items,
		Total: total,
	}, nil
}

// GetById retrieves notice by ID.
func (s *serviceImpl) GetById(ctx context.Context, id int64) (*ListItem, error) {
	noticeColumns := dao.Notice.Columns()

	var notice *NoticeEntity
	err := tenantspi.ApplyPluginTableFilter(ctx, s.pluginTableFilter(), dao.Notice.Ctx(ctx), "").
		Where(noticeColumns.Id, id).
		Scan(&notice)
	if err != nil {
		return nil, err
	}
	if notice == nil {
		return nil, bizerr.NewCode(CodeNoticeNotFound)
	}

	item := &ListItem{NoticeEntity: notice}

	// Resolve creator username
	if notice.CreatedBy > 0 {
		names, err := s.resolveCreatorNameMap(ctx, []*NoticeEntity{notice})
		if err != nil {
			return nil, err
		}
		if name := names[notice.CreatedBy]; name != "" {
			item.CreatedByName = name
		}
	}

	return item, nil
}

// Create creates a new notice.
func (s *serviceImpl) Create(ctx context.Context, in CreateInput) (int64, error) {
	var (
		bizCtx    = s.bizCtxSvc.Current(ctx)
		createdBy = int64(bizCtx.UserID)
		tenantID  = s.tenantFilterContext(ctx).TenantID
	)

	// Insert notice (GoFrame auto-fills created_at and updated_at).
	id, err := dao.Notice.Ctx(ctx).Data(do.Notice{
		TenantId:  tenantID,
		Title:     in.Title,
		Type:      in.Type,
		Content:   in.Content,
		FileIds:   in.FileIds,
		Status:    in.Status,
		Remark:    in.Remark,
		CreatedBy: createdBy,
		UpdatedBy: createdBy,
	}).InsertAndGetId()
	if err != nil {
		return 0, err
	}

	// If published, dispatch inbox notifications through the unified notify domain.
	if in.Status == noticeStatusPublished {
		if dispatchErr := s.dispatchPublishedNotice(ctx, id, in.Title, in.Content, in.Type, createdBy); dispatchErr != nil {
			logger.Errorf(ctx, "dispatch published notice failed for notice %d: %v", id, dispatchErr)
		}
	}

	return id, nil
}

// Update updates notice information.
func (s *serviceImpl) Update(ctx context.Context, in UpdateInput) error {
	noticeColumns := dao.Notice.Columns()

	// Check notice exists and get old status.
	var oldNotice *NoticeEntity
	err := tenantspi.ApplyPluginTableFilter(ctx, s.pluginTableFilter(), dao.Notice.Ctx(ctx), "").
		Where(noticeColumns.Id, in.Id).
		Scan(&oldNotice)
	if err != nil {
		return err
	}
	if oldNotice == nil {
		return bizerr.NewCode(CodeNoticeNotFound)
	}

	var (
		bizCtx    = s.bizCtxSvc.Current(ctx)
		updatedBy = int64(bizCtx.UserID)
		tenantID  = s.tenantFilterContext(ctx).TenantID
	)

	data := do.Notice{UpdatedBy: updatedBy}
	if in.Title != nil {
		data.Title = *in.Title
	}
	if in.Type != nil {
		data.Type = *in.Type
	}
	if in.Content != nil {
		data.Content = *in.Content
	}
	if in.FileIds != nil {
		data.FileIds = *in.FileIds
	}
	if in.Status != nil {
		data.Status = *in.Status
	}
	if in.Remark != nil {
		data.Remark = *in.Remark
	}

	_, err = dao.Notice.Ctx(ctx).
		OmitNilData().
		Where(tenantspi.TenantFilterColumn, tenantID).
		Where(noticeColumns.Id, in.Id).
		Data(data).
		Update()
	if err != nil {
		return err
	}

	// If status changed from draft(0) to published(1), dispatch inbox notifications.
	if in.Status != nil && *in.Status == noticeStatusPublished && oldNotice.Status == noticeStatusDraft {
		title := oldNotice.Title
		if in.Title != nil {
			title = *in.Title
		}
		content := oldNotice.Content
		if in.Content != nil {
			content = *in.Content
		}
		noticeType := oldNotice.Type
		if in.Type != nil {
			noticeType = *in.Type
		}
		if dispatchErr := s.dispatchPublishedNotice(ctx, in.Id, title, content, noticeType, oldNotice.CreatedBy); dispatchErr != nil {
			logger.Errorf(ctx, "dispatch published notice failed for notice %d: %v", in.Id, dispatchErr)
		}
	}

	return nil
}

// Delete soft-deletes notices by IDs and cascades to notify deliveries.
func (s *serviceImpl) Delete(ctx context.Context, ids []int64) error {
	idList := normalizePositiveInt64IDs(ids)
	if len(idList) == 0 {
		return bizerr.NewCode(CodeNoticeDeleteRequired)
	}

	// Soft delete using GoFrame's auto soft-delete feature.
	noticeColumns := dao.Notice.Columns()
	_, err := tenantspi.ApplyPluginTableFilter(ctx, s.pluginTableFilter(), dao.Notice.Ctx(ctx), "").
		WhereIn(noticeColumns.Id, idList).
		Delete()
	if err != nil {
		return err
	}

	sourceIDs := make([]string, 0, len(idList))
	for _, id := range idList {
		sourceIDs = append(sourceIDs, strconv.FormatInt(id, 10))
	}
	if cascadeErr := s.notifySvc.DeleteBySource(ctx, usermsgv1.SourceTypeNotice, sourceIDs); cascadeErr != nil {
		logger.Errorf(ctx, "cascade delete notify deliveries failed for notice ids %v: %v", idList, cascadeErr)
	}
	return nil
}

// normalizePositiveInt64IDs drops non-positive identifiers from a batch ID list.
func normalizePositiveInt64IDs(ids []int64) []int64 {
	if len(ids) == 0 {
		return nil
	}
	result := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id > 0 {
			result = append(result, id)
		}
	}
	return result
}

// pluginTableFilter returns the tenant table-filter slice from the injected tenant service.
func (s *serviceImpl) pluginTableFilter() tenantcap.FilterService {
	if s == nil || s.tenantSvc == nil {
		return nil
	}
	return s.tenantSvc.Filter()
}

// tenantFilterContext returns current tenant metadata for write ownership fields.
func (s *serviceImpl) tenantFilterContext(ctx context.Context) tenantcap.TenantFilterContext {
	if filter := s.pluginTableFilter(); filter != nil {
		return filter.Context(ctx)
	}
	return tenantcap.TenantFilterContext{}
}

// searchCreatorUserIDs resolves a creator keyword through the host user domain
// capability before filtering plugin-owned notice rows by creator ID.
func (s *serviceImpl) searchCreatorUserIDs(ctx context.Context, keyword string) ([]int64, error) {
	if s.userSvc == nil {
		return nil, gerror.New("linapro-content-notice requires host user capability")
	}
	result, err := s.userSvc.List(ctx, usercap.ListInput{
		Keyword: strings.TrimSpace(keyword),
		Page: capmodel.PageRequest{
			PageNum:  1,
			PageSize: noticeCreatorSearchLimit,
			Limit:    noticeCreatorSearchLimit,
		},
	})
	if err != nil || result == nil {
		return nil, err
	}
	return userInfoStorageIDs(result.Items), nil
}

// resolveCreatorNameMap resolves current-page creator display names through one
// user-domain batch call and leaves invisible or missing users blank.
func (s *serviceImpl) resolveCreatorNameMap(ctx context.Context, notices []*NoticeEntity) (map[int64]string, error) {
	names := make(map[int64]string)
	userIDs := creatorDomainIDs(notices)
	if len(userIDs) == 0 {
		return names, nil
	}
	if s.userSvc == nil {
		return nil, gerror.New("linapro-content-notice requires host user capability")
	}
	result, err := s.userSvc.BatchGet(ctx, userIDs)
	if err != nil || result == nil {
		return names, err
	}
	for id, projection := range result.Items {
		storageID, ok := userDomainIDStorageID(id)
		if !ok {
			continue
		}
		names[storageID] = userInfoDisplayName(projection)
	}
	return names, nil
}

// creatorDomainIDs converts plugin-owned notice creator storage values to
// user-domain IDs for host capability calls while de-duplicating the current page.
func creatorDomainIDs(notices []*NoticeEntity) []usercap.UserID {
	ids := make([]usercap.UserID, 0, len(notices))
	seen := make(map[int64]struct{}, len(notices))
	for _, notice := range notices {
		if notice == nil || notice.CreatedBy <= 0 {
			continue
		}
		if _, ok := seen[notice.CreatedBy]; ok {
			continue
		}
		seen[notice.CreatedBy] = struct{}{}
		ids = append(ids, usercap.UserID(strconv.FormatInt(notice.CreatedBy, 10)))
	}
	return ids
}

// userInfoStorageIDs converts visible user projections back to plugin
// notice creator IDs for database-side notice filtering.
func userInfoStorageIDs(users []*usercap.UserInfo) []int64 {
	ids := make([]int64, 0, len(users))
	for _, user := range users {
		if user == nil {
			continue
		}
		id, ok := userDomainIDStorageID(user.ID)
		if ok {
			ids = append(ids, id)
		}
	}
	return ids
}

// userDomainIDStorageID parses the current host user-domain ID encoding used by
// existing plugin-owned notice creator columns.
func userDomainIDStorageID(id usercap.UserID) (int64, bool) {
	storageID, err := strconv.ParseInt(strings.TrimSpace(string(id)), 10, 64)
	return storageID, err == nil && storageID > 0
}

// userInfoDisplayName chooses the stable notice creator display field
// from the current user-domain projection.
func userInfoDisplayName(user *usercap.UserInfo) string {
	if user == nil {
		return ""
	}
	if user.Username != "" {
		return user.Username
	}
	if user.Nickname != "" {
		return user.Nickname
	}
	if user.Label != "" {
		return user.Label
	}
	return string(user.ID)
}
