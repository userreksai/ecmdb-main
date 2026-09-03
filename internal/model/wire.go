//go:build wireinject

package model

import (
	"sync"

	"github.com/google/wire"
	"github.com/userreksai/ecmdb-main/internal/attribute"
	"github.com/userreksai/ecmdb-main/internal/model/internal/repository"
	"github.com/userreksai/ecmdb-main/internal/model/internal/repository/dao"
	"github.com/userreksai/ecmdb-main/internal/model/internal/service"
	"github.com/userreksai/ecmdb-main/internal/model/internal/web"
	"github.com/userreksai/ecmdb-main/internal/policy"
	"github.com/userreksai/ecmdb-main/internal/relation"
	"github.com/userreksai/ecmdb-main/internal/resource"
	"github.com/userreksai/ecmdb-main/internal/role"
	"github.com/userreksai/ecmdb-main/pkg/mongox"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	initMGProvider,
	initModelProvider)

func InitModule(db *mongox.Mongo, rmModule *relation.Module, attrModule *attribute.Module, resourceSvc *resource.Module,
	roleModule *role.Module, policyModule *policy.Module) (*Module, error) {
	wire.Build(
		ProviderSet,
		InitModelDAO,
		wire.FieldsOf(new(*relation.Module), "RMSvc"),
		wire.FieldsOf(new(*attribute.Module), "Svc"),
		wire.FieldsOf(new(*resource.Module), "EncryptedSvc"),
		wire.FieldsOf(new(*role.Module), "Svc"),
		wire.FieldsOf(new(*policy.Module), "Svc"),
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

func InitModelDAO(db *mongox.Mongo) dao.ModelDAO {
	InitCollectionOnce(db)
	return dao.NewModelDAO(db)
}

var initMGProvider = wire.NewSet(
	service.NewMGService,
	repository.NewMGRepository,
	dao.NewModelGroupDAO,
)

var initModelProvider = wire.NewSet(
	service.NewModelService,
	repository.NewModelRepository)
