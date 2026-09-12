package server

import (
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestApplicationPatternsRedirectsToExamples(t *testing.T) {
	s := &Server{}
	for _, partial := range []bool{false, true} {
		req := httptest.NewRequest(http.MethodGet, "/docs/application-patterns", nil)
		if partial {
			req.Header.Set("HX-Request", "true")
		}
		rec := httptest.NewRecorder()
		s.handleApplicationPatternsPage(rec, req)
		if partial {
			require.Equal(t, http.StatusOK, rec.Code)
			require.Equal(t, "/examples", rec.Header().Get("HX-Redirect"))
		} else {
			require.Equal(t, http.StatusMovedPermanently, rec.Code)
			require.Equal(t, "/examples", rec.Header().Get("Location"))
		}
	}
}
