package channel

import (
	"github.com/gotomicro/ego/core/elog"
	"github.com/userreksai/ecmdb-main/internal/pkg/notification/provider"
)

type larkTextChannel struct {
	baseChannel
}

func NewLarkTextChannel(builder provider.SelectorBuilder) Channel {
	return &larkTextChannel{
		baseChannel{
			builder: builder,
			logger:  elog.DefaultLogger,
		},
	}
}
