package relation

import (
	"github.com/userreksai/ecmdb-main/internal/relation/internal/domain"
	"github.com/userreksai/ecmdb-main/internal/relation/internal/service"
	"github.com/userreksai/ecmdb-main/internal/relation/internal/web"
)

// RR => RelationResource
// RM => RelationModel
// RT => RelationType

type RRSvc = service.RelationResourceService
type RRHandler = web.RelationResourceHandler

type RMSvc = service.RelationModelService
type RMHandler = web.RelationModelHandler

type RTSvc = service.RelationTypeService
type RTHandler = web.RelationTypeHandler

type ModelDiagram = domain.ModelDiagram

type ModelRelation = domain.ModelRelation

type RelationType = domain.RelationType

type ResourceRelation = domain.ResourceRelation
