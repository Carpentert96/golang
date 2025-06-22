package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Carpentert96/ToDoList/pkg/model"
	"github.com/Carpentert96/ToDoList/pkg/storage"
)

func main() {
	// Initialize a store instance with actor-based concurrency
	store := storage.NewStore("todos.json")

	// Set up flags
	addPtr := flag.String("add", "", "Add a new to-do")
	listPtr := flag.Bool("list", false, "List all to-dos")
	updateIDPtr := flag.Int("update", 0, "ID to update")
	descPtr := flag.String("desc", "", "New description for update")
	deleteIDPtr := flag.Int("delete", 0, "ID to delete")
	clearPtr := flag.Bool("clear", false, "Delete all to-dos")
	doneIDPtr := flag.Int("done", 0, "Mark done by ID")
	undoneIDPtr := flag.Int("undone", 0, "Mark undone by ID")
	startedIDPtr := flag.Int("started", 0, "Mark started by ID")
	flag.Parse()

	// Prepare logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	// Load current todos from store
	todos, err := store.LoadTodos()
	if err != nil {
		logger.Error("failed to load todos", "err", err)
		os.Exit(1)
	}

	switch {
	case *addPtr != "":
		// Compute next ID
		nextID := 1
		for _, t := range todos {
			if t.ID >= nextID {
				nextID = t.ID + 1
			}
		}
		newTodo := model.Todo{ID: nextID, Description: *addPtr}
		if err := store.AddTodo(newTodo); err != nil {
			logger.Error("add failed", "err", err)
			os.Exit(1)
		}
		fmt.Printf("Added todo #%d: %s\n", newTodo.ID, newTodo.Description)

	case *listPtr:
		for _, t := range todos {
			status := ""
			if t.Started {
				status += "[Started] "
			}
			if t.Done {
				status += "[Done] "
			}
			fmt.Printf("#%d: %s %s\n", t.ID, t.Description, status)
		}

	case *updateIDPtr != 0:
		id := *updateIDPtr
		if err := store.UpdateTodo(id, *descPtr, true, false); err != nil {
			logger.Error("update failed", "id", id, "err", err)
			os.Exit(1)
		}
		fmt.Printf("Updated todo #%d to %s\n", id, *descPtr)

	case *deleteIDPtr != 0:
		id := *deleteIDPtr
		if err := store.DeleteTodo(id); err != nil {
			logger.Error("delete failed", "id", id, "err", err)
			os.Exit(1)
		}
		fmt.Printf("Deleted todo #%d\n", id)

	case *clearPtr:
		for _, t := range todos {
			if err := store.DeleteTodo(t.ID); err != nil {
				logger.Error("clear failed deleting", "id", t.ID, "err", err)
				os.Exit(1)
			}
		}
		fmt.Println("Cleared all todos")

	case *doneIDPtr != 0:
		id := *doneIDPtr
		if err := store.UpdateTodo(id, "", true, true); err != nil {
			logger.Error("mark done failed", "id", id, "err", err)
			os.Exit(1)
		}
		fmt.Printf("Marked todo #%d done\n", id)

	case *undoneIDPtr != 0:
		id := *undoneIDPtr
		if err := store.UpdateTodo(id, "", false, false); err != nil {
			logger.Error("mark undone failed", "id", id, "err", err)
			os.Exit(1)
		}
		fmt.Printf("Marked todo #%d undone\n", id)

	case *startedIDPtr != 0:
		id := *startedIDPtr
		if err := store.UpdateTodo(id, "", true, false); err != nil {
			logger.Error("mark started failed", "id", id, "err", err)
			os.Exit(1)
		}
		fmt.Printf("Marked todo #%d started\n", id)

	default:
		flag.Usage()
	}

	// Wait for interrupt signal to exit
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
}
