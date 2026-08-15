package modal

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func renderModal(t *testing.T, cfg Config) string {
	t.Helper()
	var buf bytes.Buffer
	if err := Modal(cfg).Render(context.Background(), &buf); err != nil {
		t.Fatalf("render Modal: %v", err)
	}
	return buf.String()
}

func TestModalStateVarIsSafeIdentifier(t *testing.T) {
	cfg := Config{ID: "billing-modal.1"}

	if got := cfg.stateVar(); got != "billingModal1IsOpen" {
		t.Fatalf("stateVar = %q; want %q", got, "billingModal1IsOpen")
	}
}

func TestModalEmptyIDUsesDeterministicFallback(t *testing.T) {
	cfg := Config{}

	if got := cfg.stateVar(); got != "modalIsOpen" {
		t.Fatalf("stateVar = %q; want modalIsOpen", got)
	}
	if got := cfg.titleID(); got != "modalTitle" {
		t.Fatalf("titleID = %q; want modalTitle", got)
	}
}

func TestModalRenderDoesNotEmitInvalidStateIdentifier(t *testing.T) {
	html := renderModal(t, Config{
		ID:           "billing-modal.1",
		Title:        "Billing",
		TriggerLabel: "Open",
		PrimaryLabel: "Close",
	})

	if strings.Contains(html, "billing-modal.1IsOpen") {
		t.Fatalf("rendered invalid Alpine identifier:\n%s", html)
	}
	if !strings.Contains(html, "billingModal1IsOpen") {
		t.Fatalf("rendered safe Alpine identifier missing:\n%s", html)
	}
}

func TestModalRenderProvidesReducedMotionTransitionContracts(t *testing.T) {
	htmls := []string{
		renderModal(t, Config{ID: "billing", Title: "Billing", TriggerLabel: "Open"}),
		renderStructuralModal(t, AlertDialog(AlertDialogConfig{ID: "alert", Title: "Alert", TriggerLabel: "Open"})),
	}

	for _, html := range htmls {
		for _, want := range []string{
			`x-transition:enter="transition-opacity ease-out duration-200 motion-reduce:transition-none"`,
			`x-transition:enter-start="opacity-0 motion-reduce:opacity-100"`,
			`x-transition:enter-end="opacity-100"`,
			`x-transition:leave="transition-opacity ease-in duration-150 motion-reduce:transition-none"`,
			`x-transition:leave-start="opacity-100"`,
			`x-transition:leave-end="opacity-0 motion-reduce:opacity-100"`,
			`x-transition:enter="transition ease-out duration-200 delay-100 motion-reduce:transition-none motion-reduce:delay-0"`,
			`x-transition:enter-start="opacity-0 scale-50 motion-reduce:opacity-100 motion-reduce:scale-100"`,
			`x-transition:enter-end="opacity-100 scale-100"`,
			`x-transition:leave="transition ease-in duration-150 motion-reduce:transition-none"`,
			`x-transition:leave-start="opacity-100 scale-100"`,
			`x-transition:leave-end="opacity-0 scale-50 motion-reduce:opacity-100 motion-reduce:scale-100"`,
		} {
			if !strings.Contains(html, want) {
				t.Fatalf("Modal missing reduced-motion transition contract %q in %s", want, html)
			}
		}
	}
}
