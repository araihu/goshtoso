// internal/examples/todo/state_mutations.go
package todo

import (
	"slices"
	"strings"
)

func normalizePriority(p string) string {
	switch p {
	case "low", "med", "high":
		return p
	default:
		return "med"
	}
}

// Add appends a todo. Blank titles are ignored; the title is trimmed and capped
// to MaxTitleLen; the list is capped to MaxTodos. ID comes from the Seq counter.
func (s *State) Add(title, priority, due string) {
	title = strings.TrimSpace(title)
	if title == "" || len(s.Todos) >= MaxTodos {
		return
	}
	if len(title) > MaxTitleLen {
		title = title[:MaxTitleLen]
	}
	s.Seq++
	s.Todos = append(s.Todos, Todo{
		ID:       s.Seq,
		Title:    title,
		Priority: normalizePriority(priority),
		Due:      due,
		Order:    len(s.Todos),
	})
}

// indexByID returns the slice index of the todo with id, or -1.
func (s *State) indexByID(id int) int {
	return slices.IndexFunc(s.Todos, func(t Todo) bool { return t.ID == id })
}
