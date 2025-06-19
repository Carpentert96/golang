package server

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

// ctxKey is a type for context keys in this package.
type ctxKey string

const (
	TraceIDKey ctxKey = "traceID"
	LoggerKey  ctxKey = "logger"
)

// traceMiddleware injects a unique trace ID into each request's context.
func traceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := uuid.New().String()
		ctx := context.WithValue(r.Context(), TraceIDKey, traceID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// LoggerFromContext retrieves a slog.Logger (if set) or returns the default.
func LoggerFromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(LoggerKey).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}
