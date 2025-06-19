package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Carpentert96/ToDoList/pkg/model"
)

// --- actor commands & results ---
type (
	readCmd  struct{ resp chan []model.Todo }
	writeCmd struct {
		todos []model.Todo
		resp  chan error
	}
	addCmd struct {
		desc          string
		started, done bool
		resp          chan addResult
	}
)
type addResult struct {
	todo model.Todo
	err  error
}
type resetCmd struct{ resp chan struct{} }

var actorCh = make(chan interface{})

func init() {
	go func() {
		var (
			todos       []model.Todo
			initialized bool
		)
		for cmd := range actorCh {
			// first command: load from disk in the current cwd
			if !initialized {
				todos = loadFromDisk()
				initialized = true
			}
			switch c := cmd.(type) {
			case readCmd:
				// return a copy
				cp := make([]model.Todo, len(todos))
				copy(cp, todos)
				c.resp <- cp

			case writeCmd:
				todos = c.todos
				c.resp <- saveToDisk(todos)

			case addCmd:
				// compute new ID
				var maxID int
				for _, t := range todos {
					if t.ID > maxID {
						maxID = t.ID
					}
				}
				newTodo := model.Todo{
					ID:          maxID + 1,
					Description: c.desc,
					Started:     c.started,
					Done:        c.done,
				}
				todos = append(todos, newTodo)
				err := saveToDisk(todos)
				c.resp <- addResult{todo: newTodo, err: err}

				//had to add this to avoid additional writes to disk (the test was bringing back 21 todos instead of 20)
			case resetCmd:
				todos = nil
				initialized = false
				c.resp <- struct{}{}
			}
		}
	}()
}

// Calls reset inside of actor loop to avoid additional writes to disk (We're also about to call it in my test)
func Reset() {
	resp := make(chan struct{})
	actorCh <- resetCmd{resp: resp}
	<-resp
}

// LoadTodos asks the actor for the full slice.
func LoadTodos() ([]model.Todo, error) {
	resp := make(chan []model.Todo)
	actorCh <- readCmd{resp: resp}
	return <-resp, nil
}

// SaveTodos asks the actor to overwrite the slice.
func SaveTodos(todos []model.Todo) error {
	resp := make(chan error)
	actorCh <- writeCmd{todos: todos, resp: resp}
	return <-resp
}

// AddTodo asks the actor to append one new Todo and persist.
func AddTodo(desc string, started, done bool) (model.Todo, error) {
	resp := make(chan addResult)
	actorCh <- addCmd{desc: desc, started: started, done: done, resp: resp}
	res := <-resp
	return res.todo, res.err
}

// --- private disk I/O ---

const dataFile = "todos.json"

func loadFromDisk() []model.Todo {
	if _, err := os.Stat(dataFile); os.IsNotExist(err) {
		return nil
	}
	b, err := os.ReadFile(dataFile)
	if err != nil {
		panic(fmt.Errorf("storage actor read error: %w", err))
	}
	var todos []model.Todo
	if err := json.Unmarshal(b, &todos); err != nil {
		panic(fmt.Errorf("storage actor unmarshal error: %w", err))
	}
	return todos
}

func saveToDisk(todos []model.Todo) error {
	b, err := json.MarshalIndent(todos, "", "  ")
	if err != nil {
		return fmt.Errorf("storage actor marshal error: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(dataFile), 0755); err != nil {
		return fmt.Errorf("storage actor mkdir error: %w", err)
	}
	return os.WriteFile(dataFile, b, 0644)
}
