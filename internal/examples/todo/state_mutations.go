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

// Toggle flips Done for the todo with id. Unknown id is a no-op.
func (s *State) Toggle(id int) {
	if i := s.indexByID(id); i >= 0 {
		s.Todos[i].Done = !s.Todos[i].Done
	}
}

// Delete removes the todo with id. Unknown id is a no-op.
func (s *State) Delete(id int) {
	if i := s.indexByID(id); i >= 0 {
		s.Todos = slices.Delete(s.Todos, i, i+1)
	}
}

// Edit updates priority and due always; the title only when the trimmed input
// is non-empty (a blank title never overwrites an existing one).
func (s *State) Edit(id int, title, priority, due string) {
	i := s.indexByID(id)
	if i < 0 {
		return
	}
	if t := strings.TrimSpace(title); t != "" {
		if len(t) > MaxTitleLen {
			t = t[:MaxTitleLen]
		}
		s.Todos[i].Title = t
	}
	s.Todos[i].Priority = normalizePriority(priority)
	s.Todos[i].Due = due
}
