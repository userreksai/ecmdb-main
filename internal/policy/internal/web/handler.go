package web

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Duke1616/ecmdb/internal/pkg/servicetoken"
	"github.com/Duke1616/ecmdb/internal/policy/internal/domain"
	"github.com/Duke1616/ecmdb/internal/policy/internal/service"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/ecodeclub/ekit/slice"
	"github.com/ecodeclub/ginx/gctx"
	"github.com/ecodeclub/ginx/session"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc       service.Service
	sp        session.Provider
	tokenMgr  *servicetoken.Manager
	threshold time.Duration
}

func NewHandler(svc service.Service, sp session.Provider, tokenMgr *servicetoken.Manager) *Handler {
	return &Handler{
		svc:       svc,
		sp:        sp,
		tokenMgr:  tokenMgr,
		threshold: time.Minute,
	}
}

func (h *Handler) PublicRoutes(server *gin.Engine) {
	g := server.Group("/api/policy")
	g.POST("/check_login", ginx.Wrap(h.CheckLoginForSDK))
	g.POST("/check_policy", ginx.WrapBody[CheckPolicyReq](h.CheckPolicyForSDK))

	compat := server.Group("/api/permission")
	compat.POST("/check_login", ginx.Wrap(h.CheckLoginForSDK))
	compat.POST("/check_policy", ginx.WrapBody[EIAMCheckPolicyReq](h.CheckPolicyForEIAM))
}

func (h *Handler) PrivateRoutes(server *gin.Engine) {
	g := server.Group("/api/policy")
	g.POST("/add/p", ginx.WrapBody[PolicyReq](h.AddPolicies))
	g.POST("/update/p", ginx.WrapBody[PolicyReq](h.UpdatePolicies))
	g.POST("/add/g", ginx.WrapBody[AddGroupingPolicyReq](h.AddGroupingPolicy))
	g.POST("/authorize", ginx.WrapBody[AuthorizeReq](h.Authorize))
	g.POST("/user/permissions", ginx.WrapBody[GetPermissionsForUserReq](h.GetImplicitPermissionsForUser))
	g.POST("/role/permissions", ginx.WrapBody[GetPermissionsForRoleReq](h.GetPermissionsForRole))
}

func (h *Handler) CheckLoginForSDK(ctx *gin.Context) (ginx.Result, error) {
	uid, serviceToken, expiration, err := h.currentSDKIdentity(ctx)
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return ginx.Result{}, fmt.Errorf("login verification failed: %w", err)
	}

	if serviceToken {
		return ginx.Result{
			Data: CheckLoginResp{
				Uid: uid,
			},
		}, nil
	}

	now := time.Now().UnixMilli()
	if expiration <= now {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return ginx.Result{}, fmt.Errorf("token expired")
	}

	if expiration-now < h.threshold.Milliseconds() {
		_ = h.sp.RenewAccessToken(&gctx.Context{Context: ctx})
	}

	return ginx.Result{
		Data: CheckLoginResp{
			Uid: uid,
		},
	}, nil
}

func (h *Handler) CheckPolicyForSDK(ctx *gin.Context, req CheckPolicyReq) (ginx.Result, error) {
	uid, _, _, err := h.currentSDKIdentity(ctx)
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return ginx.Result{}, fmt.Errorf("login verification failed: %w", err)
	}

	userId := strconv.FormatInt(uid, 10)
	result, err := h.svc.Authorize(ctx, userId, req.Path, req.Method, policyResource(req.Service, req.Resource))
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: result,
	}, nil
}

func (h *Handler) CheckPolicyForEIAM(ctx *gin.Context, req EIAMCheckPolicyReq) (ginx.Result, error) {
	uid, _, _, err := h.currentSDKIdentity(ctx)
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return ginx.Result{}, fmt.Errorf("login verification failed: %w", err)
	}

	userId := strconv.FormatInt(uid, 10)
	result, err := h.svc.Authorize(ctx, userId, req.Path, req.Method, policyResource(req.Service, req.Resource))
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: result,
	}, nil
}

func policyResource(service, resource string) string {
	resource = strings.TrimSpace(resource)
	if resource != "" {
		return resource
	}

	normalizedService := strings.ToLower(strings.TrimSpace(service))
	if idx := strings.Index(normalizedService, ":"); idx >= 0 {
		normalizedService = normalizedService[:idx]
	}

	switch normalizedService {
	case "task", "etask":
		return "TASK"
	case "cmdb":
		return "CMDB"
	case "alert":
		return "ALERT"
	default:
		return strings.ToUpper(strings.TrimSpace(service))
	}
}

func (h *Handler) currentSDKIdentity(ctx *gin.Context) (int64, bool, int64, error) {
	if h.tokenMgr != nil && h.tokenMgr.Enabled() {
		token := servicetoken.ExtractBearerToken(ctx.GetHeader("Authorization"))
		if token != "" {
			claims, err := h.tokenMgr.Verify(token)
			if err == nil {
				return claims.UID, true, 0, nil
			}
		}
	}

	sess, err := h.sp.Get(&gctx.Context{Context: ctx})
	if err != nil {
		return 0, false, 0, err
	}
	return sess.Claims().Uid, false, sess.Claims().Expiration, nil
}

func (h *Handler) AddPolicies(ctx *gin.Context, req PolicyReq) (ginx.Result, error) {
	ok, err := h.svc.AddPolicies(ctx, h.toDomain(req))
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: ok,
	}, nil
}

func (h *Handler) GetImplicitPermissionsForUser(ctx *gin.Context, req GetPermissionsForUserReq) (ginx.Result, error) {
	resp, err := h.svc.GetImplicitPermissionsForUser(ctx, req.UserId)
	if err != nil {
		return systemErrorResult, err
	}

	policies := slice.Map(resp, func(idx int, src domain.Policy) Policy {
		return Policy{
			Path:     src.Path,
			Method:   src.Method,
			Resource: src.Resource,
			Effect:   Effect(src.Effect),
		}
	})

	return ginx.Result{
		Msg: "OK",
		Data: RetrievePolicies{
			Policies: policies,
		},
	}, nil
}

func (h *Handler) GetPermissionsForRole(ctx *gin.Context, req GetPermissionsForRoleReq) (ginx.Result, error) {
	pers, err := h.svc.GetPermissionsForRole(ctx, req.RoleCode)
	if err != nil {
		return systemErrorResult, err
	}

	policies := slice.Map(pers, func(idx int, src domain.Policy) Policy {
		return Policy{
			Path:     src.Path,
			Method:   src.Method,
			Resource: src.Resource,
			Effect:   Effect(src.Effect),
		}
	})

	return ginx.Result{
		Msg: "OK",
		Data: RetrievePolicies{
			Policies: policies,
		},
	}, nil
}

func (h *Handler) UpdatePolicies(ctx *gin.Context, req PolicyReq) (ginx.Result, error) {
	ok, err := h.svc.CreateOrUpdateFilteredPolicies(ctx, h.toDomain(req))
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: ok,
	}, nil
}

func (h *Handler) Authorize(ctx *gin.Context, req AuthorizeReq) (ginx.Result, error) {
	result, err := h.svc.Authorize(ctx, req.UserId, req.Path, req.Method, req.Resource)
	if err != nil {
		return systemErrorResult, err
	}

	if !result.Allowed {
		return ginx.Result{
			Code: 0,
			Msg:  result.Reason,
			Data: result,
		}, nil
	}

	return ginx.Result{
		Code: 0,
		Msg:  result.Reason,
		Data: result,
	}, nil
}

func (h *Handler) AddGroupingPolicy(ctx *gin.Context, req AddGroupingPolicyReq) (ginx.Result, error) {
	ok, err := h.svc.AddGroupingPolicy(ctx, h.toDomainGroup(req))
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: ok,
	}, nil
}

func (h *Handler) toDomain(req PolicyReq) domain.Policies {
	return domain.Policies{
		RoleCode: req.RoleCode,
		Policies: slice.Map(req.Policies, func(idx int, src Policy) domain.Policy {
			return domain.Policy{
				Path:     src.Path,
				Method:   src.Method,
				Resource: src.Resource,
				Effect:   domain.Effect(src.Effect),
			}
		}),
	}
}

func (h *Handler) toDomainGroup(req AddGroupingPolicyReq) domain.AddGroupingPolicy {
	return domain.AddGroupingPolicy{
		UserId:   req.UserId,
		RoleCode: req.RoleCode,
	}
}
