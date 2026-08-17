package accordion

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

func TestAccordionRendersReducedMotionContracts(t *testing.T) {
	var buf bytes.Buffer
	err := Accordion(AccordionConfig{
		ID: "reduced-motion",
		Items: []AccordionItem{
			{Title: "First", Content: templ.Raw("First content")},
		},
	}).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("render accordion: %v", err)
	}

	html := buf.String()
	for _, want := range []string{
		`class="size-5 shrink-0 transition motion-reduce:transition-none"`,
		`class="motion-reduce:transition-none!" x-cloak`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("accordion reduced-motion contract missing %q in:\n%s", want, html)
		}
	}
}
