package selectfield

import (
	"strings"
	"testing"
)

func TestTriggerClassesUseActiveSurfaceVocabulary(t *testing.T) {
	classes := Config{}.TriggerClasses()

	for _, want := range []string{
		"bg-surface",
		"text-on-surface-strong",
		"disabled:opacity-50",
	} {
		if !strings.Contains(classes, want) {
			t.Fatalf("TriggerClasses() missing %q in %q", want, classes)
		}
	}
}

func TestDisabledTriggerClassesUseDisabledVocabulary(t *testing.T) {
	classes := Config{Disabled: true}.TriggerClasses()

	for _, want := range []string{
		"bg-surface-alt",
		"text-on-surface-muted",
		"opacity-50",
		"cursor-not-allowed",
	} {
		if !strings.Contains(classes, want) {
			t.Fatalf("TriggerClasses() missing %q in %q", want, classes)
		}
	}
}
