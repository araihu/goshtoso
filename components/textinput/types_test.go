package textinput

import (
	"strings"
	"testing"
)

func TestInputClassesUseActiveSurfaceVocabulary(t *testing.T) {
	classes := Config{}.inputClasses()

	for _, want := range []string{
		"bg-surface",
		"text-on-surface-strong",
		"placeholder:text-on-surface-muted",
		"disabled:opacity-50",
		"disabled:bg-surface-alt",
		"disabled:text-on-surface-muted",
	} {
		if !strings.Contains(classes, want) {
			t.Fatalf("inputClasses() missing %q in %q", want, classes)
		}
	}
}

func TestSearchInputClassesUseActiveSurfaceVocabulary(t *testing.T) {
	classes := searchInputClasses(Config{})

	for _, want := range []string{
		"bg-surface",
		"text-on-surface-strong",
		"placeholder:text-on-surface-muted",
		"disabled:opacity-50",
		"disabled:bg-surface-alt",
		"disabled:text-on-surface-muted",
	} {
		if !strings.Contains(classes, want) {
			t.Fatalf("searchInputClasses() missing %q in %q", want, classes)
		}
	}
}
