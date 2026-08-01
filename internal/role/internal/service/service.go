package service

import (
	"context"
	"strings"

	"github.com/Duke1616/ecmdb/internal/role/internal/domain"
	"github.com/Duke1616/ecmdb/internal/role/internal/repository"
	"golang.org/x/sync/errgroup"
)

type Service interface {
	// CreateRole 创建角色
	CreateRole(ctx context.Context, req domain.Role) (int64, error)
	// ListRole 获取角色列表
	ListRole(ctx context.Context, offset, limit int64) ([]domain.Role, int64, error)
	// UpdateRole 变更角色信息
	UpdateRole(ctx context.Context, req domain.Role) (int64, error)
	// CreateOrUpdateRoleMenuIds 新增角色的菜单权限
	CreateOrUpdateRoleMenuIds(ctx context.Context, code string, menuIds []int64) (int64, error)
	// CreateOrUpdateRolePermissions updates menu permissions and model visibility in one database operation.
	CreateOrUpdateRolePermissions(ctx context.Context, code string, menuIds []int64, deniedModelUIDs []string) (int64, error)
	// FilterAccessibleModelUIDs applies the union of all role model permissions.
	FilterAccessibleModelUIDs(ctx context.Context, roleCodes []string, modelUIDs []string) ([]string, error)
	// CanAccessModel reports whether at least one role grants access to the model.
	CanAccessModel(ctx context.Context, roleCodes []string, modelUID string) (bool, error)
	// FindByExcludeCodes 查找排除当前角色编码的数据
	FindByExcludeCodes(ctx context.Context, offset, limit int64, codes []string) ([]domain.Role, int64, error)
	// FindByIncludeCodes 查找包含当前角色编码的数据
	FindByIncludeCodes(ctx context.Context, codes []string) ([]domain.Role, error)
	// FindByMenuId 查找包含菜单 ID 的角色
	FindByMenuId(ctx context.Context, menuId int64) ([]domain.Role, error)
	// FindByRoleCode 查找角色编码数据
	FindByRoleCode(ctx context.Context, code string) (domain.Role, error)
	// DeleteRole 删除角色
	DeleteRole(ctx context.Context, id int64) (int64, error)
}

type service struct {
	repo repository.RoleRepository
}

func (s *service) DeleteRole(ctx context.Context, id int64) (int64, error) {
	return s.repo.DeleteRole(ctx, id)
}

func (s *service) FindByRoleCode(ctx context.Context, code string) (domain.Role, error) {
	return s.repo.FindByRoleCode(ctx, code)
}

func (s *service) FindByMenuId(ctx context.Context, menuId int64) ([]domain.Role, error) {
	return s.repo.FindByMenuId(ctx, menuId)
}

func (s *service) CreateOrUpdateRoleMenuIds(ctx context.Context, code string, menuIds []int64) (int64, error) {
	return s.repo.CreateOrUpdateRoleMenuIds(ctx, code, menuIds)
}

func (s *service) CreateOrUpdateRolePermissions(ctx context.Context, code string, menuIds []int64, deniedModelUIDs []string) (int64, error) {
	return s.repo.CreateOrUpdateRolePermissions(ctx, code, menuIds, normalizeModelUIDs(deniedModelUIDs))
}

func (s *service) FilterAccessibleModelUIDs(ctx context.Context, roleCodes []string, modelUIDs []string) ([]string, error) {
	if len(roleCodes) == 0 || len(modelUIDs) == 0 {
		return []string{}, nil
	}

	roles, err := s.repo.FindByIncludeCodes(ctx, roleCodes)
	if err != nil {
		return nil, err
	}
	return filterAccessibleModelUIDs(roles, modelUIDs), nil
}

func filterAccessibleModelUIDs(roles []domain.Role, modelUIDs []string) []string {
	deniedByRole := make([]map[string]struct{}, 0, len(roles))
	for _, r := range roles {
		denied := make(map[string]struct{}, len(r.DeniedModelUIDs))
		for _, uid := range r.DeniedModelUIDs {
			denied[uid] = struct{}{}
		}
		deniedByRole = append(deniedByRole, denied)
	}

	result := make([]string, 0, len(modelUIDs))
	seen := make(map[string]struct{}, len(modelUIDs))
	for _, uid := range modelUIDs {
		uid = strings.TrimSpace(uid)
		if uid == "" {
			continue
		}
		if _, exists := seen[uid]; exists {
			continue
		}

		for _, denied := range deniedByRole {
			if _, isDenied := denied[uid]; !isDenied {
				result = append(result, uid)
				seen[uid] = struct{}{}
				break
			}
		}
	}

	return result
}

func (s *service) CanAccessModel(ctx context.Context, roleCodes []string, modelUID string) (bool, error) {
	allowed, err := s.FilterAccessibleModelUIDs(ctx, roleCodes, []string{modelUID})
	if err != nil {
		return false, err
	}
	return len(allowed) == 1, nil
}

func normalizeModelUIDs(modelUIDs []string) []string {
	result := make([]string, 0, len(modelUIDs))
	seen := make(map[string]struct{}, len(modelUIDs))
	for _, uid := range modelUIDs {
		uid = strings.TrimSpace(uid)
		if uid == "" {
			continue
		}
		if _, exists := seen[uid]; exists {
			continue
		}
		seen[uid] = struct{}{}
		result = append(result, uid)
	}
	return result
}

func (s *service) FindByExcludeCodes(ctx context.Context, offset, limit int64, codes []string) ([]domain.Role, int64, error) {
	var (
		eg    errgroup.Group
		rs    []domain.Role
		total int64
	)
	eg.Go(func() error {
		var err error
		rs, err = s.repo.FindByExcludeCodes(ctx, offset, limit, codes)
		return err
	})

	eg.Go(func() error {
		var err error
		total, err = s.repo.CountByExcludeCodes(ctx, codes)
		return err
	})
	if err := eg.Wait(); err != nil {
		return rs, total, err
	}
	return rs, total, nil
}

func (s *service) FindByIncludeCodes(ctx context.Context, codes []string) ([]domain.Role, error) {
	return s.repo.FindByIncludeCodes(ctx, codes)
}

func (s *service) UpdateRole(ctx context.Context, req domain.Role) (int64, error) {
	return s.repo.UpdateRole(ctx, req)
}

func (s *service) ListRole(ctx context.Context, offset, limit int64) ([]domain.Role, int64, error) {
	var (
		eg    errgroup.Group
		rs    []domain.Role
		total int64
	)
	eg.Go(func() error {
		var err error
		rs, err = s.repo.ListRole(ctx, offset, limit)
		return err
	})

	eg.Go(func() error {
		var err error
		total, err = s.repo.Total(ctx)
		return err
	})
	if err := eg.Wait(); err != nil {
		return rs, total, err
	}
	return rs, total, nil
}

func (s *service) CreateRole(ctx context.Context, req domain.Role) (int64, error) {
	return s.repo.CreateRole(ctx, req)
}

func NewService(repo repository.RoleRepository) Service {
	return &service{
		repo: repo,
	}
}
