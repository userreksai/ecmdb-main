package operationlog

import (
	"context"
	"encoding/json"
	"time"
)

const (
	OperationCreate = "CREATE"
	OperationUpdate = "UPDATE"
	OperationDelete = "DELETE"
	OperationImport = "IMPORT"
)

type Log struct {
	ID             int64           `json:"id"`
	Account        string          `json:"account"`
	OperationModel string          `json:"operation_model"`
	OperationType  string          `json:"operation_type"`
	OriginalData   json.RawMessage `json:"original_data"`
	ModifiedData   json.RawMessage `json:"modified_data"`
	OperationTime  time.Time       `json:"operation_time"`
	ModifiedCount  int64           `json:"modified_count"`
}

type Record struct {
	Account        string
	OperationModel string
	OperationType  string
	OriginalData   any
	ModifiedData   any
	ModifiedCount  int64
}

type Query struct {
	Offset         int
	Limit          int
	Account        string
	OperationModel string
	OperationType  string
	StartTime      *time.Time
	EndTime        *time.Time
}

type Service interface {
	Record(ctx context.Context, record Record) error
	List(ctx context.Context, query Query) ([]Log, int64, error)
	CleanupExpired(ctx context.Context) (int64, error)
}
