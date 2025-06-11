// model.go
package model

// Todo represents a single to-do entry in a json format
// Function now change to caps so it can be used outside the package via git import
type Todo struct {
	ID          int
	Description string
	Started     bool
	Done        bool
}
