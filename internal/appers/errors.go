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
	Op         string
}

func (e *ErrorResp) Error() string {
	return fmt.Sprintf("%d - %s, (msg: %s)", e.Code, e.Status, e.Msg)
}

type ErrorBuilder struct {
	err *ErrorResp
}

func (e *ErrorResp) Builder() *ErrorBuilder {
	return &ErrorBuilder{
		err: &ErrorResp{
			Code:       e.Code,
			Status:     e.Status,
			FieldName:  e.FieldName,
			FieldValue: e.FieldValue,
			Msg:        e.Msg,
			Op:         e.Op,
		},
	}
}

func (e *ErrorBuilder) Msg(msg string) *ErrorBuilder {
	e.err.Msg = msg
	return e
}

func (e *ErrorBuilder) Op(op string) *ErrorBuilder {
	e.err.Op = op
	return e
}

func (e *ErrorBuilder) Field(name string, value string) *ErrorBuilder {
	e.err.FieldName = name
	e.err.FieldValue = value
	return e
}

func (e *ErrorBuilder) Build() *ErrorResp {
	return e.err
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
	ErrKafkaSendEvent = &ErrorResp{
		Code:   http.StatusBadRequest,
		Status: "Send event error",
	}
	ErrSqlExecutions = &ErrorResp{
		Code:   http.StatusInternalServerError,
		Status: "Sql execution error",
	}
	ErrValidation = &ErrorResp{
		Code:   http.StatusBadRequest,
		Status: "Validation error",
	}
	ErrTypeAssertion = &ErrorResp{
		Code:   http.StatusInternalServerError,
		Status: "Type assertion error",
	}
)
