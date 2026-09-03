package middleware

import (
	"net/http"
	"time"

	"github.com/ecodeclub/ginx/gctx"
	"github.com/ecodeclub/ginx/session"
	"github.com/gin-gonic/gin"
	"github.com/gotomicro/ego/core/elog"
	"github.com/userreksai/ecmdb-main/internal/pkg/authctx"
	"github.com/userreksai/ecmdb-main/internal/pkg/servicetoken"
	"github.com/userreksai/ecmdb-main/internal/user"
)

type CheckLoginMiddlewareBuilder struct {
	threshold time.Duration
	logger    *elog.Component
	sp        session.Provider
	userSvc   user.Service
	tokenMgr  *servicetoken.Manager
}

func NewCheckLoginMiddlewareBuilder(sp session.Provider, userSvc user.Service, tokenMgr *servicetoken.Manager) *CheckLoginMiddlewareBuilder {
	return &CheckLoginMiddlewareBuilder{
		logger:    elog.DefaultLogger,
		threshold: time.Minute * 1,
		sp:        sp,
		userSvc:   userSvc,
		tokenMgr:  tokenMgr,
	}
}

func (b *CheckLoginMiddlewareBuilder) Build() gin.HandlerFunc {
	threshold := b.threshold.Milliseconds()
	return func(ctx *gin.Context) {
		if b.checkServiceToken(ctx) {
			return
		}

		gCtx := &gctx.Context{Context: ctx}
		sess, err := b.sp.Get(gCtx)
		if err != nil {
			b.logger.Error("unauthorized", elog.FieldErr(err))
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		expiration := sess.Claims().Expiration
		now := time.Now().UnixMilli()
		if expiration <= now {
			b.logger.Error("token expired", elog.Int64("expiration", expiration), elog.Int64("now", now))
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if expiration-now < threshold {
			if err = b.sp.RenewAccessToken(gCtx); err != nil {
				b.logger.Warn("renew token failed", elog.FieldErr(err))
			}
		}

		ctx.Set(session.CtxSessionKey, sess)
		uid := sess.Claims().Uid
		username := ""
		if b.userSvc != nil && shouldResolveAuditUsername(ctx) {
			if currentUser, findErr := b.userSvc.FindById(ctx, uid); findErr == nil {
				username = currentUser.Username
			} else {
				b.logger.Warn("resolve current username failed", elog.Int64("uid", uid), elog.FieldErr(findErr))
			}
		}
		authctx.Set(ctx, authctx.Identity{
			UID:      uid,
			Username: username,
			AuthType: authctx.AuthTypeSession,
		})
	}
}

func shouldResolveAuditUsername(ctx *gin.Context) bool {
	if ctx.Request.Method != http.MethodPost {
		return false
	}
	switch ctx.Request.URL.Path {
	case "/api/resource/create", "/api/resource/update", "/api/resource/set_custom_field",
		"/api/resource/delete", "/api/dataio/import":
		return true
	default:
		return false
	}
}

func (b *CheckLoginMiddlewareBuilder) checkServiceToken(ctx *gin.Context) bool {
	if b.tokenMgr == nil || !b.tokenMgr.Enabled() {
		return false
	}

	tokenString := servicetoken.ExtractBearerToken(ctx.GetHeader("Authorization"))
	if tokenString == "" {
		return false
	}

	claims, err := b.tokenMgr.Verify(tokenString)
	if err != nil {
		return false
	}

	if b.userSvc != nil {
		u, err := b.userSvc.FindById(ctx, claims.UID)
		if err != nil || u.Username != claims.Username || u.Status.ToUint8() != 1 || !b.tokenMgr.IsServiceAccount(u.Username) {
			b.logger.Warn("service token user validation failed",
				elog.Int64("uid", claims.UID),
				elog.String("username", claims.Username),
				elog.FieldErr(err))
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return true
		}
	}

	authctx.Set(ctx, authctx.Identity{
		UID:      claims.UID,
		Username: claims.Username,
		AuthType: authctx.AuthTypeServiceToken,
	})
	return true
}
