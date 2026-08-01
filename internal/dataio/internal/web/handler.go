package web

import (
	"errors"
	"fmt"

	"github.com/Duke1616/ecmdb/internal/dataio/internal/service"
	"github.com/Duke1616/ecmdb/internal/pkg/authctx"
	"github.com/Duke1616/ecmdb/internal/policy"
	"github.com/Duke1616/ecmdb/internal/resource"
	"github.com/Duke1616/ecmdb/internal/role"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/Duke1616/ecmdb/pkg/storage"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc       service.IDataIOService
	storage   *storage.S3Storage
	roleSvc   role.Service
	policySvc policy.Service
}

func NewHandler(svc service.IDataIOService, storage *storage.S3Storage, roleSvc role.Service, policySvc policy.Service) *Handler {
	return &Handler{
		svc:       svc,
		storage:   storage,
		roleSvc:   roleSvc,
		policySvc: policySvc,
	}
}

var errModelAccessDenied = errors.New("model asset access denied")

func (h *Handler) authorizeModel(ctx *gin.Context, modelUID string) error {
	if h.roleSvc == nil || h.policySvc == nil {
		return nil
	}
	uid, ok := authctx.UID(ctx)
	if !ok {
		return fmt.Errorf("get current user failed")
	}
	roleCodes, err := h.policySvc.GetRolesForUser(ctx, uid)
	if err != nil {
		return err
	}
	allowed, err := h.roleSvc.CanAccessModel(ctx, roleCodes, modelUID)
	if err != nil {
		return err
	}
	if !allowed {
		return errModelAccessDenied
	}
	return nil
}

func modelAccessError(err error) (ginx.Result, error) {
	if errors.Is(err, errModelAccessDenied) {
		return ginx.Result{Code: 403001, Msg: "无权访问该模型资产"}, nil
	}
	return systemErrorResult, err
}

func (h *Handler) PrivateRoutes(server *gin.Engine) {
	g := server.Group("/api/dataio")
	// 导出模板
	g.GET("/template/export/:model_uid", ginx.Wrap(h.ExportTemplate))
	// 导入数据 (S3 模式)
	g.POST("/import", ginx.WrapBody[ImportReq](h.Import))
	// 导出数据
	g.POST("/export", ginx.WrapBody[ExportReq](h.Export))
}

// Export 导出数据
func (h *Handler) Export(ctx *gin.Context, req ExportReq) (ginx.Result, error) {
	if err := h.authorizeModel(ctx, req.ModelUID); err != nil {
		return modelAccessError(err)
	}
	for _, field := range req.RelatedFields {
		if field.ModelUID == "" {
			return systemErrorResult, fmt.Errorf("related model uid is required")
		}
		if err := h.authorizeModel(ctx, field.ModelUID); err != nil {
			return modelAccessError(err)
		}
	}

	// 转换 FilterGroups
	groups := slice.Map(req.FilterGroups, func(idx int, src ExportFilterGroup) resource.FilterGroup {
		return resource.FilterGroup{
			Filters: slice.Map(src.Filters, func(idx int, src ExportFilterCondition) resource.FilterCondition {
				return resource.FilterCondition{
					FieldUID: src.FieldUID,
					Operator: resource.Operator(src.Operator),
					Value:    src.Value,
				}
			}),
		}
	})

	params := service.ExportParams{
		ModelUID:     req.ModelUID,
		Scope:        req.Scope.String(),
		ResourceIDs:  req.ResourceIDs,
		FilterGroups: groups,
		Fields:       req.Fields,
		RelatedFields: slice.Map(req.RelatedFields, func(idx int, src ExportRelatedField) service.RelatedFieldParam {
			return service.RelatedFieldParam{
				RelationName: src.RelationName,
				ModelUID:     src.ModelUID,
				FieldUID:     src.FieldUID,
			}
		}),
		FileName: req.FileName,
	}

	// 调用 Service 导出数据
	excelData, err := h.svc.Export(ctx.Request.Context(), params)
	if err != nil {
		return systemErrorResult, err
	}

	fileName := req.FileName
	if fileName == "" {
		fileName = req.ModelUID + "_export.xlsx"
	}
	// 确保后缀
	if len(fileName) < 5 || fileName[len(fileName)-5:] != ".xlsx" {
		fileName += ".xlsx"
	}

	// 设置 HTTP 响应头
	ctx.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	ctx.Header("Content-Disposition", "attachment; filename="+fileName)
	ctx.Header("Content-Transfer-Encoding", "binary")

	// 直接写入 Excel 数据
	ctx.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", excelData)

	return ginx.Result{}, nil
}

// ExportTemplate 导出空白导入模板
func (h *Handler) ExportTemplate(ctx *gin.Context) (ginx.Result, error) {
	// 根据请求获取模型UID
	modelUid := ctx.Param("model_uid")
	if err := h.authorizeModel(ctx, modelUid); err != nil {
		return modelAccessError(err)
	}

	// 调用 Service 生成 Excel 模板
	excelData, err := h.svc.ExportTemplate(ctx.Request.Context(), modelUid)
	if err != nil {
		return systemErrorResult, err
	}

	// 设置 HTTP 响应头,直接返回 Excel 文件
	ctx.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	ctx.Header("Content-Disposition", "attachment; filename="+modelUid+"_template.xlsx")
	ctx.Header("Content-Transfer-Encoding", "binary")

	// 直接写入 Excel 数据
	ctx.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", excelData)

	// NOTE: 返回空 Result,因为已经通过 ctx.Data 直接发送了响应
	return ginx.Result{}, nil
}

// Import 导入数据
// NOTE: 前端先通过 GenerateUploadURL 上传文件到 S3,然后调用此接口传入 file_key 进行导入
func (h *Handler) Import(ctx *gin.Context, req ImportReq) (ginx.Result, error) {
	if err := h.authorizeModel(ctx, req.ModelUID); err != nil {
		return modelAccessError(err)
	}

	// 1. 从 S3 下载文件
	fileData, err := h.storage.GetFile(ctx.Request.Context(), "ecmdb", req.FileKey)
	if err != nil {
		return systemErrorResult, err
	}

	// 2. 调用 Service 导入数据
	importedCount, err := h.svc.Import(ctx.Request.Context(), req.ModelUID, fileData)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg: "导入成功",
		Data: map[string]interface{}{
			"imported_count": importedCount,
		},
	}, nil
}
