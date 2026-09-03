package task

import (
	"github.com/userreksai/ecmdb-main/internal/task/internal/event"
	"github.com/userreksai/ecmdb-main/internal/task/internal/web"
)

type Module struct {
	Svc                  Service
	Hdl                  *web.Handler
	c                    *event.ExecuteResultConsumer
	StartTaskJob         *StartTaskJob
	PassProcessTaskJob   *PassProcessTaskJob
	TaskRecoveryJob      *TaskRecoveryJob
	TaskExecutionSyncJob *TaskExecutionSyncJob
}
