// storage.go
package storage

import (
	"encoding/json"
	"fmt"
	"os"
)

const dataFile = "todos.json"

// loadTodos reads todos.json (if it exists) and returns a slice of Todo.
// If the file doesn’t exist, it returns an empty slice.
func loadTodos() ([]Todo, error) {
	var todos []Todo

	if _, err := os.Stat(dataFile); os.IsNotExist(err) {
		return []Todo{}, nil
	}

	//I notcied a lot of older research online referenced ioutil.ReadFile which was deprecated in Go 1.16. It auto defaults to os.ReadFile now but I changed it anyway
	bytes, err := os.ReadFile(dataFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read data file: %w", err)
	}

	if err := json.Unmarshal(bytes, &todos); err != nil {
		return nil, fmt.Errorf("failed to parse data file: %w", err)
	}

	return todos, nil
}

// saveTodos writes the entire slice of Todo back to todos.json.
func saveTodos(todos []Todo) error {
	bytes, err := json.MarshalIndent(todos, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode todos: %w", err)
	}

	// Use os.WriteFile instead of ioutil.WriteFile (deprecated in Go 1.16)
	if err := os.WriteFile(dataFile, bytes, 0644); err != nil {
		return fmt.Errorf("failed to write data file: %w", err)
	}
	return nil
}
