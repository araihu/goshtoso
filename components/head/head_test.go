package head

import (
	"context"
	"io"
	"strings"
	"testing"
)

func render(t *testing.T, c interface {
	Render(context.Context, io.Writer) error
}) string {
	t.Helper()
	var sb strings.Builder
	if err := c.Render(context.Background(), &sb); err != nil {
		t.Fatalf("render: %v", err)
	}
	return sb.String()
}

func TestDependenciesUsesVersionedPaths(t *testing.T) {
	out := render(t, Dependencies())
	for _, want := range []string{
		"/assets/js/runtime/alpinejs/3.14.9/alpine.min.js",
		"/assets/js/runtime/alpinejs-collapse/3.14.9/alpine-collapse.min.js",
		"/assets/js/runtime/alpinejs-focus/3.14.9/alpine-focus.min.js",
		"/assets/js/runtime/htmx.org/2.0.8/htmx.min.js",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("Dependencies() missing versioned path %q\n%s", want, out)
		}
	}
	// Hard cut: no /vendor/ segment. Go module zips omit directories named vendor.
	for _, bad := range []string{`/assets/js/vendor/`, `vendor/alpine.min.js`, `vendor/htmx.min.js`} {
		if strings.Contains(out, bad) {
			t.Errorf("Dependencies() still emits flat path %q", bad)
		}
	}
}

func TestDependenciesMinimalUsesVersionedPaths(t *testing.T) {
	out := render(t, DependenciesMinimal())
	for _, want := range []string{
		"/assets/js/runtime/alpinejs/3.14.9/alpine.min.js",
		"/assets/js/runtime/htmx.org/2.0.8/htmx.min.js",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("DependenciesMinimal() missing %q\n%s", want, out)
		}
	}
	if strings.Contains(out, "/assets/js/vendor/") {
		t.Errorf("DependenciesMinimal() must avoid /vendor/ paths: %s", out)
	}
}
