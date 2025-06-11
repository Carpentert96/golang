package commands

import (
	"fmt"

	"github.com/Carpentert96/ToDoList/pkg/model"
)

// Clear wipes all tasks if clearAllPtr is true and persists via saveTodos.
// Returns an error if saving fails.
func Clear(clearAllPtr bool, todos []model.Todo, saveTodos func([]model.Todo) error) error {
	// no-op if flag not set
	if !clearAllPtr {
		return nil
	}

	if len(todos) == 0 {
		fmt.Println("You have no tasks to delete.")
		return nil
	}

	// clear all todos
	todos = []model.Todo{}

	// persist changes
	if err := saveTodos(todos); err != nil {
		return fmt.Errorf("failed to save todos: %w", err)
	}

	fmt.Println("All tasks have been deleted.")
	return nil
}
