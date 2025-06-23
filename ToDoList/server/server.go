package server

import (
	"context"
	"net/http"
	"time"

	"github.com/Carpentert96/ToDoList/pkg/storage"
)

// Run starts your HTTP server on :8080, wiring in traceMiddleware and your handlers.
// It returns as soon as the server shuts down (either by context cancellation or error).
func Run(ctx context.Context, store *storage.Store) error {
	// 1) Build your mux and attach routes
	mux := http.NewServeMux()
	mux.Handle("/create", traceMiddleware(http.HandlerFunc(CreateHandler(store))))
	mux.Handle("/get", traceMiddleware(http.HandlerFunc(GetHandler(store))))
	mux.Handle("/update", traceMiddleware(http.HandlerFunc(UpdateHandler(store))))
	mux.Handle("/delete", traceMiddleware(http.HandlerFunc(DeleteHandler(store))))
	mux.Handle("/list", traceMiddleware(http.HandlerFunc(ListHandler(store))))
	mux.Handle("/about/", http.StripPrefix("/about/", http.FileServer(http.Dir("static"))))

	// 2) Create the server object so we can shut it down later
	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// 3) Grab the *slog.Logger you put into ctx via InitApp()
	logger := LoggerFromContext(ctx)
	logger.InfoContext(ctx, "server.Run: listening on :8080")

	// 4) Run ListenAndServe in the background
	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	// 5) Wait for either ctx cancellation or ListenAndServe error
	select {
	case <-ctx.Done():
		// allow up to 5s for in-flight requests to finish
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		logger.InfoContext(ctx, "server.Run: shutting down HTTP server")
		return srv.Shutdown(shutCtx)

	case err := <-errCh:
		return err
	}
}
