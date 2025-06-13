package commands

import (
	"fmt"

	"github.com/Carpentert96/ToDoList/pkg/model"
)

// MarkDone sets Done=true for the given task ID and persists via saveTodos.
// Returns an error if the ID is invalid or saving fails.
func MarkDone(doneIDPtr int, todos []model.Todo, saveTodos func([]model.Todo) error) error {
	// no-op if no ID provided
	if doneIDPtr == 0 {
		return nil
	}

	// find the todo
	found := false
	for i, t := range todos {
		if t.ID == doneIDPtr {
			if todos[i].Done {
				fmt.Printf("To-do #%d is already marked as completed.\n", doneIDPtr)
				return nil
			}
			todos[i].Done = true
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("no to-do with ID %d", doneIDPtr)
	}

	//Save changes
	if err := saveTodos(todos); err != nil {
		return fmt.Errorf("failed to save todos: %w", err)
	}

	fmt.Printf("Marked to-do #%d as completed.\n", doneIDPtr)
	return nil
}
