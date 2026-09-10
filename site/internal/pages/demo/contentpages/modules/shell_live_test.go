package modulespages

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestShellLiveDocumentAndFragmentResponses(t *testing.T) {
	for _, family := range []string{"componentdocshell", "consoleshell", "landingshell"} {
		t.Run(family, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/modules/app-shells/live/"+family+"/overview", nil)
			response := httptest.NewRecorder()
			ShellLive(response, request)
			require.Equal(t, 200, response.Code)
			require.Contains(t, response.Body.String(), "<!doctype html>")
			require.Equal(t, "noindex", response.Header().Get("X-Robots-Tag"))
			if family != "landingshell" {
				request.Header.Set("HX-Request-Type", "partial")
				response = httptest.NewRecorder()
				ShellLive(response, request)
				require.Equal(t, 200, response.Code)
				require.NotContains(t, response.Body.String(), "<!doctype html>")
				require.Contains(t, response.Body.String(), "hx-swap-oob")
			}
		})
	}
	for _, route := range []string{"unknown/overview", "componentdocshell/unknown", "componentdocshell/overview/extra"} {
		response := httptest.NewRecorder()
		ShellLive(response, httptest.NewRequest("GET", "/modules/app-shells/live/"+route, nil))
		require.Equal(t, 404, response.Code)
	}
}
