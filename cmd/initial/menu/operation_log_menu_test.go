package menu

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOperationLogMenuAndIDs(t *testing.T) {
	seen := make(map[int64]struct{}, len(DefaultMenus))
	var operationLogFound bool
	var previewPermissionFound bool

	for _, item := range DefaultMenus {
		_, duplicate := seen[item.Id]
		assert.Falsef(t, duplicate, "duplicate menu ID: %d", item.Id)
		seen[item.Id] = struct{}{}

		if item.Id == 327 {
			operationLogFound = item.Name == "system-operation-log"
		}
		if item.Id == 292 {
			for _, endpoint := range item.Endpoints {
				if endpoint.Path == "/api/dataio/import/preview" && endpoint.Method == "POST" {
					previewPermissionFound = true
				}
			}
		}
	}

	require.True(t, operationLogFound)
	require.True(t, previewPermissionFound)
}
