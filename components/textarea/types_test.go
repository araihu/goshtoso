package textarea

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestTextareaClassesUseActiveSurfaceVocabulary(t *testing.T) {
	classes := Config{}.TextareaClasses()

	for _, want := range []string{
		"bg-surface",
		"text-on-surface-strong",
		"placeholder:text-on-surface-muted",
		"disabled:opacity-50",
		"disabled:bg-surface-alt",
		"disabled:text-on-surface-muted",
	} {
		if !strings.Contains(classes, want) {
			t.Fatalf("TextareaClasses() missing %q in %q", want, classes)
		}
	}
}

func TestTextareaEscapesUserControlledText(t *testing.T) {
	payload := `<img src=x onerror=alert(1)>`
	var buf bytes.Buffer
	err := Textarea(Config{
		ID:          "comment",
		Name:        "comment",
		Label:       payload,
		Placeholder: payload,
		Value:       payload,
		HelperText:  payload,
	}).Render(context.Background(), &buf)
	if err != nil {
		t.Fatal(err)
	}

	html := buf.String()
	if strings.Contains(html, payload) {
		t.Fatalf("rendered raw payload:\n%s", html)
	}
	for _, want := range []string{
		`&lt;img src=x onerror=alert(1)&gt;`,
		`placeholder="&lt;img src=x onerror=alert(1)&gt;"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("rendered HTML missing escaped payload %q:\n%s", want, html)
		}
	}
}
