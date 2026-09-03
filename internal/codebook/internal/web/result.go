package web

import (
	"github.com/userreksai/ecmdb-main/internal/codebook/internal/errs"
	"github.com/userreksai/ecmdb-main/pkg/ginx"
)

var (
	systemErrorResult = ginx.Result{
		Code: errs.SystemError.Code,
		Msg:  errs.SystemError.Msg,
	}
)
