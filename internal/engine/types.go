package engine

import (
	"github.com/userreksai/ecmdb-main/internal/engine/internal/domain"
	"github.com/userreksai/ecmdb-main/internal/engine/internal/service"
	"github.com/userreksai/ecmdb-main/internal/engine/internal/web"
)

type Service = service.Service

type Handler = web.Handler

type Instance = domain.Instance
