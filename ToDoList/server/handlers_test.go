// server/handlers_test.go
package server_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/Carpentert96/ToDoList/pkg/model"
	"github.com/Carpentert96/ToDoList/pkg/storage"
	"github.com/Carpentert96/ToDoList/server"
)

// seedJSON writes the given JSON into todos.json, resets the actor, and logs it.
func seedJSON(t *testing.T, jsonData string) {
	if err := os.WriteFile("todos.json", []byte(jsonData), 0644); err != nil {
		t.Fatalf("failed to write fixture todos.json: %v", err)
	}
	t.Logf("seeded todos.json with:\n%s", jsonData)
	storage.Reset()
}

func TestCreateAndListHandlers(t *testing.T) {
	//This is why I added the template directory to the server package
	tmp := t.TempDir()
	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	storage.Reset() //Also had to add the reset here because the storage actor is not reset between tests
	// Check if the list is empty before starting
	t.Logf("→ GET /list on empty store")
	req := httptest.NewRequest(http.MethodGet, "/list", nil)
	w := httptest.NewRecorder()
	server.ListHandler(w, req)

	resp := w.Result()
	bodyBytes, _ := io.ReadAll(resp.Body)
	body := string(bodyBytes)
	t.Logf("   Status: %d, Body: %q", resp.StatusCode, body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 OK; got %d", resp.StatusCode)
	}
	if !strings.Contains(body, "No tasks found") {
		t.Errorf("expected empty-list message, got %q", body)
	}

	// Calls the create API and adds task
	t.Logf("→ POST /create (description=Test task, started=on)")
	form := url.Values{}
	form.Set("description", "Test task")
	form.Set("started", "on")
	// done left unchecked
	req = httptest.NewRequest(http.MethodPost, "/create", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()
	server.CreateHandler(w, req)

	resp = w.Result()
	t.Logf("   Status: %d, Location: %q", resp.StatusCode, resp.Header.Get("Location"))
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected redirect (303 See Other); got %d", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/list" {
		t.Errorf("expected Location '/list'; got %q", loc)
	}

	// Check it's been saved to storage
	all, err := storage.LoadTodos()
	if err != nil {
		t.Fatalf("LoadTodos error: %v", err)
	}
	t.Logf("   Storage now has %d items", len(all))
	if len(all) != 1 || all[0].Description != "Test task" {
		t.Fatalf("unexpected todos: %+v", all)
	}

	// Call the list API to verify the task was created
	t.Logf("→ GET /list after creation")
	req = httptest.NewRequest(http.MethodGet, "/list", nil)
	w = httptest.NewRecorder()
	server.ListHandler(w, req)

	resp = w.Result()
	bodyBytes, _ = io.ReadAll(resp.Body)
	body = string(bodyBytes)
	t.Logf("   Status: %d, Body: %q", resp.StatusCode, body)

	if !strings.Contains(body, "Test task") {
		t.Errorf("expected created task in list, got %q", body)
	}
}

func TestGetHandler(t *testing.T) {
	// Setup fresh dir and reset actor
	tmp := t.TempDir()
	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	// 1) GET /get with no id → form
	req := httptest.NewRequest(http.MethodGet, "/get", nil)
	w := httptest.NewRecorder()
	server.GetHandler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /get no id status = %d; want %d", w.Code, http.StatusOK)
	}
	body, _ := io.ReadAll(w.Body)
	if !strings.Contains(string(body), `name="id"`) {
		t.Errorf("expected lookup form, got %q", body)
	}

	// 2) Seed JSON and lookup id=2
	fixture := `[
  {"ID":1,"Description":"First","Started":false,"Done":true},
  {"ID":2,"Description":"Second","Started":true,"Done":false}
]`
	seedJSON(t, fixture)

	req = httptest.NewRequest(http.MethodGet, "/get?id=2", nil)
	w = httptest.NewRecorder()
	server.GetHandler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /get?id=2 status = %d; want %d", w.Code, http.StatusOK)
	}
	body, _ = io.ReadAll(w.Body)
	t.Logf("GET /get?id=2 → %q", body)
	if !strings.Contains(string(body), "Second") {
		t.Errorf("expected to see ‘Second’, got %q", body)
	}
}

func TestUpdateHandler(t *testing.T) { //add ID 2 to prove it's not updated both
	// Setup fresh dir and reset actor
	tmp := t.TempDir()
	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	// Seed one todo with ID=1
	fixture := `[
  {"ID":1,"Description":"Old","Started":false,"Done":false}
]`
	seedJSON(t, fixture)

	// 1) GET /update?id=1 → pre-filled form
	req := httptest.NewRequest(http.MethodGet, "/update?id=1", nil)
	w := httptest.NewRecorder()
	server.UpdateHandler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /update?id=1 status = %d; want %d", w.Code, http.StatusOK)
	}
	body, _ := io.ReadAll(w.Body)
	if !strings.Contains(string(body), `value="Old"`) {
		t.Errorf("expected form with old value, got %q", body)
	}

	// 2) POST /update → change to “New”
	form := url.Values{}
	form.Set("id", "1")
	form.Set("description", "New")
	form.Set("started", "on")
	form.Set("done", "on")
	req = httptest.NewRequest(http.MethodPost, "/update", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()
	server.UpdateHandler(w, req)
	if w.Code != http.StatusSeeOther {
		t.Fatalf("POST /update status = %d; want %d", w.Code, http.StatusSeeOther)
	}
	if loc := w.Header().Get("Location"); loc != "/list" {
		t.Errorf("POST /update Location = %q; want %q", loc, "/list")
	}

	// Verify on-disk JSON was updated
	updatedBytes, _ := os.ReadFile("todos.json")
	t.Logf("after update, todos.json:\n%s", updatedBytes)
	if !strings.Contains(string(updatedBytes), `"Description": "New"`) ||
		!strings.Contains(string(updatedBytes), `"Started": true`) ||
		!strings.Contains(string(updatedBytes), `"Done": true`) {
		t.Errorf("expected updated fields in JSON, got:\n%s", updatedBytes)
	}
}

func TestDeleteHandler(t *testing.T) {
	// Setup fresh dir and reset actor
	tmp := t.TempDir()
	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	// Seed two todos
	fixture := `[
  {"ID":1,"Description":"Keep","Started":false,"Done":false},
  {"ID":2,"Description":"Delete","Started":false,"Done":false}
]`
	seedJSON(t, fixture)

	// 1) GET /delete?id=2 → confirmation
	req := httptest.NewRequest(http.MethodGet, "/delete?id=2", nil)
	w := httptest.NewRecorder()
	server.DeleteHandler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /delete?id=2 status = %d; want %d", w.Code, http.StatusOK)
	}
	body, _ := io.ReadAll(w.Body)
	if !strings.Contains(string(body), "Delete") {
		t.Errorf("expected confirmation for ‘Delete’, got %q", body)
	}

	// 2) POST /delete → remove ID=2
	form := url.Values{}
	form.Set("id", "2")
	req = httptest.NewRequest(http.MethodPost, "/delete", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()
	server.DeleteHandler(w, req)
	if w.Code != http.StatusSeeOther {
		t.Fatalf("POST /delete status = %d; want %d", w.Code, http.StatusSeeOther)
	}
	if loc := w.Header().Get("Location"); loc != "/list" {
		t.Errorf("POST /delete Location = %q; want %q", loc, "/list")
	}

	// Verify JSON now only has the first entry
	finalBytes, _ := os.ReadFile("todos.json")
	t.Logf("after delete, todos.json:\n%s", finalBytes)
	if strings.Contains(string(finalBytes), `"ID": 2`) {
		t.Errorf("expected ID 2 to be removed, but JSON still contains:\n%s", finalBytes)
	}
}

func TestConcurrentCreates(t *testing.T) { //add some logging to this

	tmp := t.TempDir()
	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	storage.Reset() //As mentioned in storage.go, reset the storage actor

	const N = 100
	errCh := make(chan error, N)

	// Evidence of goroutines running concurrently
	for i := 0; i < N; i++ {
		go func(i int) {
			form := url.Values{
				"description": {fmt.Sprintf("task-%02d", i)},
			}
			req := httptest.NewRequest(http.MethodPost, "/create", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()
			server.CreateHandler(w, req)

			if w.Result().StatusCode != http.StatusSeeOther {
				errCh <- fmt.Errorf("goroutine %d: bad status %d", i, w.Result().StatusCode)
				return
			}
			errCh <- nil
		}(i)
	}

	// wait for all goroutines to finish (log this)
	for i := 0; i < N; i++ {
		if err := <-errCh; err != nil {
			t.Error(err)
		}
	}

	// finally, verify we got exactly N todos
	todos, err := storage.LoadTodos()
	if err != nil {
		t.Fatalf("LoadTodos: %v", err)
	}
	if len(todos) != N { //expects 100!
		t.Fatalf("expected %d todos, got %d", N, len(todos))
	}
}

func TestConcurrentGetHandler(t *testing.T) {
	t.Parallel()

	// Setup
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	// Seed 10 todos
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
	seedJSON(t, sb.String())

	// Fire off 50 concurrent GETs
	const M = 50
	errCh := make(chan error, M)
	for i := 0; i < M; i++ {
		go func(i int) {
			id := (i % N) + 1
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/get?id=%d", id), nil)
			w := httptest.NewRecorder()
			server.GetHandler(w, req)
			if w.Result().StatusCode != http.StatusOK {
				errCh <- fmt.Errorf("GET id=%d status=%d", id, w.Result().StatusCode)
				return
			}
			body, _ := io.ReadAll(w.Body)
			want := fmt.Sprintf("task-%02d", id)
			if !strings.Contains(string(body), want) {
				errCh <- fmt.Errorf("GET id=%d missing %q in %q", id, want, body)
				return
			}
			errCh <- nil
		}(i)
	}

	// Collect errors
	for i := 0; i < M; i++ {
		if err := <-errCh; err != nil {
			t.Error(err)
		}
	}
}

// --- Concurrent UPDATEs ---
func TestConcurrentUpdateHandler(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	storage.Reset()

	// Seed 10 todos
	const N = 10
	var sb strings.Builder
	sb.WriteString("[")
	for i := 1; i <= N; i++ {
		sb.WriteString(fmt.Sprintf(`{"ID":%d,"Description":"Orig-%02d","Started":false,"Done":false}`, i, i))
		if i < N {
			sb.WriteString(",")
		}
	}
	sb.WriteString("]")
	seedJSON(t, sb.String())

	// Fire off N concurrent updates (each updates a distinct ID)
	errCh := make(chan error, N)
	for id := 1; id <= N; id++ {
		go func(id int) {
			form := url.Values{}
			form.Set("id", fmt.Sprint(id))
			form.Set("description", fmt.Sprintf("Upd-%02d", id))
			form.Set("started", "on")
			form.Set("done", "on")
			req := httptest.NewRequest(http.MethodPost, "/update", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()
			server.UpdateHandler(w, req)
			if w.Result().StatusCode != http.StatusSeeOther {
				errCh <- fmt.Errorf("UPDATE id=%d status=%d", id, w.Result().StatusCode)
				return
			}
			errCh <- nil
		}(id)
	}

	// Collect errors
	for i := 0; i < N; i++ {
		if err := <-errCh; err != nil {
			t.Error(err)
		}
	}

	// Verify all updates applied
	final, err := os.ReadFile("todos.json")
	if err != nil {
		t.Fatalf("read todos.json: %v", err)
	}
	var todos []model.Todo
	if err := json.Unmarshal(final, &todos); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(todos) != N {
		t.Fatalf("expected %d todos, got %d", N, len(todos))
	}
	for _, td := range todos {
		want := fmt.Sprintf("Upd-%02d", td.ID)
		if td.Description != want || !td.Started || !td.Done {
			t.Errorf("todo %#v not updated correctly", td)
		}
	}
}

func TestConcurrentDeleteHandler(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	storage.Reset()

	// Seed 10 todos
	const N = 10
	var sb strings.Builder
	sb.WriteString("[")
	for i := 1; i <= N; i++ {
		sb.WriteString(fmt.Sprintf(`{"ID":%d,"Description":"T-%02d","Started":false,"Done":false}`, i, i))
		if i < N {
			sb.WriteString(",")
		}
	}
	sb.WriteString("]")
	seedJSON(t, sb.String())

	// Fire off N concurrent deletes
	errCh := make(chan error, N)
	for id := 1; id <= N; id++ {
		go func(id int) {
			form := url.Values{"id": {fmt.Sprint(id)}}
			req := httptest.NewRequest(http.MethodPost, "/delete", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()
			server.DeleteHandler(w, req)
			if w.Result().StatusCode != http.StatusSeeOther {
				errCh <- fmt.Errorf("DELETE id=%d status=%d", id, w.Result().StatusCode)
				return
			}
			errCh <- nil
		}(id)
	}

	// Collect errors
	for i := 0; i < N; i++ {
		if err := <-errCh; err != nil {
			t.Error(err)
		}
	}

	// Verify all gone
	final, err := os.ReadFile("todos.json")
	if err != nil {
		t.Fatalf("read todos.json: %v", err)
	}
	var todos []model.Todo
	if err := json.Unmarshal(final, &todos); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(todos) != 0 {
		t.Errorf("expected 0 todos after deletes, got %d: %v", len(todos), todos)
	}
}
