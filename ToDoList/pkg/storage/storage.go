package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/Carpentert96/ToDoList/pkg/model"
)

// Store represents a single actor handling todos.json writes.
type Store struct {
	filePath string
	cmds     chan interface{}

	ctx    context.Context
	logger *slog.Logger
	tid    string
}

// Commands sent over the actor channel:
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
// You must pass in the traceID you got from InitApp.
// It starts a background goroutine to serialize all writes.
func NewStore(
	ctx context.Context,
	logger *slog.Logger,
	traceID string,
	filePath string,
) *Store {
	s := &Store{
		filePath: filePath,
		cmds:     make(chan interface{}, 100),
		ctx:      ctx,
		logger:   logger,
		tid:      traceID,
	}
	// ensure file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		os.WriteFile(filePath, []byte("[]"), 0644)
	}
	go s.actorLoop()
	return s
}

func (s *Store) actorLoop() {
	var mu sync.Mutex
	var todos []model.Todo

	// load initial
	if b, err := os.ReadFile(s.filePath); err == nil {
		json.Unmarshal(b, &todos)
	}

	for cmd := range s.cmds {
		switch c := cmd.(type) {

		case addCmd:
			mu.Lock()
			// assign next ID
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

			if err == nil {
				s.logger.InfoContext(s.ctx,
					"todos saved (add)",
					slog.String("trace", s.tid),
					slog.Int("count", len(todos)),
				)
			}
			c.resp <- err

		case updateCmd:
			mu.Lock()
			for i, t := range todos {
				if t.ID == c.ID {
					if c.Description != "" {
						todos[i].Description = c.Description
					}
					todos[i].Started = c.Started
					todos[i].Done = c.Done
					break
				}
			}
			err := s.save(todos)
			mu.Unlock()

			if err == nil {
				s.logger.InfoContext(s.ctx,
					"todos saved (update)",
					slog.String("trace", s.tid),
					slog.Int("count", len(todos)),
				)
			}
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

			if err == nil {
				s.logger.InfoContext(s.ctx,
					"todos saved (delete)",
					slog.String("trace", s.tid),
					slog.Int("count", len(todos)),
				)
			}
			c.resp <- err
		}
	}
}

// save writes todos to disk.
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

// LoadTodos reads todos from disk.
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

func (s *Store) AddTodo(todo model.Todo) error {
	ch := make(chan error)
	s.cmds <- addCmd{Description: todo.Description, Started: todo.Started, Done: todo.Done, resp: ch}
	return <-ch
}

func (s *Store) UpdateTodo(id int, description string, started, done bool) error {
	ch := make(chan error)
	s.cmds <- updateCmd{ID: id, Description: description, Started: started, Done: done, resp: ch}
	return <-ch
}

func (s *Store) DeleteTodo(id int) error {
	ch := make(chan error)
	s.cmds <- deleteCmd{ID: id, resp: ch}
	return <-ch
}
