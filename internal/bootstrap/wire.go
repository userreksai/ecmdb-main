//go:build wireinject

package bootstrap

import (
	"github.com/google/wire"
	"github.com/userreksai/ecmdb-main/internal/attribute"
	"github.com/userreksai/ecmdb-main/internal/bootstrap/internal/service"
	"github.com/userreksai/ecmdb-main/internal/model"
	"github.com/userreksai/ecmdb-main/internal/relation"
)

var ProviderSet = wire.NewSet(
	service.NewLoader,
)

func InitModule(
	modelModule *model.Module,
	attributeModule *attribute.Module,
	relationModule *relation.Module,
) (*Module, error) {
	wire.Build(
		ProviderSet,
		// 从已有的模块中提取服务
		wire.FieldsOf(new(*model.Module), "Svc", "MGSvc"),
		wire.FieldsOf(new(*attribute.Module), "Svc"),
		wire.FieldsOf(new(*relation.Module), "RMSvc", "RTSvc"),
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}
