package dao

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Duke1616/ecmdb/internal/resource/internal/domain"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const ResourceCollection = "c_resources"

type ResourceDAO interface {
	// CreateResource 创建资产
	CreateResource(ctx context.Context, resource Resource) (int64, error)

	// FindResourceById 根据 ID 和字段列表查询资产
	FindResourceById(ctx context.Context, fields []string, id int64) (Resource, error)

	// ListResource 获取指定模型的资产列表
	ListResource(ctx context.Context, fields []string, modelUid string, offset, limit int64) ([]Resource, error)

	// SearchResourcesInModel 多个搜索条件使用 AND 组合，每个条件可选择精准或模糊匹配
	SearchResourcesInModel(ctx context.Context, fields []string, modelUid string, conditions []domain.SearchCondition, offset, limit int64) (
		[]Resource, int64, error)

	// CountByModelUid 统计指定模型的资产数量
	CountByModelUid(ctx context.Context, modelUid string) (int64, error)

	// SetCustomField 设置指定资产的自定义字段值
	SetCustomField(ctx context.Context, id int64, field string, data interface{}) (int64, error)

	// ListResourcesByIds 根据 ID 列表批量查询资产
	ListResourcesByIds(ctx context.Context, fields []string, ids []int64) ([]Resource, error)

	// DeleteResource 删除指定资产
	DeleteResource(ctx context.Context, id int64) (int64, error)

	// ListExcludeAndFilterResourceByIds 排除指定 ID 并根据条件过滤资产列表
	ListExcludeAndFilterResourceByIds(ctx context.Context, fields []string, modelUid string, offset, limit int64,
		ids []int64, filter domain.Condition) ([]Resource, error)

	// TotalExcludeAndFilterResourceByIds 排除指定 ID 并根据条件统计资产总数
	TotalExcludeAndFilterResourceByIds(ctx context.Context, modelUid string, ids []int64,
		filter domain.Condition) (int64, error)

	// Search 全局搜索资产
	Search(ctx context.Context, text string) ([]SearchResource, error)

	// FindSecureData 查找指定资产的加密字段数据
	FindSecureData(ctx context.Context, id int64, fieldUid string) (string, error)

	// UpdateAttribute 更新资产属性
	UpdateAttribute(ctx context.Context, resource Resource) (int64, error)

	// CountByModelUids 统计多个模型的资产数量
	CountByModelUids(ctx context.Context, modelUids []string) (map[string]int, error)

	// BatchUpdateResources 批量更新资产
	BatchUpdateResources(ctx context.Context, resources []Resource) (int64, error)

	// BatchCreateOrUpdate 批量创建或更新资产,基于 model_uid + name 进行 upsert
	BatchCreateOrUpdate(ctx context.Context, resources []Resource) error

	// ListBeforeUtime 获取指定时间前的资产列表
	ListBeforeUtime(ctx context.Context, utime int64, fields []string, modelUid string,
		offset, limit int64) ([]Resource, error)

	// ListResourcesWithFilters 根据复杂筛选条件获取资产列表
	ListResourcesWithFilters(ctx context.Context, fields []string, modelUid string, ids []int64, offset, limit int64,
		filterGroups []domain.FilterGroup) ([]Resource, error)

	// TotalResourcesWithFilters 根据复杂筛选条件统计资产数量
	TotalResourcesWithFilters(ctx context.Context, modelUid string, ids []int64, filterGroups []domain.FilterGroup) (int64, error)
}

type resourceDAO struct {
	db *mongox.Mongo
}

func NewResourceDAO(db *mongox.Mongo) ResourceDAO {
	return &resourceDAO{
		db: db,
	}
}

func (dao *resourceDAO) ListBeforeUtime(ctx context.Context, utime int64, fields []string, modelUid string,
	offset, limit int64) ([]Resource, error) {
	col := dao.db.Collection(ResourceCollection)
	filter := bson.M{"model_uid": modelUid}
	filter["utime"] = bson.M{"$lte": utime}
	projection := buildProjection(fields)
	opts := &options.FindOptions{
		Projection: projection,
		Limit:      &limit,
		Skip:       &offset,
		Sort:       bson.D{{Key: "ctime", Value: -1}},
	}

	cursor, err := col.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("查询错误: %w", err)
	}

	var result []Resource
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *resourceDAO) BatchUpdateResources(ctx context.Context, resources []Resource) (int64, error) {
	if len(resources) == 0 {
		return 0, nil
	}

	col := dao.db.Collection(ResourceCollection)
	var totalModified int64

	// 为批量操作创建切片
	models := make([]mongo.WriteModel, 0, len(resources))

	utime := time.Now().UnixMilli()
	for _, resource := range resources {
		updateDoc := bson.M{
			"utime": utime,
		}

		// 将资源数据合并到更新文档中
		for key, value := range resource.Data {
			updateDoc[key] = value
		}

		// 创建更新模型
		filter := bson.M{"id": resource.ID}
		update := bson.M{"$set": updateDoc}

		model := mongo.NewUpdateOneModel().
			SetFilter(filter).
			SetUpdate(update).
			SetUpsert(false)

		models = append(models, model)
	}

	// 执行批量操作
	result, err := col.BulkWrite(ctx, models)
	if err != nil {
		return 0, fmt.Errorf("批量更新文档操作: %w", err)
	}

	totalModified = result.ModifiedCount
	return totalModified, nil
}

func (dao *resourceDAO) SetCustomField(ctx context.Context, id int64, field string, data interface{}) (int64, error) {
	col := dao.db.Collection(ResourceCollection)

	updateDoc := bson.M{
		"$set": bson.M{
			field:   data,
			"utime": time.Now().UnixMilli(),
		},
	}

	filter := bson.M{"id": id}
	count, err := col.UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		return 0, fmt.Errorf("修改文档操作: %w", err)
	}

	return count.ModifiedCount, nil
}

func (dao *resourceDAO) UpdateAttribute(ctx context.Context, resource Resource) (int64, error) {
	col := dao.db.Collection(ResourceCollection)

	updateDoc := bson.M{
		"utime": time.Now().UnixMilli(),
	}

	for key, value := range resource.Data {
		updateDoc[key] = value
	}

	// 构建最终的更新文档
	updateCommand := bson.M{
		"$set": updateDoc,
	}

	filter := bson.M{"id": resource.ID}
	count, err := col.UpdateOne(ctx, filter, updateCommand)
	if err != nil {
		return 0, fmt.Errorf("修改文档操作: %w", err)
	}

	return count.ModifiedCount, nil
}

func (dao *resourceDAO) FindSecureData(ctx context.Context, id int64, fieldUid string) (string, error) {
	col := dao.db.Collection(ResourceCollection)
	filter := bson.M{"id": id}
	projection := make(map[string]int, 1)
	projection[fieldUid] = 1
	opts := &options.FindOneOptions{
		Projection: projection,
	}

	var result = make(map[string]string)

	if err := col.FindOne(ctx, filter, opts).Decode(&result); err != nil {
		return "", fmt.Errorf("解码错误: %w", err)
	}

	fieldValue, ok := result[fieldUid]

	if !ok {
		return "无", nil
	}

	return fieldValue, nil
}

func (dao *resourceDAO) CreateResource(ctx context.Context, r Resource) (int64, error) {
	now := time.Now()
	r.Ctime, r.Utime = now.UnixMilli(), now.UnixMilli()
	r.ID = dao.db.GetIdGenerator(ResourceCollection)
	col := dao.db.Collection(ResourceCollection)

	_, err := col.InsertOne(ctx, r)
	if err != nil {
		return 0, fmt.Errorf("插入数据错误: %w", err)
	}

	return r.ID, nil
}

func (dao *resourceDAO) FindResourceById(ctx context.Context, fields []string, id int64) (Resource, error) {
	col := dao.db.Collection(ResourceCollection)
	filter := bson.M{"id": id}
	projection := buildProjection(fields)
	opts := &options.FindOneOptions{
		Projection: projection,
	}

	var result Resource
	if err := col.FindOne(ctx, filter, opts).Decode(&result); err != nil {
		return Resource{}, fmt.Errorf("解码错误: %w", err)
	}
	return result, nil
}

func (dao *resourceDAO) ListResource(ctx context.Context, fields []string, modelUid string, offset, limit int64) ([]Resource, error) {
	col := dao.db.Collection(ResourceCollection)
	filter := bson.M{"model_uid": modelUid}
	projection := buildProjection(fields)
	opts := &options.FindOptions{
		Projection: projection,
		Limit:      &limit,
		Skip:       &offset,
		Sort:       bson.D{{Key: "ctime", Value: -1}},
	}

	cursor, err := col.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("查询错误: %w", err)
	}

	var result []Resource
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *resourceDAO) SearchResourcesInModel(ctx context.Context, fields []string, modelUid string, conditions []domain.SearchCondition, offset, limit int64) (
	[]Resource, int64, error) {
	col := dao.db.Collection(ResourceCollection)
	searchFields := normalizeSearchFields(fields)
	projection := buildProjection(fields)
	searchFilter := buildResourceSearchFilter(modelUid, searchFields, conditions)
	total, err := col.CountDocuments(ctx, searchFilter)
	if err != nil {
		return nil, 0, fmt.Errorf("搜索匹配计数错误: %w", err)
	}
	resources, err := dao.findResourcesByFilter(ctx, searchFilter, projection, offset, limit)
	return resources, total, err
}

func (dao *resourceDAO) findResourcesByFilter(ctx context.Context, filter interface{}, projection bson.M, offset, limit int64) ([]Resource, error) {
	col := dao.db.Collection(ResourceCollection)
	opts := &options.FindOptions{
		Projection: projection,
		Limit:      &limit,
		Skip:       &offset,
		Sort:       bson.D{{Key: "ctime", Value: -1}},
	}

	cursor, err := col.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("查询错误: %w", err)
	}

	var result []Resource
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *resourceDAO) CountByModelUid(ctx context.Context, modelUid string) (int64, error) {
	col := dao.db.Collection(ResourceCollection)
	filter := bson.M{"model_uid": modelUid}

	count, err := col.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("文档计数错误: %w", err)
	}

	return count, nil
}

func (dao *resourceDAO) ListResourcesByIds(ctx context.Context, fields []string, ids []int64) ([]Resource, error) {
	col := dao.db.Collection(ResourceCollection)
	filter := bson.M{"id": bson.M{"$in": ids}}
	projection := make(map[string]int, len(fields))
	for _, v := range fields {
		projection[v] = 1
	}
	projection["_id"] = 0
	projection["id"] = 1
	projection["name"] = 1
	projection["model_uid"] = 1
	opts := &options.FindOptions{
		Projection: projection,
	}

	cursor, err := col.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("查询错误: %w", err)
	}

	var result []Resource
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *resourceDAO) DeleteResource(ctx context.Context, id int64) (int64, error) {
	col := dao.db.Collection(ResourceCollection)
	filter := bson.M{"id": id}

	result, err := col.DeleteOne(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("删除文档错误: %w", err)
	}

	return result.DeletedCount, nil
}

func (dao *resourceDAO) CountByModelUids(ctx context.Context, modelUids []string) (map[string]int, error) {
	col := dao.db.Collection(ResourceCollection)
	filter := bson.M{}
	if len(modelUids) > 0 {
		filter["model_uid"] = bson.D{{Key: "$in", Value: modelUids}}
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$model_uid"},
			{Key: "total", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
	}

	// 执行聚合查询
	cursor, err := col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("查询错误, %w", err)
	}
	defer cursor.Close(ctx)

	// 定义结果结构体
	var results []struct {
		ModelUID string `bson:"_id"`
		Total    int    `bson:"total"`
	}

	// 将游标中的数据解码到 results 变量
	if err = cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}

	// 将结果转换为 map[model_uid]total_count
	modelCountMap := make(map[string]int)
	for _, result := range results {
		modelCountMap[result.ModelUID] = result.Total
	}

	return modelCountMap, nil

}

func (dao *resourceDAO) Search(ctx context.Context, text string) ([]SearchResource, error) {
	col := dao.db.Collection(ResourceCollection)
	filter := bson.M{"$text": bson.M{"$search": text}}
	groupStage := bson.D{
		{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$model_uid"},
			{Key: "total", Value: bson.D{{Key: "$sum", Value: 1}}},
			{Key: "data", Value: bson.D{{Key: "$push", Value: "$$ROOT"}}},
		}},
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		groupStage,
		{{Key: "$sort", Value: bson.D{{Key: "total", Value: -1}}}},
	}

	cursor, err := col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("查询错误, %w", err)
	}
	defer cursor.Close(ctx)

	var result []SearchResource
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}

	return result, nil
}

func (dao *resourceDAO) ListExcludeAndFilterResourceByIds(ctx context.Context, fields []string, modelUid string,
	offset, limit int64, ids []int64, filter domain.Condition) ([]Resource, error) {
	col := dao.db.Collection(ResourceCollection)
	filters := bson.M{"model_uid": modelUid}
	if len(ids) > 0 {
		filters["id"] = bson.M{
			"$nin": ids,
		}
	}

	switch filter.Condition {
	case "not_equal":
		filters[filter.Name] = bson.M{"$ne": filter.Input}
	case "equal":
		filters[filter.Name] = filter.Input
	case "contains":
		filters[filter.Name] = bson.M{"$regex": primitive.Regex{Pattern: filter.Input, Options: "i"}}
	}

	projection := buildProjection(fields)

	opts := &options.FindOptions{
		Projection: projection,
		Limit:      &limit,
		Skip:       &offset,
	}

	cursor, err := col.Find(ctx, filters, opts)
	if err != nil {
		return nil, fmt.Errorf("查询错误: %w", err)
	}

	var result []Resource
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *resourceDAO) TotalExcludeAndFilterResourceByIds(ctx context.Context, modelUid string,
	ids []int64, filter domain.Condition) (int64, error) {
	col := dao.db.Collection(ResourceCollection)
	filters := bson.M{"model_uid": modelUid}
	if len(ids) > 0 {
		filters["id"] = bson.M{
			"$nin": ids,
		}
	}

	switch filter.Condition {
	case "not_equal":
		filters[filter.Name] = bson.M{"$ne": filter.Input}
	case "equal":
		filters[filter.Name] = filter.Input
	case "contains":
		filters[filter.Name] = bson.M{"$regex": primitive.Regex{Pattern: filter.Input, Options: "i"}}
	}

	count, err := col.CountDocuments(ctx, filters)
	if err != nil {
		return 0, fmt.Errorf("文档计数错误: %w", err)
	}

	return count, nil
}

type Resource struct {
	ID       int64         `bson:"id"`
	ModelUID string        `bson:"model_uid"`
	Data     mongox.MapStr `bson:",inline"`
	Ctime    int64         `bson:"ctime"`
	Utime    int64         `bson:"utime"`
}

type Pipeline struct {
	ModelUid string `bson:"_id"`
	Total    int    `bson:"total"`
}

type SearchResource struct {
	ModelUid string          `bson:"_id"`
	Total    int             `bson:"total"`
	Data     []mongox.MapStr `bson:"data"`
}

// BatchCreateOrUpdate 批量创建或更新资产
// NOTE: 基于 model_uid + name 进行 upsert,使用 MongoDB BulkWrite 提升性能
func (dao *resourceDAO) BatchCreateOrUpdate(ctx context.Context, resources []Resource) error {
	if len(resources) == 0 {
		return nil
	}

	col := dao.db.Collection(ResourceCollection)
	now := time.Now().UnixMilli()

	// 批量获取 ID(用于新创建的资源)
	// NOTE: 如果是更新操作,$setOnInsert 不会生效,ID 会被浪费,但这不是问题
	startID, err := dao.db.GetBatchIdGenerator(ResourceCollection, len(resources))
	if err != nil {
		return fmt.Errorf("获取批量 ID 失败: %w", err)
	}

	// 构建 BulkWrite 模型
	models := make([]mongo.WriteModel, 0, len(resources))
	currentID := startID

	for _, resource := range resources {
		// 构建 filter: 基于 model_uid + name
		filter := bson.M{
			"model_uid": resource.ModelUID,
			"name":      resource.Data["name"],
		}

		// 构建 update 文档
		updateDoc := bson.M{
			"utime": now,
		}

		// 合并资源数据
		for key, value := range resource.Data {
			updateDoc[key] = value
		}

		// 构建 upsert 操作
		update := bson.M{
			"$set": updateDoc,
			"$setOnInsert": bson.M{
				"id":    currentID,
				"ctime": now,
			},
		}

		model := mongo.NewUpdateOneModel().
			SetFilter(filter).
			SetUpdate(update).
			SetUpsert(true)

		models = append(models, model)
		currentID++
	}

	// 执行批量操作
	_, err = col.BulkWrite(ctx, models)
	if err != nil {
		return fmt.Errorf("批量创建或更新操作失败: %w", err)
	}

	return nil
}

func (dao *resourceDAO) ListResourcesWithFilters(ctx context.Context, fields []string, modelUid string, ids []int64, offset, limit int64, filterGroups []domain.FilterGroup) ([]Resource, error) {
	col := dao.db.Collection(ResourceCollection)

	// 基础过滤条件
	baseFilter := bson.M{"model_uid": modelUid}
	if len(ids) > 0 {
		baseFilter["id"] = bson.M{"$in": ids}
	}

	// 若没有筛选条件，直接使用基础条件
	if len(filterGroups) == 0 {
		projection := buildProjection(fields)
		opts := &options.FindOptions{
			Projection: projection,
			Limit:      &limit,
			Skip:       &offset,
			Sort:       bson.D{{Key: "ctime", Value: -1}},
		}
		cursor, err := col.Find(ctx, baseFilter, opts)
		if err != nil {
			return nil, fmt.Errorf("查询错误: %w", err)
		}

		var result []Resource
		if err = cursor.All(ctx, &result); err != nil {
			return nil, fmt.Errorf("解码错误: %w", err)
		}
		if err = cursor.Err(); err != nil {
			return nil, fmt.Errorf("游标遍历错误: %w", err)
		}
		return result, nil
	}

	// 构建复杂筛选条件 (Disjunctive Normal Form: (A AND B) OR (C AND D))
	var orConditions []bson.M

	for _, group := range filterGroups {
		// 跳过空组
		if len(group.Filters) == 0 {
			continue
		}

		var andConditions []bson.M
		for _, f := range group.Filters {
			cond := buildBsonCondition(f)
			if cond != nil {
				andConditions = append(andConditions, cond)
			}
		}

		if len(andConditions) > 0 {
			if len(andConditions) == 1 {
				orConditions = append(orConditions, andConditions[0])
			} else {
				orConditions = append(orConditions, bson.M{"$and": andConditions})
			}
		}
	}

	var finalFilter interface{} = baseFilter

	// 如果有有效筛选组
	if len(orConditions) > 0 {
		// 如果只有一个 OR 分支，直接合并
		if len(orConditions) == 1 {
			// 将 OR 分支的条件与 model_uid (和 ids) 合并
			// 注意: 这里使用 $and 确保安全，避免 key 冲突
			finalFilter = bson.M{
				"$and": []bson.M{
					baseFilter,
					orConditions[0],
				},
			}
		} else {
			// 多个 GROUP 之间是 OR 关系
			finalFilter = bson.M{
				"$and": []bson.M{
					baseFilter,
					bson.M{"$or": orConditions},
				},
			}
		}
	}

	projection := buildProjection(fields)
	opts := &options.FindOptions{
		Projection: projection,
		Limit:      &limit,
		Skip:       &offset,
		Sort:       bson.D{{Key: "ctime", Value: -1}},
	}

	cursor, err := col.Find(ctx, finalFilter, opts)
	if err != nil {
		return nil, fmt.Errorf("查询错误: %w", err)
	}

	var result []Resource
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func buildBsonCondition(f domain.FilterCondition) bson.M {
	key := f.FieldUID
	val := f.Value

	// MongoDB 字段名匹配 (Data inline)
	switch f.Operator {
	case "eq":
		return bson.M{key: val}
	case "ne":
		return bson.M{key: bson.M{"$ne": val}}
	case "contains":
		s, ok := val.(string)
		if !ok {
			return nil
		}
		return bson.M{key: bson.M{"$regex": primitive.Regex{Pattern: s, Options: "i"}}}
	case "gt":
		return bson.M{key: bson.M{"$gt": val}}
	case "lt":
		return bson.M{key: bson.M{"$lt": val}}
	case "gte":
		return bson.M{key: bson.M{"$gte": val}}
	case "lte":
		return bson.M{key: bson.M{"$lte": val}}
	case "in":
		return bson.M{key: bson.M{"$in": val}}
	case "nin":
		return bson.M{key: bson.M{"$nin": val}}
	default:
		return bson.M{key: val}
	}
}

func normalizeSearchFields(fields []string) []string {
	seen := make(map[string]struct{}, len(fields)+1)
	result := make([]string, 0, len(fields)+1)

	for _, field := range append([]string{"name"}, fields...) {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		if _, ok := seen[field]; ok {
			continue
		}
		seen[field] = struct{}{}
		result = append(result, field)
	}

	return result
}

func buildResourceSearchFilter(modelUid string, fields []string, conditions []domain.SearchCondition) bson.M {
	filter := bson.M{"model_uid": modelUid}
	andConditions := make(bson.A, 0, len(conditions))
	for _, condition := range conditions {
		fieldUID := strings.TrimSpace(condition.FieldUID)
		keyword := strings.TrimSpace(condition.Keyword)
		if fieldUID == "" && keyword == "" {
			continue
		}

		if keyword == "" {
			andConditions = append(andConditions, bson.M{"$or": buildEmptySearchConditions([]string{fieldUID})})
			continue
		}

		keywords := splitSearchKeywords(keyword)
		if len(keywords) == 0 {
			continue
		}

		searchFields := fields
		if fieldUID != "" {
			searchFields = []string{fieldUID}
		}
		matchType := resolveSearchMatchType(fieldUID, condition.MatchType)
		orConditions := make([]bson.M, 0, len(keywords))
		for _, searchKeyword := range keywords {
			if matchType == domain.SearchMatchTypeFuzzy {
				orConditions = append(orConditions, buildFuzzySearchConditions(searchFields, searchKeyword)...)
				continue
			}
			if fieldUID == "" {
				orConditions = append(orConditions, buildExactSearchConditions(searchFields, searchKeyword)...)
				continue
			}
			if operator, value, ok := parseNumericComparison(searchKeyword); ok {
				orConditions = append(orConditions, bson.M{
					"$expr": buildNumericComparisonExpression(fieldUID, operator, value),
				})
				continue
			}
			orConditions = append(orConditions, buildExactSearchConditions(searchFields, searchKeyword)...)
		}
		if len(orConditions) > 0 {
			if len(keywords) == 1 && len(orConditions) == 1 {
				if expression, ok := orConditions[0]["$expr"]; ok {
					andConditions = append(andConditions, bson.M{"$expr": expression})
					continue
				}
			}
			andConditions = append(andConditions, bson.M{"$or": orConditions})
		}
	}
	if len(andConditions) > 0 {
		filter["$and"] = andConditions
	}
	return filter
}

func splitSearchKeywords(keyword string) []string {
	parts := strings.Split(keyword, ",")
	keywords := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			keywords = append(keywords, part)
		}
	}
	return keywords
}

func resolveSearchMatchType(fieldUID string, matchType domain.SearchMatchType) domain.SearchMatchType {
	matchType = domain.SearchMatchType(strings.ToLower(strings.TrimSpace(string(matchType))))
	if matchType == domain.SearchMatchTypeExact || matchType == domain.SearchMatchTypeFuzzy {
		return matchType
	}
	if fieldUID == "" {
		return domain.SearchMatchTypeFuzzy
	}
	return domain.SearchMatchTypeExact
}

func buildExactSearchConditions(fields []string, keyword string) []bson.M {
	conditions := make([]bson.M, 0, len(fields)*3)
	for _, field := range fields {
		conditions = append(conditions, bson.M{field: keyword})
	}

	if integer, err := strconv.ParseInt(keyword, 10, 64); err == nil {
		for _, field := range fields {
			conditions = append(conditions, bson.M{field: integer})
		}
	}
	if decimal, err := strconv.ParseFloat(keyword, 64); err == nil {
		for _, field := range fields {
			conditions = append(conditions, bson.M{field: decimal})
		}
	}

	return conditions
}

func parseNumericComparison(keyword string) (string, float64, bool) {
	keyword = strings.TrimSpace(keyword)
	operators := []struct {
		prefix   string
		mongoKey string
	}{
		{prefix: ">=", mongoKey: "$gte"},
		{prefix: "<=", mongoKey: "$lte"},
		{prefix: ">", mongoKey: "$gt"},
		{prefix: "<", mongoKey: "$lt"},
	}

	for _, operator := range operators {
		if !strings.HasPrefix(keyword, operator.prefix) {
			continue
		}
		value, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimPrefix(keyword, operator.prefix)), 64)
		if err != nil || math.IsNaN(value) || math.IsInf(value, 0) {
			return "", 0, false
		}
		return operator.mongoKey, value, true
	}
	return "", 0, false
}

func buildNumericComparisonExpression(field, operator string, value float64) bson.M {
	convertedValue := bson.M{
		"$convert": bson.M{
			"input":   "$" + field,
			"to":      "double",
			"onError": nil,
			"onNull":  nil,
		},
	}
	return bson.M{
		"$and": bson.A{
			bson.M{"$ne": bson.A{convertedValue, nil}},
			bson.M{operator: bson.A{convertedValue, value}},
		},
	}
}

func buildEmptySearchConditions(fields []string) []bson.M {
	conditions := make([]bson.M, 0, len(fields)*3)
	for _, field := range fields {
		conditions = append(conditions,
			bson.M{field: ""},
			bson.M{field: nil},
			bson.M{field: bson.M{"$size": 0}},
		)
	}
	return conditions
}

func buildFuzzySearchConditions(fields []string, keyword string) []bson.M {
	pattern := regexp.QuoteMeta(keyword)
	conditions := make([]bson.M, 0, len(fields)*2)
	for _, field := range fields {
		conditions = append(conditions, bson.M{
			field: bson.M{"$regex": primitive.Regex{Pattern: pattern, Options: "i"}},
		})
		conditions = append(conditions, bson.M{
			"$expr": bson.M{
				"$regexMatch": bson.M{
					"input": bson.M{
						"$convert": bson.M{
							"input":   "$" + field,
							"to":      "string",
							"onError": "",
							"onNull":  "",
						},
					},
					"regex":   pattern,
					"options": "i",
				},
			},
		})
	}
	return conditions
}

func (dao *resourceDAO) TotalResourcesWithFilters(ctx context.Context, modelUid string, ids []int64, filterGroups []domain.FilterGroup) (int64, error) {
	col := dao.db.Collection(ResourceCollection)
	baseFilter := bson.M{"model_uid": modelUid}
	if len(ids) > 0 {
		baseFilter["id"] = bson.M{"$in": ids}
	}

	if len(filterGroups) == 0 {
		return dao.db.Collection(ResourceCollection).CountDocuments(ctx, baseFilter)
	}

	var orConditions []bson.M
	for _, group := range filterGroups {
		if len(group.Filters) == 0 {
			continue
		}
		var andConditions []bson.M
		for _, f := range group.Filters {
			cond := buildBsonCondition(f)
			if cond != nil {
				andConditions = append(andConditions, cond)
			}
		}
		if len(andConditions) > 0 {
			if len(andConditions) == 1 {
				orConditions = append(orConditions, andConditions[0])
			} else {
				orConditions = append(orConditions, bson.M{"$and": andConditions})
			}
		}
	}

	var finalFilter interface{} = baseFilter
	if len(orConditions) > 0 {
		if len(orConditions) == 1 {
			finalFilter = bson.M{
				"$and": []bson.M{baseFilter, orConditions[0]},
			}
		} else {
			finalFilter = bson.M{
				"$and": []bson.M{baseFilter, bson.M{"$or": orConditions}},
			}
		}
	}

	count, err := col.CountDocuments(ctx, finalFilter)
	if err != nil {
		return 0, fmt.Errorf("文档计数错误: %w", err)
	}
	return count, nil
}
