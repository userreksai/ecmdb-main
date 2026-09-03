//go:build wireinject

package codebook

import (
	"github.com/google/wire"
	repository "github.com/userreksai/ecmdb-main/internal/codebook/internal/repository"
	"github.com/userreksai/ecmdb-main/internal/codebook/internal/repository/dao"
	"github.com/userreksai/ecmdb-main/internal/codebook/internal/service"
	"github.com/userreksai/ecmdb-main/internal/codebook/internal/web"
	"github.com/userreksai/ecmdb-main/pkg/mongox"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	service.NewService,
	repository.NewCodebookRepository,
	dao.NewCodebookDAO)

func InitModule(db *mongox.Mongo) (*Module, error) {
	wire.Build(
		ProviderSet,
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}
