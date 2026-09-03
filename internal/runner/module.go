package runner

import (
	"github.com/userreksai/ecmdb-main/internal/runner/internal/service"
	"github.com/userreksai/ecmdb-main/internal/runner/internal/web"
)

type Module struct {
	Svc service.Service
	Hdl *web.Handler
}
