package web

import (
	"github.com/userreksai/ecmdb-main/internal/model/internal/errs"
	"github.com/userreksai/ecmdb-main/pkg/ginx"
)

var (
	systemErrorResult = ginx.Result{
		Code: errs.SystemError.Code,
		Msg:  errs.SystemError.Msg,
	}

	modelRelationIsNotFountResult = ginx.Result{
		Code: errs.RelationIsNotFountResult.Code,
		Msg:  errs.RelationIsNotFountResult.Msg,
	}
)
