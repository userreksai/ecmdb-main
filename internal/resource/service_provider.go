package resource

import (
	"github.com/Duke1616/ecmdb/internal/attribute"
	"github.com/Duke1616/ecmdb/internal/resource/internal/repository"
	"github.com/Duke1616/ecmdb/internal/resource/internal/service"
	"github.com/Duke1616/ecmdb/pkg/cryptox"
	"github.com/Duke1616/ecmdb/pkg/mongox"
)

// InitEncryptedService constructs the encrypted resource service without HTTP or event consumers.
func InitEncryptedService(db *mongox.Mongo, attributeModule *attribute.Module, crypto *cryptox.CryptoRegistry) EncryptedSvc {
	resourceDAO := InitResourceDAO(db)
	resourceRepository := repository.NewResourceRepository(resourceDAO)
	baseService := service.NewService(resourceRepository)
	return service.NewEncryptedResourceService(baseService, attributeModule.Svc, crypto.Resource)
}
