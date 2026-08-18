package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExamplesRootRedirectsToTicker(t *testing.T) {
	s := &Server{}
	req := httptest.NewRequest(http.MethodGet, "/examples?seed=0", nil)
	rec := httptest.NewRecorder()

	s.handleExample(rec, req)

	require.Equal(t, http.StatusMovedPermanently, rec.Code)
	require.Equal(t, "/examples/ticker?seed=0", rec.Header().Get("Location"))
}
