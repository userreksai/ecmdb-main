package web

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"github.com/userreksai/ecmdb-main/internal/pkg/authctx"
	"github.com/userreksai/ecmdb-main/internal/policy"
	"github.com/userreksai/ecmdb-main/internal/relation/internal/domain"
	"github.com/userreksai/ecmdb-main/internal/relation/internal/service"
	"github.com/userreksai/ecmdb-main/internal/role"
	"github.com/userreksai/ecmdb-main/pkg/ginx"
	"golang.org/x/sync/errgroup"
)

type RelationResourceHandler struct {
	svc       service.RelationResourceService
	roleSvc   role.Service
	policySvc policy.Service
}

func NewRelationResourceHandler(svc service.RelationResourceService, roleSvc role.Service, policySvc policy.Service) *RelationResourceHandler {
	return &RelationResourceHandler{
		svc:       svc,
		roleSvc:   roleSvc,
		policySvc: policySvc,
	}
}

var errModelAccessDenied = errors.New("model asset access denied")

func (h *RelationResourceHandler) allowedModelUIDs(ctx *gin.Context, modelUIDs []string) (map[string]struct{}, error) {
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

func (h *RelationResourceHandler) authorizeModels(ctx *gin.Context, modelUIDs ...string) error {
	allowed, err := h.allowedModelUIDs(ctx, modelUIDs)
	if err != nil {
		return err
	}
	for _, modelUID := range modelUIDs {
		if _, exists := allowed[modelUID]; !exists {
			return errModelAccessDenied
		}
	}
	return nil
}

func relationModelUIDs(relationName string) ([]string, error) {
	parts := strings.Split(relationName, "_")
	if len(parts) != 3 || parts[0] == "" || parts[2] == "" {
		return nil, fmt.Errorf("invalid resource relation name: %s", relationName)
	}
	return []string{parts[0], parts[2]}, nil
}

func (h *RelationResourceHandler) authorizeRelation(ctx *gin.Context, relationName string) error {
	modelUIDs, err := relationModelUIDs(relationName)
	if err != nil {
		return err
	}
	return h.authorizeModels(ctx, modelUIDs...)
}

func (h *RelationResourceHandler) filterRelations(ctx *gin.Context, relations []domain.ResourceRelation) ([]domain.ResourceRelation, error) {
	modelUIDs := make([]string, 0, len(relations)*2)
	for _, item := range relations {
		modelUIDs = append(modelUIDs, item.SourceModelUID, item.TargetModelUID)
	}
	allowed, err := h.allowedModelUIDs(ctx, modelUIDs)
	if err != nil {
		return nil, err
	}
	return filterRelationsByAllowed(relations, allowed), nil
}

func filterRelationsByAllowed(relations []domain.ResourceRelation, allowed map[string]struct{}) []domain.ResourceRelation {
	result := make([]domain.ResourceRelation, 0, len(relations))
	for _, item := range relations {
		_, sourceAllowed := allowed[item.SourceModelUID]
		_, targetAllowed := allowed[item.TargetModelUID]
		if sourceAllowed && targetAllowed {
			result = append(result, item)
		}
	}
	return result
}

func (h *RelationResourceHandler) filterAggregated(ctx *gin.Context, items []domain.ResourceAggregatedAssets) ([]domain.ResourceAggregatedAssets, error) {
	modelUIDs := make([]string, 0, len(items))
	for _, item := range items {
		modelUIDs = append(modelUIDs, item.ModelUid)
	}
	allowed, err := h.allowedModelUIDs(ctx, modelUIDs)
	if err != nil {
		return nil, err
	}
	return filterAggregatedByAllowed(items, allowed), nil
}

func filterAggregatedByAllowed(items []domain.ResourceAggregatedAssets, allowed map[string]struct{}) []domain.ResourceAggregatedAssets {
	result := make([]domain.ResourceAggregatedAssets, 0, len(items))
	for _, item := range items {
		if _, exists := allowed[item.ModelUid]; exists {
			result = append(result, item)
		}
	}
	return result
}

func relationModelAccessError(err error) (ginx.Result, error) {
	if errors.Is(err, errModelAccessDenied) {
		return ginx.Result{Code: 403001, Msg: "无权访问该模型资产"}, nil
	}
	return systemErrorResult, err
}

func (h *RelationResourceHandler) PrivateRoute(server *gin.Engine) {
	g := server.Group("/api/resource/relation")
	// 资源关联关系
	g.POST("/create", ginx.WrapBody[CreateResourceRelationReq](h.CreateResourceRelation))

	// TODO 暂不使用，没有根据 relationName 进行筛选，会返回所有的结果
	g.POST("/list/src", ginx.WrapBody[ListResourceDiagramReq](h.ListSrcResource))
	g.POST("/list/dst", ginx.WrapBody[ListResourceDiagramReq](h.ListDstResource))

	// 列表聚合处理、通过聚合处理，为前端友好展示
	g.POST("/pipeline/src", ginx.WrapBody[ListResourceDiagramReq](h.ListSrcAggregated))
	g.POST("/pipeline/dst", ginx.WrapBody[ListResourceDiagramReq](h.ListDstAggregated))
	g.POST("/pipeline/all", ginx.WrapBody[ListResourceDiagramReq](h.ListAllAggregated))

	g.POST("/delete", ginx.WrapBody[DeleteResourceRelationReq](h.DeleteResourceRelation))
}

func (h *RelationResourceHandler) CreateResourceRelation(ctx *gin.Context, req CreateResourceRelationReq) (ginx.Result, error) {
	if err := h.authorizeRelation(ctx, req.RelationName); err != nil {
		return relationModelAccessError(err)
	}
	resp, err := h.svc.CreateResourceRelation(ctx, domain.ResourceRelation{
		RelationName:     req.RelationName,
		SourceResourceID: req.SourceResourceID,
		TargetResourceID: req.TargetResourceID,
	})

	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg:  "创建资源关联关系成功",
		Data: resp,
	}, nil
}

func (h *RelationResourceHandler) ListSrcResource(ctx *gin.Context, req ListResourceDiagramReq) (ginx.Result, error) {
	if err := h.authorizeModels(ctx, req.ModelUid); err != nil {
		return relationModelAccessError(err)
	}
	rrs, total, err := h.svc.ListSrcResources(ctx, req.ModelUid, req.ResourceId)
	if err != nil {
		return systemErrorResult, err
	}
	rrs, err = h.filterRelations(ctx, rrs)
	if err != nil {
		return systemErrorResult, err
	}
	total = int64(len(rrs))

	return ginx.Result{
		Data: RetrieveRelationResource{
			Total: total,
			ResourceRelations: slice.Map(rrs, func(idx int, src domain.ResourceRelation) ResourceRelation {
				return h.toResourceRelationVo(src)
			}),
		},
	}, nil
}

func (h *RelationResourceHandler) ListDstResource(ctx *gin.Context, req ListResourceDiagramReq) (ginx.Result, error) {
	if err := h.authorizeModels(ctx, req.ModelUid); err != nil {
		return relationModelAccessError(err)
	}
	rrs, total, err := h.svc.ListDstResources(ctx, req.ModelUid, req.ResourceId)
	if err != nil {
		return systemErrorResult, err
	}
	rrs, err = h.filterRelations(ctx, rrs)
	if err != nil {
		return systemErrorResult, err
	}
	total = int64(len(rrs))

	return ginx.Result{
		Data: RetrieveRelationResource{
			Total: total,
			ResourceRelations: slice.Map(rrs, func(idx int, src domain.ResourceRelation) ResourceRelation {
				return h.toResourceRelationVo(src)
			}),
		},
	}, nil
}

func (h *RelationResourceHandler) ListSrcAggregated(ctx *gin.Context, req ListResourceDiagramReq) (ginx.Result, error) {
	if err := h.authorizeModels(ctx, req.ModelUid); err != nil {
		return relationModelAccessError(err)
	}
	agg, err := h.svc.ListSrcAggregated(ctx, req.ModelUid, req.ResourceId)
	if err != nil {
		return ginx.Result{}, err
	}
	agg, err = h.filterAggregated(ctx, agg)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: slice.Map(agg, func(idx int, src domain.ResourceAggregatedAssets) RetrieveAggregatedAssets {
			return h.toAggregatedAssetsVo(src)
		}),
	}, nil
}

func (h *RelationResourceHandler) ListDstAggregated(ctx *gin.Context, req ListResourceDiagramReq) (ginx.Result, error) {
	if err := h.authorizeModels(ctx, req.ModelUid); err != nil {
		return relationModelAccessError(err)
	}
	agg, err := h.svc.ListDstAggregated(ctx, req.ModelUid, req.ResourceId)
	if err != nil {
		return ginx.Result{}, err
	}
	agg, err = h.filterAggregated(ctx, agg)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: slice.Map(agg, func(idx int, src domain.ResourceAggregatedAssets) RetrieveAggregatedAssets {
			return h.toAggregatedAssetsVo(src)
		}),
	}, nil
}

func (h *RelationResourceHandler) ListAllAggregated(ctx *gin.Context, req ListResourceDiagramReq) (ginx.Result, error) {
	if err := h.authorizeModels(ctx, req.ModelUid); err != nil {
		return relationModelAccessError(err)
	}
	var (
		eg   errgroup.Group
		srcS []domain.ResourceAggregatedAssets
		dstS []domain.ResourceAggregatedAssets
	)

	eg.Go(func() error {
		var err error
		srcS, err = h.svc.ListSrcAggregated(ctx, req.ModelUid, req.ResourceId)
		return err
	})

	eg.Go(func() error {
		var err error
		dstS, err = h.svc.ListDstAggregated(ctx, req.ModelUid, req.ResourceId)
		return err
	})
	if err := eg.Wait(); err != nil {
		return systemErrorResult, err
	}
	result := append(srcS, dstS...)
	result, err := h.filterAggregated(ctx, result)
	if err != nil {
		return systemErrorResult, err
	}
	sort.Slice(result, func(i, j int) bool {
		// 根据需要的排序逻辑进行排序，这里假设你有一个字段可以用来排序，比如 id
		return result[i].Total < result[j].Total
	})

	return ginx.Result{
		Data: slice.Map(result, func(idx int, src domain.ResourceAggregatedAssets) RetrieveAggregatedAssets {
			return RetrieveAggregatedAssets{
				RelationName: src.RelationName,
				ModelUid:     src.ModelUid,
				Total:        src.Total,
				ResourceIds:  src.ResourceIds,
			}
		}),
	}, nil
}

func (h *RelationResourceHandler) DeleteResourceRelation(ctx *gin.Context, req DeleteResourceRelationReq) (ginx.Result, error) {
	if err := h.authorizeModels(ctx, req.ModelUid); err != nil {
		return relationModelAccessError(err)
	}
	if err := h.authorizeRelation(ctx, req.RelationName); err != nil {
		return relationModelAccessError(err)
	}
	var (
		id  int64
		err error
	)

	rn := strings.Split(req.RelationName, "_")
	if rn[0] == req.ModelUid {
		id, err = h.svc.DeleteSrcRelation(ctx, req.ResourceId, req.ModelUid, req.RelationName)
	} else {
		id, err = h.svc.DeleteDstRelation(ctx, req.ResourceId, req.ModelUid, req.RelationName)
	}

	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: id,
	}, nil
}

func (h *RelationResourceHandler) toResourceRelationVo(src domain.ResourceRelation) ResourceRelation {
	return ResourceRelation{
		ID:               src.ID,
		SourceModelUID:   src.SourceModelUID,
		TargetModelUID:   src.TargetModelUID,
		SourceResourceID: src.SourceResourceID,
		TargetResourceID: src.TargetResourceID,
		RelationTypeUID:  src.RelationTypeUID,
		RelationName:     src.RelationName,
	}
}

func (h *RelationResourceHandler) toAggregatedAssetsVo(src domain.ResourceAggregatedAssets) RetrieveAggregatedAssets {
	return RetrieveAggregatedAssets{
		RelationName: src.RelationName,
		ModelUid:     src.ModelUid,
		Total:        src.Total,
		ResourceIds:  src.ResourceIds,
	}
}
