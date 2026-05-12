package errors

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorWrapping(t *testing.T) {
	tests := []struct {
		name            string
		createError     func() *Error
		expectedCode    int32
		expectedReason  string
		expectedMessage string
	}{
		{
			name: "wrap NotFound with BadRequest",
			createError: func() *Error {
				notFoundErr := NotFound(fmt.Errorf("user not found"))
				return BadRequest(notFoundErr)
			},
			expectedCode:    400,
			expectedReason:  "user not found",
			expectedMessage: "user not found",
		},
		{
			name: "wrap Unauthorized with BadRequest",
			createError: func() *Error {
				unauthorizedErr := Unauthorized(fmt.Errorf("invalid token"))
				return BadRequest(unauthorizedErr)
			},
			expectedCode:    400,
			expectedReason:  "invalid token",
			expectedMessage: "invalid token",
		},
		{
			name: "wrap custom error with NotFound",
			createError: func() *Error {
				customErr := NewBadRequest("CUSTOM_REASON", "Custom message")
				return NotFound(customErr)
			},
			expectedCode:    404,
			expectedReason:  "CUSTOM_REASON",
			expectedMessage: "Custom message",
		},
		{
			name: "wrap standard error with BadRequest",
			createError: func() *Error {
				return BadRequest(fmt.Errorf("standard error"))
			},
			expectedCode:    400,
			expectedReason:  "standard error",
			expectedMessage: "standard error",
		},
		{
			name: "multiple wrapping",
			createError: func() *Error {
				notFoundErr := NotFound(fmt.Errorf("record not found"))
				forbiddenErr := Forbidden(notFoundErr)
				return BadRequest(forbiddenErr)
			},
			expectedCode:    400,
			expectedReason:  "record not found",
			expectedMessage: "record not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.createError()

			assert.Equal(t, tt.expectedCode, err.Code, "Code mismatch")
			assert.Equal(t, tt.expectedReason, err.Reason, "Reason mismatch")
			assert.Equal(t, tt.expectedMessage, err.Message, "Message mismatch")

			// Ensure Reason and Message don't contain the verbose error format
			assert.NotContains(t, err.Reason, "error: code =", "Reason should not contain verbose error format")
			assert.NotContains(t, err.Message, "error: code =", "Message should not contain verbose error format")
		})
	}
}

func TestErrorReasonAndMessage(t *testing.T) {
	tests := []struct {
		name            string
		input           error
		expectedReason  string
		expectedMessage string
	}{
		{
			name:            "nil error",
			input:           nil,
			expectedReason:  "",
			expectedMessage: "",
		},
		{
			name:            "extract from *Error",
			input:           NewNotFound("NOT_FOUND", "Resource not found"),
			expectedReason:  "NOT_FOUND",
			expectedMessage: "Resource not found",
		},
		{
			name:            "extract from standard error",
			input:           fmt.Errorf("simple error"),
			expectedReason:  "simple error",
			expectedMessage: "simple error",
		},
		{
			name:            "extract from nested *Error",
			input:           BadRequest(NotFound(fmt.Errorf("nested"))),
			expectedReason:  "nested",
			expectedMessage: "nested",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reason, message := ErrorReasonAndMessage(tt.input)
			assert.Equal(t, tt.expectedReason, reason)
			assert.Equal(t, tt.expectedMessage, message)
		})
	}
}

func TestErrorMessage_NoVerboseFormat(t *testing.T) {
	// Create nested error
	baseErr := fmt.Errorf("database connection failed")
	notFoundErr := NotFound(baseErr)
	badRequestErr := BadRequest(notFoundErr)

	// Get error message
	errMsg := badRequestErr.Message

	// Should only contain the original message, not the verbose format
	assert.Equal(t, "database connection failed", errMsg)
	assert.NotContains(t, errMsg, "code =")
	assert.NotContains(t, errMsg, "reason =")
	assert.NotContains(t, errMsg, "metadata =")
	assert.NotContains(t, errMsg, "cause =")
}

func TestAllErrorTypes(t *testing.T) {
	baseErr := fmt.Errorf("test error")

	tests := []struct {
		name         string
		createFunc   func(error) *Error
		expectedCode int32
	}{
		{"BadRequest", BadRequest, 400},
		{"Unauthorized", Unauthorized, 401},
		{"Forbidden", Forbidden, 403},
		{"NotFound", NotFound, 404},
		{"Conflict", Conflict, 409},
		{"RequestTimeout", RequestTimeout, 408},
		{"InternalServer", InternalServer, 500},
		{"ServiceUnavailable", ServiceUnavailable, 502},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.createFunc(baseErr)
			assert.Equal(t, tt.expectedCode, err.Code)
			assert.Equal(t, "test error", err.Reason)
			assert.Equal(t, "test error", err.Message)
		})
	}
}

func TestNewFromError(t *testing.T) {
	tests := []struct {
		name            string
		code            int
		input           error
		expectedCode    int32
		expectedReason  string
		expectedMessage string
	}{
		{
			name:            "from standard error",
			code:            400,
			input:           fmt.Errorf("validation failed"),
			expectedCode:    400,
			expectedReason:  "validation failed",
			expectedMessage: "validation failed",
		},
		{
			name:            "from *Error",
			code:            500,
			input:           NewNotFound("USER_NOT_FOUND", "User does not exist"),
			expectedCode:    500,
			expectedReason:  "USER_NOT_FOUND",
			expectedMessage: "User does not exist",
		},
		{
			name:            "from wrapped *Error",
			code:            502,
			input:           BadRequest(NotFound(fmt.Errorf("record missing"))),
			expectedCode:    502,
			expectedReason:  "record missing",
			expectedMessage: "record missing",
		},
		{
			name:            "custom status code",
			code:            418, // I'm a teapot
			input:           fmt.Errorf("teapot error"),
			expectedCode:    418,
			expectedReason:  "teapot error",
			expectedMessage: "teapot error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewFromError(tt.code, tt.input)

			assert.Equal(t, tt.expectedCode, err.Code)
			assert.Equal(t, tt.expectedReason, err.Reason)
			assert.Equal(t, tt.expectedMessage, err.Message)

			// Ensure no verbose format in reason/message
			assert.NotContains(t, err.Reason, "error: code =")
			assert.NotContains(t, err.Message, "error: code =")
		})
	}
}
