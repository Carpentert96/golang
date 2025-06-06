// main.go
package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	// 1) Define CLI flags, with some additional features -done, -undone and -clear
	addPtr := flag.String("add", "", "Add a new to-do:       -add=\"Task description\"")
	listPtr := flag.Bool("list", false, "List all to-do items:  -list")
	updateIDPtr := flag.Int("update", 0, "ID of the to-do to update:  -update=<id> -desc=\"New description\"")
	descPtr := flag.String("desc", "", "New description when using -update")
	deleteIDPtr := flag.Int("delete", 0, "ID of the to-do to delete:  -delete=<id>")
	clearAllPtr := flag.Bool("clear", false, "Delete ALL to-do items: -clear")
	doneIDPtr := flag.Int("done", 0, "Mark a to-do as completed:  -done=<id>")
	undoneIDPtr := flag.Int("undone", 0, "Unmark a completed to-do:  -undone=<id>")
	flag.Parse()

	// 2) Load existing todos from todos.json (internal storage evidence)
	todos, err := loadTodos()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading todos: %v\n", err)
		os.Exit(1)
	}

	// 3) If -clear is set, wipe the entire list and exit (added because it was annoying to have to delete each task one by one)
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

	// 4) If -done is set, mark that task Done=true and exit
	if *doneIDPtr != 0 {
		found := false
		for i, t := range todos {
			if t.ID == *doneIDPtr {
				if todos[i].Done {
					fmt.Printf("To-do #%d is already marked as completed.\n", *doneIDPtr)
					return
				}
				todos[i].Done = true
				found = true
				break
			}
		}
		if !found {
			fmt.Fprintf(os.Stderr, "Error: no to-do with ID %d\n", *doneIDPtr)
			os.Exit(1)
		}
		if err := saveTodos(todos); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to save todos: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Marked to-do #%d as completed.\n", *doneIDPtr)
		return
	}

	// 5) If -undone is set, mark that task Done=false and exit
	if *undoneIDPtr != 0 {
		found := false
		for i, t := range todos {
			if t.ID == *undoneIDPtr {
				if !todos[i].Done {
					fmt.Printf("To-do #%d is already marked as not completed.\n", *undoneIDPtr)
					return
				}
				todos[i].Done = false
				found = true
				break
			}
		}
		if !found {
			fmt.Fprintf(os.Stderr, "Error: no to-do with ID %d\n", *undoneIDPtr)
			os.Exit(1)
		}
		if err := saveTodos(todos); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to save todos: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Marked to-do #%d as not completed.\n", *undoneIDPtr)
		return
	}

	// 6) Otherwise, dispatch to exactly one other flag-based action
	switch {
	// 7) If -add is set, add a new task with the given description
	case *addPtr != "":
		newID := 1
		for _, t := range todos {
			if t.ID >= newID {
				newID = t.ID + 1
			}
		}
		newTodo := Todo{ID: newID, Description: *addPtr, Done: false}
		todos = append(todos, newTodo)

		if err := saveTodos(todos); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to save todos: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Added to-do #%d: %s\n", newID, newTodo.Description)

	// 8) If -list is set, print the entire list of tasks
	case *listPtr:
		if len(todos) == 0 {
			fmt.Println("You have no to-do items.")
			return
		}
		fmt.Println("Your to-do list:")
		for _, t := range todos {
			status := "[ ]"
			if t.Done {
				status = "[x]"
			}
			fmt.Printf("%d: %s %s\n", t.ID, status, t.Description)
		}

	// 9) If -update is set, modify the description of an existing task
	case *updateIDPtr != 0:
		if *descPtr == "" {
			fmt.Fprintln(os.Stderr, "Error: must supply -desc when using -update")
			os.Exit(1)
		}
		found := false
		for i, t := range todos {
			if t.ID == *updateIDPtr {
				todos[i].Description = *descPtr
				found = true
				break
			}
		}
		if !found {
			fmt.Fprintf(os.Stderr, "Error: no to-do with ID %d\n", *updateIDPtr)
			os.Exit(1)
		}
		if err := saveTodos(todos); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to save todos: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Updated to-do #%d.\n", *updateIDPtr)

	// 10) If -delete is set, remove the specified task
	case *deleteIDPtr != 0:
		idx := -1
		for i, t := range todos {
			if t.ID == *deleteIDPtr {
				idx = i
				break
			}
		}
		if idx == -1 {
			fmt.Fprintf(os.Stderr, "Error: no to-do with ID %d\n", *deleteIDPtr)
			os.Exit(1)
		}
		todos = append(todos[:idx], todos[idx+1:]...)
		if err := saveTodos(todos); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to save todos: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Deleted to-do #%d.\n", *deleteIDPtr)

	// 11) If no flags are set, print usage information
	default:
		fmt.Println("Usage:")
		flag.PrintDefaults()
	}
}
