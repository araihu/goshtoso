// internal/examples/todo/state_test.go
package todo

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEncodeDecodeRoundTrip(t *testing.T) {
	s := State{
		Todos:  []Todo{{ID: 1, Title: "Buy milk", Done: true, Priority: "high", Due: "2026-06-01", Order: 0}},
		Filter: "active",
		Seq:    2,
	}
	got, err := Decode([]byte(Encode(s)))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if len(got.Todos) != 1 || got.Todos[0].Title != "Buy milk" || got.Seq != 2 || got.Filter != "active" {
		t.Fatalf("round-trip mismatch: %+v", got)
	}
}

func TestDecodeMalformedReturnsEmpty(t *testing.T) {
	cases := []struct {
		input   string
		wantErr bool
	}{
		{"", false},             // empty input → zero state, no error
		{"not-base64!!!", true}, // invalid base64 → error + empty
		{"YWJj", true},          // base64("abc") → valid base64 but invalid JSON → error + empty
	}
	for _, tc := range cases {
		got, err := Decode([]byte(tc.input))
		if tc.wantErr && err == nil {
			t.Errorf("Decode(%q): expected error, got nil", tc.input)
		}
		if !tc.wantErr && err != nil {
			t.Errorf("Decode(%q): unexpected error: %v", tc.input, err)
		}
		if len(got.Todos) != 0 || got.Seq != 0 || got.Filter != "" {
			t.Errorf("Decode(%q): expected empty state, got %+v", tc.input, got)
		}
	}
}

func TestCookieRoundTripThroughHTTP(t *testing.T) {
	var s State
	s.Add("ship it", "high", "")

	rec := httptest.NewRecorder()
	SetCookie(rec, s)

	cookie := rec.Result().Cookies()[0]
	if cookie.Name != CookieName || cookie.Path != "/" || !cookie.HttpOnly {
		t.Fatalf("cookie attrs wrong: %+v", cookie)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(cookie)
	got := FromRequest(req)
	if len(got.Todos) != 1 || got.Todos[0].Title != "ship it" {
		t.Fatalf("did not round-trip via http: %+v", got)
	}
}

func TestFromRequestNoCookieIsEmpty(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if got := FromRequest(req); len(got.Todos) != 0 {
		t.Fatalf("missing cookie should be empty state")
	}
}

func TestFromRequestCorruptCookieIsEmpty(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: CookieName, Value: strings.Repeat("!", 20)})
	if got := FromRequest(req); len(got.Todos) != 0 {
		t.Fatalf("corrupt cookie should be empty state")
	}
}
