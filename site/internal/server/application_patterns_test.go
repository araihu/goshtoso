package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApplicationPatternsRouteRendersDirectlyAndAsFragment(t *testing.T) {
	s := &Server{}

	for _, tc := range []struct {
		name       string
		hxRequest  bool
		wantMarker string
	}{
		{name: "direct", wantMarker: "<title>Application Patterns for Goshtoso</title>"},
		{name: "htmx fragment", hxRequest: true, wantMarker: `id="application-patterns-fragment"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/docs/application-patterns", nil)
			if tc.hxRequest {
				req.Header.Set("HX-Request", "true")
			}
			rec := httptest.NewRecorder()

			s.handleApplicationPatternsPage(rec, req)

			require.Equal(t, http.StatusOK, rec.Code)
			require.Equal(t, "text/html; charset=utf-8", rec.Header().Get("Content-Type"))
			require.Contains(t, rec.Body.String(), tc.wantMarker)
			require.Contains(t, rec.Body.String(), "Compose product surfaces, not component piles")
			require.Contains(t, rec.Body.String(), "Multi-step Workflow")
		})
	}
}
