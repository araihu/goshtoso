package kbd

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

func TestKbdZeroValueRendersSemanticElement(t *testing.T) {
	var buf bytes.Buffer
	if err := Kbd("").Render(context.Background(), &buf); err != nil {
		t.Fatal(err)
	}

	html := buf.String()
	for _, want := range []string{
		"<kbd",
		"rounded-radius",
		"border-outline",
		"bg-surface-alt",
		"dark:bg-surface-dark-alt",
		"min-h-7",
		"text-sm",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("Kbd render missing %q in %s", want, html)
		}
	}
}

func TestKbdSizeClassOverridesDefault(t *testing.T) {
	var buf bytes.Buffer
	if err := Kbd("Esc", WithSize(SizeXS)).Render(context.Background(), &buf); err != nil {
		t.Fatal(err)
	}

	html := buf.String()
	if !strings.Contains(html, "min-h-5") || !strings.Contains(html, "text-[10px]") {
		t.Fatalf("Kbd xs render missing size classes in %s", html)
	}
	if strings.Contains(html, "min-h-7") {
		t.Fatalf("Kbd xs render should not include default size classes: %s", html)
	}
}

func TestKbdAttrsClassAndIconLabel(t *testing.T) {
	var buf bytes.Buffer
	icon := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := w.Write([]byte(`<svg viewBox="0 0 16 16"></svg>`))
		return err
	})

	if err := Kbd("",
		WithLabel("Command"),
		WithIcon(icon),
		WithRootClass("custom-key"),
		WithAttrs(templ.Attributes{"data-shortcut": "command"}),
	).Render(context.Background(), &buf); err != nil {
		t.Fatal(err)
	}

	html := buf.String()
	for _, want := range []string{
		`aria-label="Command"`,
		`data-shortcut="command"`,
		"custom-key",
		"sr-only",
		"Command",
		"size-4",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("Kbd icon render missing %q in %s", want, html)
		}
	}
}

func TestKbdRequiresText(t *testing.T) {
	var buf bytes.Buffer
	if err := Kbd("⌘K", WithSize(SizeSM)).Render(context.Background(), &buf); err != nil {
		t.Fatal(err)
	}

	html := buf.String()
	if !strings.Contains(html, "⌘K") {
		t.Fatalf("Kbd required text missing in %s", html)
	}
	if !strings.Contains(html, "min-h-6") {
		t.Fatalf("Kbd size option missing in %s", html)
	}
}
