package eflowcompat

import (
	runnerv1 "github.com/userreksai/ecmdb-main/api/proto/gen/ecmdb/runner/v1"
	departmentv1 "github.com/userreksai/ecmdb-main/api/proto/gen/eiam/department/v1"
	userv1 "github.com/userreksai/ecmdb-main/api/proto/gen/eiam/user/v1"
	"github.com/userreksai/ecmdb-main/internal/codebook"
	"github.com/userreksai/ecmdb-main/internal/department"
	"github.com/userreksai/ecmdb-main/internal/runner"
	"github.com/userreksai/ecmdb-main/internal/user"
	"github.com/userreksai/ecmdb-main/pkg/cryptox"
	"google.golang.org/grpc"
)

// RpcServer 汇总 eflow / ecmdb-task 需要的兼容 gRPC 服务。
type RpcServer struct {
	user       *UserServer
	department *DepartmentServer
	runner     *RunnerServer
}

// NewRpcServer 组装兼容服务；crypto 取 CryptoRegistry.Runner，与执行单元变量的加密器一致。
func NewRpcServer(users user.Service, departments department.Service, runners runner.Service,
	codebooks codebook.Service, crypto *cryptox.CryptoRegistry) *RpcServer {
	return &RpcServer{
		user:       NewUserServer(users),
		department: NewDepartmentServer(users, departments),
		runner:     NewRunnerServer(runners, codebooks, crypto.Runner),
	}
}

// Register 注册到 ECMDB 的 gRPC 服务端。
func (s *RpcServer) Register(server grpc.ServiceRegistrar) {
	userv1.RegisterUserServiceServer(server, s.user)
	departmentv1.RegisterDepartmentServiceServer(server, s.department)
	runnerv1.RegisterRunnerServiceServer(server, s.runner)
}
