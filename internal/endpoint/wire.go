//go:build wireinject

package endpoint

import (
	"sync"

	"github.com/google/wire"
	"github.com/userreksai/ecmdb-main/internal/endpoint/internal/grpc"
	"github.com/userreksai/ecmdb-main/internal/endpoint/internal/repository"
	"github.com/userreksai/ecmdb-main/internal/endpoint/internal/repository/dao"
	"github.com/userreksai/ecmdb-main/internal/endpoint/internal/service"
	"github.com/userreksai/ecmdb-main/internal/endpoint/internal/web"
	"github.com/userreksai/ecmdb-main/pkg/mongox"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	service.NewService,
	repository.NewEndpointRepository,
	grpc.NewEndpointServer,
)

func InitModule(db *mongox.Mongo) (*Module, error) {
	wire.Build(
		ProviderSet,
		InitEndpointDAO,
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}

var daoOnce = sync.Once{}

func InitCollectionOnce(db *mongox.Mongo) {
	daoOnce.Do(func() {
		err := dao.InitIndexes(db)
		if err != nil {
			panic(err)
		}
	})
}

func InitEndpointDAO(db *mongox.Mongo) dao.EndpointDAO {
	InitCollectionOnce(db)
	return dao.NewEndpointDAO(db)
}
