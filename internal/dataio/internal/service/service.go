package service

import (
	"context"

	"github.com/Duke1616/ecmdb/internal/resource"
)

// IDataIOService 数据交换服务接口
// NOTE: 提供基于 Model-Attribute-Resource 架构的数据导入导出功能,支持 Excel 格式
type IDataIOService interface {
	// PreviewImport 计算表格与模型数据差异，不写入数据
	PreviewImport(ctx context.Context, modelUID string, fileData []byte) (ImportPreview, error)

	// Import 按预览结果同步资源实例
	// modelUID: 模型唯一标识 (对应 Model.UID)
	// fileData: Excel 文件的字节数据
	Import(ctx context.Context, modelUID string, fileData []byte, confirmEmpty bool) (ImportResult, error)

	// Export 导出资源实例数据 (Resource)
	// req: 导出请求参数
	Export(ctx context.Context, req ExportParams) ([]byte, error)

	// ExportTemplate 导出模板
	// modelUID: 模型唯一标识 (对应 Model.UID)
	ExportTemplate(ctx context.Context, modelUID string) ([]byte, error)
}

type ImportAction string

const (
	ImportActionCreate    ImportAction = "create"
	ImportActionUpdate    ImportAction = "update"
	ImportActionDelete    ImportAction = "delete"
	ImportActionUnchanged ImportAction = "unchanged"
)

type ImportChange struct {
	UniqueID      string                 `json:"unique_id"`
	Action        ImportAction           `json:"action"`
	ChangedFields []string               `json:"changed_fields"`
	OriginalData  map[string]interface{} `json:"original_data"`
	ModifiedData  map[string]interface{} `json:"modified_data"`
	Resource      resource.Resource      `json:"-"`
}

type ImportPreview struct {
	ModelUID       string         `json:"model_uid"`
	UniqueField    string         `json:"unique_field"`
	SheetCount     int            `json:"sheet_count"`
	CurrentCount   int            `json:"current_count"`
	CreatedCount   int            `json:"created_count"`
	UpdatedCount   int            `json:"updated_count"`
	DeletedCount   int            `json:"deleted_count"`
	UnchangedCount int            `json:"unchanged_count"`
	IsEmpty        bool           `json:"is_empty"`
	Columns        []string       `json:"columns"`
	Rows           []ImportChange `json:"rows"`
}

type ImportResult struct {
	ImportedCount  int            `json:"imported_count"`
	CreatedCount   int            `json:"created_count"`
	UpdatedCount   int            `json:"updated_count"`
	DeletedCount   int            `json:"deleted_count"`
	UnchangedCount int            `json:"unchanged_count"`
	Changes        []ImportChange `json:"-"`
}

type ExportParams struct {
	ModelUID      string
	Scope         string // "all", "current", "selected"
	ResourceIDs   []int64
	FilterGroups  []resource.FilterGroup
	Fields        []string
	RelatedFields []RelatedFieldParam
	FileName      string
}

type RelatedFieldParam struct {
	RelationName string
	ModelUID     string
	FieldUID     string
}
