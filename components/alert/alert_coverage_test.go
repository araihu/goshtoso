package alert

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

// render is a helper that renders cfg through the public Alert entry point.
func render(t *testing.T, cfg Config) string {
	t.Helper()
	var buf bytes.Buffer
	if err := Alert(cfg).Render(context.Background(), &buf); err != nil {
		t.Fatalf("render alert: %v", err)
	}
	return buf.String()
}

func mustContainAll(t *testing.T, html string, wants ...string) {
	t.Helper()
	for _, want := range wants {
		if !strings.Contains(html, want) {
			t.Fatalf("missing %q in:\n%s", want, html)
		}
	}
}

func mustNotContain(t *testing.T, html string, unwanted ...string) {
	t.Helper()
	for _, bad := range unwanted {
		if strings.Contains(html, bad) {
			t.Fatalf("unexpected %q in:\n%s", bad, html)
		}
	}
}

// allVariants exercises the default branch (empty Tone) plus each named one.
var allVariants = []Tone{"", ToneInfo, ToneSuccess, ToneWarning, ToneDanger}

func TestDefaultAlertRendersTitleDescriptionAndRole(t *testing.T) {
	html := render(t, Config{Title: "Heads up", Description: "Body text"})
	mustContainAll(t, html,
		`role="alert"`,
		"Heads up",
		"Body text",
		"class=",
	)
	// Default (non-dismissible, no action, no link) must not emit Alpine state.
	mustNotContain(t, html, "x-data", "dismiss alert", "<a ", "hx-")
}

func TestDefaultAlertToneIcons(t *testing.T) {
	// Each tone selects a distinct icon path snippet from iconBadge.
	cases := map[Tone]string{
		ToneInfo:    "M18 10a8 8 0 1 1-16 0",
		ToneSuccess: "M10 18a8 8 0 1 0 0-16",
		ToneWarning: "a.75.75 0 0 1 .75.75v4.5",
		ToneDanger:  "M8.28 7.22a.75.75 0 0 0-1.06 1.06",
	}
	for tone, path := range cases {
		html := render(t, Config{Tone: tone, Title: "T"})
		mustContainAll(t, html, `aria-hidden="true"`, path)
	}
	// Empty tone falls back to the info icon.
	html := render(t, Config{Title: "T"})
	mustContainAll(t, html, "M18 10a8 8 0 1 1-16 0")
}

func TestListContentRendersBulletItems(t *testing.T) {
	html := render(t, Config{
		Title:     "With list",
		ListItems: []string{"first", "second", "third"},
	})
	mustContainAll(t, html, "<ul", "<li>first</li>", "<li>second</li>", "<li>third</li>")
}

func TestListContentOmittedWhenEmpty(t *testing.T) {
	html := render(t, Config{Title: "No list"})
	mustNotContain(t, html, "<ul", "<li>")
}

func TestDismissibleAlertEmitsAlpineState(t *testing.T) {
	html := render(t, Config{Title: "Dismiss me", Dismissible: true})
	mustContainAll(t, html,
		"x-data",
		"alertIsVisible",
		`x-show="alertIsVisible"`,
		`aria-label="dismiss alert"`,
		`@click="alertIsVisible = false"`,
		"x-transition:leave",
	)
}

func TestDismissibleAlertReducedMotionTransitionContract(t *testing.T) {
	html := render(t, Config{Title: "Dismiss me", Dismissible: true})
	mustContainAll(t, html,
		`x-transition:leave="transition ease-in duration-300 motion-reduce:transition-none"`,
		`x-transition:leave-start="opacity-100 scale-100"`,
		`x-transition:leave-end="opacity-0 scale-90 motion-reduce:opacity-100 motion-reduce:scale-100"`,
	)
}

func TestAlertActionControlsHonorReducedMotion(t *testing.T) {
	for _, classes := range []string{
		(Config{}).linkClasses(),
		(Config{}).primaryActionClasses(),
	} {
		if !strings.Contains(classes, "transition motion-reduce:transition-none") {
			t.Fatalf("alert action classes are missing reduced-motion transition suppression: %s", classes)
		}
	}
}

func TestDismissibleTakesPrecedenceOverActionAndLink(t *testing.T) {
	// Dismissible is checked first in Alert(), so action/link must be ignored.
	html := render(t, Config{
		Title:       "Priority",
		Dismissible: true,
		Action:      &ActionConfig{PrimaryLabel: "Go"},
		Link:        &LinkConfig{Label: "Open", Href: "/x"},
	})
	mustContainAll(t, html, "x-data", "dismiss alert")
	mustNotContain(t, html, "Go", `href="/x"`, "Open")
}

func TestLinkAlertRendersAnchor(t *testing.T) {
	html := render(t, Config{
		Title:       "Linked",
		Description: "see more",
		Link:        &LinkConfig{Label: "Learn more", Href: "/docs"},
	})
	mustContainAll(t, html, "<a ", `href="/docs"`, "Learn more")
	mustNotContain(t, html, "x-data")
}

func TestActionAlertDefaultDismissLabel(t *testing.T) {
	html := render(t, Config{
		Title:  "Action",
		Action: &ActionConfig{PrimaryLabel: "Confirm"},
	})
	mustContainAll(t, html, "Confirm", "Dismiss")
}

func TestActionAlertCustomDismissLabel(t *testing.T) {
	html := render(t, Config{
		Title:  "Action",
		Action: &ActionConfig{PrimaryLabel: "Confirm", DismissLabel: "Cancel"},
	})
	mustContainAll(t, html, "Confirm", "Cancel")
}

func TestActionAlertPrimaryOnClick(t *testing.T) {
	html := render(t, Config{
		Title:  "Action",
		Action: &ActionConfig{PrimaryLabel: "Run", PrimaryOnClick: "doThing()"},
	})
	mustContainAll(t, html, "doThing()", `@click=`)
}

func TestActionAlertPrimaryHTMXGet(t *testing.T) {
	html := render(t, Config{
		Title: "Action",
		Action: &ActionConfig{
			PrimaryLabel: "Fetch",
			PrimaryHTMX:  &HTMXConfig{Get: "/api/x", Target: "#out", Swap: "innerHTML"},
		},
	})
	mustContainAll(t, html,
		`hx-get="/api/x"`,
		`hx-target="#out"`,
		`hx-swap="innerHTML"`,
	)
	mustNotContain(t, html, "hx-post")
}

func TestActionAlertPrimaryHTMXPost(t *testing.T) {
	html := render(t, Config{
		Title: "Action",
		Action: &ActionConfig{
			PrimaryLabel: "Save",
			PrimaryHTMX:  &HTMXConfig{Post: "/api/save"},
		},
	})
	mustContainAll(t, html, `hx-post="/api/save"`)
	mustNotContain(t, html, "hx-get")
}

func TestActionAlertNoHTMXWhenNil(t *testing.T) {
	html := render(t, Config{
		Title:  "Action",
		Action: &ActionConfig{PrimaryLabel: "Plain"},
	})
	mustNotContain(t, html, "hx-get", "hx-post", "hx-target", "hx-swap")
}

func TestRootClassAppendedToContainer(t *testing.T) {
	html := render(t, Config{Title: "Custom", RootClass: "my-custom-class"})
	mustContainAll(t, html, "my-custom-class")
}

// --- Pure class-helper coverage across every variant + default branch. ---

func TestContainerClassesVariants(t *testing.T) {
	want := map[Tone]string{
		"":          "border-info",
		ToneInfo:    "border-info",
		ToneSuccess: "border-success",
		ToneWarning: "border-warning",
		ToneDanger:  "border-danger",
	}
	for variant, border := range want {
		cls := Config{Tone: variant}.containerClasses()
		if !strings.Contains(cls, border) {
			t.Fatalf("variant %q: want %q in %q", variant, border, cls)
		}
		if !strings.Contains(cls, "rounded-radius") {
			t.Fatalf("variant %q: missing base classes in %q", variant, cls)
		}
	}
}

func TestContainerClassesRootClass(t *testing.T) {
	cls := Config{Tone: ToneInfo, RootClass: "extra"}.containerClasses()
	if !strings.HasSuffix(cls, " extra") {
		t.Fatalf("RootClass not appended: %q", cls)
	}
	clsNone := Config{Tone: ToneInfo}.containerClasses()
	if strings.Contains(clsNone, "extra") {
		t.Fatalf("unexpected RootClass: %q", clsNone)
	}
}

func TestInnerClassesVariants(t *testing.T) {
	want := map[Tone]string{
		"":          "bg-info/10",
		ToneInfo:    "bg-info/10",
		ToneSuccess: "bg-success/10",
		ToneWarning: "bg-warning/10",
		ToneDanger:  "bg-danger/10",
	}
	for variant, bg := range want {
		cls := Config{Tone: variant}.innerClasses()
		if !strings.Contains(cls, bg) {
			t.Fatalf("variant %q: want %q in %q", variant, bg, cls)
		}
	}
}

func TestIconBadgeClassesVariants(t *testing.T) {
	want := map[Tone]string{
		"":          "text-info",
		ToneInfo:    "text-info",
		ToneSuccess: "text-success",
		ToneWarning: "text-warning",
		ToneDanger:  "text-danger",
	}
	for variant, text := range want {
		cls := Config{Tone: variant}.iconBadgeClasses()
		if !strings.Contains(cls, text) || !strings.Contains(cls, "rounded-full") {
			t.Fatalf("variant %q: want %q in %q", variant, text, cls)
		}
	}
}

func TestTitleClassesVariants(t *testing.T) {
	want := map[Tone]string{
		"":          "text-info-text dark:text-info-text-dark",
		ToneInfo:    "text-info-text dark:text-info-text-dark",
		ToneSuccess: "text-success-text dark:text-success-text-dark",
		ToneWarning: "text-warning-text dark:text-warning-text-dark",
		ToneDanger:  "text-danger-text dark:text-danger-text-dark",
	}
	for variant, text := range want {
		cls := Config{Tone: variant}.titleClasses()
		if !strings.Contains(cls, text) || !strings.Contains(cls, "font-semibold") {
			t.Fatalf("variant %q: want %q in %q", variant, text, cls)
		}
	}
}

func TestLinkClassesVariants(t *testing.T) {
	want := map[Tone]string{
		"":          "text-info-text dark:text-info-text-dark",
		ToneInfo:    "text-info-text dark:text-info-text-dark",
		ToneSuccess: "text-success-text dark:text-success-text-dark",
		ToneWarning: "text-warning-text dark:text-warning-text-dark",
		ToneDanger:  "text-danger-text dark:text-danger-text-dark",
	}
	for variant, text := range want {
		cls := Config{Tone: variant}.linkClasses()
		if !strings.Contains(cls, text) || !strings.Contains(cls, "whitespace-nowrap") {
			t.Fatalf("variant %q: want %q in %q", variant, text, cls)
		}
		if strings.Contains(cls, "hover:opacity") {
			t.Fatalf("variant %q: whole-control hover opacity is not contrast safe: %q", variant, cls)
		}
	}
}

func TestPrimaryActionClassesVariants(t *testing.T) {
	want := map[Tone]string{
		"":          "text-info-text dark:text-info-text-dark",
		ToneInfo:    "text-info-text dark:text-info-text-dark",
		ToneSuccess: "text-success-text dark:text-success-text-dark",
		ToneWarning: "text-warning-text dark:text-warning-text-dark",
		ToneDanger:  "text-danger-text dark:text-danger-text-dark",
	}
	for variant, text := range want {
		cls := Config{Tone: variant}.primaryActionClasses()
		if !strings.Contains(cls, text) || !strings.Contains(cls, "font-semibold") {
			t.Fatalf("variant %q: want %q in %q", variant, text, cls)
		}
		if strings.Contains(cls, "hover:opacity") {
			t.Fatalf("variant %q: whole-control hover opacity is not contrast safe: %q", variant, cls)
		}
	}
}

func TestListClassesDangerOnly(t *testing.T) {
	for _, variant := range allVariants {
		cls := Config{Tone: variant}.listClasses()
		if !strings.Contains(cls, "list-disc") {
			t.Fatalf("variant %q: missing base list classes in %q", variant, cls)
		}
		hasDanger := strings.Contains(cls, "text-danger-text dark:text-danger-text-dark")
		if variant == ToneDanger && !hasDanger {
			t.Fatalf("danger variant missing text-danger: %q", cls)
		}
		if variant != ToneDanger && hasDanger {
			t.Fatalf("variant %q should not have text-danger: %q", variant, cls)
		}
	}
}
