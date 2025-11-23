package accesslog

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/derletzte256/avito-assignment-2025-autumn/internal/pkg/logger"
	"github.com/gorilla/mux"
	"github.com/rs/xid"
	"go.uber.org/zap"
)

type contextKey string

const (
	requestIDContextKey contextKey = "X-Request-ID"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func Middleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l := logger.Get()

			requestID := r.Header.Get(string(requestIDContextKey))
			if requestID == "" {
				requestID = xid.New().String()
			}

			ctx := context.WithValue(
				r.Context(),
				requestIDContextKey,
				requestID,
			)

			r = r.WithContext(ctx)

			l = l.With(zap.String(string(requestIDContextKey), requestID))

			w.Header().Add(string(requestIDContextKey), requestID)

			lrw := newLoggingResponseWriter(w)

			r = r.WithContext(logger.WithCtx(ctx, l))

			defer func(start time.Time) {
				l = l.With(
					zap.String("method", r.Method),
					zap.String("url", r.RequestURI),
					zap.String("user_agent", r.UserAgent()),
					zap.Int("status_code", lrw.statusCode),
					zap.Duration("elapsed_ms", time.Since(start)))
				switch {
				case lrw.statusCode >= http.StatusInternalServerError:
					l.Error(fmt.Sprintf("%s request to %s failed", r.Method, r.RequestURI))
					return
				case lrw.statusCode >= http.StatusBadRequest:
					l.Warn(fmt.Sprintf("%s request to %s completed with client error", r.Method, r.RequestURI))
					return
				default:
					l.Info(fmt.Sprintf("%s request to %s completed successfully", r.Method, r.RequestURI))
					return
				}
			}(time.Now())

			next.ServeHTTP(lrw, r)
		})
	}
}
