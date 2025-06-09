package storage

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Carpentert96/ToDoList/pkg/model"
)

const dataFile = "todos.json"

// LoadTodos reads todos.json (if it exists) and returns a slice of model.Todo.
// If the file doesn’t exist, it returns an empty slice.
func LoadTodos() ([]model.Todo, error) {
	var todos []model.Todo

	if _, err := os.Stat(dataFile); os.IsNotExist(err) {
		return []model.Todo{}, nil
	}

	bytes, err := os.ReadFile(dataFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read data file: %w", err)
	}

	if err := json.Unmarshal(bytes, &todos); err != nil {
		return nil, fmt.Errorf("failed to parse data file: %w", err)
	}

	return todos, nil
}

// SaveTodos writes the entire slice of model.Todo back to todos.json.
func SaveTodos(todos []model.Todo) error {
	bytes, err := json.MarshalIndent(todos, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode todos: %w", err)
	}

	if err := os.WriteFile(dataFile, bytes, 0644); err != nil {
		return fmt.Errorf("failed to write data file: %w", err)
	}
	return nil
}
