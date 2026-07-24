package modal

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func render(t *testing.T, cfg Config) string {
	t.Helper()
	var buf bytes.Buffer
	if err := Modal(cfg).Render(context.Background(), &buf); err != nil {
		t.Fatalf("render Modal: %v", err)
	}
	return buf.String()
}

func mustContain(t *testing.T, html string, wants ...string) {
	t.Helper()
	for _, want := range wants {
		if !strings.Contains(html, want) {
			t.Fatalf("rendered HTML missing %q:\n%s", want, html)
		}
	}
}

func mustNotContain(t *testing.T, html string, unwanted ...string) {
	t.Helper()
	for _, bad := range unwanted {
		if strings.Contains(html, bad) {
			t.Fatalf("rendered HTML unexpectedly contains %q:\n%s", bad, html)
		}
	}
}

// TestCoverageDefaultModalStructure covers the default modal render path:
// trigger, dialog roles, header/body/footer, primary button.
func TestCoverageDefaultModalStructure(t *testing.T) {
	html := render(t, Config{
		ID:           "confirm",
		Title:        "Confirm action",
		Body:         "Are you sure?",
		TriggerLabel: "Open dialog",
		PrimaryLabel: "Confirm",
	})

	mustContain(t,
		html,
		`role="dialog"`,
		`aria-modal="true"`,
		`aria-labelledby="confirmTitle"`,
		`id="confirmTitle"`,
		`aria-label="close modal"`,
		"Confirm action",
		"Are you sure?",
		"Open dialog",
		"Confirm",
		"confirmIsOpen",
	)
	// No alert-only icon badge when not in alert mode.
	mustNotContain(t, html, "rounded-full")
}

// TestCoverageSecondaryButtonRendered covers secondaryButton, which is only
// emitted when SecondaryLabel is set in default mode.
func TestCoverageSecondaryButtonRendered(t *testing.T) {
	html := render(t, Config{
		ID:             "delete",
		Title:          "Delete item",
		TriggerLabel:   "Delete",
		PrimaryLabel:   "Yes",
		SecondaryLabel: "Cancel",
	})

	mustContain(t, html, "Cancel", "Yes", "deleteIsOpen = false")
}

// TestCoveragePrimaryButtonHTMXAndOnClick covers every HTMX branch and the
// OnClick branch of buttonClickExpr/primaryButton.
func TestCoveragePrimaryButtonHTMXAndOnClick(t *testing.T) {
	html := render(t, Config{
		ID:           "save",
		Title:        "Save",
		TriggerLabel: "Open",
		PrimaryLabel: "Save now",
		PrimaryAction: &ButtonAction{
			OnClick: "doSave()",
			HTMX: &HTMXConfig{
				Get:    "/api/save",
				Post:   "/api/save-post",
				Target: "#result",
				Swap:   "innerHTML",
			},
		},
	})

	mustContain(t,
		html,
		`hx-get="/api/save"`,
		`hx-post="/api/save-post"`,
		`hx-target="#result"`,
		`hx-swap="innerHTML"`,
		"saveIsOpen = false; doSave()",
	)
}

// TestCoverageSecondaryButtonHTMX covers the HTMX branches on secondaryButton.
func TestCoverageSecondaryButtonHTMX(t *testing.T) {
	html := render(t, Config{
		ID:             "form",
		TriggerLabel:   "Open",
		PrimaryLabel:   "Submit",
		SecondaryLabel: "Reset",
		SecondaryAction: &ButtonAction{
			OnClick: "resetForm()",
			HTMX: &HTMXConfig{
				Get:    "/api/reset",
				Post:   "/api/reset-post",
				Target: "#form",
				Swap:   "outerHTML",
			},
		},
	})

	mustContain(t,
		html,
		`hx-get="/api/reset"`,
		`hx-post="/api/reset-post"`,
		`hx-target="#form"`,
		`hx-swap="outerHTML"`,
		"formIsOpen = false; resetForm()",
	)
}

// TestCoverageAlertDialogTones covers alertDialog, alertIcon (every case),
// iconBadgeClasses, actionClasses, triggerClasses, and alertCTAButton.
func TestCoverageAlertDialogTones(t *testing.T) {
	cases := []struct {
		tone      Tone
		iconColor string
		ctaColor  string
		hasIcon   bool
	}{
		{ToneSuccess, "text-success", "bg-success", true},
		{ToneInfo, "text-info", "bg-info", true},
		{ToneWarning, "text-warning", "bg-warning", true},
		{ToneDanger, "text-danger", "bg-danger", true},
		{ToneDefault, "", "bg-primary", false},
	}

	for _, tc := range cases {
		t.Run(string(tc.tone), func(t *testing.T) {
			html := renderStructuralModal(t, AlertDialog(AlertDialogConfig{
				ID:           "alert",
				Title:        "Heads up",
				Body:         "Something happened",
				TriggerLabel: "Show alert",
				ActionLabel:  "Got it",
				Tone:         tc.tone,
			}))

			// Alert header has an icon badge wrapper.
			mustContain(t, html, "rounded-full", "Got it", "Heads up")
			// Action button carries the tone-specific classes.
			mustContain(t, html, tc.ctaColor)

			if tc.hasIcon {
				// Tone-specific SVG icon present (fill-rule path icons).
				mustContain(t, html, tc.iconColor, "fill-rule")
			} else {
				// Default alert tone: no colored icon, no SVG icon path.
				mustNotContain(t, html, "fill-rule")
			}
		})
	}
}

// TestCoverageAlertCTAButtonHTMX covers HTMX branches on alertCTAButton.
func TestCoverageAlertCTAButtonHTMX(t *testing.T) {
	html := renderStructuralModal(t, AlertDialog(AlertDialogConfig{
		ID:           "confirm",
		Title:        "Confirm",
		TriggerLabel: "Open",
		ActionLabel:  "Proceed",
		Tone:         ToneDanger,
		Action: &ButtonAction{
			OnClick: "proceed()",
			HTMX: &HTMXConfig{
				Get:    "/api/go",
				Post:   "/api/go-post",
				Target: "#out",
				Swap:   "beforeend",
			},
		},
	}))

	mustContain(t,
		html,
		`hx-get="/api/go"`,
		`hx-post="/api/go-post"`,
		`hx-target="#out"`,
		`hx-swap="beforeend"`,
		"confirmIsOpen = false; proceed()",
	)
}

// TestCoveragePanelClassAppended covers the PanelClass branch of DialogClasses.
func TestCoveragePanelClassAppended(t *testing.T) {
	html := render(t, Config{
		ID:           "wide",
		TriggerLabel: "Open",
		PrimaryLabel: "OK",
		PanelClass:   "max-w-3xl custom-panel",
	})

	mustContain(t, html, "custom-panel")
}

// TestCoverageTriggerClassesByPrimitive exercises modal and alert-dialog triggers.
func TestCoverageTriggerClassesByPrimitive(t *testing.T) {
	def := Config{}
	if got := def.TriggerClasses(); !strings.Contains(got, "bg-primary") {
		t.Fatalf("default TriggerClasses missing bg-primary: %q", got)
	}

	tones := map[Tone]string{
		ToneSuccess: "bg-success",
		ToneInfo:    "bg-info",
		ToneWarning: "bg-warning",
		ToneDanger:  "bg-danger",
		ToneDefault: "bg-primary",
	}
	for tone, want := range tones {
		cfg := AlertDialogConfig{Tone: tone}
		if got := cfg.triggerClasses(); !strings.Contains(got, want) {
			t.Fatalf("alert triggerClasses(%s) missing %q: %q", tone, want, got)
		}
	}
}

// TestCoverageHeaderClassesByPrimitive exercises modal and alert-dialog headers.
func TestCoverageHeaderClassesByPrimitive(t *testing.T) {
	alert := (AlertDialogConfig{}).headerClasses()
	def := Config{}.HeaderClasses()
	if alert == def {
		t.Fatalf("alert and default header classes should differ")
	}
	// Default uses p-4; alert uses px-4 py-2.
	if !strings.Contains(def, "p-4") {
		t.Fatalf("default header classes missing p-4: %q", def)
	}
	if !strings.Contains(alert, "py-2") {
		t.Fatalf("alert header classes missing py-2: %q", alert)
	}
}

// TestCoverageIconBadgeAndActionClassesDirect exercises alert-dialog helpers
// across all tones, including the default (uncolored) badge branch.
func TestCoverageIconBadgeAndActionClassesDirect(t *testing.T) {
	if got := (AlertDialogConfig{Tone: ToneDefault}).iconBadgeClasses(); strings.Contains(got, "bg-success") {
		t.Fatalf("default iconBadgeClasses should not be colored: %q", got)
	}
	for tone, want := range map[Tone]string{
		ToneSuccess: "text-success",
		ToneInfo:    "text-info",
		ToneWarning: "text-warning",
		ToneDanger:  "text-danger",
	} {
		if got := (AlertDialogConfig{Tone: tone}).iconBadgeClasses(); !strings.Contains(got, want) {
			t.Fatalf("iconBadgeClasses(%s) missing %q: %q", tone, want, got)
		}
		// Action buttons use the solid bg-<tone> fill, not the text- token.
		ctaWant := strings.Replace(want, "text-", "bg-", 1)
		if got := (AlertDialogConfig{Tone: tone}).actionClasses(); !strings.Contains(got, ctaWant) {
			t.Fatalf("actionClasses(%s) missing %q: %q", tone, ctaWant, got)
		}
	}
	if got := (AlertDialogConfig{Tone: ToneDefault}).actionClasses(); !strings.Contains(got, "bg-primary") {
		t.Fatalf("default actionClasses missing bg-primary: %q", got)
	}
}
