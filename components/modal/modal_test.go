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

	if got := cfg.StateVar(); got != "billingModal1IsOpen" {
		t.Fatalf("StateVar = %q; want %q", got, "billingModal1IsOpen")
	}
}

func TestModalEmptyIDUsesDeterministicFallback(t *testing.T) {
	cfg := Config{}

	if got := cfg.StateVar(); got != "modalIsOpen" {
		t.Fatalf("StateVar = %q; want modalIsOpen", got)
	}
	if got := cfg.TitleID(); got != "modalTitle" {
		t.Fatalf("TitleID = %q; want modalTitle", got)
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
