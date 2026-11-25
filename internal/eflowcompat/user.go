// Package eflowcompat 为独立部署的 eflow（工单流程编排）与 ecmdb-task 提供协议兼容的 gRPC 服务。
//
// eflow 面向上游 eiam / etask 的协议编程。本部署不引入 eiam，由 ECMDB 以相同的 proto 包名与字段
// 提供用户、部门服务，使 eflow 只需把 grpc.client.eiam 指向 ECMDB，业务代码无需修改：
//   - eiam.user.v1.UserService：用户查询（含是否管理员、所属部门）
//   - eiam.department.v1.DepartmentService：用户所在部门（负责人/分管领导）与部门成员
//   - ecmdb.runner.v1.RunnerService：执行单元与脚本解析，供 ecmdb-task 为 eflow 提交工作流执行
package eflowcompat

import (
	"context"
	"errors"

	userv1 "github.com/userreksai/ecmdb-main/api/proto/gen/eiam/user/v1"
	"github.com/userreksai/ecmdb-main/internal/role"
	"github.com/userreksai/ecmdb-main/internal/user"
	"github.com/ecodeclub/ekit/slice"
)

// UserServer 实现 eiam.user.v1.UserService。
type UserServer struct {
	userv1.UnimplementedUserServiceServer
	users user.Service
}

// NewUserServer 创建用户兼容服务。
func NewUserServer(users user.Service) *UserServer { return &UserServer{users: users} }

func (s *UserServer) QueryById(ctx context.Context, req *userv1.QueryByIdReq) (*userv1.QueryUserResp, error) {
	if req.GetId() <= 0 {
		return &userv1.QueryUserResp{ErrorCode: userv1.ErrorCode_INVALID_PARAMETER, ErrorMessage: "用户 ID 非法"}, nil
	}
	u, err := s.users.FindById(ctx, req.GetId())
	return singleUser(u, err)
}

func (s *UserServer) QueryByUsername(ctx context.Context, req *userv1.QueryByUsernameReq) (*userv1.QueryUserResp, error) {
	if req.GetUsername() == "" {
		return &userv1.QueryUserResp{ErrorCode: userv1.ErrorCode_INVALID_PARAMETER, ErrorMessage: "用户名不能为空"}, nil
	}
	u, err := s.users.FindByUsername(ctx, req.GetUsername())
	return singleUser(u, err)
}

func (s *UserServer) QueryByUsernames(ctx context.Context, req *userv1.QueryByUsernamesReq) (*userv1.QueryUsersResp, error) {
	if len(req.GetUsernames()) == 0 {
		return &userv1.QueryUsersResp{}, nil
	}
	users, err := s.users.FindByUsernames(ctx, req.GetUsernames())
	if err != nil {
		return nil, err
	}
	return &userv1.QueryUsersResp{Users: toProtoUsers(users)}, nil
}

func (s *UserServer) QueryByIds(ctx context.Context, req *userv1.QueryByIdsReq) (*userv1.QueryUsersResp, error) {
	if len(req.GetIds()) == 0 {
		return &userv1.QueryUsersResp{}, nil
	}
	users, err := s.users.FindByIds(ctx, req.GetIds())
	if err != nil {
		return nil, err
	}
	return &userv1.QueryUsersResp{Users: toProtoUsers(users)}, nil
}

// singleUser 把"查不到"映射为业务错误码而不是 gRPC 错误，与 eiam 的语义一致。
func singleUser(u user.User, err error) (*userv1.QueryUserResp, error) {
	if err != nil || u.Id <= 0 {
		msg := "用户不存在"
		if err != nil && !errors.Is(err, context.Canceled) {
			msg = err.Error()
		}
		return &userv1.QueryUserResp{ErrorCode: userv1.ErrorCode_USER_NOT_FOUND, ErrorMessage: msg}, nil
	}
	return &userv1.QueryUserResp{User: toProtoUser(u)}, nil
}

func toProtoUsers(users []user.User) []*userv1.User {
	return slice.Map(users, func(_ int, u user.User) *userv1.User { return toProtoUser(u) })
}

// toProtoUser 映射用户；管理员判定与 ECMDB 角色体系一致（拥有 admin 角色）。
func toProtoUser(u user.User) *userv1.User {
	displayName := u.DisplayName
	if displayName == "" {
		displayName = u.Username
	}
	return &userv1.User{
		Id:           u.Id,
		Username:     u.Username,
		DisplayName:  displayName,
		LarkUserId:   u.FeishuInfo.UserId,
		WechatUserId: u.WechatInfo.UserId,
		Email:        u.Email,
		DepartmentId: u.DepartmentId,
		IsAdmin:      slice.Contains(u.RoleCodes, role.AdminRole),
	}
}
