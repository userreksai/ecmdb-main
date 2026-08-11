package service

import (
	"testing"

	"github.com/Duke1616/ecmdb/internal/resource"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildImportPreview(t *testing.T) {
	sheet := parsedImportSheet{
		Columns: []string{"name", "ip"},
		Resources: []resource.Resource{
			{ModelUID: "host", Data: mongox.MapStr{"name": "host-a", "ip": "10.0.0.2"}},
			{ModelUID: "host", Data: mongox.MapStr{"name": "host-b", "ip": "10.0.0.3"}},
		},
	}
	current := []resource.Resource{
		{ID: 1, ModelUID: "host", Data: mongox.MapStr{"name": "host-a", "ip": "10.0.0.1", "owner": "ops"}},
		{ID: 2, ModelUID: "host", Data: mongox.MapStr{"name": "host-c", "ip": "10.0.0.4"}},
	}

	preview := buildImportPreview("host", sheet, current)

	assert.Equal(t, 1, preview.CreatedCount)
	assert.Equal(t, 1, preview.UpdatedCount)
	assert.Equal(t, 1, preview.DeletedCount)
	assert.Zero(t, preview.UnchangedCount)
	require.Len(t, preview.Rows, 3)

	updated := preview.Rows[0]
	assert.Equal(t, ImportActionUpdate, updated.Action)
	assert.Equal(t, "ops", updated.ModifiedData["owner"], "表格缺少的字段应保留原值")
	_, submittedOwner := updated.Resource.Data["owner"]
	assert.False(t, submittedOwner, "写入时只提交表格实际包含的字段")
}

func TestBuildImportPreviewEmptySheet(t *testing.T) {
	current := []resource.Resource{
		{ID: 1, ModelUID: "host", Data: mongox.MapStr{"name": "host-a"}},
		{ID: 2, ModelUID: "host", Data: mongox.MapStr{"name": "host-b"}},
	}

	preview := buildImportPreview("host", parsedImportSheet{Columns: []string{"name"}}, current)

	assert.True(t, preview.IsEmpty)
	assert.Equal(t, 2, preview.DeletedCount)
	require.Len(t, preview.Rows, 2)
	assert.Equal(t, ImportActionDelete, preview.Rows[0].Action)
}

func TestBuildImportPreviewClearsIncludedBlankField(t *testing.T) {
	sheet := parsedImportSheet{
		Columns: []string{"name", "ip"},
		Resources: []resource.Resource{
			{ModelUID: "host", Data: mongox.MapStr{"name": "host-a", "ip": ""}},
		},
	}
	current := []resource.Resource{
		{ID: 1, ModelUID: "host", Data: mongox.MapStr{"name": "host-a", "ip": "10.0.0.1", "owner": "ops"}},
	}

	preview := buildImportPreview("host", sheet, current)

	require.Len(t, preview.Rows, 1)
	assert.Equal(t, ImportActionUpdate, preview.Rows[0].Action)
	assert.Equal(t, "", preview.Rows[0].ModifiedData["ip"], "表格中存在的空单元格应清空字段")
	assert.Equal(t, "ops", preview.Rows[0].ModifiedData["owner"], "表格中缺少的字段列应保留原值")
}
