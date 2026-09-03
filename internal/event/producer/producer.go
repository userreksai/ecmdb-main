package producer

import (
	"context"

	"github.com/ecodeclub/mq-api"
	"github.com/userreksai/ecmdb-main/pkg/mqx"
)

type OrderStatusModifyEventProducer interface {
	Produce(ctx context.Context, evt OrderStatusModifyEvent) error
}

func NewOrderStatusModifyEventProducer(q mq.MQ) (OrderStatusModifyEventProducer, error) {
	return mqx.NewGeneralProducer[OrderStatusModifyEvent](q, OrderStatusModifyEventName)
}
