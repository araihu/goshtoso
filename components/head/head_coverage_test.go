package head

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/a-h/templ"
	templruntime "github.com/a-h/templ/runtime"
)

// component is the small surface every exported head entry point satisfies.
type component interface {
	Render(context.Context, io.Writer) error
}

func entryPoints() []struct {
	name string
	comp templ.Component
} {
	return []struct {
		name string
		comp templ.Component
	}{
		{"Dependencies", Dependencies()},
		{"DependenciesMinimal", DependenciesMinimal()},
	}
}

// TestRenderCancelledContext exercises the early ctx.Err() guard at the top of
// every generated template: a context that is already cancelled must short
// circuit before any markup is written.
func TestRenderCancelledContext(t *testing.T) {
	for _, ep := range entryPoints() {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		var sb strings.Builder
		err := ep.comp.Render(ctx, &sb)
		if !errors.Is(err, context.Canceled) {
			t.Errorf("%s: want context.Canceled, got %v", ep.name, err)
		}
		if sb.Len() != 0 {
			t.Errorf("%s: cancelled render wrote %d bytes, want 0", ep.name, sb.Len())
		}
	}
}

// failWriter returns an error once more than failAfter bytes have been written.
// It is paired with a 1-byte templ buffer so every generated write flushes
// straight through to the underlying writer, letting a test drive a failure at
// any byte offset in the rendered output.
type failWriter struct {
	failAfter int
	written   int
}

func (w *failWriter) Write(p []byte) (int, error) {
	if w.written+len(p) > w.failAfter {
		return 0, errors.New("write failed")
	}
	w.written += len(p)
	return len(p), nil
}

// tinyBuffer wraps w in a templ runtime buffer backed by a 1-byte bufio writer.
// Because the value is already a *templruntime.Buffer, the generated GetBuffer
// call treats it as an existing buffer (no pooled wrapper), so each WriteString
// flushes immediately and surfaces the underlying writer's error inline.
func tinyBuffer(w io.Writer) *templruntime.Buffer {
	prev := templruntime.DefaultBufferSize
	templruntime.DefaultBufferSize = 1
	defer func() { templruntime.DefaultBufferSize = prev }()

	b := &templruntime.Buffer{}
	b.Reset(w)
	return b
}

// TestRenderWriteErrorPropagates sweeps a forced write failure across the full
// byte range of each template. Every generated "if err != nil { return err }"
// branch after a write is hit at some offset, and a render that is allowed to
// complete must succeed and emit the asset contract.
func TestRenderWriteErrorPropagates(t *testing.T) {
	for _, ep := range entryPoints() {
		// Establish the full rendered length so the sweep covers every write.
		full := renderString(t, ep.comp)
		total := len(full)

		var sawError bool
		for failAfter := 0; failAfter < total; failAfter++ {
			fw := &failWriter{failAfter: failAfter}
			if err := ep.comp.Render(context.Background(), tinyBuffer(fw)); err != nil {
				sawError = true
			}
		}
		if !sawError {
			t.Errorf("%s: forced write failures never surfaced an error", ep.name)
		}

		// A writer that never fails must render the full document successfully.
		fw := &failWriter{failAfter: total + 1}
		buf := tinyBuffer(fw)
		if err := ep.comp.Render(context.Background(), buf); err != nil {
			t.Errorf("%s: unexpected error with non-failing writer: %v", ep.name, err)
		}
		// The generated code leaves this caller-owned buffer unflushed; drain the
		// final residual byte so the byte count reflects the full document.
		if err := buf.Flush(); err != nil {
			t.Errorf("%s: flush: %v", ep.name, err)
		}
		if fw.written != total {
			t.Errorf("%s: wrote %d bytes through tiny buffer, want %d", ep.name, fw.written, total)
		}
	}
}

func renderString(t *testing.T, c component) string {
	t.Helper()
	var sb strings.Builder
	if err := c.Render(context.Background(), &sb); err != nil {
		t.Fatalf("render: %v", err)
	}
	return sb.String()
}

// TestDependenciesEmitsAllRuntimeAssets locks the exact set of script/link tags
// each entry point emits, including the difference between the full and minimal
// variants (collapse + focus plugins only in the full set).
func TestDependenciesEmitsAllRuntimeAssets(t *testing.T) {
	full := renderString(t, Dependencies())
	minimal := renderString(t, DependenciesMinimal())

	shared := []string{
		`<link rel="stylesheet" href="/assets/styles.css">`,
		"/assets/js/runtime/alpinejs/",
		"/assets/js/runtime/htmx.org/",
		`<script defer src="/assets/js/combobox.js"></script>`,
	}
	for _, want := range shared {
		if !strings.Contains(full, want) {
			t.Errorf("Dependencies() missing %q", want)
		}
		if !strings.Contains(minimal, want) {
			t.Errorf("DependenciesMinimal() missing %q", want)
		}
	}

	// Collapse + focus plugins are full-only.
	for _, fullOnly := range []string{
		"/assets/js/runtime/alpinejs-collapse/",
		"/assets/js/runtime/alpinejs-focus/",
	} {
		if !strings.Contains(full, fullOnly) {
			t.Errorf("Dependencies() missing full-only asset %q", fullOnly)
		}
		if strings.Contains(minimal, fullOnly) {
			t.Errorf("DependenciesMinimal() must not emit %q", fullOnly)
		}
	}

	// HTMX must load synchronously (no defer) so hx-* attributes bind on first
	// paint; Alpine core loads deferred.
	if !strings.Contains(full, `<script src="/assets/js/runtime/htmx.org/`) {
		t.Errorf("Dependencies() must load HTMX without defer")
	}
	if !strings.Contains(full, `<script defer src="/assets/js/runtime/alpinejs/`) {
		t.Errorf("Dependencies() must defer Alpine core")
	}
}
