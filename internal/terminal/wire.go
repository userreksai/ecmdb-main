//go:build wireinject

package terminal

import (
	"github.com/google/wire"
	"github.com/userreksai/ecmdb-main/internal/attribute"
	"github.com/userreksai/ecmdb-main/internal/policy"
	"github.com/userreksai/ecmdb-main/internal/relation"
	"github.com/userreksai/ecmdb-main/internal/resource"
	"github.com/userreksai/ecmdb-main/internal/role"
	"github.com/userreksai/ecmdb-main/internal/terminal/internal/web"
)

func InitModule(relationModule *relation.Module, resourceModule *resource.Module, attributeModule *attribute.Module,
	roleModule *role.Module, policyModule *policy.Module) (*web.Handler, error) {
	wire.Build(
		web.NewHandler,
		wire.FieldsOf(new(*relation.Module), "RRSvc"),
		wire.FieldsOf(new(*resource.Module), "EncryptedSvc"),
		wire.FieldsOf(new(*attribute.Module), "Svc"),
		wire.FieldsOf(new(*role.Module), "Svc"),
		wire.FieldsOf(new(*policy.Module), "Svc"),
	)
	return new(web.Handler), nil
}
