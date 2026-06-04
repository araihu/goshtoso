package rangeinput

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func render(t *testing.T, cfg Config) string {
	t.Helper()
	var buf bytes.Buffer
	if err := Range(cfg).Render(context.Background(), &buf); err != nil {
		t.Fatalf("render range: %v", err)
	}
	return buf.String()
}

func TestRangeDefaults(t *testing.T) {
	html := render(t, Config{ID: "volume", Label: "Volume"})

	for _, want := range []string{
		`id="volume"`,
		`type="range"`,
		`value="0"`,
		`min="0"`,
		`max="100"`,
		`step="1"`,
		`for="volume"`,
		`Volume`,
		`bg-primary`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("expected %q in:\n%s", want, html)
		}
	}
}

func TestRangeConfiguredAttributes(t *testing.T) {
	html := render(t, Config{
		ID:       "brightness",
		Name:     "brightness",
		Value:    "25",
		Min:      "10",
		Max:      "90",
		Step:     "5",
		Disabled: true,
		Required: true,
	})

	for _, want := range []string{
		`name="brightness"`,
		`value="25"`,
		`min="10"`,
		`max="90"`,
		`step="5"`,
		`disabled`,
		`required`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("expected %q in:\n%s", want, html)
		}
	}
}

func TestRangeShowTicks(t *testing.T) {
	html := render(t, Config{ID: "mood", ShowTicks: true})

	if got := strings.Count(html, `<span class=`); got < 21 {
		t.Fatalf("expected default tick row, got %d spans in:\n%s", got, html)
	}
	if !strings.Contains(html, `>0</span>`) || !strings.Contains(html, `>100</span>`) {
		t.Fatalf("expected endpoint labels in:\n%s", html)
	}
	if !strings.Contains(html, `hidden sm:inline`) {
		t.Fatalf("expected responsive tick hiding in:\n%s", html)
	}
}

func TestRangeShowValueUsesAlpine(t *testing.T) {
	html := render(t, Config{ID: "volume", Value: "20", ShowValue: true})

	for _, want := range []string{
		`x-data="{ currentVal: 20 }"`,
		`x-model="currentVal"`,
		`x-text="currentVal"`,
		`>20</span>`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("expected %q in:\n%s", want, html)
		}
	}
}
