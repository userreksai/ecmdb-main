package schedule

import "github.com/userreksai/ecmdb-main/internal/rota/internal/domain"

type Scheduler interface {
	// GenerateSchedule 生成排班表
	GenerateSchedule(rule domain.RotaRule, adjustmentRules []domain.RotaAdjustmentRule,
		stime int64, etime int64) (domain.ShiftRostered, error)
	// GetCurrentSchedule 查询当期排班
	GetCurrentSchedule(rule domain.RotaRule, adjustmentRules []domain.RotaAdjustmentRule) (domain.Schedule, error)
}
