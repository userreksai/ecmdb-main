//go:build wireinject

package department

import (
	"github.com/google/wire"
	"github.com/userreksai/ecmdb-main/internal/department/internal/repository"
	"github.com/userreksai/ecmdb-main/internal/department/internal/repository/dao"
	"github.com/userreksai/ecmdb-main/internal/department/internal/service"
	"github.com/userreksai/ecmdb-main/internal/department/internal/web"
	"github.com/userreksai/ecmdb-main/pkg/mongox"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	service.NewService,
	repository.NewDepartmentRepository,
	dao.NewDepartmentDAO,
)

func InitModule(db *mongox.Mongo) (*Module, error) {
	wire.Build(
		ProviderSet,
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}
