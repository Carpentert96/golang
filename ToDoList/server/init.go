package server

import (
	"context"
	"log/slog"
	"os"

	"github.com/google/uuid"
)

const traceKey ctxKey = "traceID"

// Initilisation function to set up context and logger (parts declared in middleware.go too)
func InitApp() (context.Context, *slog.Logger, string) {
	tid := uuid.NewString()
	ctx := context.WithValue(context.Background(), traceKey, tid) //Calling the context with a new trace ID

	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{AddSource: true})
	logger := slog.New(handler)

	return ctx, logger, tid
}
