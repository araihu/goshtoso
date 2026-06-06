package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/araihu/goshtoso/site/internal/examples/wizard"
	"github.com/stretchr/testify/require"
)

func TestRenderWizardPageDoesNotSetDemoCookieWhenStorageDenied(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/examples/wizard", nil)
	req.AddCookie(&http.Cookie{Name: "gt_storage", Value: "denied"})
	rec := httptest.NewRecorder()

	s := &Server{}
	s.renderWizardPage(rec, req)

	for _, c := range rec.Result().Cookies() {
		require.NotEqual(t, wizard.CookieName, c.Name, "storage opt-out should prevent demo persistence cookies")
	}
}

func TestHandleWizardNextDoesNotSetCookieWhenStorageDenied(t *testing.T) {
	// A valid step-1 submission would normally advance + persist; storage opt-out
	// must suppress the cookie even on the mutating path.
	form := "name=Ada&email=ada%40example.com&password=hunter2hunter"
	req := httptest.NewRequest(http.MethodPost, "/api/examples/wizard/next", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "gt_storage", Value: "denied"})
	rec := httptest.NewRecorder()

	s := &Server{}
	s.handleWizardNext(rec, req)

	for _, c := range rec.Result().Cookies() {
		require.NotEqual(t, wizard.CookieName, c.Name, "storage opt-out should prevent demo persistence cookies on mutation")
	}
}
