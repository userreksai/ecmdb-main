//go:build wireinject

package startup

import (
	"github.com/google/wire"
	"github.com/userreksai/ecmdb-main/internal/policy"
	"github.com/userreksai/ecmdb-main/internal/relation"
	"github.com/userreksai/ecmdb-main/internal/role"
)

func InitRMHandler(roleModule *role.Module, policyModule *policy.Module) (*relation.RMHandler, error) {
	wire.Build(InitMongoDB,
		relation.InitModule,
		wire.FieldsOf(new(*relation.Module), "RMHdl"),
	)
	return new(relation.RMHandler), nil
}

func InitRRHandler(roleModule *role.Module, policyModule *policy.Module) (*relation.RRHandler, error) {
	wire.Build(InitMongoDB,
		relation.InitModule,
		wire.FieldsOf(new(*relation.Module), "RRHdl"),
	)
	return new(relation.RRHandler), nil
}

func InitRRSvc() relation.RRSvc {
	wire.Build(InitMongoDB, relation.InitRRService)
	return nil
}

func InitRMSvc() relation.RMSvc {
	wire.Build(InitMongoDB, relation.InitRMService)
	return nil
}
