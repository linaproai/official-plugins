// Package orgcapadapter adapts linapro-org-core services to the framework
// organization capability provider contract.
package orgcapadapter

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/orgcap"
	"lina-core/pkg/plugin/capability/orgcap/orgspi"
	"lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
	"lina-core/pkg/plugin/capability/usercap"
	"lina-plugin-linapro-org-core/backend/internal/dao"
	"lina-plugin-linapro-org-core/backend/internal/model/do"
	entitymodel "lina-plugin-linapro-org-core/backend/internal/model/entity"
	deptsvc "lina-plugin-linapro-org-core/backend/internal/service/dept"
	postsvc "lina-plugin-linapro-org-core/backend/internal/service/post"
)

const (
	// orgCapUnassignedDeptLabelKey is the runtime i18n key for the synthetic
	// Unassigned node exposed through the host orgcap contract.
	orgCapUnassignedDeptLabelKey = "plugin.linapro-org-core.post.tree.unassignedDept"
)

// Provider implements the stable host organization-capability contract.
type Provider struct {
	deptSvc   deptsvc.Service   // deptSvc resolves department tree relationships.
	postSvc   postsvc.Service   // postSvc owns post CRUD and option queries.
	tenantSvc tenantcap.Service // tenantSvc provides tenant context and plugin table filtering.
	users     usercap.Service   // users resolves host-owned user projections.
}

// Ensure Provider implements the published organization-capability provider.
var _ orgspi.Provider = (*Provider)(nil)

// New creates and returns a new provider instance.
func New(tenantSvc tenantcap.Service, users usercap.Service) *Provider {
	return &Provider{
		deptSvc:   deptsvc.New(tenantSvc, users),
		postSvc:   postsvc.New(nil, tenantSvc),
		tenantSvc: tenantSvc,
		users:     users,
	}
}

// pluginTableFilter returns the tenant table-filter slice from the injected tenant service.
func (p *Provider) pluginTableFilter() tenantcap.FilterService {
	if p == nil || p.tenantSvc == nil {
		return nil
	}
	return p.tenantSvc.Filter()
}

// tenantFilterContext returns current tenant metadata for write ownership fields.
func (p *Provider) tenantFilterContext(ctx context.Context) tenantcap.TenantFilterContext {
	if filter := p.pluginTableFilter(); filter != nil {
		return filter.Context(ctx)
	}
	return tenantcap.TenantFilterContext{}
}

// listUserDeptAssignments returns user -> department projections for the provided users.
func (p *Provider) listUserDeptAssignments(ctx context.Context, userIDs []int) (map[int]*orgcap.UserDeptAssignment, error) {
	assignments := make(map[int]*orgcap.UserDeptAssignment)
	if len(userIDs) == 0 {
		return assignments, nil
	}

	var userDepts []*entitymodel.UserDept
	if err := tenantspi.ApplyPluginTableFilter(ctx, p.pluginTableFilter(), dao.UserDept.Ctx(ctx), "").
		WhereIn(dao.UserDept.Columns().UserId, userIDs).
		Scan(&userDepts); err != nil {
		return nil, err
	}

	deptIDs := make([]int, 0, len(userDepts))
	for _, item := range userDepts {
		if item == nil {
			continue
		}
		assignments[item.UserId] = &orgcap.UserDeptAssignment{DeptID: item.DeptId}
		deptIDs = append(deptIDs, item.DeptId)
	}
	if len(deptIDs) == 0 {
		return assignments, nil
	}

	var deptList []*entitymodel.Dept
	if err := tenantspi.ApplyPluginTableFilter(ctx, p.pluginTableFilter(), dao.Dept.Ctx(ctx), "").
		WhereIn(dao.Dept.Columns().Id, deptIDs).
		Scan(&deptList); err != nil {
		return nil, err
	}
	for _, deptItem := range deptList {
		if deptItem == nil {
			continue
		}
		for userID, assignment := range assignments {
			if assignment != nil && assignment.DeptID == deptItem.Id {
				assignments[userID] = &orgcap.UserDeptAssignment{
					DeptID:   deptItem.Id,
					DeptName: deptItem.Name,
				}
			}
		}
	}
	return assignments, nil
}

// userDeptIDs returns one user's department identifier list.
func (p *Provider) userDeptIDs(ctx context.Context, userID int) ([]int, error) {
	var userDepts []*entitymodel.UserDept
	if err := tenantspi.ApplyPluginTableFilter(ctx, p.pluginTableFilter(), dao.UserDept.Ctx(ctx), "").
		Where(dao.UserDept.Columns().UserId, userID).
		Scan(&userDepts); err != nil {
		return nil, err
	}

	deptIDs := make([]int, 0, len(userDepts))
	seen := make(map[int]struct{}, len(userDepts))
	for _, item := range userDepts {
		if item == nil {
			continue
		}
		if _, ok := seen[item.DeptId]; ok {
			continue
		}
		seen[item.DeptId] = struct{}{}
		deptIDs = append(deptIDs, item.DeptId)
	}
	return deptIDs, nil
}

// BuildUserDeptScopeExists builds an EXISTS subquery for department membership
// without applying it immediately, allowing host callers to compose it with
// additional OR branches.
func (p *Provider) BuildUserDeptScopeExists(
	ctx context.Context,
	userIDColumn string,
	currentUserID int,
) (*gdb.Model, bool, error) {
	deptIDs, err := p.currentVisibleDeptIDs(ctx, currentUserID)
	if err != nil {
		return nil, false, err
	}
	if len(deptIDs) == 0 {
		return nil, true, nil
	}

	cols := dao.UserDept.Columns()
	subQuery := tenantspi.ApplyPluginTableFilter(ctx, p.pluginTableFilter(), dao.UserDept.Ctx(ctx), dao.UserDept.Table()).
		Fields(cols.UserId).
		Where(fmt.Sprintf("%s = %s", qualifiedUserDeptColumn(cols.UserId), userIDColumn)).
		WhereIn(cols.DeptId, deptIDs)
	return subQuery, false, nil
}

// ApplyUserDeptFilter constrains user rows to one department subtree with a
// correlated EXISTS query, avoiding high-cardinality user ID materialization.
func (p *Provider) ApplyUserDeptFilter(
	ctx context.Context,
	model *gdb.Model,
	userIDColumn string,
	deptID int,
) (*gdb.Model, bool, error) {
	deptIDs, err := p.deptSvc.DescendantDeptIDs(ctx, deptID)
	if err != nil {
		return nil, false, err
	}
	if len(deptIDs) == 0 {
		return model, true, nil
	}

	cols := dao.UserDept.Columns()
	subQuery := tenantspi.ApplyPluginTableFilter(ctx, p.pluginTableFilter(), dao.UserDept.Ctx(ctx), dao.UserDept.Table()).
		Fields(cols.UserId).
		Where(fmt.Sprintf("%s = %s", qualifiedUserDeptColumn(cols.UserId), userIDColumn)).
		WhereIn(cols.DeptId, deptIDs)
	return model.Where("EXISTS ?", subQuery), false, nil
}

// ApplyUserDeptUnassignedFilter constrains user rows to users without any
// department assignment in the current tenant.
func (p *Provider) ApplyUserDeptUnassignedFilter(
	ctx context.Context,
	model *gdb.Model,
	userIDColumn string,
) (*gdb.Model, bool, error) {
	cols := dao.UserDept.Columns()
	subQuery := tenantspi.ApplyPluginTableFilter(ctx, p.pluginTableFilter(), dao.UserDept.Ctx(ctx), dao.UserDept.Table()).
		Fields(cols.UserId).
		Where(fmt.Sprintf("%s = %s", qualifiedUserDeptColumn(cols.UserId), userIDColumn))
	return model.WhereNotExists(subQuery), false, nil
}

// currentVisibleDeptIDs returns the current user's department IDs plus all
// descendant department IDs with duplicates removed.
func (p *Provider) currentVisibleDeptIDs(ctx context.Context, currentUserID int) ([]int, error) {
	deptIDs, err := p.userDeptIDs(ctx, currentUserID)
	if err != nil {
		return nil, err
	}
	if len(deptIDs) == 0 {
		return []int{}, nil
	}

	seen := make(map[int]struct{})
	visibleDeptIDs := make([]int, 0, len(deptIDs))
	for _, deptID := range deptIDs {
		descendantIDs, resolveErr := p.deptSvc.DescendantDeptIDs(ctx, deptID)
		if resolveErr != nil {
			return nil, resolveErr
		}
		for _, descendantID := range descendantIDs {
			if _, ok := seen[descendantID]; ok {
				continue
			}
			seen[descendantID] = struct{}{}
			visibleDeptIDs = append(visibleDeptIDs, descendantID)
		}
	}
	return visibleDeptIDs, nil
}

// qualifiedUserDeptColumn returns one fully qualified user-department column
// name for correlated subqueries.
func qualifiedUserDeptColumn(column string) string {
	return fmt.Sprintf("%s.%s", dao.UserDept.Table(), column)
}

// BatchGetUserOrgProfiles returns stable organization profiles for provided users.
func (p *Provider) BatchGetUserOrgProfiles(
	ctx context.Context,
	userIDs []int,
) (*capmodel.BatchResult[*orgcap.UserOrgProfile, int], error) {
	result := &capmodel.BatchResult[*orgcap.UserOrgProfile, int]{
		Items:      make(map[int]*orgcap.UserOrgProfile, len(userIDs)),
		MissingIDs: make([]int, 0),
	}
	if len(userIDs) == 0 {
		return result, nil
	}
	normalized := uniquePositiveInts(userIDs)
	for _, userID := range normalized {
		result.Items[userID] = &orgcap.UserOrgProfile{UserID: userID}
	}
	assignments, err := p.listUserDeptAssignments(ctx, normalized)
	if err != nil {
		return nil, err
	}
	for userID, assignment := range assignments {
		if assignment == nil {
			continue
		}
		profile := result.Items[userID]
		if profile == nil {
			profile = &orgcap.UserOrgProfile{UserID: userID}
			result.Items[userID] = profile
		}
		profile.DeptID = assignment.DeptID
		profile.DeptName = assignment.DeptName
	}

	userPosts := make([]*entitymodel.UserPost, 0)
	if err = tenantspi.ApplyPluginTableFilter(ctx, p.pluginTableFilter(), dao.UserPost.Ctx(ctx), "").
		Fields(dao.UserPost.Columns().UserId, dao.UserPost.Columns().PostId).
		WhereIn(dao.UserPost.Columns().UserId, normalized).
		Scan(&userPosts); err != nil {
		return nil, err
	}
	postIDs := make([]int, 0, len(userPosts))
	userPostIDs := make(map[int][]int, len(normalized))
	for _, item := range userPosts {
		if item == nil {
			continue
		}
		userPostIDs[item.UserId] = append(userPostIDs[item.UserId], item.PostId)
		postIDs = append(postIDs, item.PostId)
	}
	postNames, err := p.postNames(ctx, postIDs)
	if err != nil {
		return nil, err
	}
	for userID, ids := range userPostIDs {
		profile := result.Items[userID]
		if profile == nil {
			profile = &orgcap.UserOrgProfile{UserID: userID}
			result.Items[userID] = profile
		}
		profile.PostIDs = uniquePositiveInts(ids)
		profile.PostNames = make([]string, 0, len(profile.PostIDs))
		for _, postID := range profile.PostIDs {
			if name := postNames[postID]; name != "" {
				profile.PostNames = append(profile.PostNames, name)
			}
		}
	}
	return result, nil
}

// ListDeptTree returns a bounded ordinary department tree projection.
func (p *Provider) ListDeptTree(ctx context.Context, input orgcap.DeptTreeInput) (*orgcap.DeptTreeResult, error) {
	nodes, err := p.userDeptTree(ctx)
	if err != nil {
		return nil, err
	}
	total := countDeptTreeNodes(nodes)
	maxNodes := input.MaxNodes
	if maxNodes <= 0 || maxNodes > orgcap.MaxDeptTreeNodes {
		maxNodes = orgcap.MaxDeptTreeNodes
	}
	items, truncated := truncateDeptTreeNodes(nodes, maxNodes)
	return &orgcap.DeptTreeResult{Items: items, Total: total, Truncated: truncated}, nil
}

// BatchGetDepartments returns visible department projections and opaque missing IDs.
func (p *Provider) BatchGetDepartments(
	ctx context.Context,
	deptIDs []int,
) (*capmodel.BatchResult[*orgcap.DeptInfo, int], error) {
	result := &capmodel.BatchResult[*orgcap.DeptInfo, int]{
		Items:      make(map[int]*orgcap.DeptInfo),
		MissingIDs: make([]int, 0),
	}
	normalized := uniquePositiveInts(deptIDs)
	if len(normalized) == 0 {
		return result, nil
	}
	depts := make([]*entitymodel.Dept, 0, len(normalized))
	if err := tenantspi.ApplyPluginTableFilter(ctx, p.pluginTableFilter(), dao.Dept.Ctx(ctx), "").
		WhereIn(dao.Dept.Columns().Id, normalized).
		Scan(&depts); err != nil {
		return nil, err
	}
	for _, deptItem := range depts {
		if deptItem == nil {
			continue
		}
		result.Items[deptItem.Id] = toDeptInfo(deptItem)
	}
	for _, deptID := range normalized {
		if _, ok := result.Items[deptID]; !ok {
			result.MissingIDs = append(result.MissingIDs, deptID)
		}
	}
	return result, nil
}

// SearchDepartments returns bounded department candidates.
func (p *Provider) SearchDepartments(
	ctx context.Context,
	input orgcap.DeptListInput,
) (*capmodel.PageResult[*orgcap.DeptInfo], error) {
	pageNum, pageSize := normalizeProviderPage(input.Page, orgcap.MaxDeptSearchPageSize)
	model := tenantspi.ApplyPluginTableFilter(ctx, p.pluginTableFilter(), dao.Dept.Ctx(ctx), "")
	if keyword := strings.TrimSpace(input.Keyword); keyword != "" {
		keywordFilter := model.Builder().
			WhereLike(dao.Dept.Columns().Name, "%"+keyword+"%").
			WhereOrLike(dao.Dept.Columns().Code, "%"+keyword+"%")
		model = model.Where(keywordFilter)
	}
	if input.Status != nil {
		model = model.Where(dao.Dept.Columns().Status, *input.Status)
	}
	total, err := model.Count()
	if err != nil {
		return nil, err
	}
	depts := make([]*entitymodel.Dept, 0)
	if err = model.Page(pageNum, pageSize).OrderAsc(dao.Dept.Columns().OrderNum).Scan(&depts); err != nil {
		return nil, err
	}
	items := make([]*orgcap.DeptInfo, 0, len(depts))
	for _, deptItem := range depts {
		if deptItem == nil {
			continue
		}
		items = append(items, toDeptInfo(deptItem))
	}
	return &capmodel.PageResult[*orgcap.DeptInfo]{Items: items, Total: total}, nil
}

// CreateDepartment creates one department through the plugin-owned department service.
func (p *Provider) CreateDepartment(ctx context.Context, input orgcap.DeptCreateInput) (int, error) {
	return p.deptSvc.Create(ctx, deptsvc.CreateInput{
		ParentId: input.ParentID,
		Name:     input.DeptName,
		Code:     input.DeptCode,
		OrderNum: input.OrderNum,
		Leader:   input.LeaderUserID,
		Phone:    input.Phone,
		Email:    input.Email,
		Status:   input.Status,
		Remark:   input.Remark,
	})
}

// UpdateDepartment updates one department through the plugin-owned department service.
func (p *Provider) UpdateDepartment(ctx context.Context, input orgcap.DeptUpdateInput) error {
	return p.deptSvc.Update(ctx, deptsvc.UpdateInput{
		Id:       input.DeptID,
		ParentId: input.ParentID,
		Name:     input.DeptName,
		Code:     input.DeptCode,
		OrderNum: input.OrderNum,
		Leader:   input.LeaderUserID,
		Phone:    input.Phone,
		Email:    input.Email,
		Status:   input.Status,
		Remark:   input.Remark,
	})
}

// DeleteDepartment deletes one department through the plugin-owned department service.
func (p *Provider) DeleteDepartment(ctx context.Context, deptID int) error {
	return p.deptSvc.Delete(ctx, deptID)
}

// BatchGetPosts returns visible post projections and opaque missing IDs.
func (p *Provider) BatchGetPosts(
	ctx context.Context,
	postIDs []int,
) (*capmodel.BatchResult[*orgcap.PostInfo, int], error) {
	result := &capmodel.BatchResult[*orgcap.PostInfo, int]{
		Items:      make(map[int]*orgcap.PostInfo),
		MissingIDs: make([]int, 0),
	}
	normalized := uniquePositiveInts(postIDs)
	if len(normalized) == 0 {
		return result, nil
	}
	posts := make([]*entitymodel.Post, 0, len(normalized))
	if err := tenantspi.ApplyPluginTableFilter(ctx, p.pluginTableFilter(), dao.Post.Ctx(ctx), "").
		WhereIn(dao.Post.Columns().Id, normalized).
		Scan(&posts); err != nil {
		return nil, err
	}
	for _, postItem := range posts {
		if postItem == nil {
			continue
		}
		result.Items[postItem.Id] = toPostInfo(postItem)
	}
	for _, postID := range normalized {
		if _, ok := result.Items[postID]; !ok {
			result.MissingIDs = append(result.MissingIDs, postID)
		}
	}
	return result, nil
}

// ListPosts returns bounded post projections.
func (p *Provider) ListPosts(
	ctx context.Context,
	input orgcap.PostListInput,
) (*capmodel.PageResult[*orgcap.PostInfo], error) {
	pageNum, pageSize := normalizeProviderPage(input.Page, orgcap.MaxPostOptionsPageSize)
	model := tenantspi.ApplyPluginTableFilter(ctx, p.pluginTableFilter(), dao.Post.Ctx(ctx), "")
	if input.Status != nil {
		model = model.Where(dao.Post.Columns().Status, *input.Status)
	}
	if input.DeptID != nil {
		deptIDs, err := p.deptSvc.DescendantDeptIDs(ctx, *input.DeptID)
		if err != nil {
			return nil, err
		}
		model = model.WhereIn(dao.Post.Columns().DeptId, deptIDs)
	}
	if keyword := strings.TrimSpace(input.Keyword); keyword != "" {
		keywordFilter := model.Builder().
			WhereLike(dao.Post.Columns().Name, "%"+keyword+"%").
			WhereOrLike(dao.Post.Columns().Code, "%"+keyword+"%")
		model = model.Where(keywordFilter)
	}
	total, err := model.Count()
	if err != nil {
		return nil, err
	}
	posts := make([]*entitymodel.Post, 0)
	if err = model.Page(pageNum, pageSize).OrderAsc(dao.Post.Columns().Sort).Scan(&posts); err != nil {
		return nil, err
	}
	items := make([]*orgcap.PostInfo, 0, len(posts))
	for _, postItem := range posts {
		if postItem == nil {
			continue
		}
		items = append(items, toPostInfo(postItem))
	}
	return &capmodel.PageResult[*orgcap.PostInfo]{Items: items, Total: total}, nil
}

// CreatePost creates one post through the plugin-owned post service.
func (p *Provider) CreatePost(ctx context.Context, input orgcap.PostCreateInput) (int, error) {
	return p.postSvc.Create(ctx, postsvc.CreateInput{
		DeptId: input.DeptID,
		Code:   input.PostCode,
		Name:   input.PostName,
		Sort:   input.Sort,
		Status: input.Status,
		Remark: input.Remark,
	})
}

// UpdatePost updates one post through the plugin-owned post service.
func (p *Provider) UpdatePost(ctx context.Context, input orgcap.PostUpdateInput) error {
	return p.postSvc.Update(ctx, postsvc.UpdateInput{
		Id:     input.PostID,
		DeptId: input.DeptID,
		Code:   input.PostCode,
		Name:   input.PostName,
		Sort:   input.Sort,
		Status: input.Status,
		Remark: input.Remark,
	})
}

// DeletePost deletes one post through the plugin-owned post service.
func (p *Provider) DeletePost(ctx context.Context, postID int) error {
	return p.postSvc.Delete(ctx, []int{postID})
}

// ReplaceUserAssignments rewrites one user's department and post associations.
func (p *Provider) ReplaceUserAssignments(ctx context.Context, userID int, deptID *int, postIDs []int) error {
	tenantID := p.tenantFilterContext(ctx).TenantID
	return dao.UserDept.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		if _, err := dao.UserDept.Ctx(ctx).
			Where(tenantspi.TenantFilterColumn, tenantID).
			Where(dao.UserDept.Columns().UserId, userID).
			Delete(); err != nil {
			return err
		}
		if _, err := dao.UserPost.Ctx(ctx).
			Where(tenantspi.TenantFilterColumn, tenantID).
			Where(dao.UserPost.Columns().UserId, userID).
			Delete(); err != nil {
			return err
		}

		if deptID != nil && *deptID > 0 {
			if _, err := dao.UserDept.Ctx(ctx).
				Data(do.UserDept{TenantId: tenantID, UserId: userID, DeptId: *deptID}).
				Insert(); err != nil {
				return err
			}
		}
		for _, postID := range postIDs {
			if _, err := dao.UserPost.Ctx(ctx).
				Data(do.UserPost{TenantId: tenantID, UserId: userID, PostId: postID}).
				Insert(); err != nil {
				return err
			}
		}
		return nil
	})
}

// CleanupUserAssignments deletes one user's optional organization associations.
func (p *Provider) CleanupUserAssignments(ctx context.Context, userID int) error {
	tenantID := p.tenantFilterContext(ctx).TenantID
	return dao.UserDept.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		if _, err := dao.UserDept.Ctx(ctx).
			Where(tenantspi.TenantFilterColumn, tenantID).
			Where(dao.UserDept.Columns().UserId, userID).
			Delete(); err != nil {
			return err
		}
		if _, err := dao.UserPost.Ctx(ctx).
			Where(tenantspi.TenantFilterColumn, tenantID).
			Where(dao.UserPost.Columns().UserId, userID).
			Delete(); err != nil {
			return err
		}
		return nil
	})
}

// userDeptTree returns the optional department tree used by host user management.
func (p *Provider) userDeptTree(ctx context.Context) ([]*orgcap.DeptTreeNode, error) {
	if p == nil || p.users == nil {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", "user"))
	}
	plainTree, err := p.deptSvc.Tree(ctx)
	if err != nil {
		return nil, err
	}

	var userDepts []*entitymodel.UserDept
	if err = tenantspi.ApplyPluginTableFilter(ctx, p.pluginTableFilter(), dao.UserDept.Ctx(ctx), "").
		Fields(dao.UserDept.Columns().DeptId, dao.UserDept.Columns().UserId).
		Scan(&userDepts); err != nil {
		return nil, err
	}

	userIDs := make([]usercap.UserID, 0, len(userDepts))
	seenUsers := make(map[int]struct{}, len(userDepts))
	for _, item := range userDepts {
		if item == nil || item.UserId <= 0 {
			continue
		}
		if _, ok := seenUsers[item.UserId]; ok {
			continue
		}
		seenUsers[item.UserId] = struct{}{}
		userIDs = append(userIDs, usercap.UserID(strconv.Itoa(item.UserId)))
	}
	visibleUsers, err := p.users.BatchGet(ctx, userIDs)
	if err != nil {
		return nil, err
	}
	visibleUserIDs := make(map[int]struct{}, len(userIDs))
	if visibleUsers != nil {
		for id := range visibleUsers.Items {
			parsedID, parseErr := strconv.Atoi(string(id))
			if parseErr == nil && parsedID > 0 {
				visibleUserIDs[parsedID] = struct{}{}
			}
		}
	}

	countMap := make(map[int]int, len(userDepts))
	for _, item := range userDepts {
		if item == nil {
			continue
		}
		if _, ok := visibleUserIDs[item.UserId]; !ok {
			continue
		}
		countMap[item.DeptId]++
	}

	nodes := convertDeptTreeNodes(plainTree)
	applyDeptUserCount(nodes, countMap)

	totalOut, err := p.users.List(ctx, usercap.ListInput{
		Page: capmodel.PageRequest{PageSize: 1},
	})
	if err != nil {
		return nil, err
	}
	totalUsers := 0
	if totalOut != nil {
		totalUsers = totalOut.Total
	}

	assignedUsers := 0
	for _, item := range countMap {
		assignedUsers += item
	}

	return append(nodes, newUnassignedDeptNode(totalUsers, assignedUsers)), nil
}

// toDeptInfo converts one plugin-local department row to the shared capability projection.
func toDeptInfo(deptItem *entitymodel.Dept) *orgcap.DeptInfo {
	if deptItem == nil {
		return nil
	}
	return &orgcap.DeptInfo{
		DeptID:   deptItem.Id,
		ParentID: deptItem.ParentId,
		DeptName: deptItem.Name,
		DeptCode: deptItem.Code,
		Status:   deptItem.Status,
	}
}

// toPostInfo converts one plugin-local post row to the shared capability projection.
func toPostInfo(postItem *entitymodel.Post) *orgcap.PostInfo {
	if postItem == nil {
		return nil
	}
	return &orgcap.PostInfo{
		PostID:   postItem.Id,
		DeptID:   postItem.DeptId,
		PostCode: postItem.Code,
		PostName: postItem.Name,
		Sort:     postItem.Sort,
		Status:   postItem.Status,
		Remark:   postItem.Remark,
	}
}

// convertDeptTreeNodes converts plugin-local tree nodes into the shared host contract.
func convertDeptTreeNodes(nodes []*deptsvc.TreeNode) []*orgcap.DeptTreeNode {
	result := make([]*orgcap.DeptTreeNode, 0, len(nodes))
	for _, node := range nodes {
		if node == nil {
			continue
		}
		result = append(result, &orgcap.DeptTreeNode{
			Id:       node.Id,
			Label:    node.Label,
			Children: convertDeptTreeNodes(node.Children),
		})
	}
	return result
}

// applyDeptUserCount rolls grouped department user counts up the tree and appends the count to labels.
func applyDeptUserCount(nodes []*orgcap.DeptTreeNode, countMap map[int]int) {
	for _, node := range nodes {
		if node == nil {
			continue
		}
		applyDeptUserCount(node.Children, countMap)
		node.UserCount = countMap[node.Id]
		for _, child := range node.Children {
			if child == nil {
				continue
			}
			node.UserCount += child.UserCount
		}
		node.Label = fmt.Sprintf("%s(%d)", node.Label, node.UserCount)
	}
}

// newUnassignedDeptNode creates the synthetic Unassigned projection
// with a stable label key so host controllers can localize the label.
func newUnassignedDeptNode(totalUsers int, assignedUsers int) *orgcap.DeptTreeNode {
	unassignedUsers := totalUsers - assignedUsers
	return &orgcap.DeptTreeNode{
		Id:        0,
		Label:     fmt.Sprintf("Unassigned (%d)", unassignedUsers),
		LabelKey:  orgCapUnassignedDeptLabelKey,
		UserCount: unassignedUsers,
		Children:  make([]*orgcap.DeptTreeNode, 0),
	}
}

// postNames returns post display names keyed by post identifier.
func (p *Provider) postNames(ctx context.Context, postIDs []int) (map[int]string, error) {
	normalized := uniquePositiveInts(postIDs)
	result := make(map[int]string, len(normalized))
	if len(normalized) == 0 {
		return result, nil
	}
	posts := make([]*entitymodel.Post, 0, len(normalized))
	if err := tenantspi.ApplyPluginTableFilter(ctx, p.pluginTableFilter(), dao.Post.Ctx(ctx), "").
		Fields(dao.Post.Columns().Id, dao.Post.Columns().Name).
		WhereIn(dao.Post.Columns().Id, normalized).
		Scan(&posts); err != nil {
		return nil, err
	}
	for _, postItem := range posts {
		if postItem != nil {
			result[postItem.Id] = postItem.Name
		}
	}
	return result, nil
}

// normalizeProviderPage applies bounded provider-side page defaults.
func normalizeProviderPage(page capmodel.PageRequest, maxPageSize int) (int, int) {
	pageNum := page.PageNum
	if pageNum <= 0 {
		pageNum = 1
	}
	pageSize := page.PageSize
	if pageSize <= 0 {
		pageSize = page.Limit
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if maxPageSize > 0 && pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return pageNum, pageSize
}

// uniquePositiveInts returns positive identifiers once in input order.
func uniquePositiveInts(ids []int) []int {
	if len(ids) == 0 {
		return nil
	}
	result := make([]int, 0, len(ids))
	seen := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

// countDeptTreeNodes counts all nodes in a department tree projection.
func countDeptTreeNodes(nodes []*orgcap.DeptTreeNode) int {
	total := 0
	for _, node := range nodes {
		if node == nil {
			continue
		}
		total++
		total += countDeptTreeNodes(node.Children)
	}
	return total
}

// truncateDeptTreeNodes returns a detached tree prefix capped by maxNodes.
func truncateDeptTreeNodes(nodes []*orgcap.DeptTreeNode, maxNodes int) ([]*orgcap.DeptTreeNode, bool) {
	remaining := maxNodes
	truncated := false
	var clone func([]*orgcap.DeptTreeNode) []*orgcap.DeptTreeNode
	clone = func(source []*orgcap.DeptTreeNode) []*orgcap.DeptTreeNode {
		out := make([]*orgcap.DeptTreeNode, 0, len(source))
		for _, node := range source {
			if node == nil {
				continue
			}
			if remaining <= 0 {
				truncated = true
				break
			}
			remaining--
			copyNode := *node
			copyNode.Children = clone(node.Children)
			out = append(out, &copyNode)
		}
		return out
	}
	return clone(nodes), truncated
}
