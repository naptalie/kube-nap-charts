// Package errs provides structured error handling with HTTP status mapping.
package errs

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime"
)

// ErrCode represents the type of error.
type ErrCode int

const (
	OK ErrCode = iota
	Canceled
	Unknown
	InvalidArgument
	DeadlineExceeded
	NotFound
	AlreadyExists
	PermissionDenied
	Unauthenticated
	ResourceExhausted
	FailedPrecondition
	Aborted
	OutOfRange
	Unimplemented
	Internal
	Unavailable
	DataLoss
	InternalOnlyLog // Internal error that should not be exposed to clients
)

// String implements the Stringer interface.
func (c ErrCode) String() string {
	switch c {
	case OK:
		return "OK"
	case Canceled:
		return "Canceled"
	case Unknown:
		return "Unknown"
	case InvalidArgument:
		return "InvalidArgument"
	case DeadlineExceeded:
		return "DeadlineExceeded"
	case NotFound:
		return "NotFound"
	case AlreadyExists:
		return "AlreadyExists"
	case PermissionDenied:
		return "PermissionDenied"
	case Unauthenticated:
		return "Unauthenticated"
	case ResourceExhausted:
		return "ResourceExhausted"
	case FailedPrecondition:
		return "FailedPrecondition"
	case Aborted:
		return "Aborted"
	case OutOfRange:
		return "OutOfRange"
	case Unimplemented:
		return "Unimplemented"
	case Internal:
		return "Internal"
	case Unavailable:
		return "Unavailable"
	case DataLoss:
		return "DataLoss"
	case InternalOnlyLog:
		return "InternalOnlyLog"
	default:
		return "Unknown"
	}
}

// Error represents an application error.
type Error struct {
	Code     ErrCode `json:"code"`
	Message  string  `json:"message"`
	FuncName string  `json:"-"`
	FileName string  `json:"-"`
}

// New creates a new Error with caller information.
func New(code ErrCode, err error) *Error {
	pc, filename, line, _ := runtime.Caller(1)
	funcName := runtime.FuncForPC(pc).Name()

	return &Error{
		Code:     code,
		Message:  err.Error(),
		FuncName: funcName,
		FileName: fmt.Sprintf("%s:%d", filename, line),
	}
}

// Newf creates a new Error with formatted message.
func Newf(code ErrCode, format string, v ...any) *Error {
	return New(code, fmt.Errorf(format, v...))
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.Message
}

// Encode implements the web.Encoder interface.
func (e *Error) Encode() ([]byte, string, error) {
	data, err := json.Marshal(e)
	if err != nil {
		return nil, "", fmt.Errorf("marshal error: %w", err)
	}
	return data, "application/json", nil
}

// HTTPStatus returns the HTTP status code for the error.
func (e *Error) HTTPStatus() int {
	return httpStatus[e.Code]
}

// httpStatus maps error codes to HTTP status codes.
var httpStatus = map[ErrCode]int{
	OK:                 http.StatusOK,
	Canceled:           http.StatusRequestTimeout,
	Unknown:            http.StatusInternalServerError,
	InvalidArgument:    http.StatusBadRequest,
	DeadlineExceeded:   http.StatusGatewayTimeout,
	NotFound:           http.StatusNotFound,
	AlreadyExists:      http.StatusConflict,
	PermissionDenied:   http.StatusForbidden,
	Unauthenticated:    http.StatusUnauthorized,
	ResourceExhausted:  http.StatusTooManyRequests,
	FailedPrecondition: http.StatusBadRequest,
	Aborted:            http.StatusConflict,
	OutOfRange:         http.StatusBadRequest,
	Unimplemented:      http.StatusNotImplemented,
	Internal:           http.StatusInternalServerError,
	Unavailable:        http.StatusServiceUnavailable,
	DataLoss:           http.StatusInternalServerError,
	InternalOnlyLog:    http.StatusInternalServerError,
}

// IsError checks if the error is an Error type.
func IsError(err error) bool {
	var e *Error
	return errors.As(err, &e)
}

// GetCode returns the error code from an error.
func GetCode(err error) ErrCode {
	var e *Error
	if errors.As(err, &e) {
		return e.Code
	}
	return Unknown
}
