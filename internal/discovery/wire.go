//go:build wireinject

package discovery

import (
	"github.com/google/wire"
	"github.com/userreksai/ecmdb-main/internal/discovery/internal/repository"
	"github.com/userreksai/ecmdb-main/internal/discovery/internal/repository/dao"
	"github.com/userreksai/ecmdb-main/internal/discovery/internal/service"
	"github.com/userreksai/ecmdb-main/internal/discovery/internal/web"
	"github.com/userreksai/ecmdb-main/pkg/mongox"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	service.NewService,
	repository.NewDiscoveryRepository,
	dao.NewDiscoveryDAO,
)

func InitModule(db *mongox.Mongo) (*Module, error) {
	wire.Build(
		ProviderSet,
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}
