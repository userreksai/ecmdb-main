//go:build wireinject

package tools

import (
	"github.com/google/wire"
	"github.com/userreksai/ecmdb-main/internal/tools/service"
	"github.com/userreksai/ecmdb-main/internal/tools/web"
	"github.com/userreksai/ecmdb-main/pkg/storage"
)

func InitModule(storage *storage.S3Storage) (*web.Handler, error) {
	wire.Build(
		web.NewHandler,
		service.NewService,
	)
	return new(web.Handler), nil
}
