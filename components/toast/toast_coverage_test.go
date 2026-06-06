package toast

import (
	"bytes"
	"context"
	"html"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCoverageToastVariantClasses(t *testing.T) {
	tests := []struct {
		name     string
		variant  Variant
		border   string
		bg       string
		iconBg   string
		title    string
		message  bool
		fallback bool
	}{
		{"info", Info, "border-info", "bg-info/10", "bg-info/15 text-info", "text-info", false, false},
		{"success", Success, "border-success", "bg-success/10", "bg-success/15 text-success", "text-success", false, false},
		{"warning", Warning, "border-warning", "bg-warning/10", "bg-warning/15 text-warning", "text-warning", false, false},
		{"danger", Danger, "border-danger", "bg-danger/10", "bg-danger/15 text-danger", "text-danger", false, false},
		{"message", Message, "border-outline dark:border-outline-dark", "bg-surface-alt dark:bg-surface-dark-alt", "bg-info/15 text-info", "text-info", true, false},
		{"fallback", Variant("custom"), "border-info", "bg-info/10", "bg-info/15 text-info", "text-info", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{Variant: tt.variant}
			assert.Equal(t, tt.border, cfg.BorderClass())
			assert.Equal(t, tt.bg, cfg.BgClass())
			assert.Equal(t, tt.iconBg, cfg.IconBgClass())
			assert.Equal(t, tt.title, cfg.TitleClass())
			assert.Equal(t, tt.message, cfg.Variant == Message)
			assert.Equal(t, tt.fallback, cfg.Variant != Info && cfg.Variant != Success && cfg.Variant != Warning && cfg.Variant != Danger && cfg.Variant != Message)
		})
	}
}

func TestCoverageToastEffectiveDefaultsAndActions(t *testing.T) {
	assert.Equal(t, 8000, Config{}.effectiveDuration())
	assert.Equal(t, 2500, Config{DisplayDuration: 2500}.effectiveDuration())
	assert.Equal(t, 8000, Config{DisplayDuration: -1}.effectiveDuration())
	assert.False(t, Config{}.HasAction())
	assert.True(t, Config{ActionLabel: "Undo"}.HasAction())

	assert.Equal(t, "toast-container", ContainerConfig{}.effectiveID())
	assert.Equal(t, "custom-toasts", ContainerConfig{ID: "custom-toasts"}.effectiveID())
	assert.Equal(t, 8000, ContainerConfig{}.effectiveDuration())
	assert.Equal(t, 3200, ContainerConfig{DisplayDuration: 3200}.effectiveDuration())
}

func TestCoverageRenderContainerUsesConfiguredIDsAndDuration(t *testing.T) {
	rendered := renderToastComponent(t, Container(ContainerConfig{
		ID:              "custom-toasts",
		DisplayDuration: 1500,
	}))
	browserHTML := html.UnescapeString(rendered)

	for _, want := range []string{
		`id="custom-toasts"`,
		`id="custom-toasts-oob"`,
		`displayDuration: 1500`,
		`x-on:notify.window="addNotification`,
		`x-on:toast-dismiss.window="removeNotification($event.detail.id)"`,
		`x-if="notification.variant === 'message'"`,
		`role="alert"`,
		`Reply`,
		`Dismiss`,
	} {
		assert.Contains(t, browserHTML, want)
	}
}

func TestCoverageRenderOOBToastWithActionHTMX(t *testing.T) {
	idCounter.Store(0)
	rendered := renderToastComponent(t, OOBToast(Config{
		Variant:         Warning,
		Title:           "Storage low",
		Message:         "Upgrade soon.",
		DisplayDuration: 2500,
		ActionLabel:     "Upgrade",
		ActionHTMX: &HTMXConfig{
			Post:   "/api/upgrade",
			Target: "#result",
			Swap:   "outerHTML",
		},
	}))
	browserHTML := html.UnescapeString(rendered)

	for _, want := range []string{
		`id="toast-container-oob" hx-swap-oob="beforeend"`,
		`id="server-toast-1"`,
		`data-toast-id="server-toast-1"`,
		`x-data="{`,
		`}, 2500);`,
		`border-warning`,
		`bg-warning/10`,
		`text-warning`,
		`Storage low`,
		`Upgrade soon.`,
		`hx-post="/api/upgrade"`,
		`hx-target="#result"`,
		`hx-swap="outerHTML"`,
		`>Upgrade</button>`,
	} {
		assert.Contains(t, browserHTML, want)
	}
}

func TestCoverageRenderPersistentMessageToast(t *testing.T) {
	idCounter.Store(0)
	rendered := renderToastComponent(t, Toast(Config{
		Variant:         Message,
		Message:         "Can you review this?",
		DisplayDuration: -1,
		Sender: &Sender{
			Name:   "Avery",
			Avatar: "/assets/avatar.webp",
		},
	}))
	browserHTML := html.UnescapeString(rendered)

	for _, want := range []string{
		`id="server-toast-1"`,
		`x-data="{ isVisible: true }"`,
		`border-outline bg-surface`,
		`src="/assets/avatar.webp"`,
		`Avery`,
		`Can you review this?`,
		`Reply`,
		`Dismiss`,
		`toast-dismiss`,
	} {
		assert.Contains(t, browserHTML, want)
	}
	assert.NotContains(t, browserHTML, `setTimeout(() => { this.isVisible = false`)
}

func TestCoverageJSEscapeSingleEscapesOnlyScriptStringDelimiters(t *testing.T) {
	got := jsEscapeSingle(`a'b\c "quoted"`)
	assert.Equal(t, `a\'b\\c "quoted"`, got)
	assert.True(t, strings.Contains(got, `"quoted"`))
}

func renderToastComponent(t *testing.T, component templ.Component) string {
	t.Helper()

	var buf bytes.Buffer
	require.NoError(t, component.Render(context.Background(), &buf))
	return buf.String()
}
