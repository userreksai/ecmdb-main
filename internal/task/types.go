package task

import (
	"github.com/userreksai/ecmdb-main/internal/task/internal/job"
	"github.com/userreksai/ecmdb-main/internal/task/internal/service"
	"github.com/userreksai/ecmdb-main/internal/task/internal/web"
)

type Service = service.Service

type Handler = web.Handler

type StartTaskJob = job.StartTaskJob

type PassProcessTaskJob = job.PassProcessTaskJob

type TaskRecoveryJob = job.TaskRecoveryJob

type TaskExecutionSyncJob = job.TaskExecutionSyncJob
