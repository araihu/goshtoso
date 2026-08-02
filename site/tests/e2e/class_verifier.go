//go:build e2e

package e2e

import (
	"strings"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// parseClassAttribute parses a class attribute string into individual classes
func parseClassAttribute(classAttr string) []string {
	if classAttr == "" {
		return nil
	}

	// Split by whitespace and filter empty strings
	parts := strings.Fields(classAttr)

	// Remove duplicates while preserving order
	seen := make(map[string]bool)
	var unique []string
	for _, part := range parts {
		if !seen[part] {
			seen[part] = true
			unique = append(unique, part)
		}
	}

	return unique
}

// makeSet creates a set (map) from a slice
func makeSet(items []string) map[string]bool {
	set := make(map[string]bool)
	for _, item := range items {
		set[item] = true
	}
	return set
}

// VerifyTailwindClasses checks that specific Tailwind classes are present
func VerifyTailwindClasses(t *testing.T, locator playwright.Locator, expectedClasses []string) {
	t.Helper()

	classAttr, err := locator.GetAttribute("class")
	require.NoError(t, err)

	actualClasses := parseClassAttribute(classAttr)
	actualSet := makeSet(actualClasses)

	for _, expectedClass := range expectedClasses {
		assert.True(t, actualSet[expectedClass],
			"Expected class '%s' not found. Actual classes: %v",
			expectedClass, actualClasses)
	}
}
