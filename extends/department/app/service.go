// Package app implements department use cases.
package app

import (
	"context"

	coreCache "kingfisher/core/cache"
	"kingfisher/core/errcode"
	"kingfisher/core/query"
	"kingfisher/extends/department/domain"
	"kingfisher/extends/department/port"
)

// Error 携带 errcode 的错误类型，handler 层据此映射到 HTTP 错误码
type Error struct{ Code int }

func (e *Error) Error() string { return errcode.Msg(e.Code) }

// DepartmentService 部门服务
type DepartmentService struct {
	repo  port.DepartmentRepository
	cache coreCache.Cache
}

func NewDepartmentService(repo port.DepartmentRepository, c coreCache.Cache) *DepartmentService {
	return &DepartmentService{repo: repo, cache: c}
}

// Tree 返回全量部门树（含挂载角色）。
func (s *DepartmentService) Tree(ctx context.Context) ([]domain.Department, error) {
	all, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	return buildTree(all, 0), nil
}

// List 分页部门列表（支持子树筛选）。
func (s *DepartmentService) List(ctx context.Context, q *query.Query) ([]domain.Department, int64, error) {
	return s.repo.ListPage(ctx, q)
}

// GetByID 返回单个部门（含挂载角色）。
func (s *DepartmentService) GetByID(ctx context.Context, id uint) (*domain.Department, error) {
	d, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, &Error{Code: errcode.ErrDeptNotFound}
	}
	return d, nil
}

// Create 创建部门。
func (s *DepartmentService) Create(ctx context.Context, d *domain.Department) error {
	if d.Status == 0 {
		d.Status = 1
	}
	return s.repo.Create(ctx, d)
}

// Update 更新部门基础字段（name/parent_id/sort/status/remark）。
func (s *DepartmentService) Update(ctx context.Context, id uint, updates map[string]any) error {
	return s.repo.Update(ctx, id, updates)
}

// Delete 删除部门：有子部门则拒绝；否则删部门并级联清理成员/角色关联（repo 事务内），随后清空权限缓存。
func (s *DepartmentService) Delete(ctx context.Context, id uint) error {
	has, err := s.repo.HasChildren(ctx, id)
	if err != nil {
		return err
	}
	if has {
		return &Error{Code: errcode.ErrDeptHasChildren}
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	s.flushPermCache(ctx)
	return nil
}

// AssignRoles 替换部门的角色关联，并清空权限缓存（部门角色变化会影响所有成员用户的有效权限）。
func (s *DepartmentService) AssignRoles(ctx context.Context, departmentID uint, roleIDs []uint) error {
	if err := s.repo.SetRoles(ctx, departmentID, roleIDs); err != nil {
		return err
	}
	s.flushPermCache(ctx)
	return nil
}

// flushPermCache 清空 user:perms:* 缓存（沿用角色改权限时的 DeleteByPattern 模式）。
func (s *DepartmentService) flushPermCache(ctx context.Context) {
	if s.cache != nil {
		_ = s.cache.DeleteByPattern(ctx, "user:perms:*")
	}
}

// buildTree 把扁平部门列表递归组树（copy doc/menu 模块的 buildTree 模式）。
func buildTree(depts []domain.Department, parentID uint) []domain.Department {
	var result []domain.Department
	for _, d := range depts {
		if d.ParentID == parentID {
			d.Children = buildTree(depts, d.ID)
			result = append(result, d)
		}
	}
	return result
}
