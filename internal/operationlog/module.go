package operationlog

import (
	"context"
	"time"

	"github.com/gotomicro/ego/core/elog"
	"gorm.io/gorm"
)

type Module struct {
	Svc Service
	Hdl *Handler
}

func InitModule(db *gorm.DB) (*Module, error) {
	svc, err := NewService(db)
	if err != nil {
		return nil, err
	}
	if _, err = svc.CleanupExpired(context.Background()); err != nil {
		elog.DefaultLogger.Warn("cleanup expired operation logs failed", elog.FieldErr(err))
	}
	startCleanup(svc)
	return &Module{Svc: svc, Hdl: NewHandler(svc)}, nil
}

func startCleanup(svc Service) {
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			if _, err := svc.CleanupExpired(context.Background()); err != nil {
				elog.DefaultLogger.Warn("cleanup expired operation logs failed", elog.FieldErr(err))
			}
		}
	}()
}
