package resource

import (
	"github.com/userreksai/ecmdb-main/internal/attribute"
	"github.com/userreksai/ecmdb-main/internal/resource/internal/repository"
	"github.com/userreksai/ecmdb-main/internal/resource/internal/service"
	"github.com/userreksai/ecmdb-main/pkg/cryptox"
	"github.com/userreksai/ecmdb-main/pkg/mongox"
)

// InitEncryptedService constructs the encrypted resource service without HTTP or event consumers.
func InitEncryptedService(db *mongox.Mongo, attributeModule *attribute.Module, crypto *cryptox.CryptoRegistry) EncryptedSvc {
	resourceDAO := InitResourceDAO(db)
	resourceRepository := repository.NewResourceRepository(resourceDAO)
	baseService := service.NewService(resourceRepository)
	return service.NewEncryptedResourceService(baseService, attributeModule.Svc, crypto.Resource)
}
