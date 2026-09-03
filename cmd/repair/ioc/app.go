package ioc

import (
	"github.com/userreksai/ecmdb-main/internal/attribute"
	"github.com/userreksai/ecmdb-main/internal/model"
	"github.com/userreksai/ecmdb-main/internal/resource"
)

type App struct {
	ModelSvc    model.Service
	AttrSvc     attribute.Service
	ResourceSvc resource.EncryptedSvc
}
