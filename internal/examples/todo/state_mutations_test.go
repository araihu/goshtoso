// internal/examples/todo/state_mutations_test.go
package todo

import (
	"strings"
	"testing"
)

func TestAddAssignsIncrementingIDsAndOrder(t *testing.T) {
	var s State
	s.Add("first", "low", "")
	s.Add("second", "high", "2026-06-01")
	if len(s.Todos) != 2 {
		t.Fatalf("want 2 todos, got %d", len(s.Todos))
	}
	if s.Todos[0].ID != 1 || s.Todos[1].ID != 2 {
		t.Fatalf("ids not monotonic: %+v", s.Todos)
	}
	if s.Todos[0].Order != 0 || s.Todos[1].Order != 1 {
		t.Fatalf("order not sequential: %+v", s.Todos)
	}
	if s.Seq != 2 {
		t.Fatalf("seq want 2, got %d", s.Seq)
	}
}

func TestAddTruncatesTitleAndDefaultsPriority(t *testing.T) {
	var s State
	s.Add(strings.Repeat("x", MaxTitleLen+50), "bogus", "")
	if len(s.Todos[0].Title) != MaxTitleLen {
		t.Fatalf("title not truncated: %d", len(s.Todos[0].Title))
	}
	if s.Todos[0].Priority != "med" {
		t.Fatalf("unknown priority should default to med, got %q", s.Todos[0].Priority)
	}
}

func TestAddRejectsBlankAndRespectsCap(t *testing.T) {
	var s State
	s.Add("   ", "low", "") // blank after trim → ignored
	if len(s.Todos) != 0 {
		t.Fatalf("blank title should be ignored")
	}
	for i := 0; i < MaxTodos+10; i++ {
		s.Add("t", "low", "")
	}
	if len(s.Todos) != MaxTodos {
		t.Fatalf("cap not enforced: %d", len(s.Todos))
	}
}
