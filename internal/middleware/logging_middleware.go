package middleware

import (
	"net/http"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"

	"pulse/helper/utils/toolkit/contextx"

	"github.com/zeromicro/go-zero/core/utils"
)

// responseWriter wraps http.ResponseWriter to capture the status code and bytes written.
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += n
	return n, err
}

// LoggingMiddleware returns a rest.Middleware that logs structured request/response
// information. Log level is determined by response status code:
// Info for 2xx/3xx, Warn for 4xx, Error for 5xx.
func LoggingMiddleware() rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Use elapsed timer from TraceMiddleware if available, otherwise create our own.
			timer, ok := contextx.Value[*utils.ElapsedTimer](ctx, contextx.KeyElapsedTimer)
			if !ok || timer == nil {
				timer = utils.NewElapsedTimer()
			}

			rw := newResponseWriter(w)
			next(rw, r)

			duration := timer.Duration()
			fields := []logx.LogField{
				logx.Field("method", r.Method),
				logx.Field("path", r.URL.Path),
				logx.Field("query", r.URL.RawQuery),
				logx.Field("remote_addr", r.RemoteAddr),
				logx.Field("user_agent", r.UserAgent()),
				logx.Field("content_length", r.ContentLength),
				logx.Field("status_code", rw.statusCode),
				logx.Field("response_size", rw.bytesWritten),
				logx.Field("duration_ms", duration.Truncate(time.Millisecond).Milliseconds()),
			}

			logger := logx.WithContext(ctx)
			switch {
			case rw.statusCode >= http.StatusInternalServerError:
				logger.Errorw("HTTP request", fields...)
			case rw.statusCode >= http.StatusBadRequest:
				logger.Sloww("HTTP request", fields...)
			default:
				logger.Infow("HTTP request", fields...)
			}
		}
	}
}
