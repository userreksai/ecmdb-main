//go:build wireinject

package runner

import (
	"github.com/google/wire"
	"github.com/userreksai/ecmdb-main/internal/codebook"
	"github.com/userreksai/ecmdb-main/internal/runner/internal/repository"
	"github.com/userreksai/ecmdb-main/internal/runner/internal/repository/dao"
	"github.com/userreksai/ecmdb-main/internal/runner/internal/service"
	"github.com/userreksai/ecmdb-main/internal/runner/internal/web"
	"github.com/userreksai/ecmdb-main/internal/workflow"
	"github.com/userreksai/ecmdb-main/pkg/cryptox"
	"github.com/userreksai/ecmdb-main/pkg/mongox"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	service.NewService,
	repository.NewRunnerRepository,
	dao.NewRunnerDAO,
)

func InitModule(db *mongox.Mongo, workflowSvc *workflow.Module, codebookModule *codebook.Module,
	crypto *cryptox.CryptoRegistry) (*Module, error) {
	wire.Build(
		ProviderSet,
		InitCrypto,
		wire.FieldsOf(new(*workflow.Module), "Svc"),
		wire.FieldsOf(new(*codebook.Module), "Svc"),
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}

func InitCrypto(reg *cryptox.CryptoRegistry) cryptox.Crypto {
	return reg.Runner
}
