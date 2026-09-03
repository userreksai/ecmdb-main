//go:build wireinject

package startup

import (
	"github.com/google/wire"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	notificationv1 "github.com/userreksai/ecmdb-main/api/proto/gen/ealert/notification/v1"
	teamv1 "github.com/userreksai/ecmdb-main/api/proto/gen/ealert/team"
	"github.com/userreksai/ecmdb-main/internal/department"
	"github.com/userreksai/ecmdb-main/internal/engine"
	"github.com/userreksai/ecmdb-main/internal/event"
	"github.com/userreksai/ecmdb-main/internal/event/producer"
	"github.com/userreksai/ecmdb-main/internal/event/service/easyflow"
	"github.com/userreksai/ecmdb-main/internal/order"
	"github.com/userreksai/ecmdb-main/internal/pkg/notification/sender"
	"github.com/userreksai/ecmdb-main/internal/rota"
	"github.com/userreksai/ecmdb-main/internal/task"
	"github.com/userreksai/ecmdb-main/internal/template"
	"github.com/userreksai/ecmdb-main/internal/test/ioc"
	"github.com/userreksai/ecmdb-main/internal/user"
	"github.com/userreksai/ecmdb-main/internal/workflow"
)

type TestApp struct {
	EventHandler     *easyflow.ProcessEvent
	UserModule       *user.Module
	OrderModule      *order.Module
	TaskModule       *task.Module
	TemplateModule   *template.Module
	EngineModule     *engine.Module
	WorkflowModule   *workflow.Module
	DepartmentModule *department.Module
	RotaModule       *rota.Module
}

func InitApp(
	sender sender.NotificationSender,
	teamSvc teamv1.TeamServiceClient,
	notificationSvc notificationv1.NotificationServiceClient,
	userModule *user.Module,
	orderModule *order.Module,
	taskModule *task.Module,
	templateModule *template.Module,
	engineModule *engine.Module,
	workflowModule *workflow.Module,
	departmentModule *department.Module,
	rotaModule *rota.Module,
) (*TestApp, error) {
	wire.Build(
		ioc.BaseSet,

		// External Clients (Nil for integration tests)
		wire.Value((*lark.Client)(nil)),

		// Event Module components (The core logic we are testing)
		event.InitStrategySet,
		event.InitWorkflowEngineOnce,
		producer.NewOrderStatusModifyEventProducer,

		// Extract Services from provided Modules
		wire.FieldsOf(new(*user.Module), "Svc"),
		wire.FieldsOf(new(*order.Module), "Svc"),
		wire.FieldsOf(new(*task.Module), "Svc"),
		wire.FieldsOf(new(*template.Module), "Svc"),
		wire.FieldsOf(new(*engine.Module), "Svc"),
		wire.FieldsOf(new(*workflow.Module), "Svc"),
		wire.FieldsOf(new(*department.Module), "Svc"),
		wire.FieldsOf(new(*rota.Module), "Svc"),

		wire.Struct(new(TestApp), "*"),
	)
	return nil, nil
}
