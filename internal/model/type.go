package model

import (
	"github.com/userreksai/ecmdb-main/internal/model/internal/domain"
	"github.com/userreksai/ecmdb-main/internal/model/internal/service"
	"github.com/userreksai/ecmdb-main/internal/model/internal/web"
)

type Handler = web.Handler

type Model = domain.Model

type ModelGroup = domain.ModelGroup

type Service = service.Service

type MGService = service.MGService
