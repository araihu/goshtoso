package form

import (
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/components/checkbox"
	"github.com/araihu/goshtoso/components/combobox"
	"github.com/araihu/goshtoso/components/fileinput"
	selectfield "github.com/araihu/goshtoso/components/select"
	"github.com/araihu/goshtoso/components/tagslist"
	"github.com/araihu/goshtoso/components/textarea"
	"github.com/araihu/goshtoso/components/textinput"
	"github.com/araihu/goshtoso/components/toggle"
)

func render(t *testing.T, c templ.Component) string {
	t.Helper()
	var buf strings.Builder
	if err := c.Render(context.Background(), &buf); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	return buf.String()
}

func mustContain(t *testing.T, html string, wants ...string) {
	t.Helper()
	for _, w := range wants {
		if !strings.Contains(html, w) {
			t.Fatalf("rendered HTML missing %q in:\n%s", w, html)
		}
	}
}

func mustNotContain(t *testing.T, html string, unwanted ...string) {
	t.Helper()
	for _, w := range unwanted {
		if strings.Contains(html, w) {
			t.Fatalf("rendered HTML unexpectedly contains %q in:\n%s", w, html)
		}
	}
}

// --- Form ---

func TestCoverageRenderDefaultForm(t *testing.T) {
	var buf strings.Builder
	if err := Form(Config{}).Render(context.Background(), &buf); err != nil {
		t.Fatalf("render default form: %v", err)
	}
	html := buf.String()
	mustContain(t, html, "<form", `class="relative "`)
	// default prevents enter submission
	mustContain(t, html, "keydown.enter")
	// no footer, no action, no htmx
	mustNotContain(t, html, "action=", "hx-post", "<button")
}

func TestCoverageFormNativeAction(t *testing.T) {
	html := render(t, Form(Config{
		ID:        "my-form",
		Action:    "/api/submit",
		RootClass: "max-w-xl",
	}))
	mustContain(t, html,
		`id="my-form"`,
		`method="post"`,
		`action="/api/submit"`,
		`class="relative max-w-xl"`,
	)
}

func TestCoverageFormMethodOverride(t *testing.T) {
	html := render(t, Form(Config{Action: "/x", Method: "get"}))
	mustContain(t, html, `method="get"`)
}

func TestCoverageFormAllowEnter(t *testing.T) {
	allow := false
	html := render(t, Form(Config{PreventEnterSubmit: &allow}))
	mustNotContain(t, html, "keydown.enter")
}

func TestCoverageFormPreventEnterExplicitTrue(t *testing.T) {
	prevent := true
	html := render(t, Form(Config{PreventEnterSubmit: &prevent}))
	mustContain(t, html, "keydown.enter")
}

func TestCoverageFormHTMXAllVerbs(t *testing.T) {
	html := render(t, Form(Config{HTMX: &HTMXConfig{
		Post:     "/p",
		Get:      "/g",
		Put:      "/u",
		Delete:   "/d",
		Target:   "#out",
		Swap:     "innerHTML",
		Encoding: "multipart/form-data",
	}}))
	mustContain(t, html,
		`hx-post="/p"`,
		`hx-get="/g"`,
		`hx-put="/u"`,
		`hx-delete="/d"`,
		`hx-target="#out"`,
		`hx-swap="innerHTML"`,
		`hx-encoding="multipart/form-data"`,
	)
}

func TestCoverageFormEmptyHTMXOmitsAttrs(t *testing.T) {
	html := render(t, Form(Config{HTMX: &HTMXConfig{}}))
	mustNotContain(t, html, "hx-post", "hx-get", "hx-put", "hx-delete", "hx-encoding")
}

func TestCoverageFormFooter(t *testing.T) {
	html := render(t, Form(Config{Footer: &FooterConfig{
		SubmitLabel:    "Create",
		CancelLabel:    "Cancel",
		CancelHref:     "/back",
		SubmitDisabled: "!valid",
		Sticky:         true,
	}}))
	mustContain(t, html,
		`type="submit"`,
		"Create",
		"Cancel",
		`href="/back"`,
		`x-bind:disabled="!valid"`,
		"sm:sticky sm:bottom-0",
	)
}

func TestCoverageFormFooterNoSticky(t *testing.T) {
	html := render(t, Form(Config{Footer: &FooterConfig{SubmitLabel: "Save"}}))
	mustContain(t, html, "Save", `type="submit"`)
	mustNotContain(t, html, "sm:sticky sm:bottom-0")
	// no cancel label => no anchor
	mustNotContain(t, html, "<a")
}

func TestCoverageFormFooterCancelHTMX(t *testing.T) {
	html := render(t, Form(Config{Footer: &FooterConfig{
		SubmitLabel: "Save",
		CancelLabel: "Cancel",
		CancelHref:  "/fallback",
		CancelHTMX: &CancelHTMXConfig{
			Get:     "/spa",
			Target:  "#main",
			Swap:    "innerHTML",
			PushURL: true,
		},
	}}))
	mustContain(t, html,
		`hx-get="/spa"`,
		`hx-target="#main"`,
		`hx-swap="innerHTML"`,
		`hx-push-url="true"`,
		// CancelHref still rendered when CancelHTMX present
		`href="/fallback"`,
	)
}

func TestCoverageReducedMotionTransitionContract(t *testing.T) {
	collapsible := render(t, CollapsibleSection(CollapsibleSectionConfig{
		Title: "Advanced",
	}))
	mustContain(t, collapsible,
		"transition-colors motion-reduce:transition-none",
		"transition-transform motion-reduce:transition-none",
	)

	flip := render(t, FlipSection(FlipSectionConfig{Title: "Network"}, textinput.TextInput(textinput.Config{ID: "network", Name: "network"})))
	mustContain(t, flip, "transition motion-reduce:transition-none")

	footer := render(t, Form(Config{Footer: &FooterConfig{
		CancelLabel: "Cancel",
		SubmitLabel: "Submit",
	}}))
	if got := strings.Count(footer, "transition motion-reduce:transition-none"); got != 2 {
		t.Fatalf("form footer should guard both action transitions, got %d in:\n%s", got, footer)
	}
}

// --- Section ---

func TestCoverageSectionDefault(t *testing.T) {
	html := render(t, Section(SectionConfig{Title: "Details"}))
	mustContain(t, html,
		"<h2",
		"Details",
		"md:grid-cols-2",
	)
}

func TestCoverageSectionSingleColumnOOB(t *testing.T) {
	html := render(t, Section(SectionConfig{
		ID:        "sec1",
		RootClass: "extra",
		OOB:       true,
		Columns:   "1",
	}))
	mustContain(t, html,
		`id="sec1"`,
		`hx-swap-oob="true"`,
		"grid grid-cols-1 gap-x-6 gap-y-5",
		"extra",
	)
	// no title => no h2
	mustNotContain(t, html, "<h2")
}

// --- CollapsibleSection ---

func TestCoverageCollapsibleExpanded(t *testing.T) {
	html := render(t, CollapsibleSection(CollapsibleSectionConfig{
		ID: "col1", Title: "Advanced",
		Summary: "Using defaults",
	}))
	mustContain(t, html,
		`id="col1"`,
		`x-data="{ isExpanded: true }"`,
		`aria-controls="col1-content"`,
		`id="col1-content"`,
		"Advanced",
		"Using defaults",
		`role="region"`,
	)
}

func TestCoverageCollapsibleCollapsed(t *testing.T) {
	html := render(t, CollapsibleSection(CollapsibleSectionConfig{
		Title:     "Advanced",
		Collapsed: true,
	}))
	mustContain(t, html, `x-data="{ isExpanded: false }"`)
	// no summary => no summary span text region for it
	mustNotContain(t, html, "Using defaults")
}

// --- FlipSection ---

func TestCoverageFlipSectionReadOnly(t *testing.T) {
	read := textinput.TextInput(textinput.Config{ID: "ro", Name: "ro"})
	html := render(t, FlipSection(FlipSectionConfig{
		ID: "flip1", Title: "Profile",
	}, read))
	mustContain(t, html,
		`id="flip1"`,
		`x-data="{ isEditing: false }"`,
		"Profile",
		// default labels
		"Edit",
		"Done",
		`x-show="!isEditing"`,
		`x-show="isEditing"`,
	)
}

func TestCoverageFlipSectionEditingCustomLabels(t *testing.T) {
	read := textinput.TextInput(textinput.Config{ID: "ro", Name: "ro"})
	html := render(t, FlipSection(FlipSectionConfig{
		Title:     "Profile",
		Flipped:   true,
		EditLabel: "Modify",
		DoneLabel: "Save",
	}, read))
	mustContain(t, html,
		`x-data="{ isEditing: true }"`,
		"Modify",
		"Save",
	)
}

// --- SubSection ---

func TestCoverageSubSection(t *testing.T) {
	html := render(t, SubSection(SubSectionConfig{
		ID:        "sub1",
		Title:     "Nested",
		RootClass: "px-2",
		Columns:   "1",
	}))
	mustContain(t, html,
		`id="sub1"`,
		"<h3",
		"Nested",
		"grid grid-cols-1 gap-x-6 gap-y-5",
		"px-2",
	)
}

func TestCoverageSubSectionNoTitleTwoCol(t *testing.T) {
	html := render(t, SubSection(SubSectionConfig{ID: "sub2"}))
	mustContain(t, html, "md:grid-cols-2")
	mustNotContain(t, html, "<h3")
}

// --- FieldGroup ---

func TestCoverageFieldGroupRequiredWithInput(t *testing.T) {
	html := render(t, FieldGroup(FieldGroupConfig{
		ID:       "email",
		Label:    "Email",
		Required: true,
		Errors:   []string{"is required"},
		Hints:    []string{"we never share it"},
		Input:    &textinput.Config{ID: "email", Name: "email"},
	}))
	mustContain(t, html,
		`for="email-input"`,
		"Email",
		`class="text-danger-text dark:text-danger-text-dark">*`,
		"is required",
		"we never share it",
	)
}

func TestCoverageFieldGroupValidationAndMeta(t *testing.T) {
	html := render(t, FieldGroup(FieldGroupConfig{
		ID:    "host",
		Label: "Host",
		Meta: &FieldMeta{
			FormID:    "f1",
			FieldName: "host",
			DependsOn: "region",
		},
		Validation: &ValidationConfig{
			Endpoint: "/validate",
		},
	}))
	mustContain(t, html,
		`data-goshtoso-form="f1"`,
		`data-goshtoso-field="host"`,
		`data-goshtoso-depends="region"`,
		`hx-post="/validate"`,
		`hx-trigger="change"`,
		`name="host"`,
		`hx-target="this"`,
		`hx-swap="outerHTML"`,
		`hx-include="closest form"`,
	)
}

func TestCoverageFieldGroupValidationCustomTriggerTarget(t *testing.T) {
	html := render(t, FieldGroup(FieldGroupConfig{
		Validation: &ValidationConfig{
			Endpoint: "/validate",
			Target:   "#section",
			Trigger:  "blur",
		},
	}))
	mustContain(t, html, `hx-trigger="blur"`, `hx-target="#section"`)
}

func TestCoverageFieldGroupOOBNoLabel(t *testing.T) {
	html := render(t, FieldGroup(FieldGroupConfig{
		ID:  "fg",
		OOB: true,
	}))
	mustContain(t, html, `hx-swap-oob="true"`)
	mustNotContain(t, html, "<label")
}

func TestCoverageFieldGroupBuiltinTypes(t *testing.T) {
	cases := []struct {
		name string
		cfg  FieldGroupConfig
		want string
	}{
		{"combobox", FieldGroupConfig{Combobox: &combobox.Config{ID: "cb", Name: "cb"}}, `cb`},
		{"select", FieldGroupConfig{Select: &selectfield.Config{ID: "sel", Name: "sel"}}, `sel-trigger`},
		{"textarea", FieldGroupConfig{Textarea: &textarea.Config{ID: "ta", Name: "ta"}}, `<textarea`},
		{"toggle", FieldGroupConfig{Toggle: &toggle.Config{ID: "tg", Label: "Tg"}}, `Tg`},
		{"checkbox", FieldGroupConfig{Checkbox: &checkbox.Config{ID: "ck", Label: "Ck"}}, `Ck`},
		{"tagslist", FieldGroupConfig{TagsList: &tagslist.Config{ID: "tl", Name: "tl"}}, `tl`},
		{"fileinput", FieldGroupConfig{FileInput: &fileinput.Config{ID: "fi", Name: "fi"}}, `fi`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			html := render(t, FieldGroup(tc.cfg))
			mustContain(t, html, tc.want)
		})
	}
}

func TestCoverageFieldGroupChildrenFallback(t *testing.T) {
	// No built-in field type set => renders { children... } (empty here).
	html := render(t, FieldGroup(FieldGroupConfig{ID: "fg", Label: "Custom"}))
	mustContain(t, html, "Custom")
}

// --- FormErrors + helpers ---

func TestCoverageFormErrorsEmptyRendersNothing(t *testing.T) {
	html := render(t, FormErrors(FormErrorsConfig{}))
	if strings.TrimSpace(html) != "" {
		t.Fatalf("expected empty render, got:\n%s", html)
	}
}

func TestCoverageFormErrorsSingleNoPath(t *testing.T) {
	html := render(t, FormErrors(FormErrorsConfig{
		Items: []FormErrorItem{{Message: "Something went wrong"}},
	}))
	mustContain(t, html,
		`id="form-errors"`,
		`role="alert"`,
		"Validation failed",
		"Something went wrong",
	)
	// single message, no path => paragraph form, no <ul>
	mustNotContain(t, html, "<ul")
}

func TestCoverageFormErrorsListWithPathHintCustom(t *testing.T) {
	html := render(t, FormErrors(FormErrorsConfig{
		ID:        "errs",
		Title:     "Install failed",
		Hint:      "Fix the fields",
		RootClass: "mb-4",
		Items: []FormErrorItem{
			{Path: "values.auth.password", Message: "too short"},
			{Message: "needs k8s 1.28+"},
		},
	}))
	mustContain(t, html,
		`id="errs"`,
		"Install failed",
		"Fix the fields",
		"mb-4",
		"<ul",
		"values.auth.password",
		"too short",
		"needs k8s 1.28+",
	)
}

func TestCoverageFormErrorsIDAndTitleDefaults(t *testing.T) {
	cfg := FormErrorsConfig{}
	if got := cfg.getID(); got != "form-errors" {
		t.Fatalf("getID default = %q, want form-errors", got)
	}
	if got := cfg.getTitle(); got != "Validation failed" {
		t.Fatalf("getTitle default = %q, want Validation failed", got)
	}
	custom := FormErrorsConfig{ID: "x", Title: "y"}
	if got := custom.getID(); got != "x" {
		t.Fatalf("getID = %q, want x", got)
	}
	if got := custom.getTitle(); got != "y" {
		t.Fatalf("getTitle = %q, want y", got)
	}
}

// --- pure helper unit tests ---

func TestCoverageShouldPreventEnter(t *testing.T) {
	if !(Config{}).shouldPreventEnter() {
		t.Fatal("default should prevent enter")
	}
	yes := true
	no := false
	if !(Config{PreventEnterSubmit: &yes}).shouldPreventEnter() {
		t.Fatal("explicit true should prevent")
	}
	if (Config{PreventEnterSubmit: &no}).shouldPreventEnter() {
		t.Fatal("explicit false should allow")
	}
}

func TestCoverageGetMethod(t *testing.T) {
	if got := (Config{}).getMethod(); got != "post" {
		t.Fatalf("default method = %q, want post", got)
	}
	if got := (Config{Method: "dialog"}).getMethod(); got != "dialog" {
		t.Fatalf("method = %q, want dialog", got)
	}
}

func TestCoverageGridClasses(t *testing.T) {
	if got := (SectionConfig{Columns: "1"}).gridClasses(); !strings.Contains(got, "grid-cols-1 gap-x-6 gap-y-5") {
		t.Fatalf("section single col = %q", got)
	}
	if got := (SectionConfig{}).gridClasses(); !strings.Contains(got, "md:grid-cols-2") {
		t.Fatalf("section two col = %q", got)
	}
	if got := (SubSectionConfig{Columns: "1"}).gridClasses(); !strings.Contains(got, "grid-cols-1 gap-x-6 gap-y-5") {
		t.Fatalf("subsection single col = %q", got)
	}
	if got := (SubSectionConfig{}).gridClasses(); !strings.Contains(got, "md:grid-cols-2") {
		t.Fatalf("subsection two col = %q", got)
	}
}

func TestCoverageFooterClasses(t *testing.T) {
	plain := (FooterConfig{}).footerClasses()
	if strings.Contains(plain, "sticky") {
		t.Fatalf("plain footer should not be sticky: %q", plain)
	}
	sticky := (FooterConfig{Sticky: true}).footerClasses()
	if !strings.Contains(sticky, "sm:sticky sm:bottom-0") {
		t.Fatalf("sticky footer missing sticky class: %q", sticky)
	}
}

func TestCoverageAlpineDataHelpers(t *testing.T) {
	if got := (CollapsibleSectionConfig{}).alpineData(); got != "{ isExpanded: true }" {
		t.Fatalf("collapsible expanded = %q", got)
	}
	if got := (CollapsibleSectionConfig{Collapsed: true}).alpineData(); got != "{ isExpanded: false }" {
		t.Fatalf("collapsible collapsed = %q", got)
	}
	if got := (FlipSectionConfig{}).alpineData(); got != "{ isEditing: false }" {
		t.Fatalf("flip default = %q", got)
	}
	if got := (FlipSectionConfig{Flipped: true}).alpineData(); got != "{ isEditing: true }" {
		t.Fatalf("flip flipped = %q", got)
	}
}

func TestCoverageFlipLabelDefaults(t *testing.T) {
	if got := (FlipSectionConfig{}).getEditLabel(); got != "Edit" {
		t.Fatalf("edit default = %q", got)
	}
	if got := (FlipSectionConfig{EditLabel: "Modify"}).getEditLabel(); got != "Modify" {
		t.Fatalf("edit custom = %q", got)
	}
	if got := (FlipSectionConfig{}).getDoneLabel(); got != "Done" {
		t.Fatalf("done default = %q", got)
	}
	if got := (FlipSectionConfig{DoneLabel: "Save"}).getDoneLabel(); got != "Save" {
		t.Fatalf("done custom = %q", got)
	}
}

func TestCoverageValidationGetTrigger(t *testing.T) {
	if got := (ValidationConfig{}).getTrigger(); got != "change" {
		t.Fatalf("trigger default = %q", got)
	}
	if got := (ValidationConfig{Trigger: "blur"}).getTrigger(); got != "blur" {
		t.Fatalf("trigger custom = %q", got)
	}
}
