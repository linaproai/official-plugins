// Package post implements post management for the org-center source plugin.
// It owns plugin_org_center_post CRUD, select-option queries, and tree/export helpers without
// relying on host-internal post services.
package post

import (
	"context"

	plugincontract "lina-core/pkg/pluginservice/contract"
	entitymodel "lina-plugin-org-center/backend/internal/model/entity"
)

// Table and column constants for post management.
const (
	colPostID      = "id"
	colPostDeptID  = "dept_id"
	colPostCode    = "code"
	colPostName    = "name"
	colPostSort    = "sort"
	colPostStatus  = "status"
	colPostRemark  = "remark"
	colPostCreated = "created_at"

	colDeptID       = "id"
	colDeptParentID = "parent_id"
	colDeptName     = "name"
	colDeptOrderNum = "order_num"

	colUserPostPostID = "post_id"
)

// Runtime i18n keys for backend-owned post projections and exports.
const (
	postTreeUnassignedDeptKey     = "plugin.org-center.post.tree.unassignedDept"
	postExportHeaderCodeKey       = "plugin.org-center.post.export.headers.code"
	postExportHeaderNameKey       = "plugin.org-center.post.export.headers.name"
	postExportHeaderSortKey       = "plugin.org-center.post.export.headers.sort"
	postExportHeaderStatusKey     = "plugin.org-center.post.export.headers.status"
	postExportHeaderRemarkKey     = "plugin.org-center.post.export.headers.remark"
	postExportHeaderCreatedAtKey  = "plugin.org-center.post.export.headers.createdAt"
	postExportStatusEnabledKey    = "plugin.org-center.post.export.status.enabled"
	postExportStatusDisabledKey   = "plugin.org-center.post.export.status.disabled"
	postTreeUnassignedDeptDefault = "Unassigned"
)

// Service defines tenant-scoped post management for the org-center plugin.
type Service interface {
	// List queries tenant-visible posts with pagination and department/code/name/status filters.
	List(ctx context.Context, in ListInput) (*ListOutput, error)
	// Create creates one tenant-owned post record after department scope and code checks.
	Create(ctx context.Context, in CreateInput) (int, error)
	// GetByID retrieves one tenant-visible post detail by primary key or returns not found.
	GetByID(ctx context.Context, id int) (*PostEntity, error)
	// Update updates one tenant-visible post record while preserving department
	// scope and code uniqueness.
	Update(ctx context.Context, in UpdateInput) error
	// Delete deletes one or more tenant-visible posts, rejecting rows that still
	// have user bindings.
	Delete(ctx context.Context, ids string) error
	// DeptTree returns the tenant-visible department tree decorated with post counts.
	DeptTree(ctx context.Context) ([]*DeptTreeNode, error)
	// OptionSelect returns localized, tenant-visible post options for one department subtree.
	OptionSelect(ctx context.Context, in OptionSelectInput) ([]PostOption, error)
	// Export generates one Excel file for tenant-visible posts using runtime i18n
	// fallback keys and dictionary-compatible status text.
	Export(ctx context.Context, in ExportInput) ([]byte, error)
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	i18nSvc      plugincontract.I18nService         // i18nSvc resolves plugin runtime translations.
	tenantFilter plugincontract.TenantFilterService // tenantFilter constrains plugin-owned post rows.
}

// New creates and returns a new post service instance.
func New(i18nSvc plugincontract.I18nService, tenantFilter plugincontract.TenantFilterService) Service {
	return &serviceImpl{
		i18nSvc:      i18nSvc,
		tenantFilter: tenantFilter,
	}
}

// PostEntity mirrors the plugin-local generated plugin_org_center_post entity.
type PostEntity = entitymodel.Post

// ListInput defines filters for post list queries.
type ListInput struct {
	PageNum  int
	PageSize int
	DeptId   *int
	Code     string
	Name     string
	Status   *int
}

// ListOutput defines the paged post result.
type ListOutput struct {
	List  []*PostEntity
	Total int
}

// CreateInput defines the create-post input.
type CreateInput struct {
	DeptId int
	Code   string
	Name   string
	Sort   int
	Status int
	Remark string
}

// UpdateInput defines the update-post input.
type UpdateInput struct {
	Id     int
	DeptId *int
	Code   *string
	Name   *string
	Sort   *int
	Status *int
	Remark *string
}

// DeptTreeNode represents one post-filter department node.
type DeptTreeNode struct {
	Id        int             `json:"id"`
	Label     string          `json:"label"`
	PostCount int             `json:"postCount"`
	Children  []*DeptTreeNode `json:"children"`
}

// PostOption represents one selectable post row.
type PostOption struct {
	PostId   int    `json:"postId"`
	PostName string `json:"postName"`
}

// OptionSelectInput defines the option-select input.
type OptionSelectInput struct {
	DeptId *int
}

// ExportInput defines the export filters.
type ExportInput struct {
	DeptId *int
	Code   string
	Name   string
	Status *int
}

// deptRow reuses the plugin-local generated plugin_org_center_dept entity.
type deptRow = entitymodel.Dept

// deptCountRow is the grouped post-count projection keyed by department.
type deptCountRow struct {
	DeptId int `json:"deptId"`
	Cnt    int `json:"cnt"`
}
