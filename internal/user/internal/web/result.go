package web

import (
	"github.com/userreksai/ecmdb-main/internal/user/internal/errs"
	"github.com/userreksai/ecmdb-main/pkg/ginx"
)

var (
	systemErrorResult = ginx.Result{
		Code: errs.SystemError.Code,
		Msg:  errs.SystemError.Msg,
	}
	userOrPassErrorResult = ginx.Result{
		Code: errs.UserPassError.Code,
		Msg:  errs.UserPassError.Msg,
	}
)
