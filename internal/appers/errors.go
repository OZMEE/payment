package appers

import (
	"fmt"
	"net/http"
)

type ErrorResp struct {
	Code       int
	Status     string
	FieldName  string
	FieldValue string
	Msg        string
}

func (e *ErrorResp) Error() string {
	return fmt.Sprintf("%d - %s, (msg: %s)", e.Code, e.Status, e.Msg)
}

func (e *ErrorResp) SetMsg(msg string) *ErrorResp {
	e.Msg = msg
	return e
}

func NewErrValidation(msg string, fieldName string, value string) *ErrorResp {
	return &ErrorResp{Code: http.StatusBadRequest, Status: "Error validation", Msg: msg, FieldName: fieldName, FieldValue: value}
}

func NewSqlExecutions(msg string) *ErrorResp {
	return &ErrorResp{Code: http.StatusInternalServerError, Status: "Sql execution err", Msg: msg}
}

var (
	ErrEventFound = &ErrorResp{
		Code:   http.StatusNotFound,
		Status: "payment not found",
	}
	ErrUnknown = &ErrorResp{
		Code:   http.StatusInternalServerError,
		Status: "Unknown error",
	}
	ErrDuplicatePayment = &ErrorResp{
		Code:   http.StatusBadRequest,
		Status: "Duplicate payment",
	}
	ErrParseJson = &ErrorResp{
		Code:   http.StatusBadRequest,
		Status: "Json parse error",
	}
)
