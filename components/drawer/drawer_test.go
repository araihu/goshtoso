package drawer

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

func renderDrawer(t *testing.T, cfg Config) string {
	t.Helper()
	var buf bytes.Buffer
	ctx := templ.WithChildren(context.Background(), templ.Raw(`<p>drawer body</p>`))
	if err := Drawer(cfg).Render(ctx, &buf); err != nil {
		t.Fatalf("render Drawer: %v", err)
	}
	return buf.String()
}

func TestDrawerStateVarIsSafeIdentifier(t *testing.T) {
	cfg := Config{ID: "addon-detail.1"}

	if got := cfg.stateVar(); got != "addonDetail1IsOpen" {
		t.Fatalf("stateVar = %q; want addonDetail1IsOpen", got)
	}
}

func TestDrawerEventIDExpressionEscapesQuotes(t *testing.T) {
	cfg := Config{ID: "addon'detail"}

	got := cfg.eventIDLiteral()
	if got != `addon\'detail` {
		t.Fatalf("eventIDLiteral = %q; want escaped literal", got)
	}
}

func TestDrawerRenderDoesNotEmitInvalidStateIdentifier(t *testing.T) {
	html := renderDrawer(t, Config{ID: "addon-detail.1", Title: "Addon"})

	if strings.Contains(html, "addon-detail.1IsOpen") {
		t.Fatalf("rendered invalid Alpine identifier:\n%s", html)
	}
	if !strings.Contains(html, "addonDetail1IsOpen") {
		t.Fatalf("rendered safe Alpine identifier missing:\n%s", html)
	}
}
