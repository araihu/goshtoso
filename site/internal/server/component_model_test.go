package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestComponentModelRouteRendersDirectlyAndAsFragment(t *testing.T) {
	s := &Server{}

	for _, tc := range []struct {
		name       string
		hxRequest  bool
		wantMarker string
	}{
		{name: "direct", wantMarker: "<title>Goshtoso Component Model</title>"},
		{name: "htmx fragment", hxRequest: true, wantMarker: `id="component-model-fragment"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/docs/component-model", nil)
			if tc.hxRequest {
				req.Header.Set("HX-Request", "true")
			}
			rec := httptest.NewRecorder()

			s.handleComponentModelPage(rec, req)

			require.Equal(t, http.StatusOK, rec.Code)
			require.Contains(t, rec.Body.String(), tc.wantMarker)
			require.Contains(t, rec.Body.String(), "The Goshtoso Component Model")
		})
	}
}
