//go:build wireinject

package task

import (
	"context"
	"sync"
	"time"

	"github.com/ecodeclub/mq-api"
	"github.com/google/wire"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	executorv1 "github.com/userreksai/ecmdb-main/api/proto/gen/etask/executor/v1"
	taskv1 "github.com/userreksai/ecmdb-main/api/proto/gen/etask/task/v1"
	"github.com/userreksai/ecmdb-main/internal/codebook"
	"github.com/userreksai/ecmdb-main/internal/discovery"
	"github.com/userreksai/ecmdb-main/internal/engine"
	"github.com/userreksai/ecmdb-main/internal/order"
	"github.com/userreksai/ecmdb-main/internal/pkg/notification/sender"
	"github.com/userreksai/ecmdb-main/internal/runner"
	"github.com/userreksai/ecmdb-main/internal/task/internal/event"
	"github.com/userreksai/ecmdb-main/internal/task/internal/job"
	"github.com/userreksai/ecmdb-main/internal/task/internal/repository"
	"github.com/userreksai/ecmdb-main/internal/task/internal/repository/dao"
	"github.com/userreksai/ecmdb-main/internal/task/internal/service"
	"github.com/userreksai/ecmdb-main/internal/task/internal/service/dispatch"
	"github.com/userreksai/ecmdb-main/internal/task/internal/service/scheduler"
	"github.com/userreksai/ecmdb-main/internal/task/internal/web"
	"github.com/userreksai/ecmdb-main/internal/user"
	"github.com/userreksai/ecmdb-main/internal/worker"
	"github.com/userreksai/ecmdb-main/internal/workflow"
	"github.com/userreksai/ecmdb-main/pkg/cryptox"
	"github.com/userreksai/ecmdb-main/pkg/mongox"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	dispatch.NewTaskDispatcher,
	service.NewService,
	repository.NewTaskRepository,
	InitTaskDAO,
	scheduler.NewScheduler,
)

func InitModule(q mq.MQ, db *mongox.Mongo, orderModule *order.Module, workflowModule *workflow.Module,
	engineModule *engine.Module, codebookModule *codebook.Module, workerModule *worker.Module,
	runnerModule *runner.Module, userModule *user.Module, discoveryModule *discovery.Module,
	lark *lark.Client, crypto *cryptox.CryptoRegistry, sender sender.NotificationSender,
	taskClient taskv1.TaskServiceClient, executorClient executorv1.TaskExecutionServiceClient) (*Module, error) {
	wire.Build(
		ProviderSet,
		initStartTaskJob,
		initPassProcessTaskJob,
		initTaskRecoveryJob,
		initTaskExecutionSyncJob,
		initConsumer,
		InitCrypto,
		wire.FieldsOf(new(*order.Module), "Svc"),
		wire.FieldsOf(new(*workflow.Module), "Svc"),
		wire.FieldsOf(new(*codebook.Module), "Svc"),
		wire.FieldsOf(new(*user.Module), "Svc"),
		wire.FieldsOf(new(*worker.Module), "Svc"),
		wire.FieldsOf(new(*discovery.Module), "Svc"),
		wire.FieldsOf(new(*runner.Module), "Svc"),
		wire.FieldsOf(new(*engine.Module), "Svc"),
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}

var daoOnce = sync.Once{}

func InitCollectionOnce(db *mongox.Mongo) {
	daoOnce.Do(func() {
		err := dao.InitIndexes(db)
		if err != nil {
			panic(err)
		}
	})
}

func InitTaskDAO(db *mongox.Mongo) dao.TaskDAO {
	InitCollectionOnce(db)
	return dao.NewTaskDAO(db)
}

func initConsumer(svc service.Service, q mq.MQ) *event.ExecuteResultConsumer {
	consumer, err := event.NewExecuteResultConsumer(q, svc)
	if err != nil {
		panic(err)
	}

	consumer.Start(context.Background())
	return consumer
}

func InitCrypto(reg *cryptox.CryptoRegistry) cryptox.Crypto {
	return reg.Runner
}

func initStartTaskJob(svc service.Service) *StartTaskJob {
	limit := int64(100)
	initialInterval := 10 * time.Second
	maxInterval := 30 * time.Second
	maxRetries := int32(3)
	return job.NewStartTaskJob(svc, limit, initialInterval, maxInterval, maxRetries)
}

func initTaskRecoveryJob(svc service.Service) *TaskRecoveryJob {
	limit := int64(100)
	return job.NewTaskRecoveryJob(svc, limit)
}

func initPassProcessTaskJob(svc service.Service, engineSvc engine.Service) *PassProcessTaskJob {
	minutes := int64(10)
	seconds := int64(10)
	limit := int64(100)
	return job.NewPassProcessTaskJob(svc, engineSvc, minutes, seconds, limit)
}

func initTaskExecutionSyncJob(svc service.Service, engineSvc engine.Service, executorSvc executorv1.TaskExecutionServiceClient) *TaskExecutionSyncJob {
	limit := int64(100)
	return job.NewTaskExecutionSyncJob(svc, executorSvc, limit)
}
