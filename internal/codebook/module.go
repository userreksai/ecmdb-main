package codebook

import "github.com/userreksai/ecmdb-main/internal/codebook/internal/service"

type Module struct {
	Hdl *Handler
	Svc service.Service
}
