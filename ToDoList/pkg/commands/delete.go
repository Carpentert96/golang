package commands

import (
	"fmt"

	"github.com/Carpentert96/ToDoList/pkg/model"
)

// Delete removes the task with the given ID and persists via saveTodos.
// Returns an error if the ID is not found or saving fails.
func Delete(todos []model.Todo, deleteID int, saveTodos func([]model.Todo) error) error {
	// find index of the todo to delete
	idx := -1
	for i, t := range todos {
		if t.ID == deleteID {
			idx = i
			break
		}
	}
	if idx == -1 {
		return fmt.Errorf("no to-do with ID %d", deleteID)
	}

	// remove the todo
	todos = append(todos[:idx], todos[idx+1:]...)

	// persist changes
	if err := saveTodos(todos); err != nil {
		return fmt.Errorf("failed to save todos: %w", err)
	}

	fmt.Printf("Deleted to-do #%d.\n", deleteID)
	return nil
}
