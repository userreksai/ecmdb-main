package department

import (
	"github.com/userreksai/ecmdb-main/internal/department/internal/domain"
	"github.com/userreksai/ecmdb-main/internal/department/internal/service"
	"github.com/userreksai/ecmdb-main/internal/department/internal/web"
)

type Handler = web.Handler

type Department = domain.Department

type Service = service.Service
