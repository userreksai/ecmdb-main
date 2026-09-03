package operationlog

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogEntityJSONFieldsUseTextValues(t *testing.T) {
	original, err := json.Marshal(map[string]any{"name": "before"})
	assert.NoError(t, err)
	modified, err := json.Marshal(map[string]any{"name": "after"})
	assert.NoError(t, err)

	entity := logEntity{
		OriginalData: string(original),
		ModifiedData: string(modified),
	}

	assert.JSONEq(t, `{"name":"before"}`, entity.OriginalData)
	assert.JSONEq(t, `{"name":"after"}`, entity.ModifiedData)
}

func TestRecordIgnoresAutomationAccount(t *testing.T) {
	svc := &service{}

	err := svc.Record(context.Background(), Record{
		Account:      "  SVC_ECMDB_SCRIPT  ",
		OriginalData: make(chan int),
		ModifiedData: make(chan int),
	})

	assert.NoError(t, err)
}

func TestRecordDoesNotIgnoreOtherAccounts(t *testing.T) {
	svc := &service{}

	err := svc.Record(context.Background(), Record{
		Account:      "admin",
		OriginalData: make(chan int),
	})

	assert.Error(t, err)
}
