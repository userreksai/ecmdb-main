//go:build wireinject

package strategy

import (
	"github.com/google/wire"
	"github.com/userreksai/ecmdb-main/internal/strategy/internal/service"
	"github.com/userreksai/ecmdb-main/internal/strategy/internal/web"
	"github.com/userreksai/ecmdb-main/internal/template"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	service.NewService,
)

func InitModule(templateModule *template.Module) (*Module, error) {
	wire.Build(
		ProviderSet,
		wire.FieldsOf(new(*template.Module), "Svc"),
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}
