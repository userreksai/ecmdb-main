package template

import (
	"github.com/userreksai/ecmdb-main/internal/template/internal/event"
)

type Module struct {
	Svc      Service
	c        *event.WechatApprovalCallbackConsumer
	Hdl      *Handler
	GroupHdl *GroupHdl
}
