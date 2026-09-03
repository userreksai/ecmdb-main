package attribute

import (
	"github.com/userreksai/ecmdb-main/internal/attribute/internal/domain"
	"github.com/userreksai/ecmdb-main/internal/attribute/internal/event"
	"github.com/userreksai/ecmdb-main/internal/attribute/internal/service"
	"github.com/userreksai/ecmdb-main/internal/attribute/internal/web"
)

type Handler = web.Handler

type Service = service.Service

type Attribute = domain.Attribute

type AttributeGroup = domain.AttributeGroup

type Event = event.FieldSecureAttrChange
