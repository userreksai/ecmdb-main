//go:build wireinject

package startup

import (
	"github.com/Duke1616/ecmdb/internal/attribute"
	"github.com/Duke1616/ecmdb/internal/policy"
	"github.com/Duke1616/ecmdb/internal/relation"
	"github.com/Duke1616/ecmdb/internal/resource"
	"github.com/Duke1616/ecmdb/internal/role"
	"github.com/google/wire"
)

func InitHandler(attributeModule *attribute.Module, relationModule *relation.Module,
	roleModule *role.Module, policyModule *policy.Module) (*resource.Handler, error) {
	wire.Build(InitMongoDB,
		InitMQ,
		InitCryptoRegistry,
		resource.InitModule,
		wire.FieldsOf(new(*resource.Module), "Hdl"),
	)
	return new(resource.Handler), nil
}
