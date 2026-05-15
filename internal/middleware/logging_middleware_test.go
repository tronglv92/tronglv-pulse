package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoggingMiddleware_CapturesStatusCode(t *testing.T) {
	mw := LoggingMiddleware()

	handler := mw(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"ok":true}`))
	})

	req := httptest.NewRequest(http.MethodGet, "/test?foo=bar", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, `{"ok":true}`, rec.Body.String())
}

func TestLoggingMiddleware_DefaultStatusIsOK(t *testing.T) {
	mw := LoggingMiddleware()

	handler := mw(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("hello"))
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestLoggingMiddleware_4xxDoesNotPanic(t *testing.T) {
	mw := LoggingMiddleware()

	handler := mw(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	rec := httptest.NewRecorder()

	assert.NotPanics(t, func() {
		handler(rec, req)
	})
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestLoggingMiddleware_5xxDoesNotPanic(t *testing.T) {
	mw := LoggingMiddleware()

	handler := mw(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	req := httptest.NewRequest(http.MethodPost, "/error", nil)
	rec := httptest.NewRecorder()

	assert.NotPanics(t, func() {
		handler(rec, req)
	})
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestResponseWriter_TracksBytes(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := newResponseWriter(rec)

	n, err := rw.Write([]byte("hello"))
	assert.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, 5, rw.bytesWritten)

	n, err = rw.Write([]byte(" world"))
	assert.NoError(t, err)
	assert.Equal(t, 6, n)
	assert.Equal(t, 11, rw.bytesWritten)
}
