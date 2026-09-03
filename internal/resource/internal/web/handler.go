package web

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"github.com/gotomicro/ego/core/elog"
	"github.com/userreksai/ecmdb-main/internal/attribute"
	"github.com/userreksai/ecmdb-main/internal/operationlog"
	"github.com/userreksai/ecmdb-main/internal/pkg/authctx"
	"github.com/userreksai/ecmdb-main/internal/policy"
	"github.com/userreksai/ecmdb-main/internal/relation"
	"github.com/userreksai/ecmdb-main/internal/resource/internal/domain"
	"github.com/userreksai/ecmdb-main/internal/resource/internal/service"
	"github.com/userreksai/ecmdb-main/internal/role"
	"github.com/userreksai/ecmdb-main/pkg/ginx"
)

type Handler struct {
	svc       service.EncryptedSvc
	attrSvc   attribute.Service
	RRSvc     relation.RRSvc
	roleSvc   role.Service
	policySvc policy.Service
	auditSvc  operationlog.Service
}

func NewHandler(service service.EncryptedSvc, attributeSvc attribute.Service, RRSvc relation.RRSvc,
	roleSvc role.Service, policySvc policy.Service, auditSvc operationlog.Service) *Handler {
	return &Handler{
		svc:       service,
		attrSvc:   attributeSvc,
		RRSvc:     RRSvc,
		roleSvc:   roleSvc,
		policySvc: policySvc,
		auditSvc:  auditSvc,
	}
}

var errModelAccessDenied = errors.New("model asset access denied")

func (h *Handler) authorizeModel(ctx *gin.Context, modelUID string) error {
	allowed, err := h.allowedModelUIDs(ctx, []string{modelUID})
	if err != nil {
		return err
	}
	if _, exists := allowed[modelUID]; !exists {
		return errModelAccessDenied
	}
	return nil
}

func (h *Handler) allowedModelUIDs(ctx *gin.Context, modelUIDs []string) (map[string]struct{}, error) {
	if h.roleSvc == nil || h.policySvc == nil {
		allowed := make(map[string]struct{}, len(modelUIDs))
		for _, modelUID := range modelUIDs {
			allowed[modelUID] = struct{}{}
		}
		return allowed, nil
	}
	uid, ok := authctx.UID(ctx)
	if !ok {
		return nil, fmt.Errorf("get current user failed")
	}
	roleCodes, err := h.policySvc.GetRolesForUser(ctx, uid)
	if err != nil {
		return nil, err
	}
	accessible, err := h.roleSvc.FilterAccessibleModelUIDs(ctx, roleCodes, modelUIDs)
	if err != nil {
		return nil, err
	}
	allowed := make(map[string]struct{}, len(accessible))
	for _, modelUID := range accessible {
		allowed[modelUID] = struct{}{}
	}
	return allowed, nil
}

func (h *Handler) filterAccessibleRelations(ctx *gin.Context, relations []relation.ResourceRelation) ([]relation.ResourceRelation, error) {
	modelUIDs := make([]string, 0, len(relations)*2)
	for _, item := range relations {
		modelUIDs = append(modelUIDs, item.SourceModelUID, item.TargetModelUID)
	}
	allowed, err := h.allowedModelUIDs(ctx, modelUIDs)
	if err != nil {
		return nil, err
	}

	result := make([]relation.ResourceRelation, 0, len(relations))
	for _, item := range relations {
		_, sourceAllowed := allowed[item.SourceModelUID]
		_, targetAllowed := allowed[item.TargetModelUID]
		if sourceAllowed && targetAllowed {
			result = append(result, item)
		}
	}
	return result, nil
}

func (h *Handler) filterAccessibleResources(ctx *gin.Context, resources []domain.Resource) ([]domain.Resource, error) {
	modelUIDs := make([]string, 0, len(resources))
	for _, item := range resources {
		modelUIDs = append(modelUIDs, item.ModelUID)
	}
	allowed, err := h.allowedModelUIDs(ctx, modelUIDs)
	if err != nil {
		return nil, err
	}

	result := make([]domain.Resource, 0, len(resources))
	for _, item := range resources {
		if _, exists := allowed[item.ModelUID]; exists {
			result = append(result, item)
		}
	}
	return result, nil
}

func (h *Handler) authorizeResource(ctx *gin.Context, resourceID int64) error {
	if h.roleSvc == nil || h.policySvc == nil {
		return nil
	}
	resource, err := h.svc.FindResourceById(ctx, nil, resourceID)
	if err != nil {
		return err
	}
	return h.authorizeModel(ctx, resource.ModelUID)
}

func modelAccessError(err error) (ginx.Result, error) {
	if errors.Is(err, errModelAccessDenied) {
		return ginx.Result{Code: 403001, Msg: "无权访问该模型资产"}, nil
	}
	return systemErrorResult, err
}

func (h *Handler) PrivateRoutes(server *gin.Engine) {
	g := server.Group("/api/resource")
	// 资源操作
	g.POST("/create", ginx.WrapBody[CreateResourceReq](h.CreateResource))
	// 根据 ID 查询资源列表
	g.POST("/detail", ginx.WrapBody[DetailResourceReq](h.DetailResource))
	// 根据模型 UID 查询资源列表
	g.POST("/list", ginx.WrapBody[ListResourceReq](h.ListResource))
	//查询模型内的实例 模糊搜索 + 精确匹配
	g.POST("/list/search", ginx.WrapBody[SearchModelResourceReq](h.SearchModelResource))
	g.POST("/delete", ginx.WrapBody[DeleteResourceReq](h.DeleteResource))

	// 修改资产信息
	g.POST("/update", ginx.WrapBody[UpdateResourceReq](h.UpdateResource))
	g.POST("/set_custom_field", ginx.WrapBody[SetCustomFieldReq](h.SetCustomField))
	// 资源关联关系
	g.POST("/relation/can_be_related", ginx.WrapBody[ListCanBeRelatedReqByModel](h.ListCanBeFilterRelated))
	g.POST("/relation/diagram", ginx.WrapBody[ListDiagramReq](h.FindDiagram))
	g.POST("/relation/graph", ginx.WrapBody[ListDiagramReq](h.FindAllGraph))
	g.POST("/relation/graph/add/left", ginx.WrapBody[ListDiagramReq](h.FindLeftGraph))
	g.POST("/relation/graph/add/right", ginx.WrapBody[ListDiagramReq](h.FindRightGraph))

	// 根据模型 UID 查询资源列表
	g.POST("/list/ids", ginx.WrapBody[ListResourceByIdsReq](h.ListResourceByIds))
	// 全文检索
	g.POST("/search", ginx.WrapBody[SearchReq](h.Search))
	// 查询加密字段信息
	g.POST("/secure", ginx.WrapBody[FindSecureReq](h.FindSecureData))

}

func (h *Handler) CreateResource(ctx *gin.Context, req CreateResourceReq) (ginx.Result, error) {
	if err := h.authorizeModel(ctx, req.ModelUid); err != nil {
		return modelAccessError(err)
	}
	id, err := h.svc.CreateResource(ctx, h.toDomain(req))

	if err != nil {
		return systemErrorResult, err
	}
	h.recordOperation(ctx, operationlog.Record{
		Account:        actorAccount(ctx),
		OperationModel: req.ModelUid,
		OperationType:  operationlog.OperationCreate,
		ModifiedData: map[string]interface{}{
			"id": id, "model_uid": req.ModelUid, "data": h.maskResourceData(ctx, req.ModelUid, req.Data),
		},
		ModifiedCount: 1,
	})

	return ginx.Result{
		Data: id,
		Msg:  "创建资源成功",
	}, nil
}

func (h *Handler) DetailResource(ctx *gin.Context, req DetailResourceReq) (ginx.Result, error) {
	if err := h.authorizeModel(ctx, req.ModelUid); err != nil {
		return modelAccessError(err)
	}
	if err := h.authorizeResource(ctx, req.ID); err != nil {
		return modelAccessError(err)
	}
	fields, err := h.attrSvc.SearchAttributeFieldsByModelUid(ctx, req.ModelUid)
	if err != nil {
		return systemErrorResult, err
	}

	resp, err := h.svc.FindResourceById(ctx, fields, req.ID)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: resp,
		Msg:  "查看资源详情成功",
	}, nil
}

func (h *Handler) SetCustomField(ctx *gin.Context, req SetCustomFieldReq) (ginx.Result, error) {
	if err := h.authorizeResource(ctx, req.Id); err != nil {
		return modelAccessError(err)
	}
	var original domain.Resource
	var err error
	if h.auditSvc != nil {
		original, err = h.svc.FindResourceById(ctx, nil, req.Id)
		if err != nil {
			return systemErrorResult, err
		}
	}
	count, err := h.svc.SetCustomField(ctx, req.Id, req.Field, req.Data)
	if err != nil {
		return systemErrorResult, err
	}
	if h.auditSvc != nil {
		modified := cloneResourceData(original.Data)
		modified[req.Field] = req.Data
		h.recordOperation(ctx, operationlog.Record{
			Account:        actorAccount(ctx),
			OperationModel: original.ModelUID,
			OperationType:  operationlog.OperationUpdate,
			OriginalData:   h.resourceSnapshot(ctx, original),
			ModifiedData: map[string]interface{}{
				"id": original.ID, "model_uid": original.ModelUID, "data": h.maskResourceData(ctx, original.ModelUID, modified),
			},
			ModifiedCount: count,
		})
	}

	return ginx.Result{
		Data: count,
	}, nil
}

func (h *Handler) ListResource(ctx *gin.Context, req ListResourceReq) (ginx.Result, error) {
	if err := h.authorizeModel(ctx, req.ModelUid); err != nil {
		return modelAccessError(err)
	}
	fields, err := h.attrSvc.SearchAttributeFieldsByModelUid(ctx, req.ModelUid)
	if err != nil {
		return systemErrorResult, err
	}

	resp, total, err := h.svc.ListResource(ctx, fields, req.ModelUid, req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}

	rs := slice.Map(resp, func(idx int, src domain.Resource) Resource {
		return Resource{
			ID:       src.ID,
			Name:     src.Name,
			ModelUID: src.ModelUID,
			Data:     src.Data,
		}
	})

	return ginx.Result{
		Data: RetrieveResources{
			Resources: rs,
			Total:     total,
		},
		Msg: "查看资源列表成功",
	}, nil
}

func (h *Handler) SearchModelResource(ctx *gin.Context, req SearchModelResourceReq) (ginx.Result, error) {
	if err := h.authorizeModel(ctx, req.ModelUid); err != nil {
		return modelAccessError(err)
	}
	fields, err := h.attrSvc.SearchAllAttributeFieldsByModelUid(ctx, req.ModelUid)
	if err != nil {
		return systemErrorResult, err
	}
	conditions := make([]domain.SearchCondition, 0, len(req.Conditions))
	if len(req.Conditions) == 0 {
		conditions = append(conditions, domain.SearchCondition{
			FieldUID:  req.FieldUid,
			Keyword:   req.Keyword,
			MatchType: domain.SearchMatchType(req.MatchType),
		})
	} else {
		for _, condition := range req.Conditions {
			conditions = append(conditions, domain.SearchCondition{
				FieldUID:  condition.FieldUid,
				Keyword:   condition.Keyword,
				MatchType: domain.SearchMatchType(condition.MatchType),
			})
		}
	}
	for _, condition := range conditions {
		fieldUID := strings.TrimSpace(condition.FieldUID)
		if fieldUID != "" && !contains(fields, fieldUID) {
			return ginx.Result{Code: 400001, Msg: "搜索字段不属于当前模型"}, nil
		}
		matchType := domain.SearchMatchType(strings.ToLower(strings.TrimSpace(string(condition.MatchType))))
		if matchType != "" && matchType != domain.SearchMatchTypeExact && matchType != domain.SearchMatchTypeFuzzy {
			return ginx.Result{Code: 400001, Msg: "匹配方式仅支持 exact 或 fuzzy"}, nil
		}
	}

	resp, total, err := h.svc.SearchResourcesInModel(
		ctx, fields, req.ModelUid, conditions, req.Offset, req.Limit,
	)
	if err != nil {
		return systemErrorResult, err
	}

	rs := slice.Map(resp, func(idx int, src domain.Resource) Resource {
		return Resource{
			ID:       src.ID,
			Name:     src.Name,
			ModelUID: src.ModelUID,
			Data:     src.Data,
		}
	})

	return ginx.Result{
		Data: RetrieveResources{
			Resources: rs,
			Total:     total,
		},
		Msg: "查询资源列表成功",
	}, nil
}

func (h *Handler) UpdateResource(ctx *gin.Context, req UpdateResourceReq) (ginx.Result, error) {
	if err := h.authorizeModel(ctx, req.ModelUid); err != nil {
		return modelAccessError(err)
	}
	if err := h.authorizeResource(ctx, req.Id); err != nil {
		return modelAccessError(err)
	}
	var original domain.Resource
	var err error
	if h.auditSvc != nil {
		original, err = h.svc.FindResourceById(ctx, nil, req.Id)
		if err != nil {
			return systemErrorResult, err
		}
	}
	resource := h.toDomainUpdate(req)
	t, err := h.svc.UpdateResource(ctx, resource)

	if err != nil {
		return systemErrorResult, err
	}
	if h.auditSvc != nil {
		modified := cloneResourceData(original.Data)
		for field, value := range req.Data {
			modified[field] = value
		}
		h.recordOperation(ctx, operationlog.Record{
			Account:        actorAccount(ctx),
			OperationModel: req.ModelUid,
			OperationType:  operationlog.OperationUpdate,
			OriginalData:   h.resourceSnapshot(ctx, original),
			ModifiedData: map[string]interface{}{
				"id": req.Id, "model_uid": req.ModelUid, "data": h.maskResourceData(ctx, req.ModelUid, modified),
			},
			ModifiedCount: t,
		})
	}

	return ginx.Result{
		Data: t,
	}, nil
}

func (h *Handler) ListCanBeFilterRelated(ctx *gin.Context, req ListCanBeRelatedReqByModel) (ginx.Result, error) {
	if err := h.authorizeModel(ctx, req.ModelUid); err != nil {
		return modelAccessError(err)
	}
	if err := h.authorizeResource(ctx, req.ResourceId); err != nil {
		return modelAccessError(err)
	}
	var (
		mUid       string
		err        error
		excludeIds []int64
	)
	/*
		查询已经关联的数据
		model_uid = physical
		relation_name = "physical_run_mongo"
	*/
	if req.RelationName == "" {
		return systemErrorResult, fmt.Errorf("关联名称为空")
	}

	// 传递的是当前模型UID （特别注意）
	rn := strings.Split(req.RelationName, "_")
	if rn[0] == req.ModelUid {
		mUid = rn[2]
		excludeIds, err = h.RRSvc.ListSrcRelated(ctx, req.ModelUid, req.RelationName, req.ResourceId)
	} else {
		mUid = rn[0]
		excludeIds, err = h.RRSvc.ListDstRelated(ctx, rn[2], req.RelationName, req.ResourceId)
	}
	if err != nil {
		return systemErrorResult, err
	}
	if err = h.authorizeModel(ctx, mUid); err != nil {
		return modelAccessError(err)
	}

	fields, err := h.attrSvc.SearchAttributeFieldsByModelUid(ctx, mUid)

	if err != nil {
		return systemErrorResult, err
	}

	// 排除已关联数据, 并且进行过滤，返回未关联数据
	rrs, total, err := h.svc.ListExcludeAndFilterResourceByIds(ctx, fields, mUid, req.Offset, req.Limit, excludeIds,
		domain.Condition{
			Name:      req.FilterName,
			Condition: req.FilterCondition,
			Input:     req.FilterInput,
		})
	if err != nil {
		return systemErrorResult, err
	}

	rs := slice.Map(rrs, func(idx int, src domain.Resource) Resource {
		return Resource{
			ID:       src.ID,
			Name:     src.Name,
			ModelUID: src.ModelUID,
			Data:     src.Data,
		}
	})

	return ginx.Result{
		Data: RetrieveResources{
			Resources: rs,
			Total:     total,
		},
	}, nil
}

func (h *Handler) FindAllGraph(ctx *gin.Context, req ListDiagramReq) (ginx.Result, error) {
	if err := h.authorizeModel(ctx, req.ModelUid); err != nil {
		return modelAccessError(err)
	}
	if err := h.authorizeResource(ctx, req.ResourceId); err != nil {
		return modelAccessError(err)
	}
	// 查询资产关联上下级拓扑
	graph, _, err := h.RRSvc.ListDiagram(ctx, req.ModelUid, req.ResourceId)
	if err != nil {
		return systemErrorResult, err
	}
	graph.SRC, err = h.filterAccessibleRelations(ctx, graph.SRC)
	if err != nil {
		return systemErrorResult, err
	}
	graph.DST, err = h.filterAccessibleRelations(ctx, graph.DST)
	if err != nil {
		return systemErrorResult, err
	}
	var (
		srcId []int64
		dstId []int64
	)

	rrs := append(graph.SRC, graph.DST...)
	lines := slice.Map(rrs, func(idx int, src relation.ResourceRelation) Line {
		return Line{
			From: strconv.FormatInt(src.SourceResourceID, 10),
			To:   strconv.FormatInt(src.TargetResourceID, 10),
		}
	})

	// 查询关联的所有节点 ids
	srcId = slice.Map(graph.SRC, func(idx int, src relation.ResourceRelation) int64 {
		return src.TargetResourceID
	})
	dstId = slice.Map(graph.DST, func(idx int, src relation.ResourceRelation) int64 {
		return src.SourceResourceID
	})

	ids := append(srcId, dstId...)

	// 查询节点信息
	rs, err := h.svc.ListResourceByIds(ctx, nil, ids)
	if err != nil {
		return systemErrorResult, err
	}

	nodes := slice.Map(rs, func(idx int, src domain.Resource) Node {
		data := make(map[string]any, 1)
		data["model_uid"] = src.ModelUID
		data["isNeedLoadDataFromRemoteServer"] = true
		data["childrenLoaded"] = false
		for _, id := range srcId {
			if src.ID == id {
				return Node{
					ID:                   strconv.FormatInt(src.ID, 10),
					Text:                 src.Name,
					Data:                 data,
					ExpandHolderPosition: "right",
					Expanded:             false,
				}
			}
		}
		return Node{
			ID:                   strconv.FormatInt(src.ID, 10),
			Text:                 src.Name,
			ExpandHolderPosition: "left",
			Expanded:             false,
			Data:                 data,
		}
	})

	nodes = append(nodes, Node{
		ID:       strconv.FormatInt(req.ResourceId, 10),
		Text:     req.ResourceName,
		Expanded: true,
		Data: map[string]any{
			"model_uid": req.ModelUid,
		},
	})

	return ginx.Result{
		Data: RetrieveGraph{
			Lines:  lines,
			Nodes:  nodes,
			RootId: strconv.FormatInt(req.ResourceId, 10),
		},
	}, nil
}

func (h *Handler) FindLeftGraph(ctx *gin.Context, req ListDiagramReq) (ginx.Result, error) {
	if err := h.authorizeModel(ctx, req.ModelUid); err != nil {
		return modelAccessError(err)
	}
	if err := h.authorizeResource(ctx, req.ResourceId); err != nil {
		return modelAccessError(err)
	}
	// 查询资产关联上下级拓扑
	graphLeft, _, err := h.RRSvc.ListDstResources(ctx, req.ModelUid, req.ResourceId)
	if err != nil {
		return systemErrorResult, err
	}
	graphLeft, err = h.filterAccessibleRelations(ctx, graphLeft)
	if err != nil {
		return systemErrorResult, err
	}
	var (
		srcIds []int64
	)

	lines := slice.Map(graphLeft, func(idx int, src relation.ResourceRelation) Line {
		return Line{
			From: strconv.FormatInt(src.SourceResourceID, 10),
			To:   strconv.FormatInt(src.TargetResourceID, 10),
		}
	})

	// 查询关联的所有节点 ids
	srcIds = slice.Map(graphLeft, func(idx int, src relation.ResourceRelation) int64 {
		return src.SourceResourceID
	})

	// 查询节点信息
	rs, err := h.svc.ListResourceByIds(ctx, nil, srcIds)
	if err != nil {
		return systemErrorResult, err
	}

	nodes := slice.Map(rs, func(idx int, src domain.Resource) Node {
		data := make(map[string]any, 1)
		data["model_uid"] = src.ModelUID
		data["isNeedLoadDataFromRemoteServer"] = true
		data["childrenLoaded"] = false
		return Node{
			ID:                   strconv.FormatInt(src.ID, 10),
			Text:                 src.Name,
			ExpandHolderPosition: "left",
			Expanded:             false,
			Data:                 data,
		}
	})

	return ginx.Result{
		Data: RetrieveGraph{
			Lines:  lines,
			Nodes:  nodes,
			RootId: strconv.FormatInt(req.ResourceId, 10),
		},
	}, nil
}

func (h *Handler) FindRightGraph(ctx *gin.Context, req ListDiagramReq) (ginx.Result, error) {
	if err := h.authorizeModel(ctx, req.ModelUid); err != nil {
		return modelAccessError(err)
	}
	if err := h.authorizeResource(ctx, req.ResourceId); err != nil {
		return modelAccessError(err)
	}
	// 查询资产关联上下级拓扑
	graphRight, _, err := h.RRSvc.ListSrcResources(ctx, req.ModelUid, req.ResourceId)
	if err != nil {
		return systemErrorResult, err
	}
	graphRight, err = h.filterAccessibleRelations(ctx, graphRight)
	if err != nil {
		return systemErrorResult, err
	}
	var (
		srcIds []int64
	)

	lines := slice.Map(graphRight, func(idx int, src relation.ResourceRelation) Line {
		return Line{
			From: strconv.FormatInt(src.SourceResourceID, 10),
			To:   strconv.FormatInt(src.TargetResourceID, 10),
		}
	})

	// 查询关联的所有节点 ids
	srcIds = slice.Map(graphRight, func(idx int, src relation.ResourceRelation) int64 {
		return src.TargetResourceID
	})

	// 查询节点信息
	rs, err := h.svc.ListResourceByIds(ctx, nil, srcIds)
	if err != nil {
		return systemErrorResult, err
	}

	nodes := slice.Map(rs, func(idx int, src domain.Resource) Node {
		data := make(map[string]any, 1)
		data["model_uid"] = src.ModelUID
		data["isNeedLoadDataFromRemoteServer"] = true
		data["childrenLoaded"] = false
		return Node{
			ID:                   strconv.FormatInt(src.ID, 10),
			Text:                 src.Name,
			ExpandHolderPosition: "right",
			Expanded:             false,
			Data:                 data,
		}
	})

	return ginx.Result{
		Data: RetrieveGraph{
			Lines:  lines,
			Nodes:  nodes,
			RootId: strconv.FormatInt(req.ResourceId, 10),
		},
	}, nil
}

func (h *Handler) FindDiagram(ctx *gin.Context, req ListDiagramReq) (ginx.Result, error) {
	if err := h.authorizeModel(ctx, req.ModelUid); err != nil {
		return modelAccessError(err)
	}
	if err := h.authorizeResource(ctx, req.ResourceId); err != nil {
		return modelAccessError(err)
	}
	// 查询资产关联上下级拓扑
	diagram, _, err := h.RRSvc.ListDiagram(ctx, req.ModelUid, req.ResourceId)
	if err != nil {
		return systemErrorResult, err
	}
	diagram.SRC, err = h.filterAccessibleRelations(ctx, diagram.SRC)
	if err != nil {
		return systemErrorResult, err
	}
	diagram.DST, err = h.filterAccessibleRelations(ctx, diagram.DST)
	if err != nil {
		return systemErrorResult, err
	}
	var (
		src   []ResourceRelation
		dst   []ResourceRelation
		srcId []int64
		dstId []int64
	)

	// 组合前端展示数据
	src = slice.Map(diagram.SRC, func(idx int, src relation.ResourceRelation) ResourceRelation {
		return h.toResourceRelationVo(src)
	})
	dst = slice.Map(diagram.DST, func(idx int, src relation.ResourceRelation) ResourceRelation {
		return h.toResourceRelationVo(src)
	})

	// 查询关联的所有节点 ids
	srcId = slice.Map(diagram.SRC, func(idx int, src relation.ResourceRelation) int64 {
		return src.TargetResourceID
	})
	dstId = slice.Map(diagram.DST, func(idx int, src relation.ResourceRelation) int64 {
		return src.SourceResourceID
	})
	ids := append(srcId, dstId...)

	// 查询节点信息
	rs, err := h.svc.ListResourceByIds(ctx, nil, ids)
	if err != nil {
		return systemErrorResult, err
	}

	// 组合前端返回数据
	assets := make(map[string][]ResourceAssets, len(diagram.DST)+len(diagram.SRC))
	assets = slice.ToMapV(rs, func(element domain.Resource) (string, []ResourceAssets) {
		return element.ModelUID, slice.FilterMap(rs, func(idx int, src domain.Resource) (ResourceAssets, bool) {
			if src.ModelUID == element.ModelUID {
				return ResourceAssets{
					ResourceID:   src.ID,
					ResourceName: src.Name,
				}, true
			}
			return ResourceAssets{}, false
		})
	})

	return ginx.Result{
		Data: RetrieveDiagram{
			SRC:    src,
			DST:    dst,
			Assets: assets,
		},
	}, nil
}

func (h *Handler) ListResourceByIds(ctx *gin.Context, req ListResourceByIdsReq) (ginx.Result, error) {
	if err := h.authorizeModel(ctx, req.ModelUid); err != nil {
		return modelAccessError(err)
	}
	fields, err := h.attrSvc.SearchAttributeFieldsByModelUid(ctx, req.ModelUid)
	if err != nil {
		return systemErrorResult, err
	}

	resp, err := h.svc.ListResourceByIds(ctx, fields, req.ResourceIds)
	if err != nil {
		return systemErrorResult, err
	}
	resp, err = h.filterAccessibleResources(ctx, resp)
	if err != nil {
		return systemErrorResult, err
	}

	rs := slice.Map(resp, func(idx int, src domain.Resource) Resource {
		return Resource{
			ID:       src.ID,
			Name:     src.Name,
			ModelUID: src.ModelUID,
			Data:     src.Data,
		}
	})

	return ginx.Result{
		Data: RetrieveResources{
			Resources: rs,
		},
		Msg: "根据ID查询资源成功",
	}, nil
}

func (h *Handler) Search(ctx *gin.Context, req SearchReq) (ginx.Result, error) {
	search, err := h.svc.Search(ctx, req.Text)
	if err != nil {
		return systemErrorResult, err
	}

	modelUids := slice.Map(search, func(idx int, src domain.SearchResource) string {
		return src.ModelUid
	})
	allowedModelUIDs, err := h.allowedModelUIDs(ctx, modelUids)
	if err != nil {
		return systemErrorResult, err
	}
	search = slice.FilterMap(search, func(_ int, src domain.SearchResource) (domain.SearchResource, bool) {
		_, allowed := allowedModelUIDs[src.ModelUid]
		return src, allowed
	})
	if len(search) == 0 {
		return ginx.Result{Data: []RetrieveSearchResources{}}, nil
	}
	modelUids = slice.Map(search, func(_ int, src domain.SearchResource) string {
		return src.ModelUid
	})

	fields, err := h.attrSvc.SearchAttributeFieldsBySecure(ctx, modelUids)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: slice.Map(search, func(idx int, src domain.SearchResource) RetrieveSearchResources {
			val, ok := fields[src.ModelUid]
			if ok {
				for _, name := range src.Data {
					for key := range name {
						if contains(val, key) {
							name[key] = ""
						}
					}
				}
			}
			return RetrieveSearchResources{
				ModelUid: src.ModelUid,
				Total:    src.Total,
				Data:     src.Data,
			}
		}),
	}, err
}

func (h *Handler) DeleteResource(ctx *gin.Context, req DeleteResourceReq) (ginx.Result, error) {
	if err := h.authorizeResource(ctx, req.Id); err != nil {
		return modelAccessError(err)
	}
	var original domain.Resource
	var err error
	if h.auditSvc != nil {
		original, err = h.svc.FindResourceById(ctx, nil, req.Id)
		if err != nil {
			return systemErrorResult, err
		}
	}
	count, err := h.svc.DeleteResource(ctx, req.Id)
	if err != nil {
		return systemErrorResult, err
	}

	if _, err = h.RRSvc.DeleteRelationsByResourceID(ctx, req.Id); err != nil {
		return systemErrorResult, err
	}
	if h.auditSvc != nil {
		h.recordOperation(ctx, operationlog.Record{
			Account:        actorAccount(ctx),
			OperationModel: original.ModelUID,
			OperationType:  operationlog.OperationDelete,
			OriginalData:   h.resourceSnapshot(ctx, original),
			ModifiedCount:  count,
		})
	}

	return ginx.Result{
		Data: count,
	}, nil
}

func (h *Handler) FindSecureData(ctx *gin.Context, req FindSecureReq) (ginx.Result, error) {
	if err := h.authorizeResource(ctx, req.ID); err != nil {
		return modelAccessError(err)
	}
	data, err := h.svc.FindSecureData(ctx, req.ID, req.FieldUid)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: data,
	}, err
}

func (h *Handler) toDomain(req CreateResourceReq) domain.Resource {
	return domain.Resource{
		Name:     req.Name,
		ModelUID: req.ModelUid,
		Data:     req.Data,
	}
}

func (h *Handler) toResourceRelationVo(src relation.ResourceRelation) ResourceRelation {
	return ResourceRelation{
		SourceModelUID:   src.SourceModelUID,
		TargetModelUID:   src.TargetModelUID,
		SourceResourceID: src.SourceResourceID,
		TargetResourceID: src.TargetResourceID,
		RelationTypeUID:  src.RelationTypeUID,
		RelationName:     src.RelationName,
	}
}

func (h *Handler) toDomainUpdate(src UpdateResourceReq) domain.Resource {
	return domain.Resource{
		ID:       src.Id,
		Name:     src.Name,
		ModelUID: src.ModelUid,
		Data:     src.Data,
	}
}

func contains(slice []string, elem string) bool {
	for _, e := range slice {
		if e == elem {
			return true
		}
	}
	return false
}

func actorAccount(ctx *gin.Context) string {
	identity, ok := authctx.Get(ctx)
	if !ok {
		return "unknown"
	}
	if identity.Username != "" {
		return identity.Username
	}
	return strconv.FormatInt(identity.UID, 10)
}

func (h *Handler) resourceSnapshot(ctx *gin.Context, item domain.Resource) map[string]interface{} {
	return map[string]interface{}{
		"id": item.ID, "model_uid": item.ModelUID, "data": h.maskResourceData(ctx, item.ModelUID, item.Data),
	}
}

func (h *Handler) maskResourceData(ctx *gin.Context, modelUID string, src map[string]interface{}) map[string]interface{} {
	dst := cloneResourceData(src)
	if h.attrSvc == nil {
		return dst
	}
	secureFields, err := h.attrSvc.SearchAttributeFieldsBySecure(ctx, []string{modelUID})
	if err != nil {
		return map[string]interface{}{"_redacted": "secure field metadata unavailable"}
	}
	for _, field := range secureFields[modelUID] {
		if _, exists := dst[field]; exists {
			dst[field] = "[已脱敏]"
		}
	}
	return dst
}

func cloneResourceData(src map[string]interface{}) map[string]interface{} {
	dst := make(map[string]interface{}, len(src))
	for key, value := range src {
		dst[key] = value
	}
	return dst
}

func (h *Handler) recordOperation(ctx *gin.Context, record operationlog.Record) {
	if h.auditSvc != nil {
		if err := h.auditSvc.Record(ctx, record); err != nil {
			elog.DefaultLogger.Error("record resource operation log failed", elog.FieldErr(err))
		}
	}
}
