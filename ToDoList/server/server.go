package server

import (
	"log"
	"net/http"
)

func Run() {
	mux := http.NewServeMux()

	// using mux.handle I've created a new HTTP server with the following routes injected with the traceMiddleware
	mux.Handle("/create", traceMiddleware(http.HandlerFunc(createHandler)))
	mux.Handle("/get", traceMiddleware(http.HandlerFunc(getHandler)))
	mux.Handle("/update", traceMiddleware(http.HandlerFunc(updateHandler)))
	mux.Handle("/delete", traceMiddleware(http.HandlerFunc(deleteHandler)))
	mux.Handle("/about/", http.StripPrefix("/about/", http.FileServer(http.Dir("static"))))
	mux.Handle("/list", traceMiddleware(http.HandlerFunc(listHandler)))

	log.Println("Server running on :8080")
	http.ListenAndServe(":8080", mux)
}
