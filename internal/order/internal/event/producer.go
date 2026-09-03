package event

import (
	"context"

	"github.com/ecodeclub/mq-api"
	"github.com/userreksai/ecmdb-main/pkg/mqx"
)

type CreateProcessEventProducer interface {
	Produce(ctx context.Context, evt OrderEvent) error
}

//go:generate mockgen -source=./producer.go -package=evtmocks -destination=./mocks/producer.mock.go -typed NewCreateProcessEventProducer
func NewCreateProcessEventProducer(q mq.MQ) (CreateProcessEventProducer, error) {
	return mqx.NewGeneralProducer[OrderEvent](q, CreateProcessEventName)
}
