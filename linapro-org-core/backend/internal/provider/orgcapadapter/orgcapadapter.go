// Package orgcapadapter adapts linapro-org-core services to the framework
// organization capability provider contract.
package orgcapadapter

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/orgcap"
	"lina-core/pkg/plugin/capability/orgcap/orgspi"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
	"lina-core/pkg/plugin/capability/usercap"
	"lina-plugin-linapro-org-core/backend/internal/dao"
	"lina-plugin-linapro-org-core/backend/internal/model/do"
	entitymodel "lina-plugin-linapro-org-core/backend/internal/model/entity"
	deptsvc "lina-plugin-linapro-org-core/backend/internal/service/dept"
)

const (
	// postStatusEnabled is the enabled status value used by linapro-org-core posts.
	postStatusEnabled = 1
	// orgCapUnassignedDeptLabelKey is the runtime i18n key for the synthetic
	// Unassigned node exposed through the host orgcap contract.
	orgCapUnassignedDeptLabelKey = "plugin.linapro-org-core.post.tree.unassignedDept"
)

// Provider implements the stable host organization-capability contract.
type Provider struct {
	deptSvc      deptsvc.Service                    // deptSvc resolves department tree relationships.
	tenantFilter tenantspi.PluginTableFilterService // tenantFilter constrains organization provider queries.
	users        usercap.Service                    // users resolves host-owned user projections.
}

// Ensure Provider implements the published organization-capability provider.
var _ orgspi.Provider = (*Provider)(nil)

// New creates and returns a new provider instance.
func New(tenantFilter tenantspi.PluginTableFilterService, users usercap.Service) *Provider {
	return &Provider{
		deptSvc:      deptsvc.New(tenantFilter, users),
		tenantFilter: tenantFilter,
		users:        users,
	}
}

// ListUserDeptAssignments returns user -> department projections for the provided users.
func (p *Provider) ListUserDeptAssignments(ctx context.Context, userIDs []int) (map[int]*orgcap.UserDeptAssignment, error) {
	assignments := make(map[int]*orgcap.UserDeptAssignment)
	if len(userIDs) == 0 {
		return assignments, nil
	}

	var userDepts []*entitymodel.UserDept
	if err := p.tenantFilter.Apply(ctx, dao.UserDept.Ctx(ctx), "").
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
	if err := p.tenantFilter.Apply(ctx, dao.Dept.Ctx(ctx), "").
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

// GetUserDeptInfo returns one user's department projection.
func (p *Provider) GetUserDeptInfo(ctx context.Context, userID int) (int, string, error) {
	var userDept *entitymodel.UserDept
	if err := p.tenantFilter.Apply(ctx, dao.UserDept.Ctx(ctx), "").
		Where(dao.UserDept.Columns().UserId, userID).
		Scan(&userDept); err != nil || userDept == nil {
		return 0, "", err
	}

	var deptItem *entitymodel.Dept
	if err := p.tenantFilter.Apply(ctx, dao.Dept.Ctx(ctx), "").
		Where(dao.Dept.Columns().Id, userDept.DeptId).
		Scan(&deptItem); err != nil || deptItem == nil {
		return 0, "", err
	}
	return deptItem.Id, deptItem.Name, nil
}

// GetUserDeptIDs returns one user's department identifier list.
func (p *Provider) GetUserDeptIDs(ctx context.Context, userID int) ([]int, error) {
	var userDepts []*entitymodel.UserDept
	if err := p.tenantFilter.Apply(ctx, dao.UserDept.Ctx(ctx), "").
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

// ApplyUserDeptScope injects an EXISTS-based department membership constraint
// into a host-owned query without materializing all visible user IDs in memory.
func (p *Provider) ApplyUserDeptScope(
	ctx context.Context,
	model *gdb.Model,
	userIDColumn string,
	currentUserID int,
) (*gdb.Model, bool, error) {
	subQuery, empty, err := p.BuildUserDeptScopeExists(ctx, userIDColumn, currentUserID)
	if err != nil || empty {
		return model, empty, err
	}
	return model.Where("EXISTS ?", subQuery), false, nil
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
	subQuery := p.tenantFilter.Apply(ctx, dao.UserDept.Ctx(ctx), dao.UserDept.Table()).
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
	subQuery := p.tenantFilter.Apply(ctx, dao.UserDept.Ctx(ctx), dao.UserDept.Table()).
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
	subQuery := p.tenantFilter.Apply(ctx, dao.UserDept.Ctx(ctx), dao.UserDept.Table()).
		Fields(cols.UserId).
		Where(fmt.Sprintf("%s = %s", qualifiedUserDeptColumn(cols.UserId), userIDColumn))
	return model.WhereNotExists(subQuery), false, nil
}

// currentVisibleDeptIDs returns the current user's department IDs plus all
// descendant department IDs with duplicates removed.
func (p *Provider) currentVisibleDeptIDs(ctx context.Context, currentUserID int) ([]int, error) {
	deptIDs, err := p.GetUserDeptIDs(ctx, currentUserID)
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

// GetUserPostIDs returns one user's post association list.
func (p *Provider) GetUserPostIDs(ctx context.Context, userID int) ([]int, error) {
	var userPosts []*entitymodel.UserPost
	if err := p.tenantFilter.Apply(ctx, dao.UserPost.Ctx(ctx), "").
		Where(dao.UserPost.Columns().UserId, userID).
		Scan(&userPosts); err != nil {
		return nil, err
	}

	ids := make([]int, 0, len(userPosts))
	for _, item := range userPosts {
		if item == nil {
			continue
		}
		ids = append(ids, item.PostId)
	}
	return ids, nil
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
	assignments, err := p.ListUserDeptAssignments(ctx, normalized)
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
	if err = p.tenantFilter.Apply(ctx, dao.UserPost.Ctx(ctx), "").
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
	nodes, err := p.UserDeptTree(ctx)
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

// SearchDepartments returns bounded department candidates.
func (p *Provider) SearchDepartments(
	ctx context.Context,
	input orgcap.DeptSearchInput,
) (*capmodel.PageResult[*orgcap.DeptProjection], error) {
	pageNum, pageSize := normalizeProviderPage(input.Page, orgcap.MaxDeptSearchPageSize)
	model := p.tenantFilter.Apply(ctx, dao.Dept.Ctx(ctx), "")
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
	items := make([]*orgcap.DeptProjection, 0, len(depts))
	for _, deptItem := range depts {
		if deptItem == nil {
			continue
		}
		items = append(items, &orgcap.DeptProjection{
			DeptID:   deptItem.Id,
			ParentID: deptItem.ParentId,
			DeptName: deptItem.Name,
			DeptCode: deptItem.Code,
			Status:   deptItem.Status,
		})
	}
	return &capmodel.PageResult[*orgcap.DeptProjection]{Items: items, Total: total}, nil
}

// ListPostOptionsPage returns bounded post candidates.
func (p *Provider) ListPostOptionsPage(
	ctx context.Context,
	input orgcap.PostOptionsInput,
) (*capmodel.PageResult[*orgcap.PostOption], error) {
	pageNum, pageSize := normalizeProviderPage(input.Page, orgcap.MaxPostOptionsPageSize)
	model := p.tenantFilter.Apply(ctx, dao.Post.Ctx(ctx), "")
	if input.Status != nil {
		model = model.Where(dao.Post.Columns().Status, *input.Status)
	} else {
		model = model.Where(dao.Post.Columns().Status, postStatusEnabled)
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
	options := make([]*orgcap.PostOption, 0, len(posts))
	for _, postItem := range posts {
		if postItem == nil {
			continue
		}
		options = append(options, &orgcap.PostOption{PostID: postItem.Id, PostName: postItem.Name})
	}
	return &capmodel.PageResult[*orgcap.PostOption]{Items: options, Total: total}, nil
}

// EnsureDepartmentsVisible verifies all department identifiers are visible.
func (p *Provider) EnsureDepartmentsVisible(ctx context.Context, deptIDs []int) error {
	normalized := uniquePositiveInts(deptIDs)
	if len(normalized) == 0 {
		return nil
	}
	count, err := p.tenantFilter.Apply(ctx, dao.Dept.Ctx(ctx), "").
		WhereIn(dao.Dept.Columns().Id, normalized).
		Count()
	if err != nil {
		return err
	}
	if count != len(normalized) {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	return nil
}

// EnsurePostsVisible verifies all post identifiers are visible.
func (p *Provider) EnsurePostsVisible(ctx context.Context, postIDs []int) error {
	normalized := uniquePositiveInts(postIDs)
	if len(normalized) == 0 {
		return nil
	}
	count, err := p.tenantFilter.Apply(ctx, dao.Post.Ctx(ctx), "").
		WhereIn(dao.Post.Columns().Id, normalized).
		Count()
	if err != nil {
		return err
	}
	if count != len(normalized) {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	return nil
}

// ReplaceUserAssignments rewrites one user's department and post associations.
func (p *Provider) ReplaceUserAssignments(ctx context.Context, userID int, deptID *int, postIDs []int) error {
	tenantID := p.tenantFilter.Context(ctx).TenantID
	return dao.UserDept.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err := tx.Model(dao.UserDept.Table()).
			Ctx(ctx).
			Where(tenantspi.TenantFilterColumn, tenantID).
			Where(dao.UserDept.Columns().UserId, userID).
			Delete(); err != nil {
			return err
		}
		if _, err := tx.Model(dao.UserPost.Table()).
			Ctx(ctx).
			Where(tenantspi.TenantFilterColumn, tenantID).
			Where(dao.UserPost.Columns().UserId, userID).
			Delete(); err != nil {
			return err
		}

		if deptID != nil && *deptID > 0 {
			if _, err := tx.Model(dao.UserDept.Table()).
				Ctx(ctx).
				Data(do.UserDept{TenantId: tenantID, UserId: userID, DeptId: *deptID}).
				Insert(); err != nil {
				return err
			}
		}
		for _, postID := range postIDs {
			if _, err := tx.Model(dao.UserPost.Table()).
				Ctx(ctx).
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
	tenantID := p.tenantFilter.Context(ctx).TenantID
	return dao.UserDept.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err := tx.Model(dao.UserDept.Table()).
			Ctx(ctx).
			Where(tenantspi.TenantFilterColumn, tenantID).
			Where(dao.UserDept.Columns().UserId, userID).
			Delete(); err != nil {
			return err
		}
		if _, err := tx.Model(dao.UserPost.Table()).
			Ctx(ctx).
			Where(tenantspi.TenantFilterColumn, tenantID).
			Where(dao.UserPost.Columns().UserId, userID).
			Delete(); err != nil {
			return err
		}
		return nil
	})
}

// UserDeptTree returns the optional department tree used by host user management.
func (p *Provider) UserDeptTree(ctx context.Context) ([]*orgcap.DeptTreeNode, error) {
	if p == nil || p.users == nil {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", "user"))
	}
	plainTree, err := p.deptSvc.Tree(ctx)
	if err != nil {
		return nil, err
	}

	var userDepts []*entitymodel.UserDept
	if err = p.tenantFilter.Apply(ctx, dao.UserDept.Ctx(ctx), "").
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
	visibleUsers, err := p.users.BatchGet(ctx, p.capabilityContext(ctx, "org.user_dept_tree"), userIDs)
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

	totalOut, err := p.users.Search(ctx, p.capabilityContext(ctx, "org.user_dept_tree"), usercap.SearchInput{
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

// capabilityContext creates provider-call metadata for usercap calls made by
// the organization provider.
func (p *Provider) capabilityContext(ctx context.Context, resource string) capmodel.CapabilityContext {
	tenantCtx := tenantspi.TenantFilterContext{}
	if p != nil && p.tenantFilter != nil {
		tenantCtx = p.tenantFilter.Context(ctx)
	}
	actorID := tenantCtx.ActingUserID
	if actorID == 0 {
		actorID = tenantCtx.UserID
	}
	actor := capmodel.CapabilityActor{
		Type:   capmodel.ActorTypeUser,
		UserID: int64(actorID),
	}
	if actorID == 0 {
		actor = capmodel.CapabilityActor{
			Type:         capmodel.ActorTypeSystem,
			Name:         orgcap.ProviderPluginID,
			SystemReason: "organization provider user projection",
		}
	}
	return capmodel.CapabilityContext{
		PluginID:    orgcap.ProviderPluginID,
		Actor:       actor,
		TenantID:    capmodel.DomainID(strconv.Itoa(tenantCtx.TenantID)),
		Source:      capmodel.CapabilitySourceProvider,
		SystemCall:  actor.Type == capmodel.ActorTypeSystem,
		Resource:    resource,
		RequestedAt: time.Now(),
	}
}

// ListPostOptions returns selectable post options for one department subtree.
func (p *Provider) ListPostOptions(ctx context.Context, deptID *int) ([]*orgcap.PostOption, error) {
	model := p.tenantFilter.Apply(ctx, dao.Post.Ctx(ctx), "").Where(dao.Post.Columns().Status, postStatusEnabled)
	if deptID != nil {
		deptIDs, err := p.deptSvc.DescendantDeptIDs(ctx, *deptID)
		if err != nil {
			return nil, err
		}
		model = model.WhereIn(dao.Post.Columns().DeptId, deptIDs)
	}

	var posts []*entitymodel.Post
	if err := model.OrderAsc(dao.Post.Columns().Sort).Scan(&posts); err != nil {
		return nil, err
	}

	options := make([]*orgcap.PostOption, 0, len(posts))
	for _, postItem := range posts {
		if postItem == nil {
			continue
		}
		options = append(options, &orgcap.PostOption{
			PostID:   postItem.Id,
			PostName: postItem.Name,
		})
	}
	return options, nil
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
	if err := p.tenantFilter.Apply(ctx, dao.Post.Ctx(ctx), "").
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
