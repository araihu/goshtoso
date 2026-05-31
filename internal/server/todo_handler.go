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

// writeListAndCount renders the TodoList and CountBadge OOB fragments, the two
// most common outputs shared by every mutation handler.
func writeListAndCount(r *http.Request, w http.ResponseWriter, st todo.State) {
	_ = examples.TodoList(st).Render(r.Context(), w)
	_ = examples.CountBadge(st.ActiveCount(), true).Render(r.Context(), w)
}

// writeClearButton renders the ClearButton OOB fragment so the button's
// disabled/enabled state tracks the current doneCount on every mutation.
func writeClearButton(r *http.Request, w http.ResponseWriter, st todo.State) {
	_ = examples.ClearButton(st.DoneCount(), true).Render(r.Context(), w)
}

// registerTodoRoutes wires all /api/examples/todo/* endpoints.
func (s *Server) registerTodoRoutes() {
	s.mux.HandleFunc("/api/examples/todo/add", s.handleTodoAdd)
	s.mux.HandleFunc("/api/examples/todo/toggle", s.handleTodoToggle)
	s.mux.HandleFunc("/api/examples/todo/delete", s.handleTodoDelete)
	s.mux.HandleFunc("/api/examples/todo/edit", s.handleTodoEdit)
	s.mux.HandleFunc("/api/examples/todo/filter", s.handleTodoFilter)
	s.mux.HandleFunc("/api/examples/todo/move", s.handleTodoMove)
	s.mux.HandleFunc("/api/examples/todo/clear-completed", s.handleTodoClearCompleted)
	s.mux.HandleFunc("/api/examples/todo/restore", s.handleTodoRestore)
}

// renderTodoPage is the first-load handler for /examples/todo. It reads state
// from the cookie (defaulting Filter to "all" when empty) and renders either a
// full Layout or an HTMX Fragment depending on the HX-Request header.
func (s *Server) renderTodoPage(w http.ResponseWriter, r *http.Request) {
	var st todo.State
	if _, err := r.Cookie(todo.CookieName); err != nil && r.URL.Query().Get("seed") != "0" {
		// First visit (no cookie): seed a small starter list so the example
		// never opens empty, and persist it. `?seed=0` opts out (used by e2e).
		st = todo.Sample()
		todo.SetCookie(w, st)
	} else {
		st = todo.FromRequest(r)
	}
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
	// FormValue parses the body lazily on first access; no explicit ParseForm needed.
	title := strings.TrimSpace(r.FormValue("title"))
	priority := r.FormValue("priority")
	due := r.FormValue("due")
	st.Add(title, priority, due)
	todo.SetCookie(w, st)
	writeHTML(w)
	writeListAndCount(r, w, st)
	writeClearButton(r, w, st)
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
	} else if title != "" {
		// Non-blank title but add was rejected — list is full.
		_ = toast.OOBToast(toast.Config{
			Variant: toast.Warning,
			Title:   "List full",
			Message: "Remove a task before adding more.",
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
	writeListAndCount(r, w, st)
	writeClearButton(r, w, st)
}

func (s *Server) handleTodoDelete(w http.ResponseWriter, r *http.Request) {
	if onlyPost(w, r) {
		return
	}
	st := todo.FromRequest(r)
	if st.Filter == "" {
		st.Filter = "all"
	}
	id := idParam(r)
	// Capture the todo before deleting so we can populate the undo bar.
	var deleted *todo.Todo
	for i := range st.Todos {
		if st.Todos[i].ID == id {
			cp := st.Todos[i]
			deleted = &cp
			break
		}
	}
	st.Delete(id)
	todo.SetCookie(w, st)
	writeHTML(w)
	writeListAndCount(r, w, st)
	writeClearButton(r, w, st)
	cfg := toast.Config{Variant: toast.Info, Title: "Task deleted"}
	if deleted != nil {
		// The toast itself carries the Undo action (no separate undo bar); the
		// toast's own close button handles dismissal.
		cfg.Message = deleted.Title
		cfg.ActionText = "Undo"
		cfg.ActionHxPost = examples.RestoreURL(*deleted)
		cfg.ActionHxTarget = "#todo-list"
		cfg.ActionHxSwap = "outerHTML"
	}
	_ = toast.OOBToast(cfg).Render(r.Context(), w)
}

func (s *Server) handleTodoEdit(w http.ResponseWriter, r *http.Request) {
	if onlyPost(w, r) {
		return
	}
	st := todo.FromRequest(r)
	if st.Filter == "" {
		st.Filter = "all"
	}
	st.Edit(idParam(r), r.FormValue("title"), r.FormValue("priority"), r.FormValue("due"))
	todo.SetCookie(w, st)
	writeHTML(w)
	writeListAndCount(r, w, st)
	writeClearButton(r, w, st)
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
	writeListAndCount(r, w, st)
	writeClearButton(r, w, st)
	_ = toast.OOBToast(toast.Config{
		Variant: toast.Info,
		Title:   "Cleared",
		Message: "Completed tasks removed.",
	}).Render(r.Context(), w)
}

// intParam reads a named query param as an int. Returns def on missing/invalid.
func intParam(r *http.Request, name string, def int) int {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		return def
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	return n
}

func (s *Server) handleTodoRestore(w http.ResponseWriter, r *http.Request) {
	if onlyPost(w, r) {
		return
	}
	st := todo.FromRequest(r)
	if st.Filter == "" {
		st.Filter = "all"
	}
	q := r.URL.Query()
	t := todo.Todo{
		ID:       intParam(r, "id", 0),
		Title:    q.Get("title"),
		Done:     q.Get("done") == "true",
		Priority: q.Get("priority"),
		Due:      q.Get("due"),
		Order:    intParam(r, "order", 0),
	}
	st.Restore(t)
	todo.SetCookie(w, st)
	writeHTML(w)
	writeListAndCount(r, w, st)
	writeClearButton(r, w, st)
}
