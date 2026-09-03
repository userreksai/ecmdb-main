package bootstrap

import "github.com/userreksai/ecmdb-main/internal/bootstrap/internal/service"

type Service = service.Loader

type Module struct {
	Svc Service
}
