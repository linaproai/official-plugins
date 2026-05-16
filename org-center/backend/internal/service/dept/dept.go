// Package dept implements department management for the org-center source
// plugin. It owns plugin_org_center_dept and related organization queries through plugin-local
// services instead of depending on host-internal dept packages.
package dept

import (
	"context"

	plugincontract "lina-core/pkg/pluginservice/contract"
	entitymodel "lina-plugin-org-center/backend/internal/model/entity"
)

// Table and column constants for organization storage.
const (
	colDeptID        = "id"
	colDeptParentID  = "parent_id"
	colDeptAncestors = "ancestors"
	colDeptName      = "name"
	colDeptCode      = "code"
	colDeptOrderNum  = "order_num"
	colDeptLeader    = "leader"
	colDeptPhone     = "phone"
	colDeptEmail     = "email"
	colDeptStatus    = "status"
	colDeptRemark    = "remark"

	colUserID       = "id"
	colUserUsername = "username"
	colUserNickname = "nickname"
	colUserDeptID   = "dept_id"
	colUserUserID   = "user_id"
)

// Service defines tenant-scoped department management for the org-center plugin.
type Service interface {
	// List queries tenant-visible departments with optional name and status filters.
	List(ctx context.Context, in ListInput) (*ListOutput, error)
	// Create creates one tenant-owned department record after validating parent
	// scope and code uniqueness.
	Create(ctx context.Context, in CreateInput) (int, error)
	// GetByID retrieves one tenant-visible department detail by primary key or
	// returns a business not-found error.
	GetByID(ctx context.Context, id int) (*DeptEntity, error)
	// Update updates one tenant-visible department record while preserving tree
	// integrity and code uniqueness.
	Update(ctx context.Context, in UpdateInput) error
	// Delete deletes one tenant-visible department when no child or user binding
	// blocks it.
	Delete(ctx context.Context, id int) error
	// Tree returns the tenant-visible department tree without mutating state.
	Tree(ctx context.Context) ([]*TreeNode, error)
	// Exclude returns tenant-visible department candidates excluding one subtree.
	Exclude(ctx context.Context, in ExcludeInput) ([]*DeptEntity, error)
	// Users returns selectable users for one tenant-visible department subtree
	// using keyword and limit filters.
	Users(ctx context.Context, deptID int, keyword string, limit int) ([]*DeptUser, error)
	// DescendantDeptIDs returns the given tenant-visible department plus descendants.
	DescendantDeptIDs(ctx context.Context, deptID int) ([]int, error)
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	tenantFilter plugincontract.TenantFilterService // tenantFilter constrains plugin-owned department rows.
}

// New creates and returns a new department service instance.
func New(tenantFilter plugincontract.TenantFilterService) Service {
	return &serviceImpl{tenantFilter: tenantFilter}
}

// DeptEntity mirrors the plugin-local generated plugin_org_center_dept entity.
type DeptEntity = entitymodel.Dept

// ListInput defines filters for department list queries.
type ListInput struct {
	Name   string
	Status *int
}

// ListOutput defines the department list result.
type ListOutput struct {
	List []*DeptEntity
}

// CreateInput defines the create-department input.
type CreateInput struct {
	ParentId int
	Name     string
	Code     string
	OrderNum int
	Leader   int
	Phone    string
	Email    string
	Status   int
	Remark   string
}

// UpdateInput defines the update-department input.
type UpdateInput struct {
	Id       int
	ParentId *int
	Name     *string
	Code     *string
	OrderNum *int
	Leader   *int
	Phone    *string
	Email    *string
	Status   *int
	Remark   *string
}

// ExcludeInput defines the exclude-subtree input.
type ExcludeInput struct {
	Id int
}

// TreeNode represents one department tree node.
type TreeNode struct {
	Id       int         `json:"id"`
	Label    string      `json:"label"`
	Children []*TreeNode `json:"children"`
}

// DeptUser represents one selectable user row.
type DeptUser struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
}

// userRow reuses the plugin-local generated sys_user entity.
type userRow = entitymodel.SysUser
