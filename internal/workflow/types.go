package workflow

import (
	"github.com/userreksai/ecmdb-main/internal/workflow/internal/domain"
	"github.com/userreksai/ecmdb-main/internal/workflow/internal/service"
	"github.com/userreksai/ecmdb-main/internal/workflow/internal/web"
)

type Handler = web.Handler

type Service = service.Service

type Workflow = domain.Workflow

type LogicFlow = domain.LogicFlow

type NotifyMethod = domain.NotifyMethod

// NotifyBinding 暴露 domain.NotifyBinding
type NotifyBinding = domain.NotifyBinding

// NotifyType 暴露 domain.NotifyType
type NotifyType = domain.NotifyType

const (
	NotifyTypeApproval            NotifyType = domain.NotifyTypeApproval
	NotifyTypeCC                  NotifyType = domain.NotifyTypeCC
	NotifyTypeChat                NotifyType = domain.NotifyTypeChat
	NotifyTypeProgress            NotifyType = domain.NotifyTypeProgress
	NotifyTypeProgressImageResult NotifyType = domain.NotifyTypeProgressImageResult
	NotifyTypeRevoke              NotifyType = domain.NotifyTypeRevoke
)

const (
	Feishu NotifyMethod = domain.Feishu
	Wechat NotifyMethod = domain.Wechat
)
