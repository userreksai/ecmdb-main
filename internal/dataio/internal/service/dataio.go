package service

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/Duke1616/ecmdb/internal/attribute"
	"github.com/Duke1616/ecmdb/internal/dataio/internal/domain"
	"github.com/Duke1616/ecmdb/internal/model"
	"github.com/Duke1616/ecmdb/internal/relation"
	"github.com/Duke1616/ecmdb/internal/resource"
	"github.com/ecodeclub/ekit/slice"
	"github.com/xuri/excelize/v2"
	"golang.org/x/sync/errgroup"
)

// fieldPriority 字段优先级配置
// NOTE: 用于控制 Excel 导出时的列顺序,数字越小越靠前,未配置的字段默认为 999
var fieldPriority = map[string]int{
	"name": 1, // name 字段始终在第一列
	// 可以继续添加其他需要固定顺序的字段
	// "ip":   2,
	// "port": 3,
}

// sortAttributesByPriority 按优先级排序字段
func sortAttributesByPriority(attrs []attribute.Attribute) []attribute.Attribute {
	sorted := make([]attribute.Attribute, len(attrs))
	copy(sorted, attrs)

	sort.SliceStable(sorted, func(i, j int) bool {
		pi := fieldPriority[sorted[i].FieldUid]
		pj := fieldPriority[sorted[j].FieldUid]
		if pi == 0 {
			pi = 999 // 未配置的字段默认优先级
		}
		if pj == 0 {
			pj = 999
		}
		return pi < pj
	})

	return sorted
}

// NOTE: dataIOService 实现数据交换功能,依赖三个模块的 Service
type relatedExportColumn struct {
	Key          string
	RelationName string
	ModelUID     string
	FieldUID     string
	Attr         attribute.Attribute
}

type dataIOService struct {
	attrSvc  attribute.Service
	resSvc   resource.EncryptedSvc
	modelSvc model.Service
	rmSvc    relation.RMSvc
	rrSvc    relation.RRSvc
}

// NewDataIOService 创建数据交换服务实例
func NewDataIOService(
	attrSvc attribute.Service,
	resSvc resource.EncryptedSvc,
	modelSvc model.Service,
	rmSvc relation.RMSvc,
	rrSvc relation.RRSvc,
) IDataIOService {
	return &dataIOService{
		attrSvc:  attrSvc,
		resSvc:   resSvc,
		modelSvc: modelSvc,
		rmSvc:    rmSvc,
		rrSvc:    rrSvc,
	}
}

type parsedImportSheet struct {
	Columns   []string
	Fields    []string
	Resources []resource.Resource
}

// PreviewImport 计算表格与当前模型的差异，不执行写入。
func (s *dataIOService) PreviewImport(ctx context.Context, modelUID string, fileData []byte) (ImportPreview, error) {
	sheet, err := s.parseImportSheet(ctx, modelUID, fileData)
	if err != nil {
		return ImportPreview{}, err
	}

	current, err := s.listAllResources(ctx, modelUID, sheet.Fields)
	if err != nil {
		return ImportPreview{}, err
	}
	return buildImportPreview(modelUID, sheet, current), nil
}

// Import 按唯一索引 name 同步表格与模型数据。
func (s *dataIOService) Import(ctx context.Context, modelUID string, fileData []byte, confirmEmpty bool) (ImportResult, error) {
	preview, err := s.PreviewImport(ctx, modelUID, fileData)
	if err != nil {
		return ImportResult{}, err
	}
	if preview.IsEmpty && !confirmEmpty {
		return ImportResult{}, fmt.Errorf("表格数据为空，继续导入将删除当前模型全部数据，请确认后重试")
	}

	upserts := make([]resource.Resource, 0, preview.CreatedCount+preview.UpdatedCount)
	deletes := make([]resource.Resource, 0, preview.DeletedCount)
	for _, change := range preview.Rows {
		switch change.Action {
		case ImportActionCreate, ImportActionUpdate:
			upserts = append(upserts, change.Resource)
		case ImportActionDelete:
			deletes = append(deletes, change.Resource)
		}
	}

	if err = s.resSvc.BatchCreateOrUpdate(ctx, upserts); err != nil {
		return ImportResult{}, fmt.Errorf("批量创建或更新资源失败: %w", err)
	}
	for _, item := range deletes {
		if _, err = s.resSvc.DeleteResource(ctx, item.ID); err != nil {
			return ImportResult{}, fmt.Errorf("删除表格外资源 %d 失败: %w", item.ID, err)
		}
		if s.rrSvc != nil {
			if _, err = s.rrSvc.DeleteRelationsByResourceID(ctx, item.ID); err != nil {
				return ImportResult{}, fmt.Errorf("删除资源 %d 关联关系失败: %w", item.ID, err)
			}
		}
	}

	return ImportResult{
		ImportedCount:  preview.CreatedCount + preview.UpdatedCount,
		CreatedCount:   preview.CreatedCount,
		UpdatedCount:   preview.UpdatedCount,
		DeletedCount:   preview.DeletedCount,
		UnchangedCount: preview.UnchangedCount,
		Changes:        preview.Rows,
	}, nil
}

func (s *dataIOService) parseImportSheet(ctx context.Context, modelUID string, fileData []byte) (parsedImportSheet, error) {
	f, err := excelize.OpenReader(bytes.NewReader(fileData))
	if err != nil {
		return parsedImportSheet{}, fmt.Errorf("解析 Excel 文件失败: %w", err)
	}
	defer f.Close()

	attrs, _, err := s.attrSvc.ListAttributes(ctx, modelUID)
	if err != nil {
		return parsedImportSheet{}, fmt.Errorf("获取模型字段定义失败: %w", err)
	}
	if len(attrs) == 0 {
		return parsedImportSheet{}, fmt.Errorf("模型 %s 没有定义字段", modelUID)
	}
	attributeMap := slice.ToMap(attrs, func(attr attribute.Attribute) string { return attr.FieldUid })
	fields := slice.Map(attrs, func(_ int, attr attribute.Attribute) string { return attr.FieldUid })

	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return parsedImportSheet{}, fmt.Errorf("读取 Excel 数据失败: %w", err)
	}
	if len(rows) < 3 {
		return parsedImportSheet{}, fmt.Errorf("excel 文件格式错误，至少需要导出文件中的 3 行表头")
	}

	fieldUIDRow := rows[1]
	colIndexMap := make(map[int]string)
	columns := make([]string, 0, len(fieldUIDRow))
	hasNameField := false
	seenColumns := make(map[string]struct{}, len(fieldUIDRow))
	for colIdx, fieldUID := range fieldUIDRow {
		fieldUID = strings.TrimSpace(fieldUID)
		if fieldUID != "" {
			if _, ok := attributeMap[fieldUID]; !ok {
				return parsedImportSheet{}, fmt.Errorf("表格包含模型中不存在的字段: %s", fieldUID)
			}
			if _, exists := seenColumns[fieldUID]; exists {
				return parsedImportSheet{}, fmt.Errorf("表格字段重复: %s", fieldUID)
			}
			seenColumns[fieldUID] = struct{}{}
			colIndexMap[colIdx] = fieldUID
			columns = append(columns, fieldUID)
			if fieldUID == "name" {
				hasNameField = true
			}
		}
	}
	if !hasNameField {
		return parsedImportSheet{}, fmt.Errorf("表格缺少唯一索引字段: name")
	}

	resources := make([]resource.Resource, 0, len(rows)-3)
	seenNames := make(map[string]int, len(rows)-3)
	for rowIdx, row := range rows[3:] {
		rowHasValue := false
		for _, value := range row {
			if strings.TrimSpace(value) != "" {
				rowHasValue = true
				break
			}
		}
		if !rowHasValue {
			continue
		}
		data := make(map[string]interface{})
		for colIdx, fieldUID := range colIndexMap {
			cellValue := ""
			if colIdx < len(row) {
				cellValue = strings.TrimSpace(row[colIdx])
			}
			data[fieldUID] = cellValue
		}
		name, ok := data["name"]
		nameStr := strings.TrimSpace(fmt.Sprint(name))
		if !ok || nameStr == "" {
			return parsedImportSheet{}, fmt.Errorf("excel 第 %d 行缺少唯一索引字段 name", rowIdx+4)
		}
		if firstRow, exists := seenNames[nameStr]; exists {
			return parsedImportSheet{}, fmt.Errorf("excel 第 %d 行唯一索引 %q 重复，首次出现在第 %d 行", rowIdx+4, nameStr, firstRow)
		}
		seenNames[nameStr] = rowIdx + 4
		resources = append(resources, resource.Resource{
			ModelUID: modelUID,
			Data:     data,
		})
	}
	return parsedImportSheet{Columns: columns, Fields: fields, Resources: resources}, nil
}

func (s *dataIOService) listAllResources(ctx context.Context, modelUID string, fields []string) ([]resource.Resource, error) {
	const limit int64 = 500
	all := make([]resource.Resource, 0)
	for offset := int64(0); ; offset += limit {
		items, total, err := s.resSvc.ListResourcesWithFilters(ctx, fields, modelUID, nil, offset, limit, nil)
		if err != nil {
			return nil, fmt.Errorf("获取模型现有数据失败: %w", err)
		}
		all = append(all, items...)
		if int64(len(all)) >= total || len(items) == 0 {
			break
		}
	}
	return all, nil
}

func buildImportPreview(modelUID string, sheet parsedImportSheet, current []resource.Resource) ImportPreview {
	preview := ImportPreview{
		ModelUID:     modelUID,
		UniqueField:  "name",
		SheetCount:   len(sheet.Resources),
		CurrentCount: len(current),
		IsEmpty:      len(sheet.Resources) == 0,
		Columns:      sheet.Columns,
		Rows:         make([]ImportChange, 0, len(sheet.Resources)+len(current)),
	}
	currentByName := make(map[string]resource.Resource, len(current))
	for _, item := range current {
		currentByName[strings.TrimSpace(fmt.Sprint(item.Data["name"]))] = item
	}
	seen := make(map[string]struct{}, len(sheet.Resources))

	for _, incoming := range sheet.Resources {
		name := strings.TrimSpace(fmt.Sprint(incoming.Data["name"]))
		seen[name] = struct{}{}
		existing, exists := currentByName[name]
		if !exists {
			preview.CreatedCount++
			preview.Rows = append(preview.Rows, ImportChange{
				UniqueID:      name,
				Action:        ImportActionCreate,
				ChangedFields: sortedMapKeys(incoming.Data),
				ModifiedData:  cloneData(incoming.Data),
				Resource:      incoming,
			})
			continue
		}

		merged := cloneData(existing.Data)
		changedFields := make([]string, 0, len(incoming.Data))
		for field, value := range incoming.Data {
			if old, ok := existing.Data[field]; !ok || !reflect.DeepEqual(old, value) {
				changedFields = append(changedFields, field)
			}
			merged[field] = value
		}
		sort.Strings(changedFields)
		action := ImportActionUnchanged
		if len(changedFields) > 0 {
			action = ImportActionUpdate
			preview.UpdatedCount++
		} else {
			preview.UnchangedCount++
		}
		incoming.ID = existing.ID
		preview.Rows = append(preview.Rows, ImportChange{
			UniqueID:      name,
			Action:        action,
			ChangedFields: changedFields,
			OriginalData:  cloneData(existing.Data),
			ModifiedData:  merged,
			Resource:      incoming,
		})
	}

	for _, existing := range current {
		name := strings.TrimSpace(fmt.Sprint(existing.Data["name"]))
		if _, ok := seen[name]; ok {
			continue
		}
		preview.DeletedCount++
		preview.Rows = append(preview.Rows, ImportChange{
			UniqueID:      name,
			Action:        ImportActionDelete,
			ChangedFields: sortedMapKeys(existing.Data),
			OriginalData:  cloneData(existing.Data),
			Resource:      existing,
		})
	}
	return preview
}

func cloneData(src map[string]interface{}) map[string]interface{} {
	if src == nil {
		return nil
	}
	dst := make(map[string]interface{}, len(src))
	for key, value := range src {
		dst[key] = value
	}
	return dst
}

func sortedMapKeys(src map[string]interface{}) []string {
	keys := make([]string, 0, len(src))
	for key := range src {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// Export 导出资源实例数据 (Resource)
func (s *dataIOService) Export(ctx context.Context, req ExportParams) ([]byte, error) {
	// 1. 获取数据定义
	mdl, attrs, err := s.fetchModelAndAttributes(ctx, req.ModelUID)
	if err != nil {
		return nil, err
	}

	// 2. 处理字段过滤: 未传 fields 时默认导出全部字段; 传空数组表示不导出当前模型字段
	if req.Fields != nil {
		reqFieldMap := slice.ToMap(req.Fields, func(f string) string { return f })
		attrs = slice.FilterMap(attrs, func(idx int, src attribute.Attribute) (attribute.Attribute, bool) {
			_, ok := reqFieldMap[src.FieldUid]
			return src, ok
		})
	}

	// 3. 对字段进行排序 (根据 fieldPriority)
	sortedAttrs := sortAttributesByPriority(attrs)

	relatedColumns, err := s.buildRelatedExportColumns(ctx, req.ModelUID, req.RelatedFields)
	if err != nil {
		return nil, err
	}

	// 4. 提取最终的 FieldUIDs 用于查询
	dstFields := slice.Map(sortedAttrs, func(idx int, attr attribute.Attribute) string {
		return attr.FieldUid
	})

	var allResources []resource.Resource
	offset := int64(0)
	limit := int64(100)

	for {
		resources, _, err1 := s.resSvc.ListResourcesWithFilters(ctx, dstFields, req.ModelUID, req.ResourceIDs, offset, limit, req.FilterGroups)
		if err1 != nil {
			return nil, fmt.Errorf("获取资源列表失败: %w", err1)
		}
		allResources = append(allResources, resources...)

		if len(resources) < int(limit) {
			break
		}
		offset += limit
	}

	if len(relatedColumns) > 0 {
		allResources, err = s.fillRelatedExportValues(ctx, req.ModelUID, allResources, relatedColumns)
		if err != nil {
			return nil, err
		}
	}

	exportAttrs := make([]attribute.Attribute, 0, len(sortedAttrs)+len(relatedColumns))
	exportAttrs = append(exportAttrs, sortedAttrs...)
	for _, col := range relatedColumns {
		exportAttrs = append(exportAttrs, col.Attr)
	}

	// 5. 构建 Excel
	return s.buildExcel(mdl.SheetName(), exportAttrs, allResources)
}

func (s *dataIOService) buildRelatedExportColumns(ctx context.Context, modelUID string, fields []RelatedFieldParam) ([]relatedExportColumn, error) {
	if len(fields) == 0 {
		return nil, nil
	}

	modelRelations, err := s.listAllModelRelations(ctx, modelUID)
	if err != nil {
		return nil, err
	}

	relationMap := make(map[string]relation.ModelRelation, len(modelRelations))
	relatedModelUIDSet := make(map[string]struct{}, len(modelRelations))
	for _, rel := range modelRelations {
		relationMap[rel.RelationName] = rel
		if rel.SourceModelUID == modelUID {
			relatedModelUIDSet[rel.TargetModelUID] = struct{}{}
		}
		if rel.TargetModelUID == modelUID {
			relatedModelUIDSet[rel.SourceModelUID] = struct{}{}
		}
	}

	relatedModelUIDs := stringSetKeys(relatedModelUIDSet)
	modelsByUID, err := s.getModelsByUID(ctx, relatedModelUIDs)
	if err != nil {
		return nil, err
	}

	attrsByModel := make(map[string]map[string]attribute.Attribute, len(relatedModelUIDs))
	for _, uid := range relatedModelUIDs {
		attrs, _, err1 := s.attrSvc.ListAttributes(ctx, uid)
		if err1 != nil {
			return nil, fmt.Errorf("获取关联模型字段失败: %w", err1)
		}
		attrsByModel[uid] = slice.ToMap(attrs, func(attr attribute.Attribute) string {
			return attr.FieldUid
		})
	}

	columns := make([]relatedExportColumn, 0, len(fields))
	seen := make(map[string]struct{}, len(fields))
	for _, field := range fields {
		relationName := strings.TrimSpace(field.RelationName)
		fieldUID := strings.TrimSpace(field.FieldUID)
		if relationName == "" || fieldUID == "" {
			return nil, fmt.Errorf("关联导出字段缺少 relation_name 或 field_uid")
		}

		rel, ok := relationMap[relationName]
		if !ok {
			return nil, fmt.Errorf("当前模型不存在关联关系: %s", relationName)
		}

		relatedModelUID := ""
		if rel.SourceModelUID == modelUID {
			relatedModelUID = rel.TargetModelUID
		} else if rel.TargetModelUID == modelUID {
			relatedModelUID = rel.SourceModelUID
		} else {
			return nil, fmt.Errorf("关联关系 %s 不属于当前模型 %s", relationName, modelUID)
		}

		if field.ModelUID != "" && field.ModelUID != relatedModelUID {
			return nil, fmt.Errorf("关联字段模型不匹配: relation=%s model=%s", relationName, field.ModelUID)
		}

		attrMap := attrsByModel[relatedModelUID]
		attr, ok := attrMap[fieldUID]
		if !ok {
			return nil, fmt.Errorf("关联模型 %s 不存在字段: %s", relatedModelUID, fieldUID)
		}

		key := relatedFieldKey(relationName, relatedModelUID, fieldUID)
		if _, ok = seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}

		modelName := relatedModelUID
		if mdl, ok := modelsByUID[relatedModelUID]; ok && mdl.Name != "" {
			modelName = mdl.Name
		}

		exportAttr := attr
		exportAttr.FieldUid = key
		exportAttr.FieldName = fmt.Sprintf("%s/%s/%s", relationName, modelName, attr.FieldName)
		exportAttr.Required = false
		exportAttr.Secure = false

		columns = append(columns, relatedExportColumn{
			Key:          key,
			RelationName: relationName,
			ModelUID:     relatedModelUID,
			FieldUID:     fieldUID,
			Attr:         exportAttr,
		})
	}

	return columns, nil
}

func (s *dataIOService) fillRelatedExportValues(ctx context.Context, modelUID string, resources []resource.Resource, columns []relatedExportColumn) ([]resource.Resource, error) {
	if len(resources) == 0 || len(columns) == 0 {
		return resources, nil
	}

	resourceIDs := make([]int64, 0, len(resources))
	resourceIDSet := make(map[int64]struct{}, len(resources))
	for _, res := range resources {
		if _, ok := resourceIDSet[res.ID]; ok {
			continue
		}
		resourceIDSet[res.ID] = struct{}{}
		resourceIDs = append(resourceIDs, res.ID)
	}

	relations, err := s.rrSvc.ListByResourceIDs(ctx, modelUID, resourceIDs)
	if err != nil {
		return nil, fmt.Errorf("获取关联资源关系失败: %w", err)
	}

	columnsByRelation := make(map[string][]relatedExportColumn, len(columns))
	for _, col := range columns {
		columnsByRelation[col.RelationName] = append(columnsByRelation[col.RelationName], col)
	}

	targetsByRelation := make(map[string]map[int64][]int64, len(columnsByRelation))
	modelTargetIDs := make(map[string]map[int64]struct{})
	modelFields := make(map[string]map[string]struct{})

	for _, rel := range relations {
		cols, ok := columnsByRelation[rel.RelationName]
		if !ok {
			continue
		}

		var currentID, relatedID int64
		relatedModelUID := ""
		if rel.SourceModelUID == modelUID {
			currentID = rel.SourceResourceID
			relatedID = rel.TargetResourceID
			relatedModelUID = rel.TargetModelUID
		} else if rel.TargetModelUID == modelUID {
			currentID = rel.TargetResourceID
			relatedID = rel.SourceResourceID
			relatedModelUID = rel.SourceModelUID
		} else {
			continue
		}

		if _, ok = resourceIDSet[currentID]; !ok {
			continue
		}

		if targetsByRelation[rel.RelationName] == nil {
			targetsByRelation[rel.RelationName] = make(map[int64][]int64)
		}
		targetsByRelation[rel.RelationName][currentID] = append(targetsByRelation[rel.RelationName][currentID], relatedID)

		for _, col := range cols {
			if col.ModelUID != relatedModelUID {
				continue
			}
			if modelTargetIDs[col.ModelUID] == nil {
				modelTargetIDs[col.ModelUID] = make(map[int64]struct{})
			}
			if modelFields[col.ModelUID] == nil {
				modelFields[col.ModelUID] = make(map[string]struct{})
			}
			modelTargetIDs[col.ModelUID][relatedID] = struct{}{}
			modelFields[col.ModelUID][col.FieldUID] = struct{}{}
		}
	}

	relatedResourcesByModel := make(map[string]map[int64]resource.Resource, len(modelTargetIDs))
	for relatedModelUID, idSet := range modelTargetIDs {
		ids := int64SetKeys(idSet)
		fields := stringSetKeys(modelFields[relatedModelUID])
		relatedResources, err1 := s.resSvc.ListResourceByIds(ctx, fields, ids)
		if err1 != nil {
			return nil, fmt.Errorf("获取关联资源数据失败: %w", err1)
		}

		relatedResourcesByModel[relatedModelUID] = make(map[int64]resource.Resource, len(relatedResources))
		for _, relatedResource := range relatedResources {
			relatedResourcesByModel[relatedModelUID][relatedResource.ID] = relatedResource
		}
	}

	for idx := range resources {
		if resources[idx].Data == nil {
			resources[idx].Data = make(map[string]interface{})
		}
		for _, col := range columns {
			relatedIDs := targetsByRelation[col.RelationName][resources[idx].ID]
			values := make([]string, 0, len(relatedIDs))
			for _, relatedID := range relatedIDs {
				relatedResource, ok := relatedResourcesByModel[col.ModelUID][relatedID]
				if !ok {
					continue
				}
				val, ok := relatedResource.Data[col.FieldUID]
				if !ok || val == nil {
					continue
				}
				strVal := formatExportValue(val)
				if strVal == "" {
					continue
				}
				values = append(values, strVal)
			}
			resources[idx].Data[col.Key] = strings.Join(values, ", ")
		}
	}

	return resources, nil
}

func (s *dataIOService) listAllModelRelations(ctx context.Context, modelUID string) ([]relation.ModelRelation, error) {
	const limit int64 = 200
	offset := int64(0)
	result := make([]relation.ModelRelation, 0)

	for {
		relations, total, err := s.rmSvc.ListModelUidRelation(ctx, offset, limit, modelUID)
		if err != nil {
			return nil, fmt.Errorf("获取模型关联关系失败: %w", err)
		}
		result = append(result, relations...)
		if int64(len(relations)) < limit || int64(len(result)) >= total {
			break
		}
		offset += limit
	}

	return result, nil
}

func (s *dataIOService) getModelsByUID(ctx context.Context, uids []string) (map[string]model.Model, error) {
	if len(uids) == 0 {
		return map[string]model.Model{}, nil
	}

	models, err := s.modelSvc.GetByUids(ctx, uids)
	if err != nil {
		return nil, fmt.Errorf("获取关联模型信息失败: %w", err)
	}

	return slice.ToMap(models, func(mdl model.Model) string {
		return mdl.UID
	}), nil
}

func relatedFieldKey(relationName, modelUID, fieldUID string) string {
	return fmt.Sprintf("__related__%s__%s__%s", relationName, modelUID, fieldUID)
}

func stringSetKeys(src map[string]struct{}) []string {
	keys := make([]string, 0, len(src))
	for key := range src {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func int64SetKeys(src map[int64]struct{}) []int64 {
	keys := make([]int64, 0, len(src))
	for key := range src {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	return keys
}

func formatExportValue(val interface{}) string {
	switch v := val.(type) {
	case []string:
		return strings.Join(v, ", ")
	case []interface{}:
		parts := make([]string, 0, len(v))
		for _, item := range v {
			parts = append(parts, fmt.Sprint(item))
		}
		return strings.Join(parts, ", ")
	default:
		return fmt.Sprint(v)
	}
}

// ExportTemplate 导出空白导入模板
func (s *dataIOService) ExportTemplate(ctx context.Context, modelUID string) ([]byte, error) {
	// 1. 获取数据
	mdl, attrs, err := s.fetchModelAndAttributes(ctx, modelUID)
	if err != nil {
		return nil, err
	}

	// 2. 按优先级排序字段
	sortedAttrs := sortAttributesByPriority(attrs)

	// 3. 构建 Excel (空数据)
	return s.buildExcel(mdl.SheetName(), sortedAttrs, nil)
}

// buildExcel 构建 Excel 文件
// NOTE: 通用方法,用于导出数据和导出模板
func (s *dataIOService) buildExcel(sheetName string, attrs []attribute.Attribute, resources []resource.Resource) ([]byte, error) {
	// 1. 构建 3 行表头数据
	row1 := make([]string, len(attrs)) // 字段约束
	row2 := make([]string, len(attrs)) // 字段 UID
	row3 := make([]string, len(attrs)) // 字段名称

	for i, attr := range attrs {
		row1[i] = attr.GetConstraintDescription()
		row2[i] = attr.FieldUid
		row3[i] = attr.FieldName
	}

	// 2. 创建 Builder
	builder := domain.NewBuilder(sheetName).
		With3RowHeaders(row1, row2, row3)
	defer builder.Close()

	// 3. 填充数据
	for _, res := range resources {
		row := make([]interface{}, len(attrs))
		for i, attr := range attrs {
			if val, ok := res.Data[attr.FieldUid]; ok {
				row[i] = val
			} else {
				row[i] = ""
			}
		}
		builder.AddRow(row...)
	}

	// 4. 添加数据验证(下拉列表)
	// NOTE: 如果是导出模板,验证范围预留 1000 行
	// 如果是导出数据,验证范围覆盖所有数据行 + 100 行缓冲
	validationRows := 1000
	if len(resources) > 0 {
		validationRows = len(resources) + 100
	}

	for colIdx, attr := range attrs {
		if attr.NeedsValidation() {
			builder.WithValidation(colIdx, attr.GetOptionStrings(), 4, validationRows)
		}
	}

	// 5. 导出字节数据
	return builder.ToBytes()
}

func (s *dataIOService) fetchModelAndAttributes(ctx context.Context, modelUID string) (model.Model, []attribute.Attribute, error) {
	var (
		mdl   model.Model
		attrs []attribute.Attribute
		eg    errgroup.Group
	)

	// 1. 并行获取 Model 信息和 Attribute 定义
	eg.Go(func() error {
		var err error
		mdl, err = s.modelSvc.GetByUid(ctx, modelUID)
		if err != nil {
			return fmt.Errorf("获取模型信息失败: %w", err)
		}
		return nil
	})

	eg.Go(func() error {
		var err error
		var total int64
		attrs, total, err = s.attrSvc.ListAttributes(ctx, modelUID)
		if err != nil {
			return fmt.Errorf("获取模型字段定义失败: %w", err)
		}
		if total == 0 {
			return fmt.Errorf("模型 %s 没有定义字段", modelUID)
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		return model.Model{}, nil, err
	}

	return mdl, attrs, nil
}
