package operationlog

import (
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
