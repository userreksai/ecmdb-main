//go:build wireinject

package menu

import (
	"github.com/ecodeclub/mq-api"
	"github.com/google/wire"
	"github.com/userreksai/ecmdb-main/internal/menu/internal/event"
	"github.com/userreksai/ecmdb-main/internal/menu/internal/repository"
	"github.com/userreksai/ecmdb-main/internal/menu/internal/repository/dao"
	"github.com/userreksai/ecmdb-main/internal/menu/internal/service"
	"github.com/userreksai/ecmdb-main/internal/menu/internal/web"
	"github.com/userreksai/ecmdb-main/pkg/mongox"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	service.NewService,
	repository.NewMenuRepository,
	dao.NewMenuDAO,
)

func InitModule(q mq.MQ, db *mongox.Mongo) (*Module, error) {
	wire.Build(
		event.NewMenuChangeEventProducer,
		ProviderSet,
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}
