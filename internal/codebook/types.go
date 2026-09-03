package codebook

import (
	"github.com/userreksai/ecmdb-main/internal/codebook/internal/domain"
	"github.com/userreksai/ecmdb-main/internal/codebook/internal/service"
	"github.com/userreksai/ecmdb-main/internal/codebook/internal/web"
)

type Handler = web.Handler

type Service = service.Service

type Codebook = domain.Codebook
