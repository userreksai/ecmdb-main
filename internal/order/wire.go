//go:build wireinject

package order

import (
	"context"

	"github.com/ecodeclub/mq-api"
	"github.com/google/wire"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	"github.com/userreksai/ecmdb-main/internal/engine"
	"github.com/userreksai/ecmdb-main/internal/order/internal/event"
	"github.com/userreksai/ecmdb-main/internal/order/internal/event/consumer"
	"github.com/userreksai/ecmdb-main/internal/order/internal/grpc"
	"github.com/userreksai/ecmdb-main/internal/order/internal/repository"
	"github.com/userreksai/ecmdb-main/internal/order/internal/repository/dao"
	"github.com/userreksai/ecmdb-main/internal/order/internal/service"
	"github.com/userreksai/ecmdb-main/internal/order/internal/web"
	"github.com/userreksai/ecmdb-main/internal/pkg/notification/sender"
	"github.com/userreksai/ecmdb-main/internal/template"
	"github.com/userreksai/ecmdb-main/internal/user"
	"github.com/userreksai/ecmdb-main/internal/workflow"
	"github.com/userreksai/ecmdb-main/pkg/mongox"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	service.NewService,
	service.NewProcessEngine,
	repository.NewOrderRepository,
	dao.NewTaskFormDAO,
	dao.NewOrderDAO,
	grpc.NewWorkOrderServer,
)

func InitModule(q mq.MQ, db *mongox.Mongo, workflowModule *workflow.Module, engineModule *engine.Module,
	templateModule *template.Module, userModule *user.Module, lark *lark.Client, sender sender.NotificationSender) (*Module, error) {
	wire.Build(
		ProviderSet,
		event.NewCreateProcessEventProducer,
		initWechatConsumer,
		InitProcessConsumer,
		InitModifyStatusConsumer,
		InitLardCallbackConsumer,
		wire.FieldsOf(new(*workflow.Module), "Svc"),
		wire.FieldsOf(new(*user.Module), "Svc"),
		wire.FieldsOf(new(*engine.Module), "Svc"),
		wire.FieldsOf(new(*template.Module), "Svc"),
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}

func initWechatConsumer(svc service.Service, templateSvc template.Service, userSvc user.Service, q mq.MQ) *consumer.WechatOrderConsumer {
	c, err := consumer.NewWechatOrderConsumer(svc, templateSvc, userSvc, q)
	if err != nil {
		panic(err)
	}

	c.Start(context.Background())
	return c
}

func InitProcessConsumer(q mq.MQ, workflowSvc workflow.Service, svc service.Service) *consumer.ProcessEventConsumer {
	c, err := consumer.NewProcessEventConsumer(q, workflowSvc, svc)
	if err != nil {
		return nil
	}

	c.Start(context.Background())
	return c
}

func InitModifyStatusConsumer(q mq.MQ, svc service.Service) *consumer.OrderStatusModifyEventConsumer {
	c, err := consumer.NewOrderStatusModifyEventConsumer(q, svc)
	if err != nil {
		return nil
	}

	c.Start(context.Background())
	return c
}

func InitLardCallbackConsumer(q mq.MQ, engineSvc engine.Service, lark *lark.Client, userSvc user.Service,
	templateSvc template.Service, svc service.Service, engineProcessSvc service.ProcessEngine,
	workflowSvc workflow.Service, sender sender.NotificationSender) *consumer.LarkCallbackEventConsumer {
	c, err := consumer.NewLarkCallbackEventConsumer(q, engineSvc, engineProcessSvc, svc, templateSvc,
		sender, userSvc, workflowSvc, lark)
	if err != nil {
		return nil
	}

	c.Start(context.Background())
	return c
}
