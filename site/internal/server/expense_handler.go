package server

import (
	"net/http"

	"github.com/araihu/goshtoso/components/toast"
	"github.com/araihu/goshtoso/site/internal/examples/expense"
	"github.com/araihu/goshtoso/site/internal/pages/demo"
	"github.com/araihu/goshtoso/site/internal/pages/demo/components"
	"github.com/araihu/goshtoso/site/internal/pages/demo/examples"
)

// writeExpenseListAndTotal renders the ExpenseList and the SummaryBadge OOB
// fragment, the two outputs shared by every mutation handler. The list is the
// primary swap target (#expense-list); the total is updated out-of-band.
func writeExpenseListAndTotal(r *http.Request, w http.ResponseWriter, st expense.State) {
	_ = examples.ExpenseList(st, false).Render(r.Context(), w)
	_ = examples.SummaryBadge(st, true).Render(r.Context(), w)
}

// persistExpense writes the cookie only when the visitor has not opted out of
// browser storage, mirroring persistTodo. Never persist unconditionally.
func persistExpense(r *http.Request, w http.ResponseWriter, st expense.State) {
	if storageAllowed(r) {
		expense.SetCookie(w, st)
	}
}

// registerExpenseRoutes wires all /api/examples/expense/* endpoints.
func (s *Server) registerExpenseRoutes() {
	s.mux.HandleFunc("/api/examples/expense/add", s.handleExpenseAdd)
	s.mux.HandleFunc("/api/examples/expense/delete", s.handleExpenseDelete)
	s.mux.HandleFunc("/api/examples/expense/restore", s.handleExpenseRestore)
	s.mux.HandleFunc("/api/examples/expense/filter", s.handleExpenseFilter)
	s.mux.HandleFunc("/api/examples/expense/page", s.handleExpensePage)
	s.mux.HandleFunc("/api/examples/expense/clear", s.handleExpenseClear)
}

// renderExpensePage is the first-load handler for /examples/expense. It seeds a
// small starter list on the first visit (when no cookie exists) so the example
// never opens empty, then renders either a full Layout or an HTMX Fragment.
func (s *Server) renderExpensePage(w http.ResponseWriter, r *http.Request) {
	var st expense.State
	if _, err := r.Cookie(expense.CookieName); err != nil && r.URL.Query().Get("seed") != "0" {
		st = expense.Sample()
		persistExpense(r, w, st)
	} else {
		st = expense.FromRequest(r)
	}
	if st.Page < 1 {
		st.Page = 1
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	content := examples.ExpenseApp(st)
	meta := components.MetaForKey("examples/expense")
	if r.Header.Get("HX-Request") == "true" && r.Header.Get("HX-Boosted") != "true" {
		_ = demo.ComponentDocsFragment(meta, "expense", content, storageAllowed(r)).Render(r.Context(), w)
		return
	}
	_ = demo.ComponentDocsLayout(meta, "expense", content, storageAllowed(r)).Render(r.Context(), w)
}

func (s *Server) handleExpenseAdd(w http.ResponseWriter, r *http.Request) {
	if onlyPost(w, r) {
		return
	}
	st := expense.FromRequest(r)
	seqBefore := st.Seq
	desc := r.FormValue("desc")
	st.Add(desc, r.FormValue("amount"), r.FormValue("category"), r.FormValue("date"))
	persistExpense(r, w, st)
	added := st.Seq != seqBefore
	writeHTML(w)
	// The add form is the main swap target (#expense-add-region innerHTML); the
	// list and total ride along out-of-band. On a rejected add we render the form
	// WITH hx-preserve so htmx keeps the live form (values + focus); on success we
	// render it WITHOUT preserve, replacing it with a blank form.
	_ = examples.ExpenseAddForm(!added).Render(r.Context(), w)
	_ = examples.ExpenseList(st, true).Render(r.Context(), w)
	_ = examples.SummaryBadge(st, true).Render(r.Context(), w)
	switch {
	case added:
		_ = toast.OOBToast(toast.Config{Tone: toast.ToneSuccess, Title: "Added", Message: desc}).Render(r.Context(), w)
	default:
		_ = toast.OOBToast(toast.Config{
			Tone:    toast.ToneWarning,
			Title:   "Not added",
			Message: "Enter a description and a valid amount (e.g. 12.34).",
		}).Render(r.Context(), w)
	}
}

func (s *Server) handleExpenseDelete(w http.ResponseWriter, r *http.Request) {
	if onlyPost(w, r) {
		return
	}
	st := expense.FromRequest(r)
	deleted, found := st.Find(idParam(r))
	st.Delete(idParam(r))
	persistExpense(r, w, st)
	writeHTML(w)
	writeExpenseListAndTotal(r, w, st)
	cfg := toast.Config{Tone: toast.ToneInfo, Title: "Expense deleted"}
	if found {
		cfg.Message = deleted.Desc
		cfg.ActionLabel = "Undo"
		cfg.ActionHTMX = &toast.HTMXConfig{
			Post:   examples.RestoreExpenseURL(deleted),
			Target: "#expense-list",
			Swap:   "outerHTML",
		}
	}
	_ = toast.OOBToast(cfg).Render(r.Context(), w)
}

func (s *Server) handleExpenseRestore(w http.ResponseWriter, r *http.Request) {
	if onlyPost(w, r) {
		return
	}
	st := expense.FromRequest(r)
	q := r.URL.Query()
	st.Restore(expense.Expense{
		ID:          intParam(r, "id", 0),
		Desc:        q.Get("desc"),
		AmountCents: intParam(r, "amount", 0),
		Category:    q.Get("category"),
		Date:        q.Get("date"),
		Order:       intParam(r, "order", 0),
	})
	persistExpense(r, w, st)
	writeHTML(w)
	writeExpenseListAndTotal(r, w, st)
}

func (s *Server) handleExpenseFilter(w http.ResponseWriter, r *http.Request) {
	if onlyPost(w, r) {
		return
	}
	st := expense.FromRequest(r)
	st.SetFilter(r.FormValue("search"), r.FormValue("cat"))
	persistExpense(r, w, st)
	writeHTML(w)
	writeExpenseListAndTotal(r, w, st)
}

// handleExpensePage serves pagination links. It is a GET because pagination
// renders <a hx-get> links and paging is idempotent; filters persist in the
// cookie, so only the page number travels in the query string.
func (s *Server) handleExpensePage(w http.ResponseWriter, r *http.Request) {
	st := expense.FromRequest(r)
	st.SetPage(intParam(r, "page", 1))
	persistExpense(r, w, st)
	writeHTML(w)
	writeExpenseListAndTotal(r, w, st)
}

func (s *Server) handleExpenseClear(w http.ResponseWriter, r *http.Request) {
	if onlyPost(w, r) {
		return
	}
	st := expense.FromRequest(r)
	st.Clear()
	persistExpense(r, w, st)
	writeHTML(w)
	writeExpenseListAndTotal(r, w, st)
	_ = toast.OOBToast(toast.Config{Tone: toast.ToneInfo, Title: "Cleared", Message: "All expenses removed."}).Render(r.Context(), w)
}
