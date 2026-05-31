package chat

import (
	"strings"
	"testing"
)

func TestReply_Greeting(t *testing.T) {
	got, ok := Reply("hello there")
	if !ok {
		t.Fatalf("expected a match for a greeting")
	}
	if got == "" {
		t.Fatalf("matched reply must not be empty")
	}
}

func TestReply_IFeel(t *testing.T) {
	got, ok := Reply("i feel anxious")
	if !ok {
		t.Fatalf("expected a match for 'i feel'")
	}
	want := "Why do you feel anxious?"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestReply_Question(t *testing.T) {
	got, ok := Reply("what is the time?")
	if !ok || got == "" {
		t.Fatalf("expected a canned reply for a question, got %q ok=%v", got, ok)
	}
}

func TestReply_NoMatch(t *testing.T) {
	_, ok := Reply("xyzzy plugh")
	if ok {
		t.Fatalf("expected no match for nonsense input")
	}
}

func TestReply_Deterministic(t *testing.T) {
	a, _ := Reply("i feel sad")
	b, _ := Reply("i feel sad")
	if a != b {
		t.Fatalf("Reply must be deterministic: %q != %q", a, b)
	}
}

func TestReply_EmptyCaptureDoesNotProduceBareTemplate(t *testing.T) {
	// A whitespace-only capture after "i feel" must not produce "Why do you feel ?"
	got, ok := Reply("i feel   ")
	if ok && strings.Contains(got, "feel ?") {
		t.Fatalf("bare-placeholder reply produced for empty capture: %q", got)
	}
}
