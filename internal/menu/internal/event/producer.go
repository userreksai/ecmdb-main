package event

import (
	"context"

	"github.com/ecodeclub/mq-api"
	"github.com/userreksai/ecmdb-main/pkg/mqx"
)

type MenuChangeEventProducer interface {
	Produce(ctx context.Context, evt MenuEvent) error
}

func NewMenuChangeEventProducer(q mq.MQ) (MenuChangeEventProducer, error) {
	return mqx.NewGeneralProducer[MenuEvent](q, MenuChangeEventName)
}
