package permission

import "github.com/userreksai/ecmdb-main/internal/permission/internal/event"

type Module struct {
	Hdl *Handler
	Svc Service
	c   *event.MenuChangeEventConsumer
}
