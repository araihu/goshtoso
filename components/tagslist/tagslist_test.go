package tagslist

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestTagsListUsesActiveControlVocabulary(t *testing.T) {
	var buf bytes.Buffer
	err := TagsList(Config{
		ID:     "tags",
		Name:   "tags",
		Values: []string{"production"},
	}).Render(context.Background(), &buf)
	if err != nil {
		t.Fatal(err)
	}

	html := buf.String()
	for _, want := range []string{
		"bg-primary/15",
		"border-primary/20",
		"bg-surface",
		"text-on-surface-strong",
		"placeholder:text-on-surface-muted",
		"border-outline-strong",
		"hover:bg-surface-alt",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("TagsList render missing %q in %s", want, html)
		}
	}

	if strings.Contains(html, "border-dashed") {
		t.Fatalf("TagsList Add button should not use dashed inactive styling: %s", html)
	}
}
