Hello! Welcome to my basic Golang based to do list application. Below you'll find instructions on how to use my application and any iterations it's been through. Please excuse some of my long comments, it helps to note my errors/ explain how I fixed it. I've also tried to add general comments on first implementations of code as evidence of understanding or just generally for helping me later.

V1; basic CLI flag application
Add a to-do using CLI flags (e.g. `--add "buy audi R8"`).
View all current to-dos with `--list`.
Update or delete to-dos by ID with `--update` and `--delete`.
State is persisted to disk (`todos.json`) on every change.
All to-dos are loaded from disk when the app starts.

How to run; (disclaimer - This won't work anymore as main.go starts the mux server and waits for os interupt ctrl+c)
go run main.go -add "buy Audi R8"
go run main.go -list
go run main.go -update=1 -desc "buy second Audi"
go run main.go -delete=1
go run main.go -done=1
go run main.go -undone=1

V2; Advanced functionality, logging and context tracing
Uses `log/slog` for structured logging, including traceable error messages.
Every operation is tagged with a `TraceID` using Go's `context` package.
Clean architecture: core to-do logic is separated from CLI, HTTP, and main logic.
Unit tests cover all major operations (`Add`, `Delete`, `Update`, `List`), including multi-line descriptions and error paths (for additional evidence).
Graceful shutdown via `os/signal` — the app only exits when interrupted (e.g. `Ctrl+C`).

To run test suite;
go test -v ./pkg/commands

V3; Basic API implementation (this is what replaces V1)
Exposes to-do operations via `net/http` and `ServeMux`:
http://localhost:8080/create — add a new to-do
http://localhost:8080/get — get all to-dos
http://localhost:8080/update — update a description by ID
http://localhost:8080/delete — delete a to-do by ID
- Middleware injects a `TraceID` into the request context for traceable logging.

Note! - 'go run main.go' starts the server!

V4; Front end web interface
http://localhost:8080/list - lists all to do tasks in the json file (Some colour added)
http://localhost:8080/about/about.html - created front end about page with a little demo


Note! - 'go run main.go' starts the server!




Feel free to clone, edit or use as examples if you wish

```bash
git clone https://github.com/yourusername/ToDoList.git
cd ToDoList

Built by @Carpentert96 as part of a training project