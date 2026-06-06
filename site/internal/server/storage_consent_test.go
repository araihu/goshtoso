package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/araihu/goshtoso/site/internal/examples/expense"
	"github.com/araihu/goshtoso/site/internal/examples/todo"
	"github.com/stretchr/testify/require"
)

func TestRenderTodoPageDoesNotSetDemoCookieWhenStorageDenied(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/examples/todo", nil)
	req.AddCookie(&http.Cookie{Name: "gt_storage", Value: "denied"})
	rec := httptest.NewRecorder()

	s := &Server{}
	s.renderTodoPage(rec, req)

	for _, c := range rec.Result().Cookies() {
		require.NotEqual(t, todo.CookieName, c.Name, "storage opt-out should prevent demo persistence cookies")
	}
}

func TestRenderExpensePageDoesNotSetDemoCookieWhenStorageDenied(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/examples/expense", nil)
	req.AddCookie(&http.Cookie{Name: "gt_storage", Value: "denied"})
	rec := httptest.NewRecorder()

	s := &Server{}
	s.renderExpensePage(rec, req)

	for _, c := range rec.Result().Cookies() {
		require.NotEqual(t, expense.CookieName, c.Name, "storage opt-out should prevent demo persistence cookies")
	}
}
