//go:build wireinject

package startup

import (
	"github.com/google/wire"
	"github.com/userreksai/ecmdb-main/internal/attribute"
	"github.com/userreksai/ecmdb-main/internal/model"
	"github.com/userreksai/ecmdb-main/internal/policy"
	"github.com/userreksai/ecmdb-main/internal/relation"
	"github.com/userreksai/ecmdb-main/internal/resource"
	"github.com/userreksai/ecmdb-main/internal/role"
	"github.com/userreksai/ecmdb-main/internal/test/ioc"
)

func InitHandler(rmModule *relation.Module, attrModule *attribute.Module, resourceModule *resource.Module,
	roleModule *role.Module, policyModule *policy.Module) (*model.Handler, error) {
	wire.Build(ioc.InitMongoDB,
		model.InitModule,
		wire.FieldsOf(new(*model.Module), "Hdl"),
	)
	return new(model.Handler), nil
}
