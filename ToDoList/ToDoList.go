// main.go
package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	// 1) Load existing todos (if any)
	todos, err := loadTodos()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading todos: %v\n", err)
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		// 2) Show menu
		fmt.Println()
		fmt.Println("===== ToDoList Menu =====")
		fmt.Println("1) Add a new task")
		fmt.Println("2) Show all tasks")
		fmt.Println("3) Update a task")
		fmt.Println("4) Delete a task")
		fmt.Println("5) Exit")
		fmt.Print("Enter choice [1-5]: ")

		choiceRaw, _ := reader.ReadString('\n')
		choiceRaw = strings.TrimSpace(choiceRaw)
		choice, err := strconv.Atoi(choiceRaw)
		if err != nil || choice < 1 || choice > 5 {
			fmt.Println("Invalid choice. Please enter a number between 1 and 5.")
			continue
		}

		switch choice {
		case 1:
			// ADD
			fmt.Print("Enter task description: ")
			desc, _ := reader.ReadString('\n')
			desc = strings.TrimSpace(desc)
			if desc == "" {
				fmt.Println("Description cannot be empty.")
				continue
			}

			// Determine new ID
			newID := 1
			for _, t := range todos {
				if t.ID >= newID {
					newID = t.ID + 1
				}
			}
			newTodo := Todo{
				ID:          newID,
				Description: desc,
				Done:        false,
			}
			todos = append(todos, newTodo)

			if err := saveTodos(todos); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to save todos: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Added to-do #%d.\n", newID)

		case 2:
			// LIST
			if len(todos) == 0 {
				fmt.Println("You have no to-do items.")
			} else {
				fmt.Println("Your to-do list:")
				for _, t := range todos {
					fmt.Printf("%d: %s\n", t.ID, t.Description)
				}
			}

		case 3:
			// UPDATE
			if len(todos) == 0 {
				fmt.Println("No tasks to update.")
				continue
			}
			fmt.Print("Enter ID of task to update: ")
			idRaw, _ := reader.ReadString('\n')
			idRaw = strings.TrimSpace(idRaw)
			id, err := strconv.Atoi(idRaw)
			if err != nil {
				fmt.Println("Invalid ID. Must be a number.")
				continue
			}

			index := -1
			for i, t := range todos {
				if t.ID == id {
					index = i
					break
				}
			}
			if index == -1 {
				fmt.Printf("No task found with ID %d.\n", id)
				continue
			}

			fmt.Printf("Current description: %q\n", todos[index].Description)
			fmt.Print("Enter new description: ")
			newDesc, _ := reader.ReadString('\n')
			newDesc = strings.TrimSpace(newDesc)
			if newDesc == "" {
				fmt.Println("Description cannot be empty.")
				continue
			}
			todos[index].Description = newDesc

			if err := saveTodos(todos); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to save todos: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Updated task #%d.\n", id)

		case 4:
			// DELETE
			if len(todos) == 0 {
				fmt.Println("No tasks to delete.")
				continue
			}
			fmt.Print("Enter ID of task to delete: ")
			idRaw, _ := reader.ReadString('\n')
			idRaw = strings.TrimSpace(idRaw)
			id, err := strconv.Atoi(idRaw)
			if err != nil {
				fmt.Println("Invalid ID. Must be a number.")
				continue
			}

			index := -1
			for i, t := range todos {
				if t.ID == id {
					index = i
					break
				}
			}
			if index == -1 {
				fmt.Printf("No task found with ID %d.\n", id)
				continue
			}

			// Remove from slice
			todos = append(todos[:index], todos[index+1:]...)

			if err := saveTodos(todos); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to save todos: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Deleted task #%d.\n", id)

		case 5:
			// EXIT
			fmt.Println("Goodbye!")
			return
		}
	}
}
