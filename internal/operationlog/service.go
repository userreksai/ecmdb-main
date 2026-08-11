package operationlog

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

const retentionMonths = -1

type logEntity struct {
	ID             int64     `gorm:"column:id;primaryKey;autoIncrement"`
	Account        string    `gorm:"column:account;type:varchar(128);not null;index:idx_operation_log_account"`
	OperationModel string    `gorm:"column:operation_model;type:varchar(128);not null;index:idx_operation_log_model"`
	OperationType  string    `gorm:"column:operation_type;type:varchar(16);not null;index:idx_operation_log_type"`
	OriginalData   string    `gorm:"column:original_data;type:json"`
	ModifiedData   string    `gorm:"column:modified_data;type:json"`
	OperationTime  time.Time `gorm:"column:operation_time;type:datetime(3);not null;index:idx_operation_log_time"`
	ModifiedCount  int64     `gorm:"column:modified_count;not null;default:0"`
}

func (logEntity) TableName() string { return "operation_logs" }

type service struct{ db *gorm.DB }

func NewService(db *gorm.DB) (Service, error) {
	if err := db.AutoMigrate(&logEntity{}); err != nil {
		return nil, fmt.Errorf("初始化操作日志表失败: %w", err)
	}
	return &service{db: db}, nil
}

func (s *service) Record(ctx context.Context, record Record) error {
	original, err := json.Marshal(record.OriginalData)
	if err != nil {
		return fmt.Errorf("序列化原数据失败: %w", err)
	}
	modified, err := json.Marshal(record.ModifiedData)
	if err != nil {
		return fmt.Errorf("序列化修改后数据失败: %w", err)
	}

	entity := logEntity{
		Account:        strings.TrimSpace(record.Account),
		OperationModel: strings.TrimSpace(record.OperationModel),
		OperationType:  strings.ToUpper(strings.TrimSpace(record.OperationType)),
		OriginalData:   string(original),
		ModifiedData:   string(modified),
		OperationTime:  time.Now(),
		ModifiedCount:  record.ModifiedCount,
	}
	if entity.Account == "" {
		entity.Account = "unknown"
	}
	if err = s.db.WithContext(ctx).Create(&entity).Error; err != nil {
		return fmt.Errorf("写入操作日志失败: %w", err)
	}
	return nil
}

func (s *service) List(ctx context.Context, query Query) ([]Log, int64, error) {
	limit := query.Limit
	if limit <= 0 || limit > 200 {
		limit = 20
	}
	if query.Offset < 0 {
		query.Offset = 0
	}

	db := s.db.WithContext(ctx).Model(&logEntity{})
	retentionStart := time.Now().AddDate(0, retentionMonths, 0)
	start := retentionStart
	if query.StartTime != nil && query.StartTime.After(start) {
		start = *query.StartTime
	}
	db = db.Where("operation_time >= ?", start)
	if query.EndTime != nil {
		db = db.Where("operation_time <= ?", *query.EndTime)
	}
	if account := strings.TrimSpace(query.Account); account != "" {
		db = db.Where("account LIKE ?", "%"+account+"%")
	}
	if model := strings.TrimSpace(query.OperationModel); model != "" {
		db = db.Where("operation_model LIKE ?", "%"+model+"%")
	}
	if operationType := strings.ToUpper(strings.TrimSpace(query.OperationType)); operationType != "" {
		db = db.Where("operation_type = ?", operationType)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计操作日志失败: %w", err)
	}
	var entities []logEntity
	if err := db.Order("operation_time DESC, id DESC").Offset(query.Offset).Limit(limit).Find(&entities).Error; err != nil {
		return nil, 0, fmt.Errorf("查询操作日志失败: %w", err)
	}

	logs := make([]Log, 0, len(entities))
	for _, entity := range entities {
		logs = append(logs, Log{
			ID:             entity.ID,
			Account:        entity.Account,
			OperationModel: entity.OperationModel,
			OperationType:  entity.OperationType,
			OriginalData:   json.RawMessage([]byte(entity.OriginalData)),
			ModifiedData:   json.RawMessage([]byte(entity.ModifiedData)),
			OperationTime:  entity.OperationTime,
			ModifiedCount:  entity.ModifiedCount,
		})
	}
	return logs, total, nil
}

func (s *service) CleanupExpired(ctx context.Context) (int64, error) {
	result := s.db.WithContext(ctx).Where("operation_time < ?", time.Now().AddDate(0, retentionMonths, 0)).Delete(&logEntity{})
	return result.RowsAffected, result.Error
}
