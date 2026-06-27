package authctx

import "github.com/gin-gonic/gin"

const identityKey = "ecmdb.auth.identity"

const (
	AuthTypeSession      = "session"
	AuthTypeServiceToken = "service_token"
)

type Identity struct {
	UID      int64
	Username string
	AuthType string
}

func Set(ctx *gin.Context, identity Identity) {
	ctx.Set(identityKey, identity)
}

func Get(ctx *gin.Context) (Identity, bool) {
	val, ok := ctx.Get(identityKey)
	if !ok {
		return Identity{}, false
	}

	identity, ok := val.(Identity)
	return identity, ok
}

func UID(ctx *gin.Context) (int64, bool) {
	identity, ok := Get(ctx)
	if !ok || identity.UID == 0 {
		return 0, false
	}
	return identity.UID, true
}
