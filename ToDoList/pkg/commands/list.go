package commands

import (
	"fmt"
	"io"
	"os"

	"github.com/Carpentert96/ToDoList/pkg/model"
)

var OutputList io.Writer = os.Stdout

// List prints the todo list and—because main expects it—returns an error.
// Currently it never fails, so it always returns nil.
// updated to incorporate a started status
func List(todos []model.Todo) error {
	if len(todos) == 0 {
		fmt.Fprintln(OutputList, "You have no to-do items.")
		return nil
	}
	fmt.Fprintln(OutputList, "Your to-do list:")
	for _, t := range todos {
		status := "[ ]"
		if t.Started {
			status = "[-]"
		}
		if t.Done {
			status = "[x]"
		}
		fmt.Fprintf(OutputList, "%d: %s %s\n", t.ID, status, t.Description)
	}
	return nil
}
