package appers

import (
	"fmt"
	"net/http"
)

type ErrorResp struct {
	Code int
	Msg  string
}

func (e ErrorResp) Error() string {
	return fmt.Sprintf("%d - %s", e.Code, e.Msg)
}

func NewErrValidation(msg string) *ErrorResp {
	return &ErrorResp{Code: http.StatusBadRequest, Msg: "ErrValidation: " + msg}
}

func NewSqlExecutions(msg string) *ErrorResp {
	return &ErrorResp{Code: http.StatusInternalServerError, Msg: "SqlExecutions: " + msg}
}

var (
	ErrEventFound = &ErrorResp{
		Code: http.StatusNotFound,
		Msg:  "payment not found",
	}
	ErrUnknown = &ErrorResp{
		Code: http.StatusInternalServerError,
		Msg:  "Unknown error",
	}
	ErrDuplicatePayment = &ErrorResp{
		Code: http.StatusBadRequest,
		Msg:  "Duplicate payment",
	}
)
