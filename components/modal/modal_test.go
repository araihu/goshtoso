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
