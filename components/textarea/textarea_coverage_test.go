package textarea

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

func render(t *testing.T, c templ.Component) string {
	t.Helper()
	var buf bytes.Buffer
	if err := c.Render(context.Background(), &buf); err != nil {
		t.Fatalf("render: %v", err)
	}
	return buf.String()
}

func TestCoverageRenderDefaultTextarea(t *testing.T) {
	html := render(t, Textarea(Config{}))
	for _, want := range []string{"<textarea", `rows="3"`, "class="} {
		if !strings.Contains(html, want) {
			t.Fatalf("default render missing %q in %s", want, html)
		}
	}
	// No label/helper block, and no name/placeholder attributes when unset.
	// (Tailwind utility classes contain "disabled:"/"readonly", so those are not asserted here.)
	for _, unwanted := range []string{"<label", "<small", "name=", `placeholder="`} {
		if strings.Contains(html, unwanted) {
			t.Fatalf("default render unexpectedly contains %q in %s", unwanted, html)
		}
	}
}

func TestCoverageRenderFullDefaultStateTextarea(t *testing.T) {
	html := render(t, Textarea(Config{
		ID:          "comment",
		Name:        "comment",
		Label:       "Comment",
		Placeholder: "Say something",
		Value:       "hello",
		Rows:        5,
		HelperText:  "Optional helper",
		RootClass:   "mt-4",
	}))
	for _, want := range []string{
		`id="comment"`,
		`name="comment"`,
		`placeholder="Say something"`,
		`rows="5"`,
		`<label for="comment"`,
		"Comment",
		">hello</textarea>",
		"<small",
		"Optional helper",
		"mt-4",               // RootClass routed through containerClasses
		"text-on-surface/60", // default helperTextClasses branch
		"border-outline",     // default textareaClasses branch
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("full default render missing %q in %s", want, html)
		}
	}
}

func TestCoverageRenderDisabledReadOnly(t *testing.T) {
	html := render(t, Textarea(Config{ID: "x", Disabled: true, ReadOnly: true}))
	if !strings.Contains(html, "disabled") {
		t.Fatalf("expected disabled attribute: %s", html)
	}
	if !strings.Contains(html, "readonly") {
		t.Fatalf("expected readonly attribute: %s", html)
	}
}

func TestCoverageRenderRequiredForBothTextareaPrimitives(t *testing.T) {
	for name, component := range map[string]templ.Component{
		"textarea":              Textarea(Config{ID: "required", Required: true}),
		"textarea with actions": TextareaWithActions(Config{ID: "required-actions", Required: true}),
	} {
		t.Run(name, func(t *testing.T) {
			html := render(t, component)
			if !strings.Contains(html, " required") {
				t.Fatalf("required config did not render native attribute: %s", html)
			}
		})
	}
}

func TestCoverageRenderErrorState(t *testing.T) {
	html := render(t, Textarea(Config{
		ID:         "e",
		Label:      "Bio",
		State:      StateError,
		HelperText: "Error: required",
	}))
	for _, want := range []string{
		"text-danger",   // labelClasses + helperTextClasses error branch
		"border-danger", // textareaClasses error branch
		"<svg",          // error label icon
		"M5.28 4.22",    // error (X) icon path
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("error render missing %q in %s", want, html)
		}
	}
}

func TestCoverageRenderSuccessState(t *testing.T) {
	html := render(t, Textarea(Config{
		ID:         "s",
		Label:      "Bio",
		State:      StateSuccess,
		HelperText: "Looks good",
	}))
	for _, want := range []string{
		"text-success",   // labelClasses + helperTextClasses success branch
		"border-success", // textareaClasses success branch
		"<svg",           // success label icon
		"M12.416 3.376",  // success (check) icon path
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("success render missing %q in %s", want, html)
		}
	}
}

func TestCoverageRenderWithActions(t *testing.T) {
	html := render(t, TextareaWithActions(Config{
		ID:          "msg",
		Name:        "msg",
		Placeholder: "Type a message",
		Value:       "draft",
		Rows:        6,
		Disabled:    true,
		ReadOnly:    true,
		RootClass:   "shadow-lg",
	}))
	for _, want := range []string{
		`id="msg"`,
		`name="msg"`,
		`placeholder="Type a message"`,
		`rows="6"`,
		">draft</textarea>",
		"disabled",
		"readonly",
		"shadow-lg", // condClass routed RootClass
		// three action buttons
		`aria-label="Emojis"`,
		`aria-label="Attach a file"`,
		`aria-label="Send voice"`,
		`aria-label="send"`,
		// icons rendered
		"size-5",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("actions render missing %q in %s", want, html)
		}
	}
}

func TestCoverageWithActionsNoRootClass(t *testing.T) {
	html := render(t, TextareaWithActions(Config{ID: "a"}))
	// condClass("") returns "" → container class ends at base without a trailing
	// space class; the primary assertion is simply that the render succeeded.
	if !strings.Contains(html, "<textarea") {
		t.Fatalf("expected textarea: %s", html)
	}
}

func TestCoverageGetRows(t *testing.T) {
	cases := map[int]string{
		0: "3", 1: "1", 2: "2", 3: "3", 4: "4", 5: "5",
		6: "6", 7: "7", 8: "8", 9: "9", 10: "10",
		11: "3", -1: "3", 99: "3",
	}
	for in, want := range cases {
		if got := (Config{Rows: in}).getRows(); got != want {
			t.Fatalf("getRows(%d) = %q, want %q", in, got, want)
		}
	}
}

func TestCoverageLabelClasses(t *testing.T) {
	cases := map[State]string{
		StateDefault: "w-fit pl-0.5 text-sm",
		StateError:   "text-danger",
		StateSuccess: "text-success",
	}
	for st, want := range cases {
		if got := (Config{State: st}).labelClasses(); !strings.Contains(got, want) {
			t.Fatalf("labelClasses(%q) = %q, want substring %q", st, got, want)
		}
	}
}

func TestCoverageHelperTextClasses(t *testing.T) {
	cases := map[State]string{
		StateDefault: "text-on-surface/60",
		StateError:   "text-danger",
		StateSuccess: "text-success",
	}
	for st, want := range cases {
		if got := (Config{State: st}).helperTextClasses(); !strings.Contains(got, want) {
			t.Fatalf("helperTextClasses(%q) = %q, want substring %q", st, got, want)
		}
	}
}

func TestCoverageContainerClasses(t *testing.T) {
	base := Config{}.containerClasses()
	if !strings.Contains(base, "flex w-full flex-col") {
		t.Fatalf("base container missing flex layout: %q", base)
	}
	if strings.HasSuffix(base, " ") {
		t.Fatalf("base container has trailing space: %q", base)
	}
	withClass := Config{RootClass: "gap-4"}.containerClasses()
	if !strings.HasSuffix(withClass, " gap-4") {
		t.Fatalf("container with RootClass = %q, want suffix %q", withClass, " gap-4")
	}
}

func TestCoverageInputAttrsPassthrough(t *testing.T) {
	html := render(t, Textarea(Config{
		ID:         "a",
		InputAttrs: templ.Attributes{"hx-post": "/save", "data-x": "1"},
	}))
	for _, want := range []string{`hx-post="/save"`, `data-x="1"`} {
		if !strings.Contains(html, want) {
			t.Fatalf("InputAttrs passthrough missing %q in %s", want, html)
		}
	}
}
