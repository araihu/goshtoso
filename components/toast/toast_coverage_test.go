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

func TestCoverageToastToneClasses(t *testing.T) {
	tests := []struct {
		name     string
		tone     Tone
		border   string
		bg       string
		iconBg   string
		title    string
		fallback bool
	}{
		{"info", ToneInfo, "border-info", "bg-info/10", "bg-info/15 text-info", "text-info-text dark:text-info-text-dark", false},
		{"success", ToneSuccess, "border-success", "bg-success/10", "bg-success/15 text-success", "text-success-text dark:text-success-text-dark", false},
		{"warning", ToneWarning, "border-warning", "bg-warning/10", "bg-warning/15 text-warning", "text-warning-text dark:text-warning-text-dark", false},
		{"danger", ToneDanger, "border-danger", "bg-danger/10", "bg-danger/15 text-danger", "text-danger-text dark:text-danger-text-dark", false},
		{"fallback", Tone("custom"), "border-info", "bg-info/10", "bg-info/15 text-info", "text-info-text dark:text-info-text-dark", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{Tone: tt.tone}
			assert.Equal(t, tt.border, cfg.borderClass())
			assert.Equal(t, tt.bg, cfg.bgClass())
			assert.Equal(t, tt.iconBg, cfg.iconBgClass())
			assert.Equal(t, tt.title, cfg.titleClass())
			assert.Equal(t, tt.fallback, cfg.Tone != ToneInfo && cfg.Tone != ToneSuccess && cfg.Tone != ToneWarning && cfg.Tone != ToneDanger)
		})
	}
}

func TestCoverageToastEffectiveDefaultsAndActions(t *testing.T) {
	assert.Equal(t, 8000, Config{}.effectiveDuration())
	assert.Equal(t, 2500, Config{DisplayDuration: 2500}.effectiveDuration())
	assert.Equal(t, 8000, Config{DisplayDuration: -1}.effectiveDuration())
	assert.False(t, Config{}.hasAction())
	assert.True(t, Config{ActionLabel: "Undo"}.hasAction())
	assert.Equal(t, 8000, (MessageConfig{}).effectiveDuration())
	assert.Equal(t, 2500, (MessageConfig{DisplayDuration: 2500}).effectiveDuration())
	assert.Equal(t, "Dismiss", (MessageConfig{}).effectiveDismissLabel())
	assert.Equal(t, "Close", (MessageConfig{DismissLabel: "Close"}).effectiveDismissLabel())

	assert.Equal(t, "toast-container", ContainerConfig{}.effectiveID())
	assert.Equal(t, "custom-toasts", ContainerConfig{ID: "custom-toasts"}.effectiveID())
	assert.Equal(t, 8000, ContainerConfig{}.effectiveDuration())
	assert.Equal(t, 3200, ContainerConfig{DisplayDuration: 3200}.effectiveDuration())
}

func TestCoverageRenderContainerUsesConfiguredIDsAndDuration(t *testing.T) {
	rendered := renderToastComponent(t, ToastContainer(ContainerConfig{
		ID:              "custom-toasts",
		DisplayDuration: 1500,
	}))
	browserHTML := html.UnescapeString(rendered)

	for _, want := range []string{
		`id="custom-toasts"`,
		`id="custom-toasts-oob"`,
		`displayDuration: 1500`,
		`x-on:notify.window="addNotification($event.detail)"`,
		`x-on:toast-dismiss.window="removeNotification($event.detail.id)"`,
		`notification.kind === 'toast' && notification.tone === 'success'`,
		`notification.kind === 'message-toast'`,
		`var kind = data.kind === 'message-toast' ? 'message-toast' : 'toast'`,
		`role="alert"`,
		`Dismiss`,
	} {
		assert.Contains(t, browserHTML, want)
	}
}

func TestCoverageRenderOOBToastWithActionHTMX(t *testing.T) {
	idCounter.Store(0)
	rendered := renderToastComponent(t, OOBToast(Config{
		Tone:            ToneWarning,
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
		`text-warning-text dark:text-warning-text-dark`,
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
	rendered := renderToastComponent(t, MessageToast(MessageConfig{
		Message:         "Can you review this?",
		DisplayDuration: -1,
		Sender: Sender{
			Name:   "Avery",
			Avatar: "/assets/avatar.webp",
		},
	}))
	browserHTML := html.UnescapeString(rendered)

	for _, want := range []string{
		`id="server-toast-1"`,
		`isVisible: true,`,
		`prefers-reduced-motion: reduce`,
		`addEventListener('change'`,
		`removeEventListener('change'`,
		`border-outline bg-surface`,
		`src="/assets/avatar.webp"`,
		`Avery`,
		`Can you review this?`,
		`Dismiss`,
		`toast-dismiss`,
	} {
		assert.Contains(t, browserHTML, want)
	}
	assert.NotContains(t, browserHTML, `>Reply</button>`)
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
