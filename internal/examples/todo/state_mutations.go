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

// Reorder reassigns Order so the todos named in ids come first in that sequence;
// any todo not listed keeps its relative order and trails after. Unknown ids are
// ignored.
func (s *State) Reorder(ids []int) {
	rank := make(map[int]int, len(ids))
	for pos, id := range ids {
		rank[id] = pos
	}
	next := len(ids)
	// Stable trailing order: walk current Order, assign trailing ranks in order.
	trailing := make([]Todo, 0, len(s.Todos))
	for _, td := range s.Todos {
		if _, ok := rank[td.ID]; !ok {
			trailing = append(trailing, td)
		}
	}
	slices.SortStableFunc(trailing, func(a, b Todo) int { return a.Order - b.Order })
	for _, td := range trailing {
		rank[td.ID] = next
		next++
	}
	for i := range s.Todos {
		s.Todos[i].Order = rank[s.Todos[i].ID]
	}
}

// ClearCompleted removes all done todos.
func (s *State) ClearCompleted() {
	s.Todos = slices.DeleteFunc(s.Todos, func(t Todo) bool { return t.Done })
}

// SetFilter sets the view filter; unknown values fall back to "all".
func (s *State) SetFilter(f string) {
	switch f {
	case "all", "active", "done":
		s.Filter = f
	default:
		s.Filter = "all"
	}
}

// Visible returns the todos for the current filter, sorted by Order.
func (s State) Visible() []Todo {
	out := make([]Todo, 0, len(s.Todos))
	for _, t := range s.Todos {
		switch s.Filter {
		case "active":
			if t.Done {
				continue
			}
		case "done":
			if !t.Done {
				continue
			}
		}
		out = append(out, t)
	}
	slices.SortStableFunc(out, func(a, b Todo) int { return a.Order - b.Order })
	return out
}

// ActiveCount returns the number of not-done todos.
func (s State) ActiveCount() int {
	n := 0
	for _, t := range s.Todos {
		if !t.Done {
			n++
		}
	}
	return n
}

// DoneCount returns the number of done todos.
func (s State) DoneCount() int { return len(s.Todos) - s.ActiveCount() }
