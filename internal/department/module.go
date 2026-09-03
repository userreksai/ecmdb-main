package department

import (
	"github.com/userreksai/ecmdb-main/internal/department/internal/service"
	"github.com/userreksai/ecmdb-main/internal/department/internal/web"
)

type Module struct {
	Hdl *web.Handler
	Svc service.Service
}
