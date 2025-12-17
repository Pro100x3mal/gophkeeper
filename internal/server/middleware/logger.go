package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type wrappedResponseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (w *wrappedResponseWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	size, err := w.ResponseWriter.Write(b)
	w.size += size
	return size, err
}

func (w *wrappedResponseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func Logger(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			httpLog := logger.With(zap.String("middleware", "logger"))
			start := time.Now()

			wrw := &wrappedResponseWriter{
				ResponseWriter: w,
			}

			next.ServeHTTP(wrw, r)

			httpLog.Info("http request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("remote_addr", r.RemoteAddr),
				zap.Int("status", wrw.status),
				zap.Int("size", wrw.size),
				zap.Duration("duration", time.Since(start)),
			)
		})
	}
}
