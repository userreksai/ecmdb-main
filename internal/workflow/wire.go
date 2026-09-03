//go:build wireinject

package workflow

import (
	"github.com/google/wire"
	"github.com/userreksai/ecmdb-main/internal/engine"
	"github.com/userreksai/ecmdb-main/internal/workflow/internal/repository"
	"github.com/userreksai/ecmdb-main/internal/workflow/internal/repository/dao"
	"github.com/userreksai/ecmdb-main/internal/workflow/internal/service"
	"github.com/userreksai/ecmdb-main/internal/workflow/internal/web"
	"github.com/userreksai/ecmdb-main/internal/workflow/pkg/easyflow"
	"github.com/userreksai/ecmdb-main/pkg/mongox"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	service.NewService,
	repository.NewWorkflowRepository,
	repository.NewNotifyBindingRepository,
	dao.NewWorkflowDAO,
	dao.NewSnapshotDAO,
	dao.NewNotifyBindingDAO,
	easyflow.NewLogicFlowToEngineConvert,
)

func InitModule(db *mongox.Mongo, engineModule *engine.Module) (*Module, error) {
	wire.Build(
		ProviderSet,
		wire.FieldsOf(new(*engine.Module), "Svc"),
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}
