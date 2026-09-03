package workflow

import (
	"github.com/userreksai/ecmdb-main/internal/workflow/internal/service"
	"github.com/userreksai/ecmdb-main/internal/workflow/internal/web"
)

type Module struct {
	Hdl *web.Handler
	Svc service.Service
}
