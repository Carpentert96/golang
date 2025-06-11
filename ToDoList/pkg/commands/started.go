package commands

import (
	"fmt"

	"github.com/Carpentert96/ToDoList/pkg/model"
)

// MarkStarted sets Started=true (and Done=false) for the given task ID.
// Returns an error if the ID is not found or saving fails.
func MarkStarted(startedID int, todos []model.Todo, saveTodos func([]model.Todo) error) error {
	idx := -1
	for i, t := range todos {
		if t.ID == startedID {
			idx = i
			break
		}
	}
	if idx == -1 {
		return fmt.Errorf("no to-do with ID %d", startedID)
	}

	// flip the flags
	todos[idx].Started = true
	todos[idx].Done = false

	if err := saveTodos(todos); err != nil {
		return fmt.Errorf("failed to save todos: %w", err)
	}

	fmt.Printf("Marked to-do #%d as started.\n", startedID)
	return nil
}
