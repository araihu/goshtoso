package server

import (
	"html"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestDemoHandlersEscapeRequestText(t *testing.T) {
	const payload = `<img src=x onerror=alert(1)>`
	s := &Server{}
	for _, tc := range []struct {
		name    string
		path    string
		query   string
		handler http.HandlerFunc
	}{
		{"hello", "/api/hello/" + payload, "", s.handleAPIHello},
		{"accordion path", "/api/components/accordion-content/lazy-content-b/" + payload, "", s.handleAccordionContent},
		{"accordion unknown ID", "/api/components/accordion-content/" + payload, "", s.handleAccordionContent},
		{"radio", "/api/components/radio/echo", url.Values{"value": {payload}}.Encode(), s.handleRadioEcho},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			r.URL.Path = tc.path
			r.URL.RawQuery = tc.query
			w := httptest.NewRecorder()
			tc.handler(w, r)
			body := w.Body.String()
			if strings.Contains(body, payload) || !strings.Contains(body, html.EscapeString(payload)) {
				t.Fatalf("request text must appear as escaped text: %s", body)
			}
		})
	}
}
