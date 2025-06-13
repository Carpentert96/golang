package server

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// Only need to declare this unique key once, so we can use it across the package
type ctxKey string

const (
	TraceIDKey ctxKey = "traceID"
	LoggerKey  ctxKey = "logger"
)

func traceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := uuid.New().String()
		ctx := context.WithValue(r.Context(), TraceIDKey, traceID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
