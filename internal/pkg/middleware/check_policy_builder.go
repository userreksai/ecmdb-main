package middleware

import (
	"net/http"
	"strconv"

	"github.com/ecodeclub/ginx"
	"github.com/ecodeclub/ginx/session"
	"github.com/gin-gonic/gin"
	"github.com/gotomicro/ego/core/elog"
	"github.com/userreksai/ecmdb-main/internal/pkg/authctx"
	"github.com/userreksai/ecmdb-main/internal/policy"
)

type CheckPolicyMiddlewareBuilder struct {
	svc    policy.Service
	logger *elog.Component
	sp     session.Provider
}

func NewCheckPolicyMiddlewareBuilder(svc policy.Service, sp session.Provider) *CheckPolicyMiddlewareBuilder {
	return &CheckPolicyMiddlewareBuilder{
		svc:    svc,
		logger: elog.DefaultLogger,
		sp:     sp,
	}
}

func (c *CheckPolicyMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		uid, ok := authctx.UID(ctx)
		if !ok {
			sess, err := c.sp.Get(&ginx.Context{Context: ctx})
			if err != nil {
				c.logger.Warn("user not logged in", elog.FieldErr(err))
				ctx.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			uid = sess.Claims().Uid
		}

		path := ctx.Request.URL.Path
		method := ctx.Request.Method
		result, err := c.svc.Authorize(ctx.Request.Context(), strconv.FormatInt(uid, 10), path, method, "CMDB")
		if err != nil {
			c.logger.Error("policy service failed", elog.FieldErr(err), elog.Int64("uid", uid), elog.String("path", path))
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if !result.Allowed {
			c.logger.Warn("user access denied",
				elog.Int64("uid", uid),
				elog.String("path", path),
				elog.String("method", method),
				elog.String("reason", result.Reason),
				elog.Any("roles", result.Roles),
				elog.Any("matched_policies", result.MatchedPolicies))
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}

		ctx.Next()
	}
}
