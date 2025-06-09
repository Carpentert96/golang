package commands

import (
	"fmt"
	"os"
)

// 3) If -clear is set, wipe the entire list and exit (added because it was annoying to have to delete each task one by one)
// Practicing a little error handling here
func ClearTodos(clearAllPtr *bool, todos []Todo) {
	if *clearAllPtr {
		if len(todos) == 0 {
			fmt.Println("You have no tasks to delete.")
			return
		}
		todos = []Todo{}
		if err := saveTodos(todos); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to save todos: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("All tasks have been deleted.")
		return
	}
}
