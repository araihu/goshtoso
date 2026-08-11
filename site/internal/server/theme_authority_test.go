package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFoundationRoutesUseGoshtosoThemeAsSoleAuthority(t *testing.T) {
	t.Parallel()

	tests := []struct {
		path string
		key  string
	}{
		{path: "/getting-started", key: "getting-started"},
		{path: "/components/button", key: "components/button"},
		{path: "/components/text-input", key: "components/text-input"},
		{path: "/components/table", key: "components/table"},
	}
	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			t.Parallel()
			recorder := httptest.NewRecorder()
			(&Server{}).renderDemo(recorder, httptest.NewRequest(http.MethodGet, test.path, nil), test.key)

			if recorder.Code != http.StatusOK {
				t.Fatalf("GET %s status = %d, want 200", test.path, recorder.Code)
			}
			body := recorder.Body.String()
			if got := strings.Count(body, `href="/assets/styles.css"`); got != 1 {
				t.Errorf("GET %s Goshtoso stylesheet count = %d, want 1", test.path, got)
			}
			if strings.Contains(body, `/componentdocshell/assets/araihu.css`) {
				t.Errorf("GET %s links the historical App Shell theme after Goshtoso styles", test.path)
			}
			if !strings.Contains(body, `"theme":"araihu"`) || !strings.Contains(body, `document.documentElement.setAttribute("data-theme",theme)`) {
				t.Errorf("GET %s does not bootstrap the Arai Hu theme", test.path)
			}
		})
	}
}
