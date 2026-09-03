package user

import "github.com/userreksai/ecmdb-main/internal/user/internal/web"

type Module struct {
	Hdl       *web.Handler
	Svc       Service
	RpcServer *RpcServer
}
