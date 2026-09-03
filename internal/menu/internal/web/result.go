package web

import (
	"github.com/userreksai/ecmdb-main/internal/menu/internal/errs"
	"github.com/userreksai/ecmdb-main/pkg/ginx"
)

var (
	systemErrorResult = ginx.Result{
		Code: errs.SystemError.Code,
		Msg:  errs.SystemError.Msg,
	}
	menuHasChildrenResult = ginx.Result{
		Code: errs.MenuHasChildren.Code,
		Msg:  errs.MenuHasChildren.Msg,
	}
)
