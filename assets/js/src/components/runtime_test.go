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
		"popover.js",
		"search.js",
		"sidebar.js",
		"scroll-region.js",
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

func TestPopoverRuntimePreservesLifecycleContract(t *testing.T) {
	source := readRuntimeSource(t, "popover.js")
	for _, want := range []string{
		"window.goshtosoPopover",
		"data-popover-trigger",
		"aria-controls",
		"openFromKeyboard: function",
		"closeAndFocus: function",
		"clearTimeout(this.leaveTimeout)",
		"destroy: function ()",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("popover runtime missing %q", want)
		}
	}
}

func TestDropdownRuntimeDelegatesToPopover(t *testing.T) {
	source := readRuntimeSource(t, "dropdown.js")
	for _, want := range []string{
		"window.goshtosoPopover",
		"window.goshtosoDropdown",
		"goshtosoPopover",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("dropdown runtime missing %q", want)
		}
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

	sidebar := readRuntimeSource(t, "sidebar.js")
	for _, want := range []string{
		"window.goshtosoSidebarOverlay =",
		`window.Alpine.data("goshtosoSidebarOverlay", window.goshtosoSidebarOverlay)`,
		"open: false",
		"closeAndFocus: function",
		"this.open = false",
		"this.$refs.trigger.focus()",
	} {
		if !strings.Contains(sidebar, want) {
			t.Fatalf("sidebar runtime missing %q", want)
		}
	}

	scrollRegion := readRuntimeSource(t, "scroll-region.js")
	for _, want := range []string{
		"resizeObserver.observe(viewport)",
		"resizeObserver.observe(child)",
		"mutationObserver.observe(viewport",
		`document.addEventListener("htmx:beforeCleanupElement"`,
		"nestedState.disconnect()",
	} {
		if !strings.Contains(scrollRegion, want) {
			t.Fatalf("scroll-region runtime missing %q", want)
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
