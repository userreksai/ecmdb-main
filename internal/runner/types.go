package runner

import (
	"github.com/userreksai/ecmdb-main/internal/runner/internal/domain"
	"github.com/userreksai/ecmdb-main/internal/runner/internal/service"
	"github.com/userreksai/ecmdb-main/internal/runner/internal/web"
)

type Service = service.Service

type Handler = web.Handler

type Runner = domain.Runner

type Variables = domain.Variables
