package server

import (
	"log"
	"net/http"
)

func Run() {
	mux := http.NewServeMux()

	// using mux.handle I've created a new HTTP server with the following routes injected with the traceMiddleware
	mux.Handle("/create", traceMiddleware(http.HandlerFunc(CreateHandler)))
	mux.Handle("/get", traceMiddleware(http.HandlerFunc(GetHandler)))
	mux.Handle("/update", traceMiddleware(http.HandlerFunc(UpdateHandler)))
	mux.Handle("/delete", traceMiddleware(http.HandlerFunc(DeleteHandler)))
	mux.Handle("/about/", http.StripPrefix("/about/", http.FileServer(http.Dir("static"))))
	mux.Handle("/list", traceMiddleware(http.HandlerFunc(ListHandler)))

	log.Println("Server running on :8080")
	http.ListenAndServe(":8080", mux)
}
