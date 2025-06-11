package commands

import (
	"fmt"

	"github.com/Carpentert96/ToDoList/pkg/model"
)

// AddTodo creates a new task if desc is non-empty and persists via saveTodos.
// Returns an error if saving fails.
func AddTodo(addTask string, todos []model.Todo, saveTodos func([]model.Todo) error) error {
	if addTask == "" {
		return nil
	}

	// generate next available ID
	newID := 1
	for _, t := range todos {
		if t.ID >= newID {
			newID = t.ID + 1
		}
	}

	newTodo := model.Todo{ID: newID, Description: addTask, Done: false}
	todos = append(todos, newTodo)

	// persist tasks using injected saveTodos
	if err := saveTodos(todos); err != nil {
		return fmt.Errorf("failed to save todos: %w", err)
	}

	fmt.Printf("Added to-do #%d: %s\n", newID, newTodo.Description)
	return nil
}
