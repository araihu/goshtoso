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
				Variant:   Primary,
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
			cfg:       Config{Variant: Info},
			wantParts: []string{"border-b border-info bg-info/10"},
		},
		{
			name:      "success",
			cfg:       Config{Variant: Success},
			wantParts: []string{"border-b border-success bg-success/10"},
		},
		{
			name:      "warning",
			cfg:       Config{Variant: Warning},
			wantParts: []string{"border-b border-warning bg-warning/10"},
		},
		{
			name:      "danger",
			cfg:       Config{Variant: Danger},
			wantParts: []string{"border-b border-danger bg-danger/10"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			classes := tt.cfg.ContainerClasses()
			for _, want := range tt.wantParts {
				assert.Contains(t, classes, want)
			}
		})
	}
}

func TestCoverageLinkClassesCoverEveryVariant(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want string
	}{
		{name: "default", cfg: Config{}, want: "text-primary"},
		{name: "primary", cfg: Config{Variant: Primary}, want: "text-primary"},
		{name: "info", cfg: Config{Variant: Info}, want: "text-info"},
		{name: "success", cfg: Config{Variant: Success}, want: "text-success"},
		{name: "warning", cfg: Config{Variant: Warning}, want: "text-warning"},
		{name: "danger", cfg: Config{Variant: Danger}, want: "text-danger"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			classes := tt.cfg.LinkClasses()
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
	defaultHTML := renderBanner(t, Config{
		CookieBanner: true,
		Description:  "We use cookies",
	})
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

	customHTML := renderBanner(t, Config{
		CookieBanner: true,
		Description:  "Choose your preferences",
		CookieConfig: &CookieBannerConfig{
			Title:        "Privacy choices",
			AcceptLabel:  "Allow",
			RejectLabel:  "Reject",
			AcceptAction: "allowCookies()",
			RejectAction: "show = false",
		},
		RootClass: "cookie-shadow",
	})

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
