package channel

import (
	"github.com/gotomicro/ego/core/elog"
	"github.com/userreksai/ecmdb-main/internal/pkg/notification/provider"
)

type larkCardChannel struct {
	baseChannel
}

func NewLarkCardChannel(builder provider.SelectorBuilder) Channel {
	return &larkCardChannel{
		baseChannel{
			builder: builder,
			logger:  elog.DefaultLogger,
		},
	}
}
