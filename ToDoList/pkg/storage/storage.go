// pkg/storage/storage.go
package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/Carpentert96/ToDoList/pkg/model"
)

// Default is the global store instance used by handlers.
//var Default *Store

// Store represents a single actor handling todos.json writes.
type Store struct {
	filePath string
	cmds     chan interface{}
}

// Command types for actor

type addCmd struct {
	Description string
	Started     bool
	Done        bool
	resp        chan error
}

type updateCmd struct {
	ID          int
	Description string
	Started     bool
	Done        bool
	resp        chan error
}

type deleteCmd struct {
	ID   int
	resp chan error
}

// NewStore creates and initializes a Store at filePath.
// It starts a background goroutine to serialize write commands.
func NewStore(filePath string) *Store {
	s := &Store{
		filePath: filePath,
		cmds:     make(chan interface{}, 100), // enough room for bursts
	}

	// Ensure the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		os.WriteFile(filePath, []byte("[]"), 0644)
	}

	// Launch actor
	go s.actorLoop()

	//	Default = s
	return s
}

// actorLoop runs commands sequentially to prevent races.
func (s *Store) actorLoop() {
	var mu sync.Mutex // protect in-memory slice
	var todos []model.Todo

	// load initial state
	if b, err := os.ReadFile(s.filePath); err == nil {
		json.Unmarshal(b, &todos)
	}

	for cmd := range s.cmds {
		switch c := cmd.(type) {
		case addCmd:
			mu.Lock()
			// Assign unique ID
			maxID := 0
			for _, t := range todos {
				if t.ID > maxID {
					maxID = t.ID
				}
			}
			newTodo := model.Todo{
				ID:          maxID + 1,
				Description: c.Description,
				Started:     c.Started,
				Done:        c.Done,
			}
			todos = append(todos, newTodo)
			err := s.save(todos)
			mu.Unlock()
			c.resp <- err

		case updateCmd:
			mu.Lock()
			for i, t := range todos {
				if t.ID == c.ID {
					todos[i].Description = c.Description
					todos[i].Started = c.Started
					todos[i].Done = c.Done
					break
				}
			}
			err := s.save(todos)
			mu.Unlock()
			c.resp <- err

		case deleteCmd:
			mu.Lock()
			var keep []model.Todo
			for _, t := range todos {
				if t.ID != c.ID {
					keep = append(keep, t)
				}
			}
			todos = keep
			err := s.save(todos)
			mu.Unlock()
			c.resp <- err
		}
	}
}

// save writes the current todos slice to disk.
func (s *Store) save(todos []model.Todo) error {
	b, err := json.MarshalIndent(todos, "", "  ")
	if err != nil {
		return fmt.Errorf("encode todos: %w", err)
	}
	if err := os.WriteFile(s.filePath, b, 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}

// LoadTodos reads directly from disk, bypassing actor; used for GETs and tests.
func (s *Store) LoadTodos() ([]model.Todo, error) {
	b, err := os.ReadFile(s.filePath)
	if err != nil {
		return nil, fmt.Errorf("read todos: %w", err)
	}
	var todos []model.Todo
	if err := json.Unmarshal(b, &todos); err != nil {
		return nil, fmt.Errorf("parse todos: %w", err)
	}
	return todos, nil
}

// AddTodo enqueues a new todo with internal ID generation.
func (s *Store) AddTodo(todo model.Todo) error {
	ch := make(chan error)
	s.cmds <- addCmd{
		Description: todo.Description,
		Started:     todo.Started,
		Done:        todo.Done,
		resp:        ch,
	}
	return <-ch
}

// UpdateTodo enqueues an update for the given todo ID.
func (s *Store) UpdateTodo(id int, description string, started, done bool) error {
	ch := make(chan error)
	s.cmds <- updateCmd{ID: id, Description: description, Started: started, Done: done, resp: ch}
	return <-ch
}

// DeleteTodo enqueues a deletion of the given todo ID.
func (s *Store) DeleteTodo(id int) error {
	ch := make(chan error)
	s.cmds <- deleteCmd{ID: id, resp: ch}
	return <-ch
}
