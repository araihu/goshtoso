package badge

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// renderComponent renders any templ component into a string for assertions.
func renderComponent(t *testing.T, c templ.Component) string {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, c.Render(context.Background(), &buf))
	return buf.String()
}

// allVariants is every Variant the component understands, used to exercise the
// switch arms of the class helpers and templ entry points.
var allVariants = []Variant{
	Default, Inverse, Primary, Secondary, Info, Success, Warning, Danger,
}

func TestSizeClasses(t *testing.T) {
	cases := map[Size]string{
		SizeSM:     "text-[10px] px-1.5 py-0.5",
		SizeLG:     "text-sm px-3 py-1.5",
		SizeMD:     "text-xs px-2 py-1",
		Size("xx"): "text-xs px-2 py-1", // unknown size falls back to default
	}
	for size, want := range cases {
		got := Config{Size: size}.SizeClasses()
		assert.Equalf(t, want, got, "SizeClasses for %q", size)
	}
}

func TestSizeTextClass(t *testing.T) {
	cases := map[Size]string{
		SizeSM:     "text-[10px]",
		SizeLG:     "text-sm",
		SizeMD:     "text-xs",
		Size("xx"): "text-xs", // unknown size falls back to default
	}
	for size, want := range cases {
		got := Config{Size: size}.SizeTextClass()
		assert.Equalf(t, want, got, "SizeTextClass for %q", size)
	}
}

func TestVariantClasses_AllArms(t *testing.T) {
	// Each variant maps to a distinct background token; assert the token is present
	// so every switch arm is both executed and verified.
	wantBG := map[Variant]string{
		Default:   "bg-surface-alt",
		Inverse:   "bg-surface-dark-alt",
		Primary:   "bg-primary",
		Secondary: "bg-secondary",
		Info:      "bg-info",
		Success:   "bg-success",
		Warning:   "bg-warning",
		Danger:    "bg-danger",
	}
	for _, v := range allVariants {
		got := Config{Variant: v}.VariantClasses()
		assert.Containsf(t, got, "border", "VariantClasses %q should set a border", v)
		assert.Containsf(t, got, wantBG[v], "VariantClasses %q background", v)
	}
	// Unknown variant falls back to the default arm.
	assert.Equal(t,
		Config{Variant: Default}.VariantClasses(),
		Config{Variant: Variant("nope")}.VariantClasses(),
	)
}

func TestSoftVariantClasses_AllArms(t *testing.T) {
	wantText := map[Variant]string{
		Default:   "text-on-surface",
		Inverse:   "text-on-surface",
		Primary:   "text-primary",
		Secondary: "text-secondary",
		Info:      "text-info",
		Success:   "text-success",
		Warning:   "text-warning",
		Danger:    "text-danger",
	}
	for _, v := range allVariants {
		got := Config{Variant: v}.SoftVariantClasses()
		assert.Containsf(t, got, "bg-surface", "SoftVariantClasses %q uses surface bg", v)
		assert.Containsf(t, got, wantText[v], "SoftVariantClasses %q text", v)
	}
	assert.Equal(t,
		Config{Variant: Default}.SoftVariantClasses(),
		Config{Variant: Variant("nope")}.SoftVariantClasses(),
	)
}

func TestSoftInnerClasses_AllArms(t *testing.T) {
	wantBG := map[Variant]string{
		Default:   "bg-surface-alt/10",
		Inverse:   "bg-surface-dark-alt/10",
		Primary:   "bg-primary/10",
		Secondary: "bg-secondary/10",
		Info:      "bg-info/10",
		Success:   "bg-success/10",
		Warning:   "bg-warning/10",
		Danger:    "bg-danger/10",
	}
	for _, v := range allVariants {
		got := Config{Variant: v}.SoftInnerClasses()
		assert.Containsf(t, got, wantBG[v], "SoftInnerClasses %q", v)
	}
	assert.Equal(t,
		Config{Variant: Default}.SoftInnerClasses(),
		Config{Variant: Variant("nope")}.SoftInnerClasses(),
	)
}

func TestIndicatorClasses_AllArms(t *testing.T) {
	wantBG := map[Variant]string{
		Default:   "bg-on-surface",
		Inverse:   "bg-on-surface",
		Primary:   "bg-primary",
		Secondary: "bg-secondary",
		Info:      "bg-info",
		Success:   "bg-success",
		Warning:   "bg-warning",
		Danger:    "bg-danger",
	}
	for _, v := range allVariants {
		got := Config{Variant: v}.IndicatorClasses()
		assert.Containsf(t, got, "size-1.5 rounded-full", "IndicatorClasses %q base", v)
		assert.Containsf(t, got, wantBG[v], "IndicatorClasses %q bg", v)
	}
	// IndicatorColor override short-circuits the variant switch.
	got := Config{Variant: Danger, IndicatorColor: "bg-pink-500"}.IndicatorClasses()
	assert.Equal(t, "size-1.5 rounded-full bg-pink-500", got)
	assert.NotContains(t, got, "bg-danger")
}

func TestIsSoft(t *testing.T) {
	assert.True(t, Config{Style: StyleSoft}.IsSoft())
	assert.False(t, Config{Style: StyleSolid}.IsSoft())
	assert.False(t, Config{}.IsSoft())
}

func TestBadge_SimplePath(t *testing.T) {
	// No soft/icon/indicator => simpleBadge path.
	html := renderComponent(t, Badge(Config{Label: "New", Variant: Primary, RootClass: "ml-2"}))
	assert.Contains(t, html, "New")
	assert.Contains(t, html, "bg-primary")
	assert.Contains(t, html, "ml-2")           // RootClass threaded through
	assert.Contains(t, html, "rounded-radius") // simpleBadge container marker
	assert.NotContains(t, html, "inline-flex") // not the inner-span path
}

func TestBadge_SoftPath(t *testing.T) {
	html := renderComponent(t, Badge(Config{Label: "Active", Variant: Success, Style: StyleSoft, RootClass: "mr-1"}))
	assert.Contains(t, html, "Active")
	assert.Contains(t, html, "inline-flex")   // badgeWithInner container
	assert.Contains(t, html, "text-success")  // soft variant text color
	assert.Contains(t, html, "bg-success/10") // soft inner bg
	assert.Contains(t, html, "mr-1")
}

func TestBadge_IndicatorPath(t *testing.T) {
	html := renderComponent(t, Badge(Config{Label: "Live", Variant: Danger, Indicator: true}))
	assert.Contains(t, html, "inline-flex")
	assert.Contains(t, html, `aria-hidden="true"`)
	assert.Contains(t, html, "size-1.5 rounded-full")
	assert.Contains(t, html, "bg-danger")
}

func TestBadge_IconPath(t *testing.T) {
	icon := templ.Raw(`<svg data-testid="star"></svg>`)
	html := renderComponent(t, Badge(Config{Label: "Starred", Variant: Warning, Icon: icon}))
	assert.Contains(t, html, "inline-flex")
	assert.Contains(t, html, `data-testid="star"`)
	assert.Contains(t, html, "shrink-0") // icon wrapper
	assert.Contains(t, html, "Starred")
}

func TestBadge_AllVariantsRender(t *testing.T) {
	for _, v := range allVariants {
		for _, style := range []Style{StyleSolid, StyleSoft} {
			html := renderComponent(t, Badge(Config{Label: "x", Variant: v, Style: style}))
			assert.NotEmptyf(t, html, "variant %q style %q rendered empty", v, style)
			assert.Containsf(t, html, "class=", "variant %q style %q missing class", v, style)
		}
	}
}

func TestBadge_LabelEscaped(t *testing.T) {
	// templ must escape HTML in user-provided labels.
	html := renderComponent(t, Badge(Config{Label: `<script>alert(1)</script>`}))
	assert.NotContains(t, html, "<script>alert(1)</script>")
	assert.Contains(t, html, "&lt;script&gt;")
}

func TestNotificationBadge(t *testing.T) {
	// count <= 0 renders nothing.
	assert.Empty(t, strings.TrimSpace(renderComponent(t, NotificationBadge(0))))
	assert.Empty(t, strings.TrimSpace(renderComponent(t, NotificationBadge(-5))))

	// 1..99 shows the number.
	html := renderComponent(t, NotificationBadge(7))
	assert.Contains(t, html, "7")
	assert.Contains(t, html, "bg-danger")

	// >99 caps at "99+".
	html = renderComponent(t, NotificationBadge(150))
	assert.Contains(t, html, "99+")
	assert.NotContains(t, html, "150")
}

func TestNotificationDot(t *testing.T) {
	html := renderComponent(t, NotificationDot())
	assert.Contains(t, html, "rounded-full")
	assert.Contains(t, html, "bg-danger")
	assert.Contains(t, html, "size-3")
}

func TestAnimatingDot_AllVariants(t *testing.T) {
	wantColor := map[Variant]string{
		Default:   "bg-primary", // unmatched arms fall through to the primary default
		Inverse:   "bg-primary", // no case for Inverse => default color
		Primary:   "bg-primary",
		Secondary: "bg-secondary",
		Info:      "bg-info",
		Success:   "bg-success",
		Warning:   "bg-warning",
		Danger:    "bg-danger",
	}
	for _, v := range allVariants {
		html := renderComponent(t, AnimatingDot(v))
		assert.Containsf(t, html, `aria-label="notification"`, "AnimatingDot %q aria-label", v)
		assert.Containsf(t, html, "animate-ping", "AnimatingDot %q ping", v)
		assert.Containsf(t, html, "motion-reduce:animate-none", "AnimatingDot %q reduced motion", v)
		assert.Containsf(t, html, wantColor[v], "AnimatingDot %q color", v)
	}
}
