package policy

import (
	"github.com/userreksai/ecmdb-main/internal/policy/internal/domain"
	"github.com/userreksai/ecmdb-main/internal/policy/internal/grpc"
	"github.com/userreksai/ecmdb-main/internal/policy/internal/service"
	"github.com/userreksai/ecmdb-main/internal/policy/internal/web"
)

type Handler = web.Handler

type RpcServer = grpc.PolicyServer

type Service = service.Service

type Policy = domain.Policy

type Policies = domain.Policies

type BatchPolicies = domain.BatchPolicies
