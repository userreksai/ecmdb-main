package template

import (
	"github.com/userreksai/ecmdb-main/internal/template/internal/domain"
	"github.com/userreksai/ecmdb-main/internal/template/internal/service"
	"github.com/userreksai/ecmdb-main/internal/template/internal/web"
)

type Handler = web.Handler

type GroupHdl = web.GroupHandler

type Service = service.Service

type Template = domain.Template
