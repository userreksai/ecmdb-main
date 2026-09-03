//go:build wireinject

package role

import (
	"sync"

	"github.com/google/wire"
	"github.com/userreksai/ecmdb-main/internal/role/internal/repository"
	"github.com/userreksai/ecmdb-main/internal/role/internal/repository/dao"
	"github.com/userreksai/ecmdb-main/internal/role/internal/service"
	"github.com/userreksai/ecmdb-main/internal/role/internal/web"
	"github.com/userreksai/ecmdb-main/pkg/mongox"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	service.NewService,
	repository.NewRoleRepository,
)

func InitModule(db *mongox.Mongo) (*Module, error) {
	wire.Build(
		ProviderSet,
		InitRoleDAO,
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

func InitRoleDAO(db *mongox.Mongo) dao.RoleDAO {
	InitCollectionOnce(db)
	return dao.NewRoleDAO(db)
}
