package operationlog

import (
	"fmt"
	"time"

	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/gin-gonic/gin"
)

type Handler struct{ svc Service }

func NewHandler(svc Service) *Handler { return &Handler{svc: svc} }

type listReq struct {
	Offset         int    `json:"offset"`
	Limit          int    `json:"limit"`
	Account        string `json:"account"`
	OperationModel string `json:"operation_model"`
	OperationType  string `json:"operation_type"`
	StartTime      string `json:"start_time"`
	EndTime        string `json:"end_time"`
}

func (h *Handler) PrivateRoutes(server *gin.Engine) {
	server.Group("/api/operation-log").POST("/list", ginx.WrapBody[listReq](h.List))
}

func (h *Handler) List(ctx *gin.Context, req listReq) (ginx.Result, error) {
	start, err := parseOptionalTime(req.StartTime)
	if err != nil {
		return ginx.Result{}, fmt.Errorf("开始日期格式错误: %w", err)
	}
	end, err := parseOptionalTime(req.EndTime)
	if err != nil {
		return ginx.Result{}, fmt.Errorf("结束日期格式错误: %w", err)
	}
	logs, total, err := h.svc.List(ctx, Query{
		Offset:         req.Offset,
		Limit:          req.Limit,
		Account:        req.Account,
		OperationModel: req.OperationModel,
		OperationType:  req.OperationType,
		StartTime:      start,
		EndTime:        end,
	})
	if err != nil {
		return ginx.Result{}, err
	}
	return ginx.Result{Data: map[string]any{"logs": logs, "total": total}}, nil
}

func parseOptionalTime(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}
