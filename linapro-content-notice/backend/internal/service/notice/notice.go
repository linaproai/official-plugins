// Package notice implements the notice-management, publication, and
// delete-cascade services for the linapro-content-notice source plugin. It owns the
// plugin_linapro_content_notice table access and consumes host capability seams for business
// context and notify fan-out.
package notice

import (
	"context"

	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/notifycap"
	"lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/capability/usercap"
)

// Notice type values (matching sys_notice_type dictionary)
const (
	noticeTypeNotice       = 1 // Notice
	noticeTypeAnnouncement = 2 // Announcement
)

// Notice status values (matching sys_notice_status dictionary)
const (
	noticeStatusDraft     = 0 // Draft
	noticeStatusPublished = 1 // Published
)

// Service defines tenant-scoped notice CRUD and publication fan-out for the plugin.
type Service interface {
	// List returns notices visible to ctx's tenant using the supplied page and
	// fuzzy title/creator/type filters. It returns DAO or tenant-scope errors.
	List(ctx context.Context, in ListInput) (*ListOutput, error)
	// GetById returns one tenant-visible notice by primary key with creator
	// display metadata, or CodeNoticeNotFound when the row is outside scope or absent.
	GetById(ctx context.Context, id int64) (*ListItem, error)
	// Create creates a new notice and, when the status is published, fans the
	// notice into the inbox pipeline through the host notify bridge. The row is
	// owned by the current tenant and user from ctx; dispatch failures are logged
	// without rolling back the persisted notice.
	Create(ctx context.Context, in CreateInput) (int64, error)
	// Update updates the notice fields. A transition from draft to published
	// triggers inbox publication through the host notify bridge. It only updates
	// the current tenant's row and returns CodeNoticeNotFound for missing scope.
	Update(ctx context.Context, in UpdateInput) error
	// Delete soft-deletes notices by IDs and cascades the deletion into the
	// notify delivery records so inboxes stay consistent. Empty IDs return a
	// business error; cascade failures are logged because the notice deletion is authoritative.
	Delete(ctx context.Context, ids string) error
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	bizCtxSvc bizctxcap.Service // Business context bridge
	notifySvc notifycap.Service // Notification capability
	tenantSvc tenantcap.Service // Tenant capability bridge
	userSvc   usercap.Service   // User domain projection capability
}

// New creates and returns a new Service instance.
func New(
	bizCtxSvc bizctxcap.Service,
	notifySvc notifycap.Service,
	tenantSvc tenantcap.Service,
	userSvc usercap.Service,
) Service {
	return &serviceImpl{
		bizCtxSvc: bizCtxSvc,
		notifySvc: notifySvc,
		tenantSvc: tenantSvc,
		userSvc:   userSvc,
	}
}

// ListInput defines input for List function.
type ListInput struct {
	PageNum   int    // Page number, starting from 1
	PageSize  int    // Page size
	Title     string // Title, supports fuzzy search
	Type      int    // Type: 1=Notice 2=Announcement
	CreatedBy string // Creator username, supports fuzzy search
}

// ListItem defines a single list item.
type ListItem struct {
	*NoticeEntity        // Notice entity
	CreatedByName string `json:"createdByName"` // Creator username
}

// ListOutput defines output for List function.
type ListOutput struct {
	List  []*ListItem // List items
	Total int         // Total count
}

// CreateInput defines input for Create function.
type CreateInput struct {
	Title   string // Title
	Type    int    // Type: 1=Notice 2=Announcement
	Content string // Content
	FileIds string // Attachment file IDs, comma-separated
	Status  int    // Status: 0=Draft 1=Published
	Remark  string // Remark
}

// UpdateInput defines input for Update function.
type UpdateInput struct {
	Id      int64   // Notice ID
	Title   *string // Title
	Type    *int    // Type: 1=Notice 2=Announcement
	Content *string // Content
	FileIds *string // Attachment file IDs, comma-separated
	Status  *int    // Status: 0=Draft 1=Published
	Remark  *string // Remark
}
