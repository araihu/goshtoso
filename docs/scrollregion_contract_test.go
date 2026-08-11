package docs_test

import (
	"os"
	"strings"
	"testing"
)

func TestUsageDocumentsScrollRegionPublicContract(t *testing.T) {
	content, err := os.ReadFile("USAGE.md")
	if err != nil {
		t.Fatalf("read USAGE.md: %v", err)
	}
	usage := string(content)
	for _, want := range []string{
		"components/scrollregion",
		"scrollregion.ScrollRegion(scrollregion.Config{",
		"DisableIndicators",
		"keyboard",
		"touch",
	} {
		if !strings.Contains(usage, want) {
			t.Errorf("USAGE.md missing Scroll Region contract %q", want)
		}
	}
}
