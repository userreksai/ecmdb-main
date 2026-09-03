package endpoint

import (
	"github.com/userreksai/ecmdb-main/internal/endpoint/internal/domain"
	"github.com/userreksai/ecmdb-main/internal/endpoint/internal/grpc"
	"github.com/userreksai/ecmdb-main/internal/endpoint/internal/service"
	"github.com/userreksai/ecmdb-main/internal/endpoint/internal/web"
)

type Handler = web.Handler

type Service = service.Service

type Endpoint = domain.Endpoint

type RpcServer = grpc.EndpointServer
