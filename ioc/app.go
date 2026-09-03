package ioc

import (
	"github.com/gotomicro/ego/server/egin"
	"github.com/gotomicro/ego/task/ecron"
	"github.com/userreksai/ecmdb-main/internal/endpoint"
	"github.com/userreksai/ecmdb-main/internal/event/service/easyflow"
	grpcpkg "github.com/userreksai/ecmdb-task/pkg/grpc"
)

type App struct {
	Web    *egin.Component
	Server *grpcpkg.Server
	Event  *easyflow.ProcessEvent
	Jobs   []*ecron.Component
	Svc    endpoint.Service
}
