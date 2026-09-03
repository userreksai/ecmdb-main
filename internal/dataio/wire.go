//go:build wireinject

package dataio

import (
	"github.com/google/wire"
	"github.com/userreksai/ecmdb-main/internal/attribute"
	"github.com/userreksai/ecmdb-main/internal/dataio/internal/service"
	"github.com/userreksai/ecmdb-main/internal/dataio/internal/web"
	"github.com/userreksai/ecmdb-main/internal/model"
	"github.com/userreksai/ecmdb-main/internal/operationlog"
	"github.com/userreksai/ecmdb-main/internal/policy"
	"github.com/userreksai/ecmdb-main/internal/relation"
	"github.com/userreksai/ecmdb-main/internal/resource"
	"github.com/userreksai/ecmdb-main/internal/role"
	"github.com/userreksai/ecmdb-main/pkg/storage"
)

// ProviderSet 数据交换模块依赖集合
var ProviderSet = wire.NewSet(
	web.NewHandler,
	service.NewDataIOService,
)

func InitModule(attributeModule *attribute.Module, resourceModule *resource.Module, storage *storage.S3Storage,
	modelModule *model.Module, relationModule *relation.Module, roleModule *role.Module,
	policyModule *policy.Module, operationLogModule *operationlog.Module) (*Module, error) {
	wire.Build(
		ProviderSet,
		wire.FieldsOf(new(*attribute.Module), "Svc"),
		wire.FieldsOf(new(*resource.Module), "EncryptedSvc"),
		wire.FieldsOf(new(*model.Module), "Svc"),
		wire.FieldsOf(new(*relation.Module), "RMSvc", "RRSvc"),
		wire.FieldsOf(new(*role.Module), "Svc"),
		wire.FieldsOf(new(*policy.Module), "Svc"),
		wire.FieldsOf(new(*operationlog.Module), "Svc"),
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}
