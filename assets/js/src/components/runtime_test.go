package componentruntime

import (
	"os"
	"strings"
	"testing"

	"github.com/araihu/goshtoso/internal/jstooling"
)

func TestAuthoredComponentRuntimeSourcesParse(t *testing.T) {
	t.Parallel()

	paths := map[string]string{
		"../action-group.js": "assets/js/src/action-group.js",
		"combobox-client.js": "assets/js/src/components/combobox-client.js",
		"carousel.js":        "assets/js/src/components/carousel.js",
		"dropdown.js":        "assets/js/src/components/dropdown.js",
		"navigation.js":      "assets/js/src/components/navigation.js",
		"popover.js":         "assets/js/src/components/popover.js",
		"search.js":          "assets/js/src/components/search.js",
		"sidebar.js":         "assets/js/src/components/sidebar.js",
		"scroll-region.js":   "assets/js/src/components/scroll-region.js",
		"table.js":           "assets/js/src/components/table.js",
	}
	sources := make(map[string][]byte, len(paths))
	for path, sourcePath := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		sources[sourcePath] = content
	}
	if err := jstooling.ValidateJavaScript(sources); err != nil {
		t.Fatalf("validate component runtime JavaScript: %v", err)
	}
}

func TestActionGroupRuntimeDeclaresObservableLayoutContract(t *testing.T) {
	t.Parallel()

	actionGroup := readRuntimeSource(t, "../action-group.js")
	for _, want := range []string{
		`root.dataset.actionGroupLayoutState = "pending"`,
		`root.dataset.actionGroupLayoutRevision = "0"`,
		`root.dataset.actionGroupLayoutState = state`,
		`root.dataset.actionGroupLayoutRevision = String(layoutRevision)`,
		"layoutRevision += 1",
		"collapsed.every(Boolean)",
		"collapsed.some(Boolean)",
	} {
		if !strings.Contains(actionGroup, want) {
			t.Fatalf("Action Group runtime missing observable layout contract %q", want)
		}
	}
}

func TestPopoverRuntimeRestoresEscapeFocusAfterTrapCleanup(t *testing.T) {
	t.Parallel()

	popover := readRuntimeSource(t, "popover.js")
	for _, want := range []string{
		"focusTriggerIfOwned: function",
		"deferFocusRestoreUntilMenuHidden: function",
		"restoreTriggerAfterMenuHidden: function",
		`window.getComputedStyle(menu).display !== "none"`,
		`new MutationObserver(function ()`,
		`menu.addEventListener("transitionend"`,
		"requestAnimationFrame",
		`attributeFilter: ["class", "hidden", "style"]`,
		"this.isOpen || this.openedWithKeyboard",
		"closingFocus === trigger",
		"active === closingFocus",
		"focusRestoreGeneration: 0",
		"destroyed: false",
		"state.destroyed",
		"state.focusRestoreGeneration !== generation",
		"state.focusTriggerIfOwned(trigger, menu, closingFocus)",
		"state.deferFocusRestoreUntilMenuHidden(trigger, menu, closingFocus, generation)",
		"this.cancelFocusRestore()",
		"trigger.focus()",
	} {
		if !strings.Contains(popover, want) {
			t.Fatalf("Popover runtime missing guarded Escape focus restoration %q", want)
		}
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
		"element.dataset.searchPriority",
		"priority: Number(raw.priority !== undefined ? raw.priority : raw.Priority) || 0",
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
