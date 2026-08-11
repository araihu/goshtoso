package banner

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCoverageContainerClassesCoverPositionVariantsAndRootClass(t *testing.T) {
	tests := []struct {
		name      string
		cfg       Config
		wantParts []string
	}{
		{
			name: "default relative",
			cfg:  Config{},
			wantParts: []string{
				"flex w-full p-4",
				"border-b border-outline bg-surface-alt",
			},
		},
		{
			name: "fixed primary with root class",
			cfg: Config{
				Position:  PositionFixed,
				Tone:      TonePrimary,
				RootClass: "shadow-lg",
			},
			wantParts: []string{
				"fixed inset-x-0 top-0 z-50",
				"border-b border-primary bg-primary/10",
				"shadow-lg",
			},
		},
		{
			name:      "info",
			cfg:       Config{Tone: ToneInfo},
			wantParts: []string{"border-b border-info bg-info/10"},
		},
		{
			name:      "success",
			cfg:       Config{Tone: ToneSuccess},
			wantParts: []string{"border-b border-success bg-success/10"},
		},
		{
			name:      "warning",
			cfg:       Config{Tone: ToneWarning},
			wantParts: []string{"border-b border-warning bg-warning/10"},
		},
		{
			name:      "danger",
			cfg:       Config{Tone: ToneDanger},
			wantParts: []string{"border-b border-danger bg-danger/10"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			classes := tt.cfg.containerClasses()
			for _, want := range tt.wantParts {
				assert.Contains(t, classes, want)
			}
		})
	}
}

func TestCoverageLinkClassesCoverEveryTone(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want string
	}{
		{name: "default", cfg: Config{}, want: "text-primary"},
		{name: "primary", cfg: Config{Tone: TonePrimary}, want: "text-primary"},
		{name: "info", cfg: Config{Tone: ToneInfo}, want: "text-info-text dark:text-info-text-dark"},
		{name: "success", cfg: Config{Tone: ToneSuccess}, want: "text-success-text dark:text-success-text-dark"},
		{name: "warning", cfg: Config{Tone: ToneWarning}, want: "text-warning-text dark:text-warning-text-dark"},
		{name: "danger", cfg: Config{Tone: ToneDanger}, want: "text-danger-text dark:text-danger-text-dark"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			classes := tt.cfg.linkClasses()
			assert.Contains(t, classes, "font-medium")
			assert.Contains(t, classes, "hover:underline")
			assert.Contains(t, classes, tt.want)
		})
	}
}

func TestCoverageRenderSimpleBannerBranches(t *testing.T) {
	t.Run("default dismiss action and role", func(t *testing.T) {
		rendered := renderBanner(t, Config{Description: "Scheduled maintenance"})

		assert.Contains(t, rendered, `role="banner"`)
		assert.Contains(t, rendered, `x-data="{ show: true }"`)
		assert.Contains(t, rendered, `@click="show = false"`)
		assert.Contains(t, rendered, `aria-label="dismiss banner"`)
		assert.Contains(t, rendered, "Scheduled maintenance")
	})

	t.Run("custom dismiss action and persistent mode", func(t *testing.T) {
		rendered := renderBanner(t, Config{
			Description:   "Pinned notice",
			Persistent:    true,
			DismissAction: "dismissBanner()",
		})

		assert.NotContains(t, rendered, `aria-label="dismiss banner"`)
		assert.NotContains(t, rendered, `dismissBanner()`)
		assert.Contains(t, rendered, "Pinned notice")
	})

	t.Run("cta link and cta button branches", func(t *testing.T) {
		linkHTML := renderBanner(t, Config{
			Description: "Read the release notes",
			CTA: &CTAConfig{
				ActionLabel: "Open notes",
				Href:        "/releases",
			},
		})
		assert.Contains(t, linkHTML, `<a href="/releases"`)
		assert.Contains(t, linkHTML, "Open notes")
		assert.NotContains(t, linkHTML, `<button type="button" class="whitespace-nowrap bg-primary`)

		buttonHTML := renderBanner(t, Config{
			Description: "Start trial",
			CTA: &CTAConfig{
				ActionLabel: "Start now",
				OnClick:     "startTrial()",
			},
		})
		assert.Contains(t, buttonHTML, `<button type="button"`)
		assert.Contains(t, buttonHTML, `@click="startTrial()"`)
		assert.Contains(t, buttonHTML, "Start now")
	})
}

func TestCoverageRenderCookieBannerDefaultsAndCustomActions(t *testing.T) {
	defaultHTML := renderStructuralBanner(t, CookieBanner(CookieBannerConfig{
		Description: "We use cookies",
	}))
	for _, want := range []string{
		`role="dialog"`,
		`aria-label="Cookie consent"`,
		"Cookie Consent",
		"Accept",
		"Decline",
		"We use cookies",
	} {
		require.Contains(t, defaultHTML, want)
	}

	customHTML := renderStructuralBanner(t, CookieBanner(CookieBannerConfig{
		Title:        "Privacy choices",
		Description:  "Choose your preferences",
		AcceptLabel:  "Allow",
		RejectLabel:  "Reject",
		AcceptAction: "allowCookies()",
		RejectAction: "show = false",
		RootClass:    "cookie-shadow",
	}))

	for _, want := range []string{
		"Privacy choices",
		"Allow",
		"Reject",
		`@click="allowCookies()"`,
		`@click="show = false"`,
		"cookie-shadow",
	} {
		require.Contains(t, customHTML, want)
	}
	assert.False(t, strings.Contains(customHTML, "Cookie Consent"), "custom config should replace the default title")
}

func TestBannerHTMXActionsPreferButtonsAndKeepLegacyAlpine(t *testing.T) {
	dismiss := renderBanner(t, Config{
		Description: "Maintenance", DismissAction: "show = false",
		DismissHTMX: &HTMXConfig{Post: "/api/banner/dismiss", Target: "#notice", Swap: "outerHTML"},
	})
	assert.Contains(t, dismiss, `@click="show = false"`)
	assert.Contains(t, dismiss, `hx-post="/api/banner/dismiss"`)
	assert.Contains(t, dismiss, `hx-target="#notice"`)

	cta := renderBanner(t, Config{
		Description: "Deploy ready",
		CTA: &CTAConfig{ActionLabel: "Deploy", Href: "/deploy", OnClick: "trackDeploy()", HTMX: &HTMXConfig{
			Post: "/api/deploy", Target: "#deploy-result", Swap: "innerHTML", Trigger: "click", Vals: `{"release":"v1"}`, Confirm: "Deploy?",
		}},
	})
	for _, want := range []string{
		`<button type="button"`, `@click="trackDeploy()"`, `hx-post="/api/deploy"`,
		`hx-target="#deploy-result"`, `hx-swap="innerHTML"`, `hx-trigger="click"`,
		`hx-vals="{&#34;release&#34;:&#34;v1&#34;}"`, `hx-confirm="Deploy?"`,
	} {
		assert.Contains(t, cta, want)
	}
	assert.NotContains(t, cta, `href="/deploy"`)

	cookies := renderStructuralBanner(t, CookieBanner(CookieBannerConfig{
		Description: "Cookies", AcceptAction: "show = false",
		AcceptHTMX: &HTMXConfig{Post: "/api/consent", Target: "#consent", Swap: "outerHTML"},
		RejectHTMX: &HTMXConfig{Post: "/api/consent/reject", Confirm: "Reject optional cookies?"},
	}))
	assert.Contains(t, cookies, `@click="show = false"`)
	assert.Contains(t, cookies, `hx-post="/api/consent"`)
	assert.Contains(t, cookies, `hx-target="#consent"`)
	assert.Contains(t, cookies, `hx-post="/api/consent/reject"`)
	assert.Contains(t, cookies, `hx-confirm="Reject optional cookies?"`)
}

func TestBannerHTMXPostTakesPrecedenceOverGet(t *testing.T) {
	rendered := renderBanner(t, Config{
		Description: "Deploy ready",
		CTA:         &CTAConfig{ActionLabel: "Deploy", HTMX: &HTMXConfig{Get: "/preview", Post: "/deploy"}},
	})

	assert.Contains(t, rendered, `hx-post="/deploy"`)
	assert.NotContains(t, rendered, `hx-get="/preview"`)
}
