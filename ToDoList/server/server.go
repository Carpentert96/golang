package server

import (
	"log"
	"net/http"

	"github.com/Carpentert96/ToDoList/pkg/storage"
)

func Run() {

	mux := http.NewServeMux()
	store := storage.NewStore("todos.json")

	// using mux.handle I've created a new HTTP server with the following routes injected with the traceMiddleware
	mux.Handle("/create", traceMiddleware(CreateHandler(store)))
	mux.Handle("/get", traceMiddleware(GetHandler(store)))
	mux.Handle("/update", traceMiddleware(UpdateHandler(store)))
	mux.Handle("/delete", traceMiddleware(DeleteHandler(store)))
	mux.Handle("/list", traceMiddleware(ListHandler(store)))
	mux.Handle("/about/", http.StripPrefix("/about/", http.FileServer(http.Dir("static"))))

	log.Println("Server running on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
