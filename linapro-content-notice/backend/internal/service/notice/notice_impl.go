// notice_impl.go implements notice CRUD, publish state transitions, and
// tenant-scoped query/export behavior for the linapro-content-notice plugin. It keeps
// host i18n and tenant-filter dependencies injected so plugin-owned notice rows
// remain isolated from host-internal services.

package notice

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/notifycap"
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

	m := s.tenantFilter.Apply(ctx, dao.Notice.Ctx(ctx), "")

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
	err := s.tenantFilter.Apply(ctx, dao.Notice.Ctx(ctx), "").
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
	bizCtx := s.bizCtxSvc.Current(ctx)
	createdBy := int64(bizCtx.UserID)
	tenantID := s.tenantFilter.Context(ctx).TenantID

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
	if in.Status == NoticeStatusPublished {
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
	err := s.tenantFilter.Apply(ctx, dao.Notice.Ctx(ctx), "").
		Where(noticeColumns.Id, in.Id).
		Scan(&oldNotice)
	if err != nil {
		return err
	}
	if oldNotice == nil {
		return bizerr.NewCode(CodeNoticeNotFound)
	}

	bizCtx := s.bizCtxSvc.Current(ctx)
	updatedBy := int64(bizCtx.UserID)
	tenantID := s.tenantFilter.Context(ctx).TenantID

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
	if in.Status != nil && *in.Status == NoticeStatusPublished && oldNotice.Status == NoticeStatusDraft {
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
func (s *serviceImpl) Delete(ctx context.Context, ids string) error {
	idList := normalizeNoticeDeleteIDs(ids)
	if len(idList) == 0 {
		return bizerr.NewCode(CodeNoticeDeleteRequired)
	}

	// Soft delete using GoFrame's auto soft-delete feature.
	noticeColumns := dao.Notice.Columns()
	_, err := s.tenantFilter.Apply(ctx, dao.Notice.Ctx(ctx), "").
		WhereIn(noticeColumns.Id, idList).
		Delete()
	if err != nil {
		return err
	}

	if cascadeErr := s.notifySvc.DeleteBySource(ctx, s.capabilityContext(ctx, "notice.delete"), notifycap.SourceTypeNotice, idList); cascadeErr != nil {
		logger.Errorf(ctx, "cascade delete notify deliveries failed for notice ids %s: %v", ids, cascadeErr)
	}
	return nil
}

// normalizeNoticeDeleteIDs trims comma-separated notice IDs and removes empty
// entries before passing them to the DAO layer.
func normalizeNoticeDeleteIDs(ids string) []string {
	rawIDs := strings.Split(ids, ",")
	result := make([]string, 0, len(rawIDs))
	for _, id := range rawIDs {
		normalizedID := strings.TrimSpace(id)
		if normalizedID == "" {
			continue
		}
		result = append(result, normalizedID)
	}
	return result
}

// searchCreatorUserIDs resolves a creator keyword through the host user domain
// capability before filtering plugin-owned notice rows by creator ID.
func (s *serviceImpl) searchCreatorUserIDs(ctx context.Context, keyword string) ([]int64, error) {
	if s.userSvc == nil {
		return nil, gerror.New("linapro-content-notice requires host user capability")
	}
	result, err := s.userSvc.Search(ctx, s.capabilityContext(ctx, noticeCreatorCapabilityResource), usercap.SearchInput{
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
	return userProjectionStorageIDs(result.Items), nil
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
	result, err := s.userSvc.BatchGet(ctx, s.capabilityContext(ctx, noticeCreatorCapabilityResource), userIDs)
	if err != nil || result == nil {
		return names, err
	}
	for id, projection := range result.Items {
		storageID, ok := userDomainIDStorageID(id)
		if !ok {
			continue
		}
		names[storageID] = userProjectionDisplayName(projection)
	}
	return names, nil
}

// capabilityContext builds the audited domain-call context used by host user
// capabilities without exposing host-private request or storage objects.
func (s *serviceImpl) capabilityContext(ctx context.Context, resource string) capmodel.CapabilityContext {
	var current bizctxcap.CurrentContext
	if s.bizCtxSvc != nil {
		current = s.bizCtxSvc.Current(ctx)
	}
	tenantCtx := tenantspi.TenantFilterContext{TenantID: current.TenantID}
	if s.tenantFilter != nil {
		tenantCtx = s.tenantFilter.Context(ctx)
	}
	actorUserID := tenantCtx.ActingUserID
	if actorUserID == 0 {
		actorUserID = current.ActingUserID
	}
	if actorUserID == 0 {
		actorUserID = tenantCtx.UserID
	}
	if actorUserID == 0 {
		actorUserID = current.UserID
	}
	tenantID := tenantCtx.TenantID
	if tenantID == 0 {
		tenantID = current.TenantID
	}
	return capmodel.CapabilityContext{
		PluginID: pluginID,
		Actor: capmodel.CapabilityActor{
			Type:   capmodel.ActorTypeUser,
			UserID: int64(actorUserID),
			Name:   current.Username,
		},
		TenantID:    capmodel.DomainID(strconv.Itoa(tenantID)),
		Source:      capmodel.CapabilitySourceHTTP,
		Resource:    resource,
		RequestedAt: time.Now(),
	}
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

// userProjectionStorageIDs converts visible user projections back to plugin
// notice creator IDs for database-side notice filtering.
func userProjectionStorageIDs(users []*usercap.UserProjection) []int64 {
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

// userProjectionDisplayName chooses the stable notice creator display field
// from the current user-domain projection.
func userProjectionDisplayName(user *usercap.UserProjection) string {
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
