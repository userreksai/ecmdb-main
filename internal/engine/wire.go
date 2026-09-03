//go:build wireinject

package engine

import (
	"github.com/google/wire"
	"github.com/userreksai/ecmdb-main/internal/engine/internal/repository"
	"github.com/userreksai/ecmdb-main/internal/engine/internal/repository/dao"
	"github.com/userreksai/ecmdb-main/internal/engine/internal/service"
	"github.com/userreksai/ecmdb-main/internal/engine/internal/web"
	"gorm.io/gorm"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	service.NewService,
	repository.NewProcessEngineRepository,
	dao.NewProcessEngineDAO,
)

func InitModule(db *gorm.DB) (*Module, error) {
	wire.Build(
		ProviderSet,
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}
