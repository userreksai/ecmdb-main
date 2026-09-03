//go:build wireinject

package ioc

import (
	"time"

	"github.com/google/wire"
	"github.com/spf13/viper"
	templatev1 "github.com/userreksai/ecmdb-main/api/proto/gen/ealert/template/v1"
	"github.com/userreksai/ecmdb-main/cmd/initial/version"
	"github.com/userreksai/ecmdb-main/internal/attribute"
	"github.com/userreksai/ecmdb-main/internal/bootstrap"
	"github.com/userreksai/ecmdb-main/internal/department"
	"github.com/userreksai/ecmdb-main/internal/engine"
	"github.com/userreksai/ecmdb-main/internal/menu"
	"github.com/userreksai/ecmdb-main/internal/model"
	"github.com/userreksai/ecmdb-main/internal/operationlog"
	"github.com/userreksai/ecmdb-main/internal/permission"
	"github.com/userreksai/ecmdb-main/internal/pkg/servicetoken"
	"github.com/userreksai/ecmdb-main/internal/policy"
	"github.com/userreksai/ecmdb-main/internal/relation"
	"github.com/userreksai/ecmdb-main/internal/resource"
	"github.com/userreksai/ecmdb-main/internal/role"
	"github.com/userreksai/ecmdb-main/internal/user"
	"github.com/userreksai/ecmdb-main/internal/workflow"
	"github.com/userreksai/ecmdb-main/ioc"
	grpcpkg "github.com/userreksai/ecmdb-task/pkg/grpc"
	"github.com/userreksai/ecmdb-task/pkg/grpc/registry"
	"github.com/userreksai/ecmdb-task/pkg/grpc/registry/etcd"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

var BaseSet = wire.NewSet(ioc.InitMongoDB, ioc.InitMySQLDB, ioc.InitRedis, ioc.InitRedisSearch,
	ioc.InitMQ, ioc.InitEtcdClient, ioc.InitLdapConfig, ioc.InitModuleCrypto)

func InitApp() (*App, error) {
	wire.Build(wire.Struct(new(App), "*"),
		BaseSet,
		servicetoken.NewManager,
		ioc.InitCasbin,
		ioc.InitSession,
		user.InitModule,
		version.NewService,
		version.NewDao,
		wire.FieldsOf(new(*user.Module), "Svc"),
		department.InitModule,
		role.InitModule,
		wire.FieldsOf(new(*role.Module), "Svc"),
		menu.InitModule,
		wire.FieldsOf(new(*menu.Module), "Svc"),
		permission.InitModule,
		wire.FieldsOf(new(*permission.Module), "Svc"),
		policy.InitModule,
		wire.FieldsOf(new(*policy.Module), "Svc"),
		// 操作日志模块（初始化 MySQL 表，并供资源模块记录审计日志）
		operationlog.InitModule,
		// 关联关系模块（无依赖，最先初始化）
		relation.InitModule,
		// 字段模块（依赖 MQ）
		attribute.InitModule,
		// 资源模块（model 依赖它）
		resource.InitModule,
		// 模型模块（依赖 relation、attribute、resource）
		model.InitModule,
		// Bootstrap 模块（依赖 model、attribute、relation，统一管理 CMDB 相关服务）
		bootstrap.InitModule,
		wire.FieldsOf(new(*bootstrap.Module), "Svc"),

		// 流程引擎模块
		engine.InitModule,
		// 工作流模块 (依赖 engine)
		workflow.InitModule,
		wire.FieldsOf(new(*workflow.Module), "Svc"),

		// grpc set
		InitRegistry,
		InitEALERTGrpcClient,
		InitTemplateServiceClient,
	)
	return new(App), nil
}

// InitRegistry 初始化统一的服务注册中心
func InitRegistry(etcdClient *clientv3.Client) registry.Registry {
	r, err := etcd.NewRegistry(etcdClient)
	if err != nil {
		panic(err)
	}
	return r
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

// InitTemplateServiceClient 初始化 template 服务客户端
func InitTemplateServiceClient(cc grpc.ClientConnInterface) templatev1.TemplateServiceClient {
	return templatev1.NewTemplateServiceClient(cc)
}
