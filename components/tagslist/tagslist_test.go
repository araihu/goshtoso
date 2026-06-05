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

func TestTagsListInitialValuesStayInsideAlpineStrings(t *testing.T) {
	payload := `'); alert(1); ('`
	var buf bytes.Buffer
	err := TagsList(Config{
		ID:     "tags",
		Name:   `tags'); alert(2); ('`,
		Values: []string{payload},
	}).Render(context.Background(), &buf)
	if err != nil {
		t.Fatal(err)
	}

	html := buf.String()
	if strings.Contains(html, `'); alert(1); ('`) || strings.Contains(html, `'); alert(2); ('`) {
		t.Fatalf("rendered unescaped Alpine string payload:\n%s", html)
	}
	for _, want := range []string{`items: [&#39;\&#39;); alert(1); (\&#39;&#39;]`, `name: &#39;tags\&#39;); alert(2); (\&#39;&#39;`} {
		if !strings.Contains(html, want) {
			t.Fatalf("TagsList render missing escaped Alpine payload %q in %s", want, html)
		}
	}
}
