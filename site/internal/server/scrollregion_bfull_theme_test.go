package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// The B-FULL theme axis must be part of the routed response, rather than a
// Playwright-side root mutation or a persisted preference claim. The special
// query is deliberately limited to the ScrollRegion evidence route.
func TestScrollRegionBFullRoutedThemeBindsInitialHTML(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/components/scroll-region?t-gs-011-theme=goshtoso", nil)

	(&Server{}).renderDemo(recorder, request, "components/scroll-region")

	require.Equal(t, http.StatusOK, recorder.Code)
	body := recorder.Body.String()
	require.Contains(t, body, `<html`)
	require.Contains(t, body, `data-theme="goshtoso"`)
	require.Contains(t, body, `data-goshtoso-theme-initial-source="server-routed-html"`)
	require.Contains(t, body, `"theme":"goshtoso"`)
	require.Contains(t, body, `"persistTheme":false`)
	require.NotContains(t, body, `id="site-theme-trigger"`)
}

func TestScrollRegionBFullRoutedThemeRejectsInvalidOrForeignQuery(t *testing.T) {
	for _, test := range []struct {
		name string
		path string
		key  string
	}{
		{name: "unknown theme", path: "/components/scroll-region?t-gs-011-theme=modern", key: "components/scroll-region"},
		{name: "foreign route", path: "/components/button?t-gs-011-theme=goshtoso", key: "components/button"},
		{name: "fragment", path: "/components/scroll-region?t-gs-011-theme=goshtoso", key: "components/scroll-region"},
	} {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			if test.name == "fragment" {
				request.Header.Set("HX-Request", "true")
			}

			(&Server{}).renderDemo(recorder, request, test.key)

			require.Equal(t, http.StatusBadRequest, recorder.Code)
			require.True(t, strings.Contains(recorder.Body.String(), "T-GS-011 routed theme"))
		})
	}
}
