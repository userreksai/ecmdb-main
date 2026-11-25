package eflowcompat

import (
	"context"

	departmentv1 "github.com/userreksai/ecmdb-main/api/proto/gen/eiam/department/v1"
	userv1 "github.com/userreksai/ecmdb-main/api/proto/gen/eiam/user/v1"
	"github.com/userreksai/ecmdb-main/internal/department"
	"github.com/userreksai/ecmdb-main/internal/user"
)

const defaultMemberLimit = 500

// DepartmentServer 实现 eiam.department.v1.DepartmentService。
type DepartmentServer struct {
	departmentv1.UnimplementedDepartmentServiceServer
	users       user.Service
	departments department.Service
}

// NewDepartmentServer 创建部门兼容服务。
func NewDepartmentServer(users user.Service, departments department.Service) *DepartmentServer {
	return &DepartmentServer{users: users, departments: departments}
}

// QueryByUserId 返回用户所在部门（含部门负责人与分管领导），供 eflow 解析"部门负责人/分管领导"审批人。
func (s *DepartmentServer) QueryByUserId(ctx context.Context, req *departmentv1.QueryByUserIdReq) (*departmentv1.QueryDepartmentsResp, error) {
	u, err := s.users.FindById(ctx, req.GetUserId())
	if err != nil || u.Id <= 0 {
		return &departmentv1.QueryDepartmentsResp{ErrorCode: departmentv1.ErrorCode_INVALID_PARAMETER,
			ErrorMessage: "用户不存在"}, nil
	}
	if u.DepartmentId <= 0 {
		return &departmentv1.QueryDepartmentsResp{}, nil
	}
	dept, err := s.departments.FindById(ctx, u.DepartmentId)
	if err != nil || dept.Id <= 0 {
		return &departmentv1.QueryDepartmentsResp{ErrorCode: departmentv1.ErrorCode_DEPARTMENT_NOT_FOUND,
			ErrorMessage: "用户所在部门不存在"}, nil
	}
	return &departmentv1.QueryDepartmentsResp{Departments: []*departmentv1.Department{{
		Id: dept.Id, ParentId: dept.Pid, Name: dept.Name, Sort: dept.Sort,
		Leaders: dept.Leaders, MainLeader: dept.MainLeader,
	}}}, nil
}

// ListMembers 分页返回部门成员。
func (s *DepartmentServer) ListMembers(ctx context.Context, req *departmentv1.ListMembersReq) (*departmentv1.ListMembersResp, error) {
	limit := req.GetLimit()
	if limit <= 0 {
		limit = defaultMemberLimit
	}
	users, total, err := s.users.FindByDepartmentId(ctx, req.GetOffset(), limit, req.GetDepartmentId())
	if err != nil {
		return nil, err
	}
	members := make([]*userv1.User, 0, len(users))
	for _, u := range users {
		members = append(members, toProtoUser(u))
	}
	return &departmentv1.ListMembersResp{Members: members, Total: total}, nil
}
