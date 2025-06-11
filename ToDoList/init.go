// init.go
package main

import (
	"context"
	"os"

	"github.com/google/uuid"
	"golang.org/x/exp/slog"
)

type ctxKey string

const traceKey ctxKey = "traceID"

// initApp creates your root Context (with a TraceID) and your structured Logger.
func initApp() (context.Context, *slog.Logger, string) {
	tid := uuid.NewString()
	ctx := context.WithValue(context.Background(), traceKey, tid)

	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{AddSource: true})
	logger := slog.New(handler)

	return ctx, logger, tid
}
