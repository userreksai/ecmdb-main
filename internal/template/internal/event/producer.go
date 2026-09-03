package event

import (
	"context"

	"github.com/ecodeclub/mq-api"
	"github.com/userreksai/ecmdb-main/pkg/mqx"
	"github.com/xen0n/go-workwx"
)

type WechatOrderEventProducer interface {
	Produce(ctx context.Context, evt *workwx.OAApprovalDetail) error
}

func NewWechatOrderEventProducer(q mq.MQ) (WechatOrderEventProducer, error) {
	return mqx.NewGeneralProducer[*workwx.OAApprovalDetail](q, WechatOrderEventName)
}
