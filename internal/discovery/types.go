package discovery

import (
	"github.com/userreksai/ecmdb-main/internal/discovery/internal/domain"
	"github.com/userreksai/ecmdb-main/internal/discovery/internal/service"
	"github.com/userreksai/ecmdb-main/internal/discovery/internal/web"
)

type Handler = web.Handler

type Service = service.Service

type Discovery = domain.Discovery
