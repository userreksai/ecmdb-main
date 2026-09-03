//go:build wireinject

package ioc

import (
	"github.com/google/wire"
	"github.com/userreksai/ecmdb-main/internal/attribute"
	"github.com/userreksai/ecmdb-main/internal/model"
	"github.com/userreksai/ecmdb-main/internal/resource"
	"github.com/userreksai/ecmdb-main/ioc"
)

var BaseSet = wire.NewSet(ioc.InitMongoDB, ioc.InitMQ, ioc.InitModuleCrypto)

func InitApp() (*App, error) {
	wire.Build(wire.Struct(new(App), "*"),
		BaseSet,
		model.InitService,
		attribute.InitModule,
		wire.FieldsOf(new(*attribute.Module), "Svc"),
		resource.InitEncryptedService,
	)
	return new(App), nil
}
