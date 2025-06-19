package server

import (
	"html/template"
	"net/http"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/Carpentert96/ToDoList/pkg/model"
	"github.com/Carpentert96/ToDoList/pkg/storage"
)

var templateDir string

func init() {
	// Kept running into an issue where this runtime function couldn't find where the template was so, I told it where to look
	// but it kept trying to create /server/server/templates and panic. I believe was because templates wasn't in the same directory as server.go but I wanted to keep it refactored
	_, b, _, ok := runtime.Caller(0)
	if !ok {
		panic("could not determine caller info for handlers.go")
	}
	// Base path
	base := filepath.Dir(b)
	// Go find "templates" (before I had it as "server/templates" but that was causing issuesthe double server/server)
	templateDir = filepath.Join(base, "templates")
}

func loadTemplate(name string) *template.Template {
	full := filepath.Join(templateDir, name)
	return template.Must(template.ParseFiles(full))
}

// ListHandler renders the to-do list.
func ListHandler(w http.ResponseWriter, r *http.Request) {
	logger := LoggerFromContext(r.Context())
	todos, err := storage.LoadTodos()
	if err != nil {
		logger.Error("load todos", "err", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	tmpl := loadTemplate("list.html")
	tmpl.Execute(w, todos)
}

// CreateHandler shows a blank form on GET, and processes it on POST.
func CreateHandler(w http.ResponseWriter, r *http.Request) {
	logger := LoggerFromContext(r.Context())

	switch r.Method {
	case http.MethodGet:
		// Render the blank create form
		loadTemplate("create.html").Execute(w, nil)

	case http.MethodPost:
		// Parse form values
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad form", http.StatusBadRequest)
			return
		}
		desc := r.FormValue("description")
		started := r.FormValue("started") == "on"
		done := r.FormValue("done") == "on"

		// Atomically add the new todo via the actor
		if _, err := storage.AddTodo(desc, started, done); err != nil {
			logger.Error("add todo failed", "err", err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		// Redirect back to the list view
		http.Redirect(w, r, "/list", http.StatusSeeOther)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GetHandler shows a lookup form (no id) or displays a single to-do when ?id= is set.
func GetHandler(w http.ResponseWriter, r *http.Request) {
	logger := LoggerFromContext(r.Context())
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		loadTemplate("get.html").Execute(w, nil)
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	todos, err := storage.LoadTodos()
	if err != nil {
		logger.Error("load todos", "err", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	var found *model.Todo
	for _, t := range todos {
		if t.ID == id {
			found = &t
			break
		}
	}
	if found == nil {
		http.NotFound(w, r)
		return
	}
	loadTemplate("get.html").Execute(w, found)
}

// UpdateHandler shows the edit form on GET and applies changes on POST.
func UpdateHandler(w http.ResponseWriter, r *http.Request) {
	logger := LoggerFromContext(r.Context())
	switch r.Method {
	case http.MethodGet:
		id, err := strconv.Atoi(r.URL.Query().Get("id"))
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}
		todos, err := storage.LoadTodos()
		if err != nil {
			logger.Error("load todos", "err", err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		var current *model.Todo
		for _, t := range todos {
			if t.ID == id {
				current = &t
				break
			}
		}
		if current == nil {
			http.NotFound(w, r)
			return
		}
		loadTemplate("update.html").Execute(w, current)

	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad form", http.StatusBadRequest)
			return
		}
		id, err := strconv.Atoi(r.FormValue("id"))
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}
		desc := r.FormValue("description")
		started := r.FormValue("started") == "on"
		done := r.FormValue("done") == "on"

		todos, err := storage.LoadTodos()
		if err != nil {
			logger.Error("load todos", "err", err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		for i := range todos {
			if todos[i].ID == id {
				todos[i].Description = desc
				todos[i].Started = started
				todos[i].Done = done
				break
			}
		}
		if err := storage.SaveTodos(todos); err != nil {
			logger.Error("save todos", "err", err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/list", http.StatusSeeOther)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// DeleteHandler shows a confirmation on GET and removes on POST.
func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	logger := LoggerFromContext(r.Context())
	switch r.Method {
	case http.MethodGet:
		idStr := r.URL.Query().Get("id")
		if idStr == "" {
			http.Error(w, "Missing ID", http.StatusBadRequest)
			return
		}
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}
		todos, _ := storage.LoadTodos()
		var current *model.Todo
		for _, t := range todos {
			if t.ID == id {
				current = &t
				break
			}
		}
		if current == nil {
			http.NotFound(w, r)
			return
		}
		loadTemplate("delete.html").Execute(w, current)

	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad form", http.StatusBadRequest)
			return
		}
		id, err := strconv.Atoi(r.FormValue("id"))
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}
		todos, err := storage.LoadTodos()
		if err != nil {
			logger.Error("load todos", "err", err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		newList := make([]model.Todo, 0, len(todos))
		for _, t := range todos {
			if t.ID != id {
				newList = append(newList, t)
			}
		}
		if err := storage.SaveTodos(newList); err != nil {
			logger.Error("save todos", "err", err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/list", http.StatusSeeOther)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
