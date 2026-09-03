package resource

import "github.com/userreksai/ecmdb-main/internal/resource/internal/event"

type Module struct {
	Svc          Service
	EncryptedSvc EncryptedSvc
	Hdl          *Handler
	c            *event.FieldSecureAttrChangeConsumer
}
