//go:build wireinject

package ioc

import (
	"time"

	"github.com/google/wire"
	"github.com/spf13/viper"
	notificationv1 "github.com/userreksai/ecmdb-main/api/proto/gen/ealert/notification/v1"
	teamv1 "github.com/userreksai/ecmdb-main/api/proto/gen/ealert/team"
	templatev1 "github.com/userreksai/ecmdb-main/api/proto/gen/ealert/template/v1"
	"github.com/userreksai/ecmdb-main/internal/attribute"
	"github.com/userreksai/ecmdb-main/internal/codebook"
	"github.com/userreksai/ecmdb-main/internal/dataio"
	"github.com/userreksai/ecmdb-main/internal/department"
	"github.com/userreksai/ecmdb-main/internal/discovery"
	"github.com/userreksai/ecmdb-main/internal/endpoint"
	"github.com/userreksai/ecmdb-main/internal/engine"
	"github.com/userreksai/ecmdb-main/internal/event"
	"github.com/userreksai/ecmdb-main/internal/menu"
	"github.com/userreksai/ecmdb-main/internal/model"
	"github.com/userreksai/ecmdb-main/internal/operationlog"
	"github.com/userreksai/ecmdb-main/internal/order"
	"github.com/userreksai/ecmdb-main/internal/permission"
	"github.com/userreksai/ecmdb-main/internal/pkg/middleware"
	"github.com/userreksai/ecmdb-main/internal/pkg/servicetoken"
	"github.com/userreksai/ecmdb-main/internal/policy"
	"github.com/userreksai/ecmdb-main/internal/relation"
	"github.com/userreksai/ecmdb-main/internal/resource"
	"github.com/userreksai/ecmdb-main/internal/role"
	"github.com/userreksai/ecmdb-main/internal/rota"
	"github.com/userreksai/ecmdb-main/internal/runner"
	"github.com/userreksai/ecmdb-main/internal/strategy"
	"github.com/userreksai/ecmdb-main/internal/task"
	"github.com/userreksai/ecmdb-main/internal/template"
	"github.com/userreksai/ecmdb-main/internal/terminal"
	"github.com/userreksai/ecmdb-main/internal/tools"
	"github.com/userreksai/ecmdb-main/internal/user"
	"github.com/userreksai/ecmdb-main/internal/worker"
	"github.com/userreksai/ecmdb-main/internal/workflow"
	"github.com/userreksai/ecmdb-main/pkg/storage"
	grpcpkg "github.com/userreksai/ecmdb-task/pkg/grpc"
	"github.com/userreksai/ecmdb-task/pkg/grpc/registry"
	"google.golang.org/grpc"
)

var BaseSet = wire.NewSet(InitMongoDB, InitMySQLDB, InitRedis, InitMinioClient, InitMQ,
	InitRedisSearch, InitEtcdClient, InitWorkWx, InitLarkClient, InitModuleCrypto, InitRegistry)

func InitApp() (*App, error) {
	wire.Build(wire.Struct(new(App), "*"),
		BaseSet,
		servicetoken.NewManager,
		InitSession,
		InitCasbin,
		InitLdapConfig,
		storage.NewS3Storage,
		InitEALERTGrpcClient,
		InitNotificationServiceClient,
		InitTeamServiceClient,
		model.InitModule,
		wire.FieldsOf(new(*model.Module), "Hdl"),
		attribute.InitModule,
		wire.FieldsOf(new(*attribute.Module), "Hdl"),
		resource.InitModule,
		wire.FieldsOf(new(*resource.Module), "Hdl"),
		relation.InitModule,
		wire.FieldsOf(new(*relation.Module), "RRHdl", "RMHdl", "RTHdl"),
		user.InitModule,
		wire.FieldsOf(new(*user.Module), "Hdl", "Svc", "RpcServer"),
		template.InitModule,
		wire.FieldsOf(new(*template.Module), "Hdl", "GroupHdl"),
		codebook.InitModule,
		wire.FieldsOf(new(*codebook.Module), "Hdl"),
		worker.InitModule,
		runner.InitModule,
		wire.FieldsOf(new(*runner.Module), "Hdl"),
		order.InitModule,
		wire.FieldsOf(new(*order.Module), "Hdl", "RpcServer"),
		strategy.InitModule,
		wire.FieldsOf(new(*strategy.Module), "Hdl"),
		workflow.InitModule,
		wire.FieldsOf(new(*workflow.Module), "Hdl", "Svc"),
		engine.InitModule,
		wire.FieldsOf(new(*engine.Module), "Hdl"),
		event.InitModule,
		wire.FieldsOf(new(*event.Module), "Event"),
		task.InitModule,
		wire.FieldsOf(new(*task.Module), "Hdl", "StartTaskJob", "PassProcessTaskJob", "TaskExecutionSyncJob", "TaskRecoveryJob"),
		policy.InitModule,
		wire.FieldsOf(new(*policy.Module), "Hdl", "Svc", "RpcServer"),
		operationlog.InitModule,
		wire.FieldsOf(new(*operationlog.Module), "Hdl"),
		menu.InitModule,
		wire.FieldsOf(new(*menu.Module), "Hdl"),
		endpoint.InitModule,
		wire.FieldsOf(new(*endpoint.Module), "Hdl", "Svc", "RpcServer"),
		department.InitModule,
		wire.FieldsOf(new(*department.Module), "Hdl"),
		role.InitModule,
		wire.FieldsOf(new(*role.Module), "Hdl"),
		permission.InitModule,
		wire.FieldsOf(new(*permission.Module), "Hdl"),
		rota.InitModule,
		wire.FieldsOf(new(*rota.Module), "Hdl", "RpcServer"),
		discovery.InitModule,
		wire.FieldsOf(new(*discovery.Module), "Hdl"),
		tools.InitModule,
		terminal.InitModule,
		dataio.InitModule,
		wire.FieldsOf(new(*dataio.Module), "Hdl"),
		InitTASKGrpcClient,
		InitTaskServiceClient,
		InitTaskExecutionServiceClient,
		middleware.NewCheckPolicyMiddlewareBuilder,
		middleware.NewCheckLoginMiddlewareBuilder,
		initCronJobs,
		InitWebServer,
		InitListener,
		InitGrpcServer,
		InitGinMiddlewares,

		// 消息通知
		InitSender,
	)
	return new(App), nil
}

// InitEALERTGrpcClient 初始化 EALERT gRPC 客户端
func InitEALERTGrpcClient(reg registry.Registry) grpc.ClientConnInterface {
	var cfg grpcpkg.ClientConfig
	if err := viper.UnmarshalKey("grpc.client.ealert", &cfg); err != nil {
		panic(err)
	}

	// 通过 WaitForReady 控制，如果地址不通直接返回错误
	cc, err := grpcpkg.NewClientConn(
		reg,
		grpcpkg.WithServiceName(cfg.Name),
		grpcpkg.WithClientJWTAuth(cfg.AuthToken),
		grpcpkg.WithDialOption(grpc.WithConnectParams(grpc.ConnectParams{
			MinConnectTimeout: 3 * time.Second,
		}),
			grpc.WithDefaultCallOptions(
				grpc.WaitForReady(false),
			)),
	)
	if err != nil {
		panic(err)
	}

	return cc
}

// InitNotificationServiceClient 初始化 notification 服务客户端
func InitNotificationServiceClient(cc grpc.ClientConnInterface) notificationv1.NotificationServiceClient {
	return notificationv1.NewNotificationServiceClient(cc)
}

// InitTeamServiceClient 初始化 team 服务客户端
func InitTeamServiceClient(cc grpc.ClientConnInterface) teamv1.TeamServiceClient {
	return teamv1.NewTeamServiceClient(cc)
}

// InitTemplateServiceClient 初始化 template 服务客户端
func InitTemplateServiceClient(cc grpc.ClientConnInterface) templatev1.TemplateServiceClient {
	return templatev1.NewTemplateServiceClient(cc)
}
