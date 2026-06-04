package textarea

import (
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
