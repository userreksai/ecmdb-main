package model

import (
	"github.com/userreksai/ecmdb-main/internal/model/internal/repository"
	"github.com/userreksai/ecmdb-main/internal/model/internal/service"
	"github.com/userreksai/ecmdb-main/pkg/mongox"
)

// InitService constructs the model service without the HTTP handler dependencies.
func InitService(db *mongox.Mongo) Service {
	modelDAO := InitModelDAO(db)
	modelRepository := repository.NewModelRepository(modelDAO)
	return service.NewModelService(modelRepository)
}
