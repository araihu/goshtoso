package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestComponentsRootRedirectsToAccordion(t *testing.T) {
	s := New(".")
	req := httptest.NewRequest(http.MethodGet, "/components/", nil)
	rec := httptest.NewRecorder()

	s.ServeHTTP(rec, req)

	require.Equal(t, http.StatusMovedPermanently, rec.Code)
	require.Equal(t, "/components/accordion", rec.Header().Get("Location"))
}
