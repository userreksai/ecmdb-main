//go:build wireinject

package event

import (
	"log"
	"sync"

	"github.com/ecodeclub/mq-api"
	"github.com/google/wire"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	easyEngine "github.com/userreksai/easy-workflow/workflow/engine"
	notificationv1 "github.com/userreksai/ecmdb-main/api/proto/gen/ealert/notification/v1"
	teamv1 "github.com/userreksai/ecmdb-main/api/proto/gen/ealert/team"
	"github.com/userreksai/ecmdb-main/internal/department"
	"github.com/userreksai/ecmdb-main/internal/engine"
	"github.com/userreksai/ecmdb-main/internal/event/producer"
	"github.com/userreksai/ecmdb-main/internal/event/service/assignees"
	"github.com/userreksai/ecmdb-main/internal/event/service/easyflow"
	"github.com/userreksai/ecmdb-main/internal/event/service/strategy"
	"github.com/userreksai/ecmdb-main/internal/event/service/strategy/automation"
	"github.com/userreksai/ecmdb-main/internal/event/service/strategy/carbon_copy"
	"github.com/userreksai/ecmdb-main/internal/event/service/strategy/chat"
	"github.com/userreksai/ecmdb-main/internal/event/service/strategy/start"
	userstrategy "github.com/userreksai/ecmdb-main/internal/event/service/strategy/user"
	"github.com/userreksai/ecmdb-main/internal/order"
	"github.com/userreksai/ecmdb-main/internal/pkg/notification/sender"
	"github.com/userreksai/ecmdb-main/internal/rota"
	"github.com/userreksai/ecmdb-main/internal/task"
	"github.com/userreksai/ecmdb-main/internal/template"
	"github.com/userreksai/ecmdb-main/internal/user"
	"github.com/userreksai/ecmdb-main/internal/workflow"
	"github.com/userreksai/ecmdb-main/pkg/resolve"
	"gorm.io/gorm"
)

var InitStrategySet = wire.NewSet(
	strategy.NewService,
	userstrategy.NewNotification,
	automation.NewNotification,
	start.NewNotification,
	chat.NewNotification,
	carbon_copy.NewNotification,
	ProviderNewDispatcher,
	wire.Bind(new(strategy.SendStrategy), new(*strategy.Dispatcher)),

	// Resolvers
	assignees.NewAppointResolver,
	assignees.NewFounderResolver,
	assignees.NewLeaderResolver,
	assignees.NewMainLeaderResolver,
	assignees.NewOnCallResolver,
	assignees.NewTemplateResolver,
	assignees.NewTeamResolver,
	InitResolveEngine,
)

func InitResolveEngine(
	appoint *assignees.AppointResolver,
	founder *assignees.FounderResolver,
	leader *assignees.LeaderResolver,
	mainLeader *assignees.MainLeaderResolver,
	onCall *assignees.OnCallResolver,
	template *assignees.TemplateResolver,
	team *assignees.TeamResolver,
) resolve.Engine {
	return resolve.NewEngine().Register(
		appoint,
		founder,
		leader,
		mainLeader,
		onCall,
		template,
		team,
	)
}

func InitModule(q mq.MQ, db *gorm.DB, engineModule *engine.Module, taskModule *task.Module, orderModule *order.Module,
	templateModule *template.Module, userModule *user.Module, workflowModule *workflow.Module, sender sender.NotificationSender,
	departmentModule *department.Module, rotaModule *rota.Module, lark *lark.Client, notificationSvc notificationv1.NotificationServiceClient,
	teamSvc teamv1.TeamServiceClient) (*Module, error) {
	wire.Build(
		producer.NewOrderStatusModifyEventProducer,
		InitStrategySet,
		InitWorkflowEngineOnce,
		wire.FieldsOf(new(*engine.Module), "Svc"),
		wire.FieldsOf(new(*department.Module), "Svc"),
		wire.FieldsOf(new(*task.Module), "Svc"),
		wire.FieldsOf(new(*template.Module), "Svc"),
		wire.FieldsOf(new(*order.Module), "Svc"),
		wire.FieldsOf(new(*workflow.Module), "Svc"),
		wire.FieldsOf(new(*user.Module), "Svc"),
		wire.FieldsOf(new(*rota.Module), "Svc"),
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}

var engineOnce = sync.Once{}

func InitWorkflowEngineOnce(db *gorm.DB, engineSvc engine.Service, producer producer.OrderStatusModifyEventProducer,
	taskSvc task.Service, orderSvc order.Service, workflowSvc workflow.Service,
	strategy strategy.SendStrategy) *easyflow.ProcessEvent {
	event, err := easyflow.NewProcessEvent(producer, engineSvc, taskSvc, orderSvc, workflowSvc, strategy)
	if err != nil {
		panic(err)
	}

	engineOnce.Do(func() {
		easyEngine.DB = db
		if err = easyEngine.DatabaseInitialize(); err != nil {
			log.Fatalln("easy workflow 初始化数据表失败，错误:", err)
		}
		// 是否忽略事件错误
		easyEngine.IgnoreEventError = false
	})

	return event
}

func ProviderNewDispatcher(
	user *userstrategy.Notification,
	auto *automation.Notification,
	start *start.Notification,
	chat *chat.Notification,
	cc *carbon_copy.Notification,
	base strategy.Service,
) *strategy.Dispatcher {
	return strategy.NewDispatcher(user, auto, start, chat, cc, base)
}
