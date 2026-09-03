package resource

import (
	"github.com/userreksai/ecmdb-main/internal/resource/internal/domain"
	"github.com/userreksai/ecmdb-main/internal/resource/internal/service"
	"github.com/userreksai/ecmdb-main/internal/resource/internal/web"
)

type Handler = web.Handler

type Service = service.Service

type Resource = domain.Resource

type EncryptedSvc = service.EncryptedSvc

type Operator = domain.Operator
type FilterGroup = domain.FilterGroup

type FilterCondition = domain.FilterCondition
