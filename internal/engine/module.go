package engine

import (
	"github.com/userreksai/ecmdb-main/internal/engine/internal/web"
)

type Module struct {
	Svc Service
	Hdl *web.Handler
}
