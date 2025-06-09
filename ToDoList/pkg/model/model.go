// model.go
package model

// Todo represents a single to-do entry in a json format.
type Todo struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
	Done        bool   `json:"done"`
}
