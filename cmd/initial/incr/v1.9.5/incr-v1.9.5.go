package v195

import (
	"context"

	"github.com/gotomicro/ego/core/elog"
	"github.com/userreksai/ecmdb-main/cmd/initial/backup"
	"github.com/userreksai/ecmdb-main/cmd/initial/incr"
	"github.com/userreksai/ecmdb-main/cmd/initial/ioc"
	"github.com/userreksai/ecmdb-main/cmd/initial/menu"
)

type incrV195 struct {
	App        *ioc.App
	ChangeSync *menu.ChangeSync
	logger     elog.Component
}

func NewIncrV195(app *ioc.App) incr.InitialIncr {
	return &incrV195{
		App:        app,
		ChangeSync: menu.NewChange(app),
		logger:     *elog.DefaultLogger,
	}
}

func (i *incrV195) Version() string {
	return "v1.9.5"
}

func (i *incrV195) Before(ctx context.Context) error {
	backupManager := backup.NewBackupManager(i.App)
	_, err := backupManager.BackupMongoCollection(ctx, "c_menu", backup.Options{
		Version:     i.Version(),
		Description: "v1.9.5 菜单更新前备份",
		Tags: map[string]string{
			"type":   "version_upgrade",
			"module": "operation_log_menu",
		},
	})
	return err
}

func (i *incrV195) Commit(ctx context.Context) error {
	return i.ChangeSync.UpdateMenu(ctx)
}

func (i *incrV195) After(ctx context.Context) error {
	return i.App.VerSvc.CreateOrUpdateVersion(ctx, i.Version())
}

func (i *incrV195) Rollback(context.Context) error {
	return nil
}
