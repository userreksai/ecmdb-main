package rota

import (
	"github.com/userreksai/ecmdb-main/internal/rota/internal/grpc"
	"github.com/userreksai/ecmdb-main/internal/rota/internal/service"
	"github.com/userreksai/ecmdb-main/internal/rota/internal/web"
)

type Handler = web.Handler

type Service = service.Service

type RpcServer = grpc.RotaServer
