package user

import (
	"github.com/userreksai/ecmdb-main/internal/user/internal/domain"
	"github.com/userreksai/ecmdb-main/internal/user/internal/grpc"
	"github.com/userreksai/ecmdb-main/internal/user/internal/service"
	"github.com/userreksai/ecmdb-main/internal/user/internal/web"
)

type Handler = web.Handler

type Service = service.Service

type User = domain.User

type FeishuInfo = domain.FeishuInfo

type WechatInfo = domain.WechatInfo

type RpcServer = grpc.UserServer
