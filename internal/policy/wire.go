//go:build wireinject

package policy

import (
	"github.com/casbin/casbin/v2"
	"github.com/ecodeclub/ginx/session"
	"github.com/google/wire"
	"github.com/userreksai/ecmdb-main/internal/pkg/servicetoken"
	"github.com/userreksai/ecmdb-main/internal/policy/internal/grpc"
	"github.com/userreksai/ecmdb-main/internal/policy/internal/service"
	"github.com/userreksai/ecmdb-main/internal/policy/internal/web"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	service.NewService,
	grpc.NewPolicyServer,
)

func InitModule(enforcer *casbin.SyncedEnforcer, sp session.Provider, tokenMgr *servicetoken.Manager) (*Module, error) {
	wire.Build(
		ProviderSet,
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}
