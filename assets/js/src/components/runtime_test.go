package componentruntime

import (
	"os"
	"strings"
	"testing"

	"github.com/araihu/goshtoso/internal/jstooling"
)

func TestAuthoredComponentRuntimeSourcesParse(t *testing.T) {
	t.Parallel()

	paths := []string{
		"combobox-client.js",
		"navigation.js",
		"search.js",
		"table.js",
	}
	sources := make(map[string][]byte, len(paths))
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		sources["assets/js/src/components/"+path] = content
	}
	if err := jstooling.ValidateJavaScript(sources); err != nil {
		t.Fatalf("validate component runtime JavaScript: %v", err)
	}
}

func TestAuthoredComponentRuntimePreservesLifecycleAndDataContracts(t *testing.T) {
	t.Parallel()

	search := readRuntimeSource(t, "search.js")
	for _, want := range []string{
		"root.dataset.searchId",
		"root.dataset.searchMaxResults",
		"root.dataset.searchMatchMode",
		"fuzzyScore: function",
		"rankedMatches: function",
		"destroy: function ()",
		"this.requestController.abort()",
	} {
		if !strings.Contains(search, want) {
			t.Fatalf("search runtime missing %q", want)
		}
	}

	table := readRuntimeSource(t, "table.js")
	for _, want := range []string{
		"root.dataset.tableFilterEndpoint",
		"head.dataset.tableSortBy",
		"url.searchParams.set(\"order_by\"",
		"document.removeEventListener(\"htmx:configRequest\"",
		"document.addEventListener(\"htmx:load\"",
		"document.addEventListener(\"htmx:beforeCleanupElement\"",
		"window.goshtosoSafeNavigationTarget",
	} {
		if !strings.Contains(table, want) {
			t.Fatalf("table runtime missing %q", want)
		}
	}

	navigation := readRuntimeSource(t, "navigation.js")
	for _, forbidden := range []string{`protocol === "javascript:"`, `protocol === "data:"`} {
		if strings.Contains(navigation, forbidden) {
			t.Fatalf("navigation runtime permits executable protocol through %q", forbidden)
		}
	}
}

func readRuntimeSource(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(content)
}
