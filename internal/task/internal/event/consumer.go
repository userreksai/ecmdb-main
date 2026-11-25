package event

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ecodeclub/mq-api"
	"github.com/gotomicro/ego/core/elog"
	"github.com/userreksai/ecmdb-main/internal/task/domain"
	"github.com/userreksai/ecmdb-main/internal/task/internal/service"
)

type ExecuteResultConsumer struct {
	consumer mq.Consumer
	svc      service.Service
	logger   *elog.Component
}

func NewExecuteResultConsumer(q mq.MQ, svc service.Service) (*ExecuteResultConsumer, error) {
	groupID := "task_receive_execute"
	consumer, err := q.Consumer(ExecuteResultEventName, groupID)
	if err != nil {
		return nil, err
	}

	return &ExecuteResultConsumer{
		consumer: consumer,
		svc:      svc,
		logger:   elog.DefaultLogger,
	}, nil
}

func (c *ExecuteResultConsumer) Start(ctx context.Context) {
	go func() {
		for {
			err := c.Consume(ctx)
			if err != nil {
				c.logger.Error("同步修改任务执行状态失败", elog.Any("错误信息", err))
				time.Sleep(time.Second)
			}
		}
	}()
}

func (c *ExecuteResultConsumer) Consume(ctx context.Context) error {
	cm, err := c.consumer.Consume(ctx)
	if err != nil {
		return fmt.Errorf("获取消息失败: %w", err)
	}
	var evt ExecuteResultEvent
	if err = json.Unmarshal(cm.Value, &evt); err != nil {
		return fmt.Errorf("解析消息失败: %w", err)
	}

	// 非正数 task_id 属于 ecmdb-task 为 eflow 提交的工作流执行（task_id 为负的执行 ID），
	// 由 ecmdb-task 自己处理；这里跳过，避免误更新或误发失败通知
	if evt.TaskId <= 0 {
		return nil
	}

	triggerPosition := domain.TriggerPositionTaskExecutionSuccess
	if domain.Status(evt.Status) == domain.FAILED {
		triggerPosition = domain.TriggerPositionTaskExecutionFailed
	}

	_, err = c.svc.UpdateTaskResult(ctx, domain.TaskResult{
		Id:              evt.TaskId,
		Result:          evt.Result,
		WantResult:      evt.WantResult,
		TriggerPosition: triggerPosition.ToString(),
		Status:          domain.Status(evt.Status),
	})

	return err
}

func (c *ExecuteResultConsumer) Stop(_ context.Context) error {
	return c.consumer.Close()
}
