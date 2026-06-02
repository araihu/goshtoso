package profile

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSetCookieThenFromRequest(t *testing.T) {
	rec := httptest.NewRecorder()
	SetCookie(rec, State{Name: "Grace", Bio: "Compiler pioneer."})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	for _, c := range rec.Result().Cookies() {
		req.AddCookie(c)
	}
	got := FromRequest(req)
	if got.Name != "Grace" || got.Bio != "Compiler pioneer." {
		t.Fatalf("round trip via cookie failed: %+v", got)
	}
}

func TestFromRequestNoCookieYieldsZero(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if got := FromRequest(req); got.Name != "" || got.Bio != "" {
		t.Fatalf("expected zero State, got %+v", got)
	}
}

func TestSampleIsNonEmptyAndWithinBudget(t *testing.T) {
	s := Sample()
	if s.Name == "" || s.Bio == "" {
		t.Fatal("Sample should seed a non-empty profile")
	}
	if len(Encode(s)) > maxCookieBytes {
		t.Fatalf("Sample over budget: %d", len(Encode(s)))
	}
}
