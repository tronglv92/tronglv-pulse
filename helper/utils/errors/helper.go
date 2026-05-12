package errors

import (
	"fmt"
	"net/http"
)

const (
	DefaultMsg = "Something went wrong"
)

var (
	NoPermission = Forbidden(
		fmt.Errorf("missing permission"),
	)

	DuplicateData = Forbidden(
		fmt.Errorf("duplicate data"),
	)
)

func BadRequest(err error) *Error {
	return NewBadRequest(ErrorReasonAndMessage(err))
}

func NewBadRequest(reason, message string) *Error {
	return New(http.StatusBadRequest, reason, message)
}

func Unauthorized(err error) *Error {
	return NewUnauthorized(ErrorReasonAndMessage(err))
}

func NewUnauthorized(reason, message string) *Error {
	return New(http.StatusUnauthorized, reason, message)
}

func Forbidden(err error) *Error {
	return NewForbidden(ErrorReasonAndMessage(err))
}

func NewForbidden(reason, message string) *Error {
	return New(http.StatusForbidden, reason, message)
}

func NotFound(err error) *Error {
	return NewNotFound(ErrorReasonAndMessage(err))
}

func NewNotFound(reason, message string) *Error {
	return New(http.StatusNotFound, reason, message)
}

func Conflict(err error) *Error {
	return NewConflict(ErrorReasonAndMessage(err))
}

func NewConflict(reason, message string) *Error {
	return New(http.StatusConflict, reason, message)
}

func RequestTimeout(err error) *Error {
	return NewRequestTimeout(ErrorReasonAndMessage(err))
}

func NewRequestTimeout(reason, message string) *Error {
	return New(http.StatusRequestTimeout, reason, message)
}

func InternalServer(err error) *Error {
	return NewInternalServer(ErrorReasonAndMessage(err))
}

func NewInternalServer(reason, message string) *Error {
	return New(http.StatusInternalServerError, reason, message)
}

func ServiceUnavailable(err error) *Error {
	return NewServiceUnavailable(ErrorReasonAndMessage(err))
}

func NewServiceUnavailable(reason, message string) *Error {
	return New(http.StatusBadGateway, reason, message)
}

func Database(err error) *Error {
	if IsRecordNotFound(err) {
		return NewNotFound("RECORD_NOT_FOUND", err.Error())
	}
	return NewDatabase(err.Error(), "A database error occurred.")
}

func NewDatabase(reason, message string) *Error {
	return New(http.StatusInternalServerError, reason, message)
}
