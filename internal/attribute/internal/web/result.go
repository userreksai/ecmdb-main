package web

import (
	"github.com/userreksai/ecmdb-main/internal/attribute/internal/errs"
	"github.com/userreksai/ecmdb-main/pkg/ginx"
)

var (
	systemErrorResult = ginx.Result{
		Code: errs.SystemError.Code,
		Msg:  errs.SystemError.Msg,
	}

	duplicateErrorResult = ginx.Result{
		Code: 500001,
		Msg:  "唯一标识冲突",
	}
)
