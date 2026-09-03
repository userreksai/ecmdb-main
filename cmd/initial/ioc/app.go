package ioc

import (
	templatev1 "github.com/userreksai/ecmdb-main/api/proto/gen/ealert/template/v1"
	"github.com/userreksai/ecmdb-main/cmd/initial/version"
	"github.com/userreksai/ecmdb-main/internal/bootstrap"
	"github.com/userreksai/ecmdb-main/internal/menu"
	"github.com/userreksai/ecmdb-main/internal/permission"
	"github.com/userreksai/ecmdb-main/internal/policy"
	"github.com/userreksai/ecmdb-main/internal/role"
	"github.com/userreksai/ecmdb-main/internal/user"
	"github.com/userreksai/ecmdb-main/internal/workflow"
	"github.com/userreksai/ecmdb-main/pkg/mongox"
	"gorm.io/gorm"
)

type App struct {
	UserSvc        user.Service
	RoleSvc        role.Service
	MenuSvc        menu.Service
	PermissionSvc  permission.Service
	policySvc      policy.Service
	VerSvc         version.Service
	BootstrapSvc   bootstrap.Service
	TemplateClient templatev1.TemplateServiceClient
	WorkflowSvc    workflow.Service
	GormDB         *gorm.DB
	DB             *mongox.Mongo
}
