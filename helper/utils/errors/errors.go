package errors

import (
	"fmt"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"net"
)

type Error struct {
	Status
	cause error
}

func (e *Error) Error() string {
	return fmt.Sprintf("error: code = %d reason = %s message = %s metadata = %v cause = %v", e.Code, e.Reason, e.Message, e.Metadata, e.cause)
}

func (e *Error) WithCause(cause error) *Error {
	err := Clone(e)
	err.cause = cause
	return err
}

func (e *Error) WithMetadata(md map[string]string) *Error {
	err := Clone(e)
	err.Metadata = md
	return err
}

func (e *Error) GetMeta(key string) string {
	if v, ok := e.Metadata[key]; ok {
		return v
	}
	return ""
}

func (e *Error) GRPCStatus() *status.Status {
	s, _ := status.New(ToGRPCCode(int(e.Code)), e.Reason).
		WithDetails(&errdetails.ErrorInfo{
			Reason:   e.Reason,
			Metadata: e.Metadata,
		})
	return s
}

func New(code int, reason, message string) *Error {
	return &Error{
		Status: Status{
			Code:    int32(code),
			Message: message,
			Reason:  reason,
		},
	}
}

// NewFromError creates an error with the given code, extracting reason and message from the error
// If err is already an *Error, it preserves the reason and message
// Otherwise, it uses err.Error() for both reason and message
func NewFromError(code int, err error) *Error {
	reason, message := ErrorReasonAndMessage(err)
	return New(code, reason, message)
}

func Newf(code int, reason, format string, a ...any) *Error {
	return New(code, reason, fmt.Sprintf(format, a...))
}

func Clone(err *Error) *Error {
	if err == nil {
		return nil
	}
	metadata := make(map[string]string, len(err.Metadata))
	for k, v := range err.Metadata {
		metadata[k] = v
	}
	return &Error{
		cause: err.cause,
		Status: Status{
			Code:     err.Code,
			Reason:   err.Reason,
			Message:  err.Message,
			Metadata: metadata,
		},
	}
}

func ToError(result any) error {
	var err error
	switch e := result.(type) {
	default:
		err = fmt.Errorf("%s", result)
	case error:
		err = e
	}
	return err
}

func From(err error) *Error {
	if err == nil {
		return InternalServer(fmt.Errorf(DefaultMsg))
	}
	if IsRecordNotFound(err) {
		return NotFound(err)
	}
	if e := IsError(err); e != nil {
		return e
	}
	return InternalServer(err)
}

func FromNotNil(err error) *Error {
	if err == nil {
		return nil
	}
	return From(err)
}

func IsError(err error) *Error {
	var r *Error
	if As(err, &r) {
		return r
	}
	return nil
}

func IsTimeout(err error) bool {
	var e net.Error
	if As(err, &e) && e.Timeout() {
		return true
	}
	return false
}

func IsRecordNotFound(err error) bool {
	return Is(err, gorm.ErrRecordNotFound)
}

// ErrorReasonAndMessage returns a normalized reason and message from an error.
// If err is a typed *Error, its Reason and Message are returned.
// Otherwise, err.Error() is used for both fields.
func ErrorReasonAndMessage(err error) (reason, message string) {
	if err == nil {
		return "", ""
	}
	if e := IsError(err); e != nil {
		return e.Reason, e.Message
	}
	msg := err.Error()
	return msg, msg
}
