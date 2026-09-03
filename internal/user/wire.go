//go:build wireinject

package user

import (
	"github.com/RediSearch/redisearch-go/v2/redisearch"
	"github.com/ecodeclub/ginx/session"
	"github.com/google/wire"
	"github.com/userreksai/ecmdb-main/internal/department"
	"github.com/userreksai/ecmdb-main/internal/pkg/servicetoken"
	"github.com/userreksai/ecmdb-main/internal/policy"
	"github.com/userreksai/ecmdb-main/internal/user/internal/grpc"
	"github.com/userreksai/ecmdb-main/internal/user/internal/repository"
	"github.com/userreksai/ecmdb-main/internal/user/internal/repository/cache"
	"github.com/userreksai/ecmdb-main/internal/user/internal/repository/dao"
	"github.com/userreksai/ecmdb-main/internal/user/internal/service"
	"github.com/userreksai/ecmdb-main/internal/user/internal/web"
	"github.com/userreksai/ecmdb-main/internal/user/ldapx"
	"github.com/userreksai/ecmdb-main/pkg/cryptox"
	"github.com/userreksai/ecmdb-main/pkg/mongox"
)

var ProviderSet = wire.NewSet(
	service.NewLdapService,
	service.NewService,
	repository.NewResourceRepository,
	grpc.NewUserServer,
	dao.NewUserDao,
	web.NewHandler,
)

func InitLdapUserCache(conn *redisearch.Client) cache.RedisearchLdapUserCache {
	return cache.NewRedisearchLdapUserCache(conn)
}

func InitCrypto(reg *cryptox.CryptoRegistry) cryptox.Crypto {
	return reg.User
}

func InitModule(db *mongox.Mongo, redisClient *redisearch.Client, ldapConfig ldapx.Config, policyModule *policy.Module,
	departmentModule *department.Module, sp session.Provider, crypto *cryptox.CryptoRegistry, tokenMgr *servicetoken.Manager) (*Module, error) {
	wire.Build(
		ProviderSet,
		InitLdapUserCache,
		InitCrypto,
		wire.Struct(new(Module), "*"),
		wire.FieldsOf(new(*department.Module), "Svc"),
		wire.FieldsOf(new(*policy.Module), "Svc"),
	)
	return new(Module), nil
}
