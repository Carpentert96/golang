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
	_, b, _, ok := runtime.Caller(0)
	if !ok {
		panic("could not determine caller info for handlers.go")
	}
	base := filepath.Dir(b)
	templateDir = filepath.Join(base, "templates")
}

func loadTemplate(name string) *template.Template {
	full := filepath.Join(templateDir, name)
	return template.Must(template.ParseFiles(full))
}

// ListHandler renders the to-do list.
func ListHandler(store *storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl := loadTemplate("list.html")
		todos, err := store.LoadTodos()
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, todos)
	}
}

// CreateHandler shows blank form on GET, processes a new todo via actor on POST.
func CreateHandler(store *storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			loadTemplate("create.html").Execute(w, nil)

		case http.MethodPost:
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Bad form", http.StatusBadRequest)
				return
			}
			desc := r.FormValue("description")
			started := r.FormValue("started") == "on"
			done := r.FormValue("done") == "on"

			newTodo := model.Todo{
				Description: desc,
				Started:     started,
				Done:        done,
			}

			if err := store.AddTodo(newTodo); err != nil {
				http.Error(w, "Internal error", http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/list", http.StatusSeeOther)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// GetHandler shows lookup form or displays a single todo.
func GetHandler(store *storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		todos, err := store.LoadTodos()
		if err != nil {
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
}

// UpdateHandler renders edit form on GET, enqueues update on POST.
func UpdateHandler(store *storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			id, err := strconv.Atoi(r.URL.Query().Get("id"))
			if err != nil {
				http.Error(w, "Invalid ID", http.StatusBadRequest)
				return
			}
			todos, err := store.LoadTodos()
			if err != nil {
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
			id, _ := strconv.Atoi(r.FormValue("id"))
			desc := r.FormValue("description")
			started := r.FormValue("started") == "on"
			done := r.FormValue("done") == "on"
			if err := store.UpdateTodo(id, desc, started, done); err != nil {
				http.Error(w, "Internal error", http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/list", http.StatusSeeOther)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// DeleteHandler confirms on GET, enqueues delete on POST.
func DeleteHandler(store *storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			id, err := strconv.Atoi(r.URL.Query().Get("id"))
			if err != nil {
				http.Error(w, "Invalid ID", http.StatusBadRequest)
				return
			}
			todos, err := store.LoadTodos()
			if err != nil {
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
			loadTemplate("delete.html").Execute(w, current)

		case http.MethodPost:
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Bad form", http.StatusBadRequest)
				return
			}
			id, _ := strconv.Atoi(r.FormValue("id"))
			if err := store.DeleteTodo(id); err != nil {
				http.Error(w, "Internal error", http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/list", http.StatusSeeOther)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
