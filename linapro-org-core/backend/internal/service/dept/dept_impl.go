// dept_impl.go implements department tree, CRUD, import/export, and user
// assignment helpers for the linapro-org-core plugin. It maintains plugin-owned
// organization rows and protects hierarchy mutations from cycles or orphaned
// assignments.

package dept

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
	"lina-core/pkg/plugin/capability/usercap"
	"lina-plugin-linapro-org-core/backend/internal/dao"
	"lina-plugin-linapro-org-core/backend/internal/model/do"
	entitymodel "lina-plugin-linapro-org-core/backend/internal/model/entity"
)

const (
	deptCapabilityPluginID = "linapro-org-core"
	deptUserDefaultLimit   = 10
	deptUserMaxLimit       = 200
)

// List queries the department list with optional filters.
func (s *serviceImpl) List(ctx context.Context, in ListInput) (*ListOutput, error) {
	model := s.tenantFilter.Apply(ctx, dao.Dept.Ctx(ctx), "")
	if in.Name != "" {
		model = model.WhereLike(colDeptName, "%"+in.Name+"%")
	}
	if in.Status != nil {
		model = model.Where(colDeptStatus, *in.Status)
	}

	list := make([]*DeptEntity, 0)
	if err := model.OrderAsc(colDeptOrderNum).Scan(&list); err != nil {
		return nil, err
	}
	return &ListOutput{List: list}, nil
}

// Create creates one department record.
func (s *serviceImpl) Create(ctx context.Context, in CreateInput) (int, error) {
	if in.Code != "" {
		if err := s.checkCodeUnique(ctx, in.Code, 0); err != nil {
			return 0, err
		}
	}

	ancestors := "0"
	if in.ParentId != 0 {
		parent, err := s.GetByID(ctx, in.ParentId)
		if err != nil {
			return 0, err
		}
		ancestors = fmt.Sprintf("%s,%d", parent.Ancestors, in.ParentId)
	}

	tenantID := s.tenantFilter.Context(ctx).TenantID
	id, err := dao.Dept.Ctx(ctx).Data(do.Dept{
		TenantId:  tenantID,
		ParentId:  in.ParentId,
		Ancestors: ancestors,
		Name:      in.Name,
		Code:      in.Code,
		OrderNum:  in.OrderNum,
		Leader:    in.Leader,
		Phone:     in.Phone,
		Email:     in.Email,
		Status:    in.Status,
		Remark:    in.Remark,
	}).InsertAndGetId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

// GetByID retrieves one department detail by primary key.
func (s *serviceImpl) GetByID(ctx context.Context, id int) (*DeptEntity, error) {
	var dept *DeptEntity
	err := s.tenantFilter.Apply(ctx, dao.Dept.Ctx(ctx), "").
		Where(colDeptID, id).
		Scan(&dept)
	if err != nil {
		return nil, err
	}
	if dept == nil {
		return nil, bizerr.NewCode(CodeDeptNotFound)
	}
	return dept, nil
}

// Update updates one department record.
func (s *serviceImpl) Update(ctx context.Context, in UpdateInput) error {
	dept, err := s.GetByID(ctx, in.Id)
	if err != nil {
		return err
	}

	buildData := func(includeParent bool, newParentID int, newAncestors string) (do.Dept, error) {
		data := do.Dept{}
		if includeParent {
			data.ParentId = newParentID
			data.Ancestors = newAncestors
		}
		if in.Name != nil {
			data.Name = *in.Name
		}
		if in.Code != nil {
			if *in.Code != "" {
				if err := s.checkCodeUnique(ctx, *in.Code, in.Id); err != nil {
					return data, err
				}
			}
			data.Code = *in.Code
		}
		if in.OrderNum != nil {
			data.OrderNum = *in.OrderNum
		}
		if in.Leader != nil {
			data.Leader = *in.Leader
		}
		if in.Phone != nil {
			data.Phone = *in.Phone
		}
		if in.Email != nil {
			data.Email = *in.Email
		}
		if in.Status != nil {
			data.Status = *in.Status
		}
		if in.Remark != nil {
			data.Remark = *in.Remark
		}
		return data, nil
	}

	if in.ParentId != nil && *in.ParentId != dept.ParentId {
		newParentID := *in.ParentId
		newAncestors := "0"
		if newParentID != 0 {
			parent, err := s.GetByID(ctx, newParentID)
			if err != nil {
				return err
			}
			newAncestors = fmt.Sprintf("%s,%d", parent.Ancestors, newParentID)
		}

		oldPrefix := fmt.Sprintf("%s,%d", dept.Ancestors, in.Id)
		newPrefix := fmt.Sprintf("%s,%d", newAncestors, in.Id)
		children := make([]*DeptEntity, 0)
		err = s.tenantFilter.Apply(ctx, dao.Dept.Ctx(ctx), "").
			WhereLike(colDeptAncestors, oldPrefix+",%").
			WhereOr(colDeptParentID, in.Id).
			Scan(&children)
		if err != nil {
			return err
		}

		tenantID := s.tenantFilter.Context(ctx).TenantID
		err = dao.Dept.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
			for _, child := range children {
				if child == nil {
					continue
				}
				childAncestors := gstr.Replace(child.Ancestors, oldPrefix, newPrefix, 1)
				_, err = tx.Model(dao.Dept.Table()).Safe().Ctx(ctx).
					OmitNilData().
					Where(tenantspi.TenantFilterColumn, tenantID).
					Where(colDeptID, child.Id).
					Data(do.Dept{Ancestors: childAncestors}).
					Update()
				if err != nil {
					return err
				}
			}
			data, buildErr := buildData(true, newParentID, newAncestors)
			if buildErr != nil {
				return buildErr
			}
			_, err = tx.Model(dao.Dept.Table()).Safe().Ctx(ctx).
				OmitNilData().
				Where(tenantspi.TenantFilterColumn, tenantID).
				Where(colDeptID, in.Id).
				Data(data).
				Update()
			return err
		})
		return err
	}

	data, err := buildData(false, 0, "")
	if err != nil {
		return err
	}
	tenantID := s.tenantFilter.Context(ctx).TenantID
	_, err = dao.Dept.Ctx(ctx).
		OmitNilData().
		Where(tenantspi.TenantFilterColumn, tenantID).
		Where(colDeptID, in.Id).
		Data(data).
		Update()
	return err
}

// Delete deletes one department when no child or user binding blocks it.
func (s *serviceImpl) Delete(ctx context.Context, id int) error {
	childCount, err := s.tenantFilter.Apply(ctx, dao.Dept.Ctx(ctx), "").
		Where(colDeptParentID, id).
		Count()
	if err != nil {
		return err
	}
	if childCount > 0 {
		return bizerr.NewCode(CodeDeptHasChildrenDeleteDenied)
	}

	userCount, err := s.tenantFilter.Apply(ctx, dao.UserDept.Ctx(ctx), "").
		Where(colUserDeptID, id).
		Count()
	if err != nil {
		return err
	}
	if userCount > 0 {
		return bizerr.NewCode(CodeDeptHasUsersDeleteDenied)
	}

	_, err = s.tenantFilter.Apply(ctx, dao.Dept.Ctx(ctx), "").
		Where(colDeptID, id).
		Delete()
	return err
}

// Tree returns the plain department tree.
func (s *serviceImpl) Tree(ctx context.Context) ([]*TreeNode, error) {
	deptList := make([]*DeptEntity, 0)
	err := s.tenantFilter.Apply(ctx, dao.Dept.Ctx(ctx), "").
		OrderAsc(colDeptOrderNum).
		Scan(&deptList)
	if err != nil {
		return nil, err
	}

	nodeMap := make(map[int]*TreeNode, len(deptList))
	for _, deptItem := range deptList {
		if deptItem == nil {
			continue
		}
		nodeMap[deptItem.Id] = &TreeNode{Id: deptItem.Id, Label: deptItem.Name, Children: make([]*TreeNode, 0)}
	}

	roots := make([]*TreeNode, 0)
	for _, deptItem := range deptList {
		if deptItem == nil {
			continue
		}
		node := nodeMap[deptItem.Id]
		if parent, ok := nodeMap[deptItem.ParentId]; ok {
			parent.Children = append(parent.Children, node)
			continue
		}
		roots = append(roots, node)
	}
	return roots, nil
}

// Exclude returns department candidates excluding one subtree.
func (s *serviceImpl) Exclude(ctx context.Context, in ExcludeInput) ([]*DeptEntity, error) {
	dept, err := s.GetByID(ctx, in.Id)
	if err != nil {
		return nil, err
	}

	prefix := fmt.Sprintf("%s,%d", dept.Ancestors, in.Id)
	list := make([]*DeptEntity, 0)
	err = s.tenantFilter.Apply(ctx, dao.Dept.Ctx(ctx), "").
		WhereNot(colDeptID, in.Id).
		WhereNotLike(colDeptAncestors, prefix+",%").
		WhereNotLike(colDeptAncestors, prefix).
		OrderAsc(colDeptOrderNum).
		Scan(&list)
	if err != nil {
		return nil, err
	}
	return list, nil
}

// Users returns selectable users for one department subtree.
func (s *serviceImpl) Users(ctx context.Context, deptID int, keyword string, limit int) ([]*DeptUser, error) {
	if s == nil || s.users == nil {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", "user"))
	}
	limit = normalizeDeptUserLimit(limit)
	if deptID == 0 {
		out, err := s.users.SearchUsers(ctx, s.capabilityContext(ctx, "dept.users"), usercap.SearchInput{
			Keyword: keyword,
			Page:    capmodel.PageRequest{PageSize: limit},
		})
		if err != nil {
			return nil, err
		}
		return toDeptUsers(out.Items, limit), nil
	}

	deptIDs, err := s.DescendantDeptIDs(ctx, deptID)
	if err != nil {
		return nil, err
	}
	if keyword != "" {
		return s.usersByKeywordAndDeptIDs(ctx, keyword, limit, deptIDs)
	}

	userDeptRows := make([]*entitymodel.UserDept, 0)
	err = s.tenantFilter.Apply(ctx, dao.UserDept.Ctx(ctx), "").
		Fields(colUserUserID).
		WhereIn(colUserDeptID, deptIDs).
		Group(colUserUserID).
		Limit(limit).
		Scan(&userDeptRows)
	if err != nil {
		return nil, err
	}
	if len(userDeptRows) == 0 {
		return make([]*DeptUser, 0), nil
	}

	seen := make(map[int]struct{}, len(userDeptRows))
	userIDs := make([]int, 0, len(userDeptRows))
	for _, row := range userDeptRows {
		if row == nil {
			continue
		}
		if _, ok := seen[row.UserId]; ok {
			continue
		}
		seen[row.UserId] = struct{}{}
		userIDs = append(userIDs, row.UserId)
	}

	projections, err := s.batchGetUsers(ctx, userIDs)
	if err != nil {
		return nil, err
	}
	return toDeptUsers(projections, limit), nil
}

// DescendantDeptIDs returns the given department plus all descendants.
func (s *serviceImpl) DescendantDeptIDs(ctx context.Context, deptID int) ([]int, error) {
	deptIDs := []int{deptID}
	parentIDs := []int{deptID}
	for len(parentIDs) > 0 {
		childValues, err := s.tenantFilter.Apply(ctx, dao.Dept.Ctx(ctx), "").
			WhereIn(colDeptParentID, parentIDs).
			Fields(colDeptID).
			Array()
		if err != nil {
			return nil, err
		}
		childIDs := gconv.Ints(childValues)
		deptIDs = append(deptIDs, childIDs...)
		parentIDs = childIDs
	}
	return deptIDs, nil
}

// checkCodeUnique checks whether one department code already exists.
func (s *serviceImpl) checkCodeUnique(ctx context.Context, code string, excludeID int) error {
	model := s.tenantFilter.Apply(ctx, dao.Dept.Ctx(ctx), "").Where(colDeptCode, code)
	if excludeID > 0 {
		model = model.WhereNot(colDeptID, excludeID)
	}
	count, err := model.Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return bizerr.NewCode(CodeDeptCodeExists)
	}
	return nil
}

// usersByKeywordAndDeptIDs searches visible users through usercap and intersects
// the bounded result with plugin-owned department assignments.
func (s *serviceImpl) usersByKeywordAndDeptIDs(ctx context.Context, keyword string, limit int, deptIDs []int) ([]*DeptUser, error) {
	searchLimit := deptUserMaxLimit
	out, err := s.users.SearchUsers(ctx, s.capabilityContext(ctx, "dept.users"), usercap.SearchInput{
		Keyword: keyword,
		Page:    capmodel.PageRequest{PageSize: searchLimit},
	})
	if err != nil {
		return nil, err
	}
	if out == nil || len(out.Items) == 0 {
		return []*DeptUser{}, nil
	}
	candidateIDs := make([]int, 0, len(out.Items))
	projectionsByID := make(map[int]*usercap.UserProjection, len(out.Items))
	for _, projection := range out.Items {
		userID, ok := parseUserProjectionID(projection)
		if !ok {
			continue
		}
		candidateIDs = append(candidateIDs, userID)
		projectionsByID[userID] = projection
	}
	if len(candidateIDs) == 0 {
		return []*DeptUser{}, nil
	}

	userDeptRows := make([]*entitymodel.UserDept, 0)
	err = s.tenantFilter.Apply(ctx, dao.UserDept.Ctx(ctx), "").
		Fields(colUserUserID).
		WhereIn(colUserDeptID, deptIDs).
		WhereIn(colUserUserID, candidateIDs).
		Group(colUserUserID).
		Scan(&userDeptRows)
	if err != nil {
		return nil, err
	}
	assigned := make(map[int]struct{}, len(userDeptRows))
	for _, row := range userDeptRows {
		if row != nil {
			assigned[row.UserId] = struct{}{}
		}
	}
	filtered := make([]*usercap.UserProjection, 0, limit)
	for _, userID := range candidateIDs {
		if _, ok := assigned[userID]; !ok {
			continue
		}
		filtered = append(filtered, projectionsByID[userID])
		if len(filtered) >= limit {
			break
		}
	}
	return toDeptUsers(filtered, limit), nil
}

// batchGetUsers resolves user projections through usercap while preserving
// assignment order and hiding missing or invisible users.
func (s *serviceImpl) batchGetUsers(ctx context.Context, userIDs []int) ([]*usercap.UserProjection, error) {
	ids := make([]usercap.UserID, 0, len(userIDs))
	for _, userID := range userIDs {
		if userID > 0 {
			ids = append(ids, usercap.UserID(strconv.Itoa(userID)))
		}
	}
	out, err := s.users.BatchGetUsers(ctx, s.capabilityContext(ctx, "dept.users"), ids)
	if err != nil {
		return nil, err
	}
	result := make([]*usercap.UserProjection, 0, len(ids))
	if out == nil {
		return result, nil
	}
	for _, id := range ids {
		if item := out.Items[id]; item != nil {
			result = append(result, item)
		}
	}
	return result, nil
}

// capabilityContext creates the plugin-visible metadata for organization HTTP
// reads without exposing host request internals.
func (s *serviceImpl) capabilityContext(ctx context.Context, resource string) capmodel.CapabilityContext {
	tenantCtx := tenantspi.TenantFilterContext{}
	if s != nil && s.tenantFilter != nil {
		tenantCtx = s.tenantFilter.Context(ctx)
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
			Name:         deptCapabilityPluginID,
			SystemReason: "department user projection",
		}
	}
	return capmodel.CapabilityContext{
		PluginID:    deptCapabilityPluginID,
		Actor:       actor,
		TenantID:    capmodel.DomainID(strconv.Itoa(tenantCtx.TenantID)),
		Source:      capmodel.CapabilitySourceHTTP,
		SystemCall:  actor.Type == capmodel.ActorTypeSystem,
		Resource:    resource,
		RequestedAt: time.Now(),
	}
}

// normalizeDeptUserLimit applies the bounded selector contract.
func normalizeDeptUserLimit(limit int) int {
	if limit <= 0 {
		return deptUserDefaultLimit
	}
	if limit > deptUserMaxLimit {
		return deptUserMaxLimit
	}
	return limit
}

// toDeptUsers converts usercap projections into selectable user rows.
func toDeptUsers(rows []*usercap.UserProjection, limit int) []*DeptUser {
	result := make([]*DeptUser, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		id, ok := parseUserProjectionID(row)
		if !ok {
			continue
		}
		result = append(result, &DeptUser{Id: id, Username: row.Username, Nickname: row.Nickname})
		if limit > 0 && len(result) >= limit {
			break
		}
	}
	return result
}

// parseUserProjectionID decodes the domain string ID back to the API's legacy
// numeric response shape at the controller boundary.
func parseUserProjectionID(row *usercap.UserProjection) (int, bool) {
	if row == nil {
		return 0, false
	}
	id, err := strconv.Atoi(string(row.ID))
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}
