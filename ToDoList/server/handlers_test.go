// server/handlers_test.go
package server_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Carpentert96/ToDoList/pkg/model"
	"github.com/Carpentert96/ToDoList/pkg/storage"
	"github.com/Carpentert96/ToDoList/server"
)

// seedJSON writes jsonData into filePath and logs it.
func seedJSON(t *testing.T, filePath, jsonData string) {
	if err := os.WriteFile(filePath, []byte(jsonData), 0644); err != nil {
		t.Fatalf("failed to write fixture %s: %v", filePath, err)
	}
	t.Logf("seeded %s with: %s", filePath, jsonData)
}

// This test starts with an empty file (unlike the ones below) and creates a JSON file by calling the CreateHandler
func TestCreateAndListHandlers(t *testing.T) {
	t.Parallel()

	// Prepare isolated JSON file and store
	tmp := t.TempDir()
	dataFile := filepath.Join(tmp, "todos.json")
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{})) //using os.Stout to display all writes and saves for evidence
	store := storage.NewStore(context.Background(), logger, dataFile)

	// GET /list on empty store
	t.Logf("→ GET /list on empty store")
	req := httptest.NewRequest(http.MethodGet, "/list", nil)
	w := httptest.NewRecorder()
	server.ListHandler(store)(w, req)
	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	t.Logf("Status=%d, Body=%q", resp.StatusCode, body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK; got %d", resp.StatusCode)
	}
	if !strings.Contains(string(body), "No tasks found") {
		t.Errorf("expected empty message; got %q", body)
	}

	// POST /create
	t.Logf("→ POST /create")
	form := url.Values{"description": {"Test task"}}
	req = httptest.NewRequest(http.MethodPost, "/create", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()
	server.CreateHandler(store)(w, req)
	resp = w.Result()
	t.Logf("Status=%d, Location=%q", resp.StatusCode, resp.Header.Get("Location"))
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected 303; got %d", resp.StatusCode)
	}
	if resp.Header.Get("Location") != "/list" {
		t.Errorf("expected redirect to /list; got %q", resp.Header.Get("Location"))
	}

	// Ensure storage has one
	todos, err := store.LoadTodos()
	if err != nil {
		t.Fatalf("LoadTodos error: %v", err)
	}
	if len(todos) != 1 || todos[0].Description != "Test task" {
		t.Fatalf("unexpected todos: %+v", todos)
	}

	// GET /list after create
	t.Logf("→ GET /list after create")
	req = httptest.NewRequest(http.MethodGet, "/list", nil)
	w = httptest.NewRecorder()
	server.ListHandler(store)(w, req)
	resp = w.Result()
	body, _ = io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Test task") {
		t.Errorf("list did not show created task; got %q", body)
	}
}

func TestGetHandler(t *testing.T) {
	t.Parallel()
	// Seed 2 todos before actor
	tmp := t.TempDir()
	dataFile := filepath.Join(tmp, "todos.json")
	seedJSON(t, dataFile, `[
  {"ID":1,"Description":"First","Started":false,"Done":true},
  {"ID":2,"Description":"Second","Started":true,"Done":false}
]`)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	store := storage.NewStore(context.Background(), logger, dataFile)

	// GET form
	req := httptest.NewRequest(http.MethodGet, "/get", nil)
	w := httptest.NewRecorder()
	server.GetHandler(store)(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /get status %d; want %d", w.Code, http.StatusOK)
	}
	body, _ := io.ReadAll(w.Body)
	if !strings.Contains(string(body), `name="id"`) {
		t.Errorf("expected lookup form; got %q", body)
	}

	// GET id=2
	req = httptest.NewRequest(http.MethodGet, "/get?id=2", nil)
	w = httptest.NewRecorder()
	server.GetHandler(store)(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /get?id=2 status %d; want %d", w.Code, http.StatusOK)
	}
	body, _ = io.ReadAll(w.Body)
	if !strings.Contains(string(body), "Second") {
		t.Errorf("expected 'Second'; got %q", body)
	}
}

func TestUpdateHandler(t *testing.T) {
	t.Parallel()
	// Seed 1 todo
	tmp := t.TempDir()
	dataFile := filepath.Join(tmp, "todos.json")
	seedJSON(t, dataFile, `[{"ID":1,"Description":"Old","Started":false,"Done":false}]`)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	store := storage.NewStore(context.Background(), logger, dataFile)

	// GET pre-filled
	req := httptest.NewRequest(http.MethodGet, "/update?id=1", nil)
	w := httptest.NewRecorder()
	server.UpdateHandler(store)(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /update?id=1 status %d; want %d", w.Code, http.StatusOK)
	}
	body, _ := io.ReadAll(w.Body)
	if !strings.Contains(string(body), `value="Old"`) {
		t.Errorf("expected Old value; got %q", body)
	}

	// POST update
	form := url.Values{"id": {"1"}, "description": {"New"}, "started": {"on"}, "done": {"on"}}
	req = httptest.NewRequest(http.MethodPost, "/update", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()
	server.UpdateHandler(store)(w, req)
	if w.Code != http.StatusSeeOther {
		t.Fatalf("POST /update status %d; want %d", w.Code, http.StatusSeeOther)
	}

	// Verify file
	final, _ := os.ReadFile(dataFile)
	var todos []model.Todo
	if err := json.Unmarshal(final, &todos); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if len(todos) != 1 || todos[0].Description != "New" || !todos[0].Started || !todos[0].Done {
		t.Errorf("expected updated todo; got %+v", todos[0])
	}
}

func TestDeleteHandler(t *testing.T) {
	t.Parallel()
	// Seed 2 todos
	tmp := t.TempDir()
	dataFile := filepath.Join(tmp, "todos.json")
	seedJSON(t, dataFile, `[
  {"ID":1,"Description":"Keep","Started":false,"Done":false},
  {"ID":2,"Description":"Delete","Started":false,"Done":false}
]`)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	store := storage.NewStore(context.Background(), logger, dataFile)

	// GET confirm
	req := httptest.NewRequest(http.MethodGet, "/delete?id=2", nil)
	w := httptest.NewRecorder()
	server.DeleteHandler(store)(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /delete?id=2 status %d; want %d", w.Code, http.StatusOK)
	}

	// POST delete
	form := url.Values{"id": {"2"}}
	req = httptest.NewRequest(http.MethodPost, "/delete", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()
	server.DeleteHandler(store)(w, req)
	if w.Code != http.StatusSeeOther {
		t.Fatalf("POST /delete status %d; want %d", w.Code, http.StatusSeeOther)
	}

	// Verify
	final, _ := os.ReadFile(dataFile)
	var remain []model.Todo
	if err := json.Unmarshal(final, &remain); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if len(remain) != 1 || remain[0].ID != 1 {
		t.Errorf("expected only ID=1, got %v", remain)
	}
}

func TestConcurrentCreates(t *testing.T) {
	t.Parallel()
	// fresh actor
	tmp := t.TempDir()
	dataFile := filepath.Join(tmp, "todos.json")
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	store := storage.NewStore(context.Background(), logger, dataFile)

	const N = 100
	errCh := make(chan error, N)
	for i := 0; i < N; i++ {
		go func(i int) {
			req := httptest.NewRequest(http.MethodPost, "/create",
				strings.NewReader(url.Values{"description": {fmt.Sprintf("task-%02d", i)}}.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()
			server.CreateHandler(store)(w, req)
			if w.Result().StatusCode != http.StatusSeeOther {
				errCh <- fmt.Errorf("%d: %d", i, w.Result().StatusCode)
				return
			}
			errCh <- nil
		}(i)
	}
	for i := 0; i < N; i++ {
		if err := <-errCh; err != nil {
			t.Error(err)
		}
	}

	todos, _ := store.LoadTodos()
	if len(todos) != N {
		t.Fatalf("expected %d todos, got %d", N, len(todos))
	}
}

func TestConcurrentGetHandler(t *testing.T) {
	t.Parallel()
	// seed then actor
	tmp := t.TempDir()
	dataFile := filepath.Join(tmp, "todos.json")
	const N = 10
	var sb strings.Builder
	sb.WriteString("[")
	for i := 1; i <= N; i++ {
		sb.WriteString(fmt.Sprintf(`{"ID":%d,"Description":"task-%02d","Started":false,"Done":false}`, i, i))
		if i < N {
			sb.WriteString(",")
		}
	}
	sb.WriteString("]")
	seedJSON(t, dataFile, sb.String())
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	store := storage.NewStore(context.Background(), logger, dataFile)

	const M = 50
	errCh := make(chan error, M)
	for i := 0; i < M; i++ {
		go func(i int) {
			id := (i % N) + 1
			req := httptest.NewRequest(http.MethodGet,
				fmt.Sprintf("/get?id=%d", id), nil)
			w := httptest.NewRecorder()
			server.GetHandler(store)(w, req)
			if res := w.Result(); res.StatusCode != http.StatusOK {
				errCh <- fmt.Errorf("%d: %d", id, res.StatusCode)
				return
			}
			body, _ := io.ReadAll(w.Body)
			want := fmt.Sprintf("task-%02d", id)
			if !strings.Contains(string(body), want) {
				errCh <- fmt.Errorf("%d missing %s", id, want)
				return
			}
			errCh <- nil
		}(i)
	}
	for i := 0; i < M; i++ {
		if err := <-errCh; err != nil {
			t.Error(err)
		}
	}
}

func TestConcurrentUpdateHandler(t *testing.T) {
	t.Parallel()

	// seed then actor
	tmp := t.TempDir()
	dataFile := filepath.Join(tmp, "todos.json")
	const N = 10
	var sb2 strings.Builder
	sb2.WriteString("[")
	for i := 1; i <= N; i++ {
		sb2.WriteString(fmt.Sprintf(`{"ID":%d,"Description":"Orig-%02d","Started":false,"Done":false}`, i, i))
		if i < N {
			sb2.WriteString(",")
		}
	}
	sb2.WriteString("]")
	seedJSON(t, dataFile, sb2.String())
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	store := storage.NewStore(context.Background(), logger, dataFile)

	errCh := make(chan error, N)
	for id := 1; id <= N; id++ {
		go func(id int) {
			form := url.Values{
				"id":          {fmt.Sprint(id)},
				"description": {fmt.Sprintf("Upd-%02d", id)},
				"started":     {"on"},
				"done":        {"on"},
			}
			req := httptest.NewRequest(http.MethodPost, "/update", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()
			server.UpdateHandler(store)(w, req)
			if w.Result().StatusCode != http.StatusSeeOther {
				errCh <- fmt.Errorf("%d: %d", id, w.Result().StatusCode)
				return
			}
			errCh <- nil
		}(id)
	}
	for i := 0; i < N; i++ {
		if err := <-errCh; err != nil {
			t.Error(err)
		}
	}

	final, _ := os.ReadFile(dataFile)
	var got []model.Todo
	json.Unmarshal(final, &got)
	if len(got) != N {
		t.Fatalf("expected %d todos, got %d", N, len(got))
	}
	for _, td := range got {
		want := fmt.Sprintf("Upd-%02d", td.ID)
		if td.Description != want || !td.Started || !td.Done {
			t.Errorf("todo %+v not updated correctly", td)
		}
	}
}

func TestConcurrentDeleteHandler(t *testing.T) {
	t.Parallel()

	// seed then actor
	tmp := t.TempDir()
	dataFile := filepath.Join(tmp, "todos.json")
	const Ndel = 10
	var sb3 strings.Builder
	sb3.WriteString("[")
	for i := 1; i <= Ndel; i++ {
		sb3.WriteString(fmt.Sprintf(`{"ID":%d,"Description":"T-%02d","Started":false,"Done":false}`, i, i))
		if i < Ndel {
			sb3.WriteString(",")
		}
	}
	sb3.WriteString("]")
	seedJSON(t, dataFile, sb3.String())
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	store := storage.NewStore(context.Background(), logger, dataFile)

	errCh := make(chan error, Ndel)
	for id := 1; id <= Ndel; id++ {
		go func(id int) {
			form := url.Values{"id": {fmt.Sprint(id)}}
			req := httptest.NewRequest(http.MethodPost, "/delete", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()
			server.DeleteHandler(store)(w, req)
			if w.Result().StatusCode != http.StatusSeeOther {
				errCh <- fmt.Errorf("%d: %d", id, w.Result().StatusCode)
				return
			}
			errCh <- nil
		}(id)
	}
	for i := 0; i < Ndel; i++ {
		if err := <-errCh; err != nil {
			t.Error(err)
		}
	}

	final, _ := os.ReadFile(dataFile)
	var gotDel []model.Todo
	json.Unmarshal(final, &gotDel)
	if len(gotDel) != 0 {
		t.Errorf("expected 0 todos, got %d", len(gotDel))
	}
}
