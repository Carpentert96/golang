package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/Carpentert96/ToDoList/pkg/commands"
	"github.com/Carpentert96/ToDoList/pkg/storage"
)

func main() {
	//Implementation of the init.go function (containing logging and context setup)
	ctx, logger, tid := initApp()

	addPtr := flag.String("add", "", "Add a new to-do")
	listPtr := flag.Bool("list", false, "List all to-dos")
	updateIDPtr := flag.Int("update", 0, "ID to update")
	descPtr := flag.String("desc", "", "New description")
	deleteIDPtr := flag.Int("delete", 0, "ID to delete")
	clearPtr := flag.Bool("clear", false, "Delete all to-dos")
	doneIDPtr := flag.Int("done", 0, "Mark done by ID")
	undoneIDPtr := flag.Int("undone", 0, "Mark undone by ID")
	startedIDPtr := flag.Int("started", 0, "Mark started by ID")
	flag.Parse()

	// As per feedback, I have added one single dereference in which I can call
	addTask := *addPtr
	listFlag := *listPtr
	updateID := *updateIDPtr
	updateDesc := *descPtr
	deleteID := *deleteIDPtr
	clearAll := *clearPtr
	doneID := *doneIDPtr
	undoneID := *undoneIDPtr
	startedID := *startedIDPtr

	// load existing todos from storage and added my first iteration of logging
	todos, err := storage.LoadTodos()
	if err != nil {
		logger.ErrorContext(ctx, "failed to load todos",
			slog.String("trace", tid),
			slog.Any("err", err),
		)
		os.Exit(1)
	}

	// dispatch based on flags
	switch {
	case addPtr != nil && addTask != "":
		logger.InfoContext(ctx, "adding todo",
			slog.String("trace", tid),
			slog.String("task", addTask),
		)
		if err := commands.AddTodo(addTask, todos, storage.SaveTodos); err != nil {
			logger.ErrorContext(ctx, "add failed",
				slog.String("trace", tid),
				slog.Any("err", err),
			)
			os.Exit(1)
		}

	case listPtr != nil && listFlag:
		logger.InfoContext(ctx, "listing todos",
			slog.String("trace", tid),
		)
		if err := commands.List(todos); err != nil {
			logger.ErrorContext(ctx, "list failed",
				slog.String("trace", tid),
				slog.Any("err", err),
			)
			os.Exit(1)
		}

	case updateIDPtr != nil && updateID != 0:
		logger.InfoContext(ctx, "updating todo",
			slog.String("trace", tid),
			slog.Int("id", updateID),
			slog.String("desc", updateDesc),
		)
		if err := commands.Update(updateID, updateDesc, todos, storage.SaveTodos); err != nil {
			logger.ErrorContext(ctx, "update failed",
				slog.String("trace", tid),
				slog.Any("err", err),
			)
			os.Exit(1)
		}

	case deleteIDPtr != nil && deleteID != 0:
		logger.InfoContext(ctx, "deleting todo",
			slog.String("trace", tid),
			slog.Int("id", deleteID),
		)
		if err := commands.Delete(todos, deleteID, storage.SaveTodos); err != nil {
			logger.ErrorContext(ctx, "delete failed",
				slog.String("trace", tid),
				slog.Any("err", err),
			)
			os.Exit(1)
		}

	case clearPtr != nil && clearAll:
		logger.InfoContext(ctx, "clearing all todos",
			slog.String("trace", tid),
		)
		if err := commands.Clear(clearAll, todos, storage.SaveTodos); err != nil {
			logger.ErrorContext(ctx, "clear failed",
				slog.String("trace", tid),
				slog.Any("err", err),
			)
			os.Exit(1)
		}

	case doneIDPtr != nil && doneID != 0:
		logger.InfoContext(ctx, "marking todo done",
			slog.String("trace", tid),
			slog.Int("id", doneID),
		)
		if err := commands.MarkDone(doneID, todos, storage.SaveTodos); err != nil {
			logger.ErrorContext(ctx, "mark done failed",
				slog.String("trace", tid),
				slog.Any("err", err),
			)
			os.Exit(1)
		}

	case undoneIDPtr != nil && undoneID != 0:
		logger.InfoContext(ctx, "marking todo undone",
			slog.String("trace", tid),
			slog.Int("id", undoneID),
		)
		if err := commands.MarkUndone(undoneID, todos, storage.SaveTodos); err != nil {
			logger.ErrorContext(ctx, "mark undone failed",
				slog.String("trace", tid),
				slog.Any("err", err),
			)
			os.Exit(1)
		}

	case startedIDPtr != nil && startedID != 0:
		logger.InfoContext(ctx, "marking todo started",
			slog.String("trace", tid),
			slog.Int("id", startedID),
		)
		if err := commands.MarkStarted(startedID, todos, storage.SaveTodos); err != nil {
			logger.ErrorContext(ctx, "mark started failed",
				slog.String("trace", tid),
				slog.Any("err", err),
			)
			os.Exit(1)
		}

	default:
		fmt.Println("Usage:")
		flag.PrintDefaults()
	}
}
