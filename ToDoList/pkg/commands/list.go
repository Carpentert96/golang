package commands

import (
	"fmt"

	"github.com/Carpentert96/ToDoList/pkg/model"
)

// List prints the todo list and—because main expects it—returns an error.
// Currently it never fails, so it always returns nil.
// updated to incorporate a started status
func List(todos []model.Todo) error {
	if len(todos) == 0 {
		fmt.Println("You have no to-do items.")
		return nil
	}
	fmt.Println("Your to-do list:")
	for _, t := range todos {
		status := "[ ]"
		if t.Started {
			status = "[-]"
		}
		if t.Done {
			status = "[x]"
		}
		fmt.Printf("%d: %s %s\n", t.ID, status, t.Description)
	}
	return nil
}
