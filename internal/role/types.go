package role

import (
	"github.com/userreksai/ecmdb-main/internal/role/internal/domain"
	"github.com/userreksai/ecmdb-main/internal/role/internal/service"
	"github.com/userreksai/ecmdb-main/internal/role/internal/web"
)

type Handler = web.Handler

type Service = service.Service

const (
	AdminRole = domain.AdminRole
)

type Role = domain.Role
