//go:build wireinject

package resource

import (
	"context"
	"sync"

	"github.com/ecodeclub/mq-api"
	"github.com/google/wire"
	"github.com/userreksai/ecmdb-main/internal/attribute"
	"github.com/userreksai/ecmdb-main/internal/operationlog"
	"github.com/userreksai/ecmdb-main/internal/policy"
	"github.com/userreksai/ecmdb-main/internal/relation"
	"github.com/userreksai/ecmdb-main/internal/resource/internal/event"
	"github.com/userreksai/ecmdb-main/internal/resource/internal/repository"
	"github.com/userreksai/ecmdb-main/internal/resource/internal/repository/dao"
	"github.com/userreksai/ecmdb-main/internal/resource/internal/service"
	"github.com/userreksai/ecmdb-main/internal/resource/internal/web"
	"github.com/userreksai/ecmdb-main/internal/role"
	"github.com/userreksai/ecmdb-main/pkg/cryptox"
	"github.com/userreksai/ecmdb-main/pkg/mongox"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	repository.NewResourceRepository)

func InitModule(db *mongox.Mongo, attributeModule *attribute.Module, relationModule *relation.Module,
	q mq.MQ, crypto *cryptox.CryptoRegistry, roleModule *role.Module, policyModule *policy.Module,
	operationLogModule *operationlog.Module) (*Module, error) {
	wire.Build(
		ProviderSet,
		NewEncryptedService,
		InitResourceDAO,
		NewService,
		InitCrypto,
		initConsumer,
		wire.FieldsOf(new(*attribute.Module), "Svc"),
		wire.FieldsOf(new(*relation.Module), "RRSvc"),
		wire.FieldsOf(new(*role.Module), "Svc"),
		wire.FieldsOf(new(*policy.Module), "Svc"),
		wire.FieldsOf(new(*operationlog.Module), "Svc"),
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

func InitResourceDAO(db *mongox.Mongo) dao.ResourceDAO {
	InitCollectionOnce(db)
	return dao.NewResourceDAO(db)
}

func NewService(repo repository.ResourceRepository) Service {
	return service.NewService(repo)
}

func NewEncryptedService(baseSvc service.Service, attrSvc attribute.Service,
	cryptox cryptox.Crypto) EncryptedSvc {
	return service.NewEncryptedResourceService(baseSvc, attrSvc, cryptox)
}

func InitCrypto(reg *cryptox.CryptoRegistry) cryptox.Crypto {
	return reg.Resource
}

func initConsumer(q mq.MQ, svc service.EncryptedSvc, cryptox cryptox.Crypto) *event.FieldSecureAttrChangeConsumer {
	consumer, err := event.NewFieldSecureAttrChangeConsumer(q, svc, 20, cryptox)
	if err != nil {
		panic(err)
	}

	consumer.Start(context.Background())
	return consumer
}
