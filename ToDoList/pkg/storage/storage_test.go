package storage_test

//ran outside of the storage package just for additonal practice
import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/google/uuid"

	"github.com/Carpentert96/ToDoList/pkg/model"
	"github.com/Carpentert96/ToDoList/pkg/storage"
)

// TestStoreBasic exercises AddTodo, LoadTodos, UpdateTodo, and DeleteTodo in sequence.
func TestStoreBasic(t *testing.T) {
	t.Parallel()

	// 1) Prepare a temp file
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "todos.json")
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	tid := uuid.New().String()
	store := storage.NewStore(context.Background(), logger, tid, filePath)

	// 2) Initially, LoadTodos should return an empty slice
	todos, err := store.LoadTodos()
	if err != nil {
		t.Fatalf("LoadTodos failed: %v", err)
	}
	if len(todos) != 0 {
		t.Errorf("expected 0 todos, got %d", len(todos))
	}

	// 3) Add two todos
	wantDescs := []string{"first task", "second task"}
	for _, desc := range wantDescs {
		if err := store.AddTodo(model.Todo{Description: desc}); err != nil {
			t.Fatalf("AddTodo(%q) returned error: %v", desc, err)
		}
	}

	// 4) Verify LoadTodos sees them (in-order IDs)
	todos, err = store.LoadTodos()
	if err != nil {
		t.Fatalf("LoadTodos after add failed: %v", err)
	}
	if len(todos) != len(wantDescs) {
		t.Fatalf("expected %d todos, got %d", len(wantDescs), len(todos))
	}
	for i, td := range todos {
		if td.ID != i+1 {
			t.Errorf("todo #%d ID = %d; want %d", i, td.ID, i+1)
		}
		if td.Description != wantDescs[i] {
			t.Errorf("todo #%d desc = %q; want %q", i, td.Description, wantDescs[i])
		}
	}

	// 5) Update the second todo’s description
	newDesc := "updated second"
	if err := store.UpdateTodo(2, newDesc, false, false); err != nil {
		t.Fatalf("UpdateTodo failed: %v", err)
	}
	todos, _ = store.LoadTodos()
	if todos[1].Description != newDesc {
		t.Errorf("after update, desc = %q; want %q", todos[1].Description, newDesc)
	}

	// 6) Delete the first todo
	if err := store.DeleteTodo(1); err != nil {
		t.Fatalf("DeleteTodo failed: %v", err)
	}
	todos, _ = store.LoadTodos()
	if len(todos) != 1 || todos[0].ID != 2 {
		t.Errorf("after delete, remaining = %+v; want only ID=2", todos)
	}
}

// TestStoreConcurrency spawns multiple AddTodo operations in parallel to verify safety.
func TestStoreConcurrency(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "todos.json") //having written these tests last, I forgot I used "dataFile" instead of "filePath" and was perplexed as to why it was in error. Kept it in anyway
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	tid := uuid.New().String()
	store := storage.NewStore(context.Background(), logger, tid, filePath)

	const N = 50
	var wg sync.WaitGroup
	wg.Add(N)
	for i := 0; i < N; i++ {
		go func(i int) {
			defer wg.Done()
			desc := fmt.Sprintf("task-%02d", i)
			if err := store.AddTodo(model.Todo{Description: desc}); err != nil {
				t.Errorf("AddTodo %q failed: %v", desc, err)
			}
		}(i)
	}
	wg.Wait()

	todos, err := store.LoadTodos()
	if err != nil {
		t.Fatalf("LoadTodos failed: %v", err)
	}
	if len(todos) != N {
		t.Fatalf("expected %d todos, got %d", N, len(todos))
	}
	idSet := make(map[int]bool)
	for _, td := range todos {
		if td.Description == "" {
			t.Errorf("found empty description on ID %d", td.ID)
		}
		if idSet[td.ID] {
			t.Errorf("duplicate ID %d", td.ID)
		}
		idSet[td.ID] = true
	}
}
