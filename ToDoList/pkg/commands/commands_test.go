package commands

import (
	"bytes"
	"errors"
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
	orig := Output
	Output = buf
	defer func() { Output = orig }()

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
