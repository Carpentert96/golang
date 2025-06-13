package server

//server handlers including list, create, get, update, delete with logging and template rendering
import (
	"context"
	"encoding/json"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/Carpentert96/ToDoList/pkg/storage"
)

func LoggerFromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(LoggerKey).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}
func listHandler(w http.ResponseWriter, r *http.Request) {
	logger := LoggerFromContext(r.Context())

	todos, err := storage.LoadTodos()
	if err != nil {
		logger.Error("failed to load todos", "err", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("server/templates/list.html")
	if err != nil {
		logger.Error("failed to load template", "err", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, todos); err != nil {
		logger.Error("failed to execute template", "err", err)
		http.Error(w, "Render error", http.StatusInternalServerError) //Actually got this a lot before I figured out the correct template
	}
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	traceID := r.Context().Value(TraceIDKey).(string)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "create called",
		"traceID": traceID,
	})
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	traceID := r.Context().Value(TraceIDKey).(string)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "get called",
		"traceID": traceID,
	})
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	traceID := r.Context().Value(TraceIDKey).(string)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "update called",
		"traceID": traceID,
	})
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	traceID := r.Context().Value(TraceIDKey).(string)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "delete called",
		"traceID": traceID,
	})
}
