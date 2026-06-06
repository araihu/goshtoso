package drawer

import (
	"strings"
	"testing"
)

func TestCoveragePanelClassesCoverSidesWidthsAndCustomClass(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want []string
	}{
		{
			name: "default right medium",
			cfg:  Config{},
			want: []string{"right-0", "border-l", "max-w-[420px]"},
		},
		{
			name: "left small",
			cfg:  Config{Side: SideLeft, Width: WidthSM},
			want: []string{"left-0", "border-r", "max-w-[320px]"},
		},
		{
			name: "large",
			cfg:  Config{Width: WidthLG},
			want: []string{"right-0", "max-w-[560px]"},
		},
		{
			name: "extra large with custom class",
			cfg:  Config{Width: WidthXL, PanelClass: "data-[state=open]:ring-2"},
			want: []string{"max-w-[720px]", "data-[state=open]:ring-2"},
		},
		{
			name: "full",
			cfg:  Config{Width: WidthFull},
			want: []string{"max-w-full", "md:max-w-[90vw]"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cfg.PanelClasses()
			for _, want := range tt.want {
				if !strings.Contains(got, want) {
					t.Fatalf("PanelClasses() = %q; missing %q", got, want)
				}
			}
		})
	}
}

func TestCoverageBodyAndTransitionDefaults(t *testing.T) {
	cfg := Config{ID: "filters"}
	if got := cfg.GetBodyID(); got != "filters-body" {
		t.Fatalf("GetBodyID() = %q; want filters-body", got)
	}
	if got := (Config{BodyID: "custom-body"}).GetBodyID(); got != "custom-body" {
		t.Fatalf("custom GetBodyID() = %q; want custom-body", got)
	}
	if got := (Config{Side: SideLeft}).EnterStart(); got != "-translate-x-full" {
		t.Fatalf("left EnterStart() = %q; want -translate-x-full", got)
	}
	if got := (Config{Side: SideRight}).EnterStart(); got != "translate-x-full" {
		t.Fatalf("right EnterStart() = %q; want translate-x-full", got)
	}
	if got := (Config{}).EnterEnd(); got != "translate-x-0" {
		t.Fatalf("EnterEnd() = %q; want translate-x-0", got)
	}
}

func TestCoverageJSIdentifierAndLiteralEscaping(t *testing.T) {
	identifierCases := map[string]string{
		"":          "drawer",
		"  ":        "drawer",
		"123 menu":  "drawer123Menu",
		"nav item":  "navItem",
		"---drawer": "drawer",
		"snake_id":  "snake_id",
	}
	for raw, want := range identifierCases {
		if got := safeJSIdentifier(raw, "drawer"); got != want {
			t.Fatalf("safeJSIdentifier(%q) = %q; want %q", raw, got, want)
		}
	}

	literal := jsStringSingle("quote' slash\\ line\nreturn\r\u2028\u2029")
	for _, want := range []string{`quote\'`, `slash\\`, `line\n`, `return\r`, `\u2028`, `\u2029`} {
		if !strings.Contains(literal, want) {
			t.Fatalf("jsStringSingle() = %q; missing %q", literal, want)
		}
	}
}

func TestCoverageRenderPersistentDrawerWithCustomBody(t *testing.T) {
	html := renderDrawer(t, Config{
		ID:         "settings-drawer",
		Title:      "Settings",
		BodyID:     "settings-panel-body",
		Persistent: true,
		PanelClass: "ring-2",
		Side:       SideLeft,
		Width:      WidthFull,
	})

	for _, want := range []string{
		`x-data="{settingsDrawerIsOpen: false}"`,
		`aria-labelledby="settings-drawerTitle"`,
		`id="settings-panel-body"`,
		`role="dialog"`,
		`aria-modal="true"`,
		`ring-2`,
		`left-0`,
		`md:max-w-[90vw]`,
		`<p>drawer body</p>`,
		`<svg`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("rendered drawer missing %q:\n%s", want, html)
		}
	}

	for _, forbidden := range []string{`x-on:click="$dispatch('drawer:close-request'`, `x-on:keydown.esc.window`} {
		if strings.Contains(html, forbidden) {
			t.Fatalf("persistent drawer should omit %q:\n%s", forbidden, html)
		}
	}
}
