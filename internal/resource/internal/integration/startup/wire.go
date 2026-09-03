//go:build wireinject

package startup

import (
	"github.com/google/wire"
	"github.com/userreksai/ecmdb-main/internal/attribute"
	"github.com/userreksai/ecmdb-main/internal/operationlog"
	"github.com/userreksai/ecmdb-main/internal/policy"
	"github.com/userreksai/ecmdb-main/internal/relation"
	"github.com/userreksai/ecmdb-main/internal/resource"
	"github.com/userreksai/ecmdb-main/internal/role"
)

func InitHandler(attributeModule *attribute.Module, relationModule *relation.Module,
	roleModule *role.Module, policyModule *policy.Module) (*resource.Handler, error) {
	wire.Build(InitMongoDB,
		InitMQ,
		InitCryptoRegistry,
		InitOperationLogModule,
		resource.InitModule,
		wire.FieldsOf(new(*resource.Module), "Hdl"),
	)
	return new(resource.Handler), nil
}

func InitOperationLogModule() *operationlog.Module { return &operationlog.Module{} }
