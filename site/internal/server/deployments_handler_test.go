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
	for _, partial := range []bool{false, true} {
		req := httptest.NewRequest(http.MethodGet, "/examples/deployments", nil)
		if partial {
			req.Header.Set("HX-Request-Type", "partial")
		}
		rec := httptest.NewRecorder()
		s.renderDeploymentConsole(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Empty(t, rec.Header().Get("HX-Redirect"))
		require.Contains(t, rec.Body.String(), `id="deployment-console"`)
		require.Contains(t, rec.Body.String(), `id="deployment-navigation"`)
		require.Equal(t, !partial, strings.Contains(strings.ToLower(rec.Body.String()), "<!doctype html>"))
		require.Contains(t, rec.Body.String(), `id="componentdocshell-sidebar-content"`)
	}

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

func TestDeploymentConsoleFragmentMutations(t *testing.T) {
	s := &Server{}
	post := func(values url.Values) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/examples/deployments", strings.NewReader(values.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("HX-Request-Type", "partial")
		rec := httptest.NewRecorder()
		s.renderDeploymentConsole(rec, req)
		require.Empty(t, rec.Header().Get("HX-Redirect"))
		require.Empty(t, rec.Header().Get("Location"))
		require.NotContains(t, strings.ToLower(rec.Body.String()), "<!doctype html>")
		return rec
	}

	invalid := post(url.Values{"action": {"create"}, "service": {"billing"}, "environment": {"Staging"}})
	require.Equal(t, http.StatusUnprocessableEntity, invalid.Code)
	require.Contains(t, invalid.Body.String(), `value="billing"`)
	require.Contains(t, invalid.Body.String(), "Enter a version")
	require.Empty(t, invalid.Header().Get("HX-Push-Url"))

	created := post(url.Values{"action": {"create"}, "service": {"billing"}, "version": {"v1"}, "environment": {"Staging"}})
	require.Equal(t, http.StatusOK, created.Code)
	require.Contains(t, created.Body.String(), "Deployment created")
	target, err := url.Parse(created.Header().Get("HX-Push-Url"))
	require.NoError(t, err)
	require.Equal(t, "detail", target.Query().Get("view"))
	state := deployments.Decode(target.Query().Get("state"))
	require.Len(t, state.Filter("billing"), 1)

	approved := post(url.Values{"action": {"approve"}, "id": {target.Query().Get("id")}, "state": {state.Encode()}})
	require.Equal(t, http.StatusOK, approved.Code)
	require.Contains(t, approved.Body.String(), "Deployment approved")
	target, err = url.Parse(approved.Header().Get("HX-Push-Url"))
	require.NoError(t, err)
	state = deployments.Decode(target.Query().Get("state"))
	require.Equal(t, "Approved", state.Filter("billing")[0].Status)
}
