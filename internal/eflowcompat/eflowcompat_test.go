package eflowcompat

import (
	"context"
	"errors"
	"testing"

	runnerv1 "github.com/userreksai/ecmdb-main/api/proto/gen/ecmdb/runner/v1"
	departmentv1 "github.com/userreksai/ecmdb-main/api/proto/gen/eiam/department/v1"
	userv1 "github.com/userreksai/ecmdb-main/api/proto/gen/eiam/user/v1"
	"github.com/userreksai/ecmdb-main/internal/codebook"
	"github.com/userreksai/ecmdb-main/internal/department"
	"github.com/userreksai/ecmdb-main/internal/runner"
	"github.com/userreksai/ecmdb-main/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var people = map[int64]user.User{
	1: {Id: 1, Username: "admin", RoleCodes: []string{"admin"}, DepartmentId: 5},
	2: {Id: 2, Username: "alice", DisplayName: "Alice", DepartmentId: 5,
		FeishuInfo: user.FeishuInfo{UserId: "ou_a"}, WechatInfo: user.WechatInfo{UserId: "wx_a"}},
	3: {Id: 3, Username: "bob"},
}

type fakeUsers struct{ user.Service }

func (fakeUsers) FindById(_ context.Context, id int64) (user.User, error) {
	if u, ok := people[id]; ok {
		return u, nil
	}
	return user.User{}, errors.New("not found")
}
func (fakeUsers) FindByUsername(_ context.Context, name string) (user.User, error) {
	for _, u := range people {
		if u.Username == name {
			return u, nil
		}
	}
	return user.User{}, errors.New("not found")
}
func (fakeUsers) FindByDepartmentId(_ context.Context, _, _ int64, dept int64) ([]user.User, int64, error) {
	var out []user.User
	for _, u := range people {
		if u.DepartmentId == dept {
			out = append(out, u)
		}
	}
	return out, int64(len(out)), nil
}

type fakeDepartments struct{ department.Service }

func (fakeDepartments) FindById(_ context.Context, id int64) (department.Department, error) {
	if id != 5 {
		return department.Department{}, errors.New("not found")
	}
	return department.Department{Id: 5, Name: "运维部", Leaders: []string{"alice"}, MainLeader: "admin"}, nil
}

func TestUserServer(t *testing.T) {
	srv := NewUserServer(fakeUsers{})
	testCases := []struct {
		name      string
		id        int64
		wantCode  userv1.ErrorCode
		wantAdmin bool
		wantName  string
	}{
		{name: "管理员角色映射为 is_admin", id: 1, wantAdmin: true, wantName: "admin"},
		{name: "普通用户携带显示名与三方账号", id: 2, wantName: "Alice"},
		{name: "显示名为空时回退为用户名", id: 3, wantName: "bob"},
		{name: "用户不存在返回业务错误码", id: 9, wantCode: userv1.ErrorCode_USER_NOT_FOUND},
		{name: "非法 ID", id: 0, wantCode: userv1.ErrorCode_INVALID_PARAMETER},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := srv.QueryById(context.Background(), &userv1.QueryByIdReq{Id: tc.id})
			require.NoError(t, err)
			assert.Equal(t, tc.wantCode, resp.GetErrorCode())
			if tc.wantCode != userv1.ErrorCode_ERROR_CODE_UNSPECIFIED {
				return
			}
			assert.Equal(t, tc.wantAdmin, resp.GetUser().GetIsAdmin())
			assert.Equal(t, tc.wantName, resp.GetUser().GetDisplayName())
			if tc.id == 2 {
				assert.Equal(t, "ou_a", resp.GetUser().GetLarkUserId())
				assert.Equal(t, int64(5), resp.GetUser().GetDepartmentId())
			}
		})
	}
}

func TestDepartmentServer(t *testing.T) {
	srv := NewDepartmentServer(fakeUsers{}, fakeDepartments{})
	ctx := context.Background()

	resp, err := srv.QueryByUserId(ctx, &departmentv1.QueryByUserIdReq{UserId: 2})
	require.NoError(t, err)
	require.Len(t, resp.GetDepartments(), 1)
	assert.Equal(t, []string{"alice"}, resp.GetDepartments()[0].GetLeaders())
	assert.Equal(t, "admin", resp.GetDepartments()[0].GetMainLeader())

	resp, err = srv.QueryByUserId(ctx, &departmentv1.QueryByUserIdReq{UserId: 3})
	require.NoError(t, err)
	assert.Empty(t, resp.GetDepartments(), "未分配部门的用户返回空列表")

	members, err := srv.ListMembers(ctx, &departmentv1.ListMembersReq{DepartmentId: 5})
	require.NoError(t, err)
	assert.Equal(t, int64(2), members.GetTotal())
}

type fakeRunners struct{ runner.Service }

func (fakeRunners) FindById(_ context.Context, id int64) (runner.Runner, error) {
	switch id {
	case 7:
		return runner.Runner{Id: 7, Name: "deploy", CodebookUid: "cb-deploy", Kind: "GRPC", Target: "execute",
			Handler: "shell", Variables: []runner.Variables{
				{Key: "TOKEN", Value: "enc(s3cr3t)", Secret: true}, {Key: "ENV", Value: "prod"}}}, nil
	case 8:
		return runner.Runner{Id: 8, CodebookUid: "missing"}, nil
	}
	return runner.Runner{}, errors.New("not found")
}

type fakeCodebooks struct{ codebook.Service }

func (fakeCodebooks) FindByUid(_ context.Context, uid string) (codebook.Codebook, error) {
	if uid != "cb-deploy" {
		return codebook.Codebook{}, errors.New("not found")
	}
	return codebook.Codebook{Id: 42, Identifier: "cb-deploy", Code: "echo hi", Language: "shell"}, nil
}

type fakeCrypto struct{}

func (fakeCrypto) Encrypt(s string) (string, error) { return "enc(" + s + ")", nil }
func (fakeCrypto) Decrypt(s string) (string, error) {
	if len(s) > 5 && s[:4] == "enc(" {
		return s[4 : len(s)-1], nil
	}
	return "", errors.New("bad cipher")
}

func TestRunnerServer(t *testing.T) {
	srv := NewRunnerServer(fakeRunners{}, fakeCodebooks{}, fakeCrypto{})
	ctx := context.Background()

	t.Run("查询执行单元换算代码模板数字 ID 且不返回变量", func(t *testing.T) {
		resp, err := srv.FindRunnerByID(ctx, &runnerv1.FindRunnerByIDRequest{Id: 7})
		require.NoError(t, err)
		assert.Equal(t, int64(42), resp.GetRunner().GetCodebookId())
		assert.Equal(t, "GRPC", resp.GetRunner().GetKind())
	})

	t.Run("解析执行信息返回脚本与解密后的变量", func(t *testing.T) {
		resp, err := srv.ResolveExecution(ctx, &runnerv1.ResolveExecutionRequest{RunnerId: 7})
		require.NoError(t, err)
		exec := resp.GetExecution()
		assert.Equal(t, "echo hi", exec.GetCode())
		got := map[string]string{}
		for _, v := range exec.GetVariables() {
			got[v.GetKey()] = v.GetValue()
		}
		assert.Equal(t, "s3cr3t", got["TOKEN"])
		assert.Equal(t, "prod", got["ENV"])
	})

	testCases := []struct {
		name string
		id   int64
		code codes.Code
	}{
		{name: "执行单元不存在", id: 99, code: codes.NotFound},
		{name: "绑定的代码模板不存在", id: 8, code: codes.FailedPrecondition},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := srv.ResolveExecution(ctx, &runnerv1.ResolveExecutionRequest{RunnerId: tc.id})
			assert.Equal(t, tc.code, status.Code(err))
		})
	}
}
