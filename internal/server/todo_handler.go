package server

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/araihu/goshtoso/components/toast"
	"github.com/araihu/goshtoso/internal/examples/todo"
	"github.com/araihu/goshtoso/internal/pages/demo"
	"github.com/araihu/goshtoso/internal/pages/demo/examples"
)

// registerTodoRoutes wires all /api/examples/todo/* endpoints.
func (s *Server) registerTodoRoutes() {
	s.mux.HandleFunc("/api/examples/todo/add", s.handleTodoAdd)
	s.mux.HandleFunc("/api/examples/todo/toggle", s.handleTodoToggle)
	s.mux.HandleFunc("/api/examples/todo/delete", s.handleTodoDelete)
	s.mux.HandleFunc("/api/examples/todo/edit", s.handleTodoEdit)
	s.mux.HandleFunc("/api/examples/todo/filter", s.handleTodoFilter)
	s.mux.HandleFunc("/api/examples/todo/move", s.handleTodoMove)
	s.mux.HandleFunc("/api/examples/todo/clear-completed", s.handleTodoClearCompleted)
	s.mux.HandleFunc("/api/examples/todo/reorder", s.handleTodoReorder)
}

// renderTodoPage is the first-load handler for /examples/todo. It reads state
// from the cookie (defaulting Filter to "all" when empty) and renders either a
// full Layout or an HTMX Fragment depending on the HX-Request header.
func (s *Server) renderTodoPage(w http.ResponseWriter, r *http.Request) {
	st := todo.FromRequest(r)
	if st.Filter == "" {
		st.Filter = "all"
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	content := examples.TodoApp(st)
	if r.Header.Get("HX-Request") == "true" && r.Header.Get("HX-Boosted") != "true" {
		_ = demo.Fragment("Todo List", "todo", content).Render(r.Context(), w)
		return
	}
	_ = demo.Layout("Todo List", "todo", content).Render(r.Context(), w)
}

// onlyPost returns true and writes a 405 if the request is not a POST.
func onlyPost(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return true
	}
	return false
}

// idParam reads the ?id= query parameter as an int. Returns 0 on missing/invalid.
func idParam(r *http.Request) int {
	raw := r.URL.Query().Get("id")
	if raw == "" {
		return 0
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0
	}
	return n
}

// parseIDs splits a comma-separated string of ints.
func parseIDs(raw string) []int {
	parts := strings.Split(raw, ",")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		n, err := strconv.Atoi(p)
		if err == nil {
			out = append(out, n)
		}
	}
	return out
}

// moveByButton moves the todo with the given id one step up or down within the
// currently visible list, then reorders the full state to match.
func moveByButton(st *todo.State, id int, dir string) {
	visible := st.Visible()
	// Find position in visible list.
	pos := -1
	for i, t := range visible {
		if t.ID == id {
			pos = i
			break
		}
	}
	if pos < 0 {
		return
	}
	var swapPos int
	switch dir {
	case "up":
		if pos == 0 {
			return
		}
		swapPos = pos - 1
	case "down":
		if pos == len(visible)-1 {
			return
		}
		swapPos = pos + 1
	default:
		return
	}
	// Swap in visible list.
	visible[pos], visible[swapPos] = visible[swapPos], visible[pos]
	// Build the new id order from the swapped visible list, followed by hidden todos.
	ids := make([]int, 0, len(st.Todos))
	for _, t := range visible {
		ids = append(ids, t.ID)
	}
	// Append todos not in the visible set (they stay in relative order after Reorder).
	visibleSet := make(map[int]bool, len(visible))
	for _, t := range visible {
		visibleSet[t.ID] = true
	}
	for _, t := range st.Todos {
		if !visibleSet[t.ID] {
			ids = append(ids, t.ID)
		}
	}
	st.Reorder(ids)
}

// writeHTML sets Content-Type for all todo fragment responses.
func writeHTML(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
}

// --- Handlers ---

func (s *Server) handleTodoAdd(w http.ResponseWriter, r *http.Request) {
	if onlyPost(w, r) {
		return
	}
	st := todo.FromRequest(r)
	if st.Filter == "" {
		st.Filter = "all"
	}
	seqBefore := st.Seq
	_ = r.ParseForm()
	title := r.FormValue("title")
	priority := r.FormValue("priority")
	due := r.FormValue("due")
	st.Add(title, priority, due)
	todo.SetCookie(w, st)
	writeHTML(w)
	_ = examples.TodoList(st).Render(r.Context(), w)
	_ = examples.CountBadge(st.ActiveCount()).Render(r.Context(), w)
	if st.Seq != seqBefore {
		// A todo was actually added.
		msg := title
		if msg == "" {
			msg = "Task added."
		}
		_ = toast.OOBToast(toast.Config{
			Variant: toast.Success,
			Title:   "Added",
			Message: msg,
		}).Render(r.Context(), w)
	}
}

func (s *Server) handleTodoToggle(w http.ResponseWriter, r *http.Request) {
	if onlyPost(w, r) {
		return
	}
	st := todo.FromRequest(r)
	if st.Filter == "" {
		st.Filter = "all"
	}
	st.Toggle(idParam(r))
	todo.SetCookie(w, st)
	writeHTML(w)
	_ = examples.TodoList(st).Render(r.Context(), w)
	_ = examples.CountBadge(st.ActiveCount()).Render(r.Context(), w)
}

func (s *Server) handleTodoDelete(w http.ResponseWriter, r *http.Request) {
	if onlyPost(w, r) {
		return
	}
	st := todo.FromRequest(r)
	if st.Filter == "" {
		st.Filter = "all"
	}
	st.Delete(idParam(r))
	todo.SetCookie(w, st)
	writeHTML(w)
	_ = examples.TodoList(st).Render(r.Context(), w)
	_ = examples.CountBadge(st.ActiveCount()).Render(r.Context(), w)
	_ = toast.OOBToast(toast.Config{
		Variant: toast.Info,
		Title:   "Deleted",
		Message: "Task removed.",
	}).Render(r.Context(), w)
}

func (s *Server) handleTodoEdit(w http.ResponseWriter, r *http.Request) {
	if onlyPost(w, r) {
		return
	}
	st := todo.FromRequest(r)
	if st.Filter == "" {
		st.Filter = "all"
	}
	_ = r.ParseForm()
	st.Edit(idParam(r), r.FormValue("title"), r.FormValue("priority"), r.FormValue("due"))
	todo.SetCookie(w, st)
	writeHTML(w)
	_ = examples.TodoList(st).Render(r.Context(), w)
	_ = examples.CountBadge(st.ActiveCount()).Render(r.Context(), w)
}

func (s *Server) handleTodoFilter(w http.ResponseWriter, r *http.Request) {
	if onlyPost(w, r) {
		return
	}
	st := todo.FromRequest(r)
	if st.Filter == "" {
		st.Filter = "all"
	}
	st.SetFilter(r.URL.Query().Get("f"))
	todo.SetCookie(w, st)
	writeHTML(w)
	_ = examples.TodoList(st).Render(r.Context(), w)
}

func (s *Server) handleTodoMove(w http.ResponseWriter, r *http.Request) {
	if onlyPost(w, r) {
		return
	}
	st := todo.FromRequest(r)
	if st.Filter == "" {
		st.Filter = "all"
	}
	dir := r.URL.Query().Get("dir")
	moveByButton(&st, idParam(r), dir)
	todo.SetCookie(w, st)
	writeHTML(w)
	_ = examples.TodoList(st).Render(r.Context(), w)
}

func (s *Server) handleTodoClearCompleted(w http.ResponseWriter, r *http.Request) {
	if onlyPost(w, r) {
		return
	}
	st := todo.FromRequest(r)
	if st.Filter == "" {
		st.Filter = "all"
	}
	st.ClearCompleted()
	todo.SetCookie(w, st)
	writeHTML(w)
	_ = examples.TodoList(st).Render(r.Context(), w)
	_ = examples.CountBadge(st.ActiveCount()).Render(r.Context(), w)
	_ = toast.OOBToast(toast.Config{
		Variant: toast.Info,
		Title:   "Cleared",
		Message: "Completed tasks removed.",
	}).Render(r.Context(), w)
}

func (s *Server) handleTodoReorder(w http.ResponseWriter, r *http.Request) {
	if onlyPost(w, r) {
		return
	}
	st := todo.FromRequest(r)
	if st.Filter == "" {
		st.Filter = "all"
	}
	_ = r.ParseForm()
	ids := parseIDs(r.FormValue("ids"))
	st.Reorder(ids)
	todo.SetCookie(w, st)
	writeHTML(w)
	_ = examples.CountBadge(st.ActiveCount()).Render(r.Context(), w)
}
