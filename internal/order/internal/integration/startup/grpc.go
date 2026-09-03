package startup

import (
	"github.com/gotomicro/ego/core/elog"
	"github.com/spf13/viper"
	"github.com/userreksai/ecmdb-main/internal/order"
	"github.com/userreksai/ecmdb-main/pkg/grpcx"
	"github.com/userreksai/ecmdb-main/pkg/grpcx/interceptors/jwt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

func InitGrpcServer(orderRpc *order.RpcServer, ecli *clientv3.Client) *grpcx.Server {
	type Config struct {
		Name    string `mapstructure:"name"`
		Port    int    `mapstructure:"port"`
		EtcdTTL int64  `mapstructure:"etcd_ttl"`
		Key     string `mapstructure:"key"`
	}

	var cfg Config
	err := viper.UnmarshalKey("grpc.server", &cfg)
	if err != nil {
		panic(err)
	}

	jwtInterceptor := jwt.NewJwtAuth(cfg.Key)
	server := grpc.NewServer(grpc.UnaryInterceptor(
		jwtInterceptor.JwtAuthInterceptor(),
	))
	orderRpc.Register(server)

	return &grpcx.Server{
		Server:     server,
		Port:       cfg.Port,
		Name:       cfg.Name,
		L:          elog.DefaultLogger,
		EtcdClient: ecli,
		EtcdTTL:    cfg.EtcdTTL,
	}
}
