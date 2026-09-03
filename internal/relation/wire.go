//go:build wireinject

package relation

import (
	"sync"

	"github.com/google/wire"
	"github.com/userreksai/ecmdb-main/internal/policy"
	"github.com/userreksai/ecmdb-main/internal/relation/internal/repository"
	"github.com/userreksai/ecmdb-main/internal/relation/internal/repository/dao"
	"github.com/userreksai/ecmdb-main/internal/relation/internal/service"
	"github.com/userreksai/ecmdb-main/internal/relation/internal/web"
	"github.com/userreksai/ecmdb-main/internal/role"
	"github.com/userreksai/ecmdb-main/pkg/mongox"
)

var ProviderSet = wire.NewSet(
	web.NewRelationResourceHandler,
	web.NewRelationModelHandler,
	web.NewRelationTypeHandler,
	service.NewRelationTypeService,
	repository.NewRelationTypeRepository,
	repository.NewRelationModelRepository,
	repository.NewRelationResourceRepository,
	initRmDAO,
	intRrDAO,
)

func InitModule(db *mongox.Mongo, roleModule *role.Module, policyModule *policy.Module) (*Module, error) {
	wire.Build(
		ProviderSet,
		InitRelationTypeDAO,
		InitRRService,
		InitRMService,
		wire.FieldsOf(new(*role.Module), "Svc"),
		wire.FieldsOf(new(*policy.Module), "Svc"),
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}

var (
	rmDaoOnce = sync.Once{}
	rrDaoOnce = sync.Once{}
	rmd       dao.RelationModelDAO
	rrd       dao.RelationResourceDAO
)

func initRmDAO(db *mongox.Mongo) dao.RelationModelDAO {
	rmDaoOnce.Do(func() {
		rmd = dao.NewRelationModelDAO(db)
	})
	return rmd
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

func InitRelationTypeDAO(db *mongox.Mongo) dao.RelationTypeDAO {
	InitCollectionOnce(db)
	return dao.NewRelationTypeDAO(db)
}

func InitRMService(db *mongox.Mongo) RMSvc {
	wire.Build(
		initRmDAO,
		intRrDAO,
		repository.NewRelationModelRepository,
		repository.NewRelationResourceRepository,
		service.NewRelationModelService,
	)
	return nil
}

func intRrDAO(db *mongox.Mongo) dao.RelationResourceDAO {
	rrDaoOnce.Do(func() {
		rrd = dao.NewRelationResourceDAO(db)
	})
	return rrd
}

func InitRRService(db *mongox.Mongo) RRSvc {
	wire.Build(
		intRrDAO,
		repository.NewRelationResourceRepository,
		service.NewRelationResourceService,
	)
	return nil
}
