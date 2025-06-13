package commands

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/Carpentert96/ToDoList/pkg/model"
)

func TestAddTodo_MultipleAdds(t *testing.T) {
	var saved []model.Todo
	fakeSave := func(ts []model.Todo) error {
		saved = append([]model.Todo(nil), ts...)
		return nil
	}

	// Delicration of output in my add.go file for testing
	buf := &bytes.Buffer{}
	orig := Output                   //Original Output variable
	Output = buf                     // Override
	defer func() { Output = orig }() //reset

	todos := []model.Todo{}
	titles := []string{"first task", "second task"}

	for i, title := range titles {
		if err := AddTodo(title, todos, fakeSave); err != nil {
			t.Fatalf("AddTodo(%q) returned error: %v", title, err)
		}
		//Mock todos slice or adding to fake save function
		todos = append([]model.Todo(nil), saved...)

		if len(saved) != i+1 {
			t.Errorf("after adding %q, saved has %d entries; want %d", title, len(saved), i+1)
		}
		last := saved[len(saved)-1]
		if last.ID != i+1 {
			t.Errorf("ID = %d; want %d", last.ID, i+1)
		}
		if last.Description != title {
			t.Errorf("Desc = %q; want %q", last.Description, title)
		}
	}

	out := buf.String()
	for _, title := range titles {
		if !strings.Contains(out, title) {
			t.Errorf("stdout missing %q; got %q", title, out)
		}
	}
}

func TestAddTodo_MultiLineDescription(t *testing.T) {
	multi := "line one\nline two\nline three"

	var saved []model.Todo
	fakeSave := func(ts []model.Todo) error {
		saved = append([]model.Todo(nil), ts...)
		return nil
	}

	buf := &bytes.Buffer{}
	orig := Output
	Output = buf
	defer func() { Output = orig }()

	if err := AddTodo(multi, nil, fakeSave); err != nil {
		t.Fatalf("AddTodo multi-line returned error: %v", err)
	}

	if len(saved) != 1 {
		t.Fatalf("expected 1 todo saved; got %d", len(saved))
	}
	if saved[0].Description != multi {
		t.Errorf("saved desc = %q; want %q", saved[0].Description, multi)
	}
	if !strings.Contains(buf.String(), "Added to-do #1") {
		t.Errorf("stdout = %q; want contains Added to-do #1", buf.String())
	}
}

func TestAddTodo_SaveError(t *testing.T) {
	want := errors.New("disk full")
	fakeSave := func(_ []model.Todo) error { return want }
	if err := AddTodo("anything", nil, fakeSave); !errors.Is(err, want) {
		t.Errorf("error = %v; want %v", err, want)
	}
}

// Simple delete test using hardcoded data rather than a mock (except for save function)
func TestDelete_RemovesCorrectTodo(t *testing.T) {
	initial := []model.Todo{
		{ID: 1, Description: "I hope"},
		{ID: 2, Description: "This works"},
	}
	var saved []model.Todo
	fakeSave := func(ts []model.Todo) error {
		saved = append([]model.Todo(nil), ts...)
		return nil
	}

	err := Delete(initial, 1, fakeSave)
	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	if len(saved) != 1 || saved[0].ID != 2 {
		t.Errorf("expected remaining todo with ID 2, got %+v", saved)
	}
}

func TestUpdate_ChangesDescription(t *testing.T) {
	initial := []model.Todo{
		{ID: 1, Description: "Before update", Done: false},
	}
	var saved []model.Todo
	fakeSave := func(ts []model.Todo) error {
		saved = append([]model.Todo(nil), ts...)
		return nil
	}

	err := Update(1, "After update", initial, fakeSave)
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}

	if saved[0].Description != "After update" {
		t.Errorf("expected description to be updated, got %q", saved[0].Description)
	}
}

// The list test kept failing because I forgot to add it's own mocking, using at one from add.go (which is why it wasn't in error or panic was it?)
// Fixed the test to use a mock OutputList variable (in list.go) instead of the original Output variable
func TestList_PrintsTodosWithStatus(t *testing.T) {
	todos := []model.Todo{
		{ID: 1, Description: "buy Audi R8", Started: false, Done: false},
		{ID: 2, Description: "Audi R8 in progress", Started: true, Done: false},
		{ID: 3, Description: "Huge debt due to R8 completed", Started: true, Done: true},
	}

	buf := &bytes.Buffer{}
	orig := OutputList
	OutputList = buf
	defer func() { OutputList = orig }()

	err := List(todos)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	out := buf.String()

	if !strings.Contains(out, "Your to-do list:") {
		t.Errorf("Expected header 'Your to-do list:' not found\nFull output:\n%s", out)
	}

	expected := map[int]string{
		1: "[ ]", // Not started
		2: "[-]", // Started
		3: "[x]", // Done (also overrides Started)
	}

	for _, todo := range todos {
		wantLine := fmt.Sprintf("%d: %s %s\n", todo.ID, expected[todo.ID], todo.Description)
		if !strings.Contains(out, wantLine) {
			t.Errorf("Expected line %q not found in output:\n%s", wantLine, out)
		}
	}
}
