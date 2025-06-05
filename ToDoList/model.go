// model.go
package main

// Todo represents a single to-do entry.
type Todo struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
	Done        bool   `json:"done"` // currently unused, but available if you want a “completed” toggle later
}
