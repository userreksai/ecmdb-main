//go:build wireinject

package rota

import (
	"fmt"
	"time"

	"github.com/google/wire"
	"github.com/userreksai/ecmdb-main/internal/rota/internal/grpc"
	"github.com/userreksai/ecmdb-main/internal/rota/internal/repository"
	"github.com/userreksai/ecmdb-main/internal/rota/internal/repository/dao"
	"github.com/userreksai/ecmdb-main/internal/rota/internal/service"
	"github.com/userreksai/ecmdb-main/internal/rota/internal/service/schedule"
	"github.com/userreksai/ecmdb-main/internal/rota/internal/web"
	"github.com/userreksai/ecmdb-main/pkg/mongox"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	service.NewService,
	repository.NewRotaRepository,
	grpc.NewRotaServer,
	dao.NewRotaDao,
)

func InitModule(db *mongox.Mongo) (*Module, error) {
	wire.Build(
		ProviderSet,
		InitScheduleRule,
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}

func InitScheduleRule() schedule.Scheduler {
	// 创建一个位置对象，表示中国北京的位置
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		fmt.Print()
	}

	return schedule.NewRruleSchedule(location)
}
