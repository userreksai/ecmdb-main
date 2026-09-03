package web

import (
	"github.com/userreksai/ecmdb-main/internal/engine/internal/errs"
	"github.com/userreksai/ecmdb-main/pkg/ginx"
)

var (
	systemErrorResult = ginx.Result{
		Code: errs.SystemError.Code,
		Msg:  errs.SystemError.Msg,
	}

	validateErrorResult = ginx.Result{
		Code: errs.ValidationError.Code,
		Msg:  errs.ValidationError.Msg,
	}
)
