package model

import (
	"github.com/Duke1616/ecmdb/internal/model/internal/repository"
	"github.com/Duke1616/ecmdb/internal/model/internal/service"
	"github.com/Duke1616/ecmdb/pkg/mongox"
)

// InitService constructs the model service without the HTTP handler dependencies.
func InitService(db *mongox.Mongo) Service {
	modelDAO := InitModelDAO(db)
	modelRepository := repository.NewModelRepository(modelDAO)
	return service.NewModelService(modelRepository)
}
