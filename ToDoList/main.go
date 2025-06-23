package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Carpentert96/ToDoList/pkg/model"
	"github.com/Carpentert96/ToDoList/pkg/storage"
	"github.com/Carpentert96/ToDoList/server"
)

func main() {
	// 1) Init context + logger
	ctx, logger, tid := server.InitApp()

	//Last change for the store function to add logging after every write/save
	store := storage.NewStore(ctx, logger, tid, "todos.json")

	// 3) Define flags
	servePtr := flag.Bool("serve", false, "Run as HTTP server")
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

	// 4) Dereference once
	addTask := *addPtr
	listFlag := *listPtr
	updateID := *updateIDPtr
	updateDesc := *descPtr
	deleteID := *deleteIDPtr
	clearAll := *clearPtr
	doneID := *doneIDPtr
	undoneID := *undoneIDPtr
	startedID := *startedIDPtr

	// 5) HTTP server mode
	if *servePtr {
		logger.InfoContext(ctx, "▶️  starting HTTP server")

		// shut down on SIGINT/SIGTERM
		ctx, cancel := context.WithCancel(ctx)
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-sigCh
			cancel()
		}()

		if err := server.Run(ctx, store); err != nil {
			logger.ErrorContext(ctx, "server.Run failed", "err", err)
			os.Exit(1)
		}
		return
	}

	// 6) CLI mode: load todos
	todos, err := store.LoadTodos()
	if err != nil {
		logger.ErrorContext(ctx, "failed to load todos", "err", err)
		os.Exit(1)
	}

	// 7) Dispatch based on flags (using the dereferenced vars)
	switch {
	case addTask != "":
		id := nextID(todos)
		t := model.Todo{ID: id, Description: addTask}
		logger.InfoContext(ctx, "adding todo", "id", id, "desc", addTask)
		if err := store.AddTodo(t); err != nil {
			logger.ErrorContext(ctx, "AddTodo failed", "err", err)
			os.Exit(1)
		}
		fmt.Printf("Added todo #%d: %s\n", id, addTask)

	case listFlag:
		logger.InfoContext(ctx, "listing todos")
		for _, t := range todos {
			status := ""
			if t.Started {
				status += "[Started] "
			}
			if t.Done {
				status += "[Done]"
			}
			fmt.Printf("#%d: %s %s\n", t.ID, t.Description, status)
		}

	case updateID != 0:
		if updateDesc == "" {
			fmt.Fprintln(os.Stderr, "❌ must provide -desc when using -update")
			os.Exit(1)
		}
		logger.InfoContext(ctx, "updating todo", "id", updateID, "desc", updateDesc)
		if err := store.UpdateTodo(updateID, updateDesc, true, false); err != nil {
			logger.ErrorContext(ctx, "UpdateTodo failed", "err", err)
			os.Exit(1)
		}
		fmt.Printf("Updated todo #%d to %s\n", updateID, updateDesc)

	case deleteID != 0:
		logger.InfoContext(ctx, "deleting todo", "id", deleteID)
		if err := store.DeleteTodo(deleteID); err != nil {
			logger.ErrorContext(ctx, "DeleteTodo failed", "err", err)
			os.Exit(1)
		}
		fmt.Printf("Deleted todo #%d\n", deleteID)

	case clearAll:
		logger.InfoContext(ctx, "clearing all todos")
		for _, t := range todos {
			if err := store.DeleteTodo(t.ID); err != nil {
				logger.ErrorContext(ctx, "DeleteTodo failed", "id", t.ID, "err", err)
				os.Exit(1)
			}
		}
		fmt.Println("Cleared all todos")

	case doneID != 0:
		logger.InfoContext(ctx, "marking done", "id", doneID)
		if err := store.UpdateTodo(doneID, "", true, true); err != nil {
			logger.ErrorContext(ctx, "MarkDone failed", "err", err)
			os.Exit(1)
		}
		fmt.Printf("Marked todo #%d done\n", doneID)

	case undoneID != 0:
		logger.InfoContext(ctx, "marking undone", "id", undoneID)
		if err := store.UpdateTodo(undoneID, "", false, false); err != nil {
			logger.ErrorContext(ctx, "MarkUndone failed", "err", err)
			os.Exit(1)
		}
		fmt.Printf("Marked todo #%d undone\n", undoneID)

	case startedID != 0:
		logger.InfoContext(ctx, "marking started", "id", startedID)
		if err := store.UpdateTodo(startedID, "", true, false); err != nil {
			logger.ErrorContext(ctx, "MarkStarted failed", "err", err)
			os.Exit(1)
		}
		fmt.Printf("Marked todo #%d started\n", startedID)

	default:
		flag.Usage()
	}
}

// nextID finds the next available ID.
func nextID(todos []model.Todo) int {
	max := 0
	for _, t := range todos {
		if t.ID > max {
			max = t.ID
		}
	}
	return max + 1
}
