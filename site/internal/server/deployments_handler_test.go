package server

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/araihu/goshtoso/site/internal/examples/deployments"
	"github.com/stretchr/testify/require"
)

func TestDeploymentConsoleRequests(t *testing.T) {
	s := &Server{}
	req := httptest.NewRequest(http.MethodGet, "/examples/deployments", nil)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	s.renderDeploymentConsole(rec, req)
	require.Equal(t, "/examples/deployments", rec.Header().Get("HX-Redirect"))

	for _, tc := range []struct {
		name   string
		values url.Values
		status int
	}{
		{"invalid", url.Values{"action": {"create"}, "service": {"billing"}, "environment": {"Moon"}}, 422},
		{"created", url.Values{"action": {"create"}, "service": {"billing"}, "version": {"v1"}, "environment": {"Staging"}}, 303},
		{"unknown action", url.Values{"action": {"delete"}}, 400},
		{"missing record", url.Values{"action": {"approve"}, "id": {"999"}}, 404},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/examples/deployments", strings.NewReader(tc.values.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rec := httptest.NewRecorder()
			s.renderDeploymentConsole(rec, req)
			require.Equal(t, tc.status, rec.Code)
			if tc.status == 422 {
				require.Contains(t, rec.Body.String(), `value="billing"`)
				require.Contains(t, rec.Body.String(), "Enter a version")
			}
			if tc.status == 303 {
				target, err := url.Parse(rec.Header().Get("Location"))
				require.NoError(t, err)
				require.Equal(t, "detail", target.Query().Get("view"))
				state := deployments.Decode(target.Query().Get("state"))
				require.Len(t, state.Filter("billing"), 1)
			}
		})
	}
}
