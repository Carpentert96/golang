package commands

import (
	"fmt"

	"github.com/Carpentert96/ToDoList/pkg/model"
)

// MarkUndone sets Done=false for the given task ID and persists via saveTodos.
// Returns an error if the ID is not found or saving fails.
func MarkUndone(undoneIDPtr int, todos []model.Todo, saveTodos func([]model.Todo) error) error {
	// no-op if no ID provided
	if undoneIDPtr == 0 {
		return nil
	}

	// find the todo
	found := false
	for i, t := range todos {
		if t.ID == undoneIDPtr {
			if !todos[i].Done {
				fmt.Printf("To-do #%d is already marked as not completed.\n", undoneIDPtr)
				return nil
			}
			todos[i].Done = false
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("no to-do with ID %d", undoneIDPtr)
	}

	// persist changes
	if err := saveTodos(todos); err != nil {
		return fmt.Errorf("failed to save todos: %w", err)
	}

	fmt.Printf("Marked to-do #%d as not completed.\n", undoneIDPtr)
	return nil
}
