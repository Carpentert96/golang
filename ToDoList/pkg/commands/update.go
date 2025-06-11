package commands

import (
	"fmt"

	"github.com/Carpentert96/ToDoList/pkg/model"
)

// Update modifies the description of an existing task and persists via saveTodos.
// Returns an error if the input is invalid or saving fails.
func Update(updateIDPtr int, descPtr string, todos []model.Todo, saveTodos func([]model.Todo) error) error {
	// no-op if no update ID provided
	if updateIDPtr == 0 {
		return nil
	}

	// require a new description
	if descPtr == "" {
		return fmt.Errorf("must supply -desc when using -update")
	}

	// find and update the todo
	found := false
	for i, t := range todos {
		if t.ID == updateIDPtr {
			todos[i].Description = descPtr
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("no to-do with ID %d", updateIDPtr)
	}

	// persist changes
	if err := saveTodos(todos); err != nil {
		return fmt.Errorf("failed to save todos: %w", err)
	}

	fmt.Printf("Updated to-do #%d.\n", updateIDPtr)
	return nil
}
