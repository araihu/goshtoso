// internal/examples/wizard/state_test.go
package wizard

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEncodeDecodeRoundTrip(t *testing.T) {
	s := WizardState{
		Step:    3,
		Account: Account{Name: "Ada", Email: "ada@example.com", Password: "hunter2hunter"},
		Address: Address{Line1: "1 Analytical Way", City: "London", Country: "UK", Postal: "EC1"},
		Plan:    "pro",
		Done:    false,
	}
	got, err := Decode([]byte(Encode(s)))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if got.Step != 3 || got.Account.Email != "ada@example.com" || got.Address.City != "London" || got.Plan != "pro" {
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
		if got.Step != 0 || got.Plan != "" || got.Done {
			t.Errorf("Decode(%q): expected empty state, got %+v", tc.input, got)
		}
	}
}

func TestNormalizedClampsStep(t *testing.T) {
	cases := []struct {
		in   int
		want int
	}{
		{0, FirstStep}, {-5, FirstStep}, {1, 1}, {4, 4}, {9, LastStep},
	}
	for _, tc := range cases {
		if got := (WizardState{Step: tc.in}).Normalized().Step; got != tc.want {
			t.Errorf("Normalized(step=%d).Step = %d, want %d", tc.in, got, tc.want)
		}
	}
}

func TestCookieRoundTripThroughHTTP(t *testing.T) {
	s := WizardState{Step: 2}
	s.SetAccount("Ada", "ada@example.com", "hunter2hunter")

	rec := httptest.NewRecorder()
	SetCookie(rec, s)

	cookie := rec.Result().Cookies()[0]
	if cookie.Name != CookieName || cookie.Path != "/" || !cookie.HttpOnly {
		t.Fatalf("cookie attrs wrong: %+v", cookie)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(cookie)
	got := FromRequest(req)
	if got.Step != 2 || got.Account.Name != "Ada" {
		t.Fatalf("did not round-trip via http: %+v", got)
	}
}

func TestFromRequestNoCookieIsEmpty(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if got := FromRequest(req); got.Step != 0 || got.Account.Name != "" {
		t.Fatalf("missing cookie should be empty state, got %+v", got)
	}
}

func TestFromRequestCorruptCookieIsEmpty(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: CookieName, Value: strings.Repeat("!", 20)})
	if got := FromRequest(req); got.Step != 0 {
		t.Fatalf("corrupt cookie should be empty state, got %+v", got)
	}
}
