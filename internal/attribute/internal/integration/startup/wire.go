//go:build wireinject

package startup

import (
	"github.com/google/wire"
	"github.com/userreksai/ecmdb-main/internal/attribute"
)

func InitHandler() (*attribute.Handler, error) {
	wire.Build(
		InitMongoDB,
		InitMQ,
		attribute.InitModule,
		wire.FieldsOf(new(*attribute.Module), "Hdl"),
	)
	return new(attribute.Handler), nil
}
