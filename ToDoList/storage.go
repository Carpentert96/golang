// storage.go
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

const dataFile = "todos.json"

// loadTodos reads todos.json (if it exists) and returns a slice of Todo.
// If the file doesn’t exist, it returns an empty slice (no error).
func loadTodos() ([]Todo, error) {
	var todos []Todo
	if _, err := os.Stat(dataFile); os.IsNotExist(err) {
		return []Todo{}, nil
	}
	bytes, err := ioutil.ReadFile(dataFile)
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
	if err := ioutil.WriteFile(dataFile, bytes, 0644); err != nil {
		return fmt.Errorf("failed to write data file: %w", err)
	}
	return nil
}
