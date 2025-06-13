package commands

import (
	"fmt"
	"io"
	"os"

	"github.com/Carpentert96/ToDoList/pkg/model"
)

// added for testing purposes (mocking overide)
var Output io.Writer = os.Stdout

// AddTodo creates a new task and returns an error if saving fails.
func AddTodo(addTask string, todos []model.Todo, saveTodos func([]model.Todo) error) error {
	if addTask == "" {
		return nil
	}

	// Incriment +1
	newID := 1
	for _, t := range todos {
		if t.ID >= newID {
			newID = t.ID + 1
		}
	}

	newTodo := model.Todo{ID: newID, Description: addTask, Done: false}
	todos = append(todos, newTodo)

	// Saving to disc
	if err := saveTodos(todos); err != nil {
		return fmt.Errorf("failed to save todos: %w", err)
	}

	fmt.Fprintf(Output, "Added to-do #%d: %s\n", newID, newTodo.Description)
	return nil
}
