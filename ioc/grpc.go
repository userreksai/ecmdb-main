package ioc

import (
	"github.com/userreksai/ecmdb-main/internal/endpoint"
	"github.com/userreksai/ecmdb-main/internal/order"
	"github.com/userreksai/ecmdb-main/internal/policy"
	"github.com/userreksai/ecmdb-main/internal/rota"
	"github.com/userreksai/ecmdb-main/internal/user"

	grpcpkg "github.com/userreksai/ecmdb-task/pkg/grpc"
	registrysdk "github.com/userreksai/ecmdb-task/pkg/grpc/registry"

	"github.com/spf13/viper"
)

func InitGrpcServer(registry registrysdk.Registry, orderRpc *order.RpcServer, policyRpc *policy.RpcServer,
	endpointRpc *endpoint.RpcServer, userRpc *user.RpcServer, rotaRpc *rota.RpcServer) *grpcpkg.Server {
	var cfg grpcpkg.ServerConfig
	if err := viper.UnmarshalKey("grpc.server.ecmdb", &cfg); err != nil {
		panic(err)
	}

	server := grpcpkg.NewServer(cfg, registry, grpcpkg.WithJWTAuth(cfg.AuthToken))

	orderRpc.Register(server)
	policyRpc.Register(server)
	endpointRpc.Register(server)
	userRpc.Register(server)
	rotaRpc.Register(server)

	return server
}
