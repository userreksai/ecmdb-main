//go:build wireinject

package permission

import (
	"context"

	"github.com/ecodeclub/mq-api"
	"github.com/google/wire"
	"github.com/userreksai/ecmdb-main/internal/menu"
	"github.com/userreksai/ecmdb-main/internal/model"
	"github.com/userreksai/ecmdb-main/internal/permission/internal/event"
	"github.com/userreksai/ecmdb-main/internal/permission/internal/service"
	"github.com/userreksai/ecmdb-main/internal/permission/internal/web"
	"github.com/userreksai/ecmdb-main/internal/policy"
	"github.com/userreksai/ecmdb-main/internal/role"
	"github.com/userreksai/ecmdb-main/pkg/mongox"
)

func InitModule(db *mongox.Mongo, q mq.MQ, roleModule *role.Module, menuModule *menu.Module,
	policyModule *policy.Module, modelModule *model.Module) (*Module, error) {
	wire.Build(
		web.NewHandler,
		service.NewService,
		InitMenuChangeEventConsumer,
		wire.Struct(new(Module), "*"),
		wire.FieldsOf(new(*menu.Module), "Svc"),
		wire.FieldsOf(new(*role.Module), "Svc"),
		wire.FieldsOf(new(*policy.Module), "Svc"),
		wire.FieldsOf(new(*model.Module), "Svc", "MGSvc"),
	)
	return new(Module), nil
}

func InitMenuChangeEventConsumer(q mq.MQ, svc service.Service) *event.MenuChangeEventConsumer {
	c, err := event.NewMenuChangeEventConsumer(q, svc)
	if err != nil {
		return nil
	}

	c.Start(context.Background())
	return c
}
