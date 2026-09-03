package full

import (
	"context"
	"fmt"
	"time"

	"github.com/userreksai/ecmdb-main/cmd/initial/menu"
	"github.com/userreksai/ecmdb-main/internal/role"
	userStruct "github.com/userreksai/ecmdb-main/internal/user"
	"go.mongodb.org/mongo-driver/mongo"
)

func (i *fullInitial) InitUser() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	fmt.Printf("👤 开始初始化用户数据...\n")

	// 创建用户
	fmt.Printf("🔧 创建系统管理员用户...\n")
	start := time.Now()
	user, err := i.App.UserSvc.FindOrCreateBySystem(ctx, userStruct.User{
		Password:    Password,
		Username:    UserName,
		DisplayName: DisPlayName,
	})
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("❌ 用户创建失败: %v\n", err)
		return 0, err
	}

	fmt.Printf("✅ 用户初始化完成! 耗时: %v\n", duration)
	return user.Id, nil
}

func (i *fullInitial) InitRole() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	fmt.Printf("🔐 开始初始化角色数据...\n")

	// 创建角色
	fmt.Printf("🔧 创建超级管理员角色...\n")
	start := time.Now()
	_, err := i.App.RoleSvc.CreateRole(ctx, role.Role{
		Name:   "超级管理员",
		Code:   RoleCode,
		Status: true,
	})
	duration := time.Since(start)

	if err != nil {
		// 检查是否为 MongoDB 重复键错误
		if mongo.IsDuplicateKeyError(err) {
			fmt.Printf("⚠️  角色已存在，跳过创建。耗时: %v\n", duration)
			return nil
		}

		fmt.Printf("❌ 角色创建失败: %v\n", err)
		return err
	}

	fmt.Printf("✅ 角色初始化完成! 耗时: %v\n", duration)
	return nil
}

func (i *fullInitial) InitMenu() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	ms := menu.GetInjectMenus()
	fmt.Printf("🔄 开始初始化菜单数据...\n")
	fmt.Printf("📊 菜单数据统计: 共 %d 个菜单项\n", len(ms))

	start := time.Now()
	_, err := i.App.MenuSvc.InjectMenu(ctx, ms)
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("❌ 菜单初始化失败: %v\n", err)
		return err
	}

	// 计算并存储菜单文件的 MD5 哈希值
	hashCalculator := menu.NewMenuHashCalculator()
	menuHash, err := hashCalculator.CalculateProjectMenuHash()
	if err != nil {
		fmt.Printf("⚠️  计算菜单哈希失败: %v\n", err)
	}

	// 存储菜单哈希值到版本信息中
	if err = i.App.VerSvc.SetMenuHash(ctx, menuHash); err != nil {
		fmt.Printf("⚠️  存储菜单哈希失败: %v\n", err)
	}

	fmt.Printf("✅ 菜单初始化完成! 耗时: %v\n", duration)
	return nil
}

func (i *fullInitial) InitPermission(userId int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	menuIds := menu.GetAllMenuIDs()
	fmt.Printf("🔄 开始初始化用户权限...\n")
	fmt.Printf("📊 权限数据统计: 共 %d 个菜单权限\n", len(menuIds))

	start := time.Now()
	// 角色添加菜单
	if _, err := i.App.RoleSvc.CreateOrUpdateRoleMenuIds(ctx, RoleCode, menuIds); err != nil {
		fmt.Printf("❌ 权限初始化失败: %v\n", err)
		return err
	}

	// 用户绑定角色
	if _, err := i.App.UserSvc.AddRoleBind(ctx, userId, []string{RoleCode}); err != nil {
		fmt.Printf("❌ 用户绑定角色失败: %v\n", err)
		return err
	}

	// casbin 刷新后端接口权限
	if err := i.App.PermissionSvc.AddPermissionForRole(ctx, RoleCode, menuIds); err != nil {
		fmt.Printf("❌ 权限初始化失败: %v\n", err)
		return err
	}

	duration := time.Since(start)
	fmt.Printf("✅ 权限初始化完成! 耗时: %v\n", duration)

	return nil
}
