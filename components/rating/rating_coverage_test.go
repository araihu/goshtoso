package rating

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

func renderRating(t *testing.T, cfg Config) string {
	t.Helper()
	var buf bytes.Buffer
	if err := Rating(cfg).Render(context.Background(), &buf); err != nil {
		t.Fatalf("render rating: %v", err)
	}
	return buf.String()
}

func TestResolvedIDFallbacks(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want string
	}{
		{"explicit id", Config{ID: "explicit", Name: "n"}, "explicit"},
		{"name fallback", Config{Name: "field"}, "field"},
		{"default", Config{}, "rating"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.ResolvedID(); got != tt.want {
				t.Fatalf("ResolvedID() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolvedName(t *testing.T) {
	if got := (Config{Name: "field"}).ResolvedName(); got != "field" {
		t.Fatalf("ResolvedName with Name = %q, want %q", got, "field")
	}
	// No name: falls back to ResolvedID which falls back to "rating".
	if got := (Config{ID: "rid"}).ResolvedName(); got != "rid" {
		t.Fatalf("ResolvedName fallback to ID = %q, want %q", got, "rid")
	}
	if got := (Config{}).ResolvedName(); got != "rating" {
		t.Fatalf("ResolvedName default = %q, want %q", got, "rating")
	}
}

func TestResolvedMax(t *testing.T) {
	if got := (Config{Max: 10}).ResolvedMax(); got != 10 {
		t.Fatalf("ResolvedMax(10) = %d, want 10", got)
	}
	if got := (Config{}).ResolvedMax(); got != 5 {
		t.Fatalf("ResolvedMax default = %d, want 5", got)
	}
	if got := (Config{Max: -2}).ResolvedMax(); got != 5 {
		t.Fatalf("ResolvedMax(-2) = %d, want default 5", got)
	}
}

func TestResolvedValueClamping(t *testing.T) {
	if got := (Config{Value: -5}).ResolvedValue(); got != 0 {
		t.Fatalf("ResolvedValue(-5) = %d, want 0", got)
	}
	if got := (Config{Value: 99, Max: 5}).ResolvedValue(); got != 5 {
		t.Fatalf("ResolvedValue(99) = %d, want 5", got)
	}
	if got := (Config{Value: 3}).ResolvedValue(); got != 3 {
		t.Fatalf("ResolvedValue(3) = %d, want 3", got)
	}
}

func TestResolvedLabel(t *testing.T) {
	if got := (Config{Label: "Stars"}).ResolvedLabel(); got != "Stars" {
		t.Fatalf("ResolvedLabel(Stars) = %q", got)
	}
	if got := (Config{}).ResolvedLabel(); got != "Rating" {
		t.Fatalf("ResolvedLabel default = %q, want Rating", got)
	}
}

func TestRootClasses(t *testing.T) {
	base := (Config{}).RootClasses()
	if !strings.Contains(base, "inline-flex flex-col gap-2") {
		t.Fatalf("RootClasses missing base: %q", base)
	}
	if strings.Contains(base, "custom") {
		t.Fatalf("RootClasses should not contain custom when Class empty: %q", base)
	}
	withClass := (Config{Class: "custom-class"}).RootClasses()
	if !strings.Contains(withClass, "custom-class") {
		t.Fatalf("RootClasses missing appended Class: %q", withClass)
	}
}

func TestControlClasses(t *testing.T) {
	interactive := (Config{}).ControlClasses()
	if !strings.Contains(interactive, "focus-within:outline-2") {
		t.Fatalf("interactive ControlClasses should add focus outline: %q", interactive)
	}
	readOnly := (Config{ReadOnly: true}).ControlClasses()
	if strings.Contains(readOnly, "focus-within:outline-2") {
		t.Fatalf("read-only ControlClasses should not add focus outline: %q", readOnly)
	}
}

func TestIconClassesSizes(t *testing.T) {
	tests := []struct {
		size Size
		want string
	}{
		{SizeSM, "size-5 text-lg"},
		{SizeLG, "size-8 text-3xl"},
		{SizeXL, "size-10 text-4xl"},
		{SizeMD, "size-6 text-2xl"},
		{Size("unknown"), "size-6 text-2xl"},
	}
	for _, tt := range tests {
		t.Run(string(tt.size), func(t *testing.T) {
			got := (Config{Size: tt.size}).IconClasses()
			if !strings.Contains(got, tt.want) {
				t.Fatalf("IconClasses(%q) = %q, want substring %q", tt.size, got, tt.want)
			}
			if strings.Contains(got, "opacity-60") {
				t.Fatalf("IconClasses should not dim when enabled: %q", got)
			}
		})
	}
	disabled := (Config{Disabled: true}).IconClasses()
	if !strings.Contains(disabled, "opacity-60") {
		t.Fatalf("disabled IconClasses should dim: %q", disabled)
	}
}

func TestActiveInactiveIconClasses(t *testing.T) {
	star := Config{Style: StyleStars}
	if got := star.ActiveIconClasses(); got != "text-warning" {
		t.Fatalf("star ActiveIconClasses = %q", got)
	}
	if !strings.Contains(star.InactiveIconClasses(), "text-on-surface-muted") {
		t.Fatalf("star InactiveIconClasses = %q", star.InactiveIconClasses())
	}
	emoji := Config{Style: StyleEmoji}
	if !strings.Contains(emoji.ActiveIconClasses(), "scale-110") {
		t.Fatalf("emoji ActiveIconClasses = %q", emoji.ActiveIconClasses())
	}
	if !strings.Contains(emoji.InactiveIconClasses(), "grayscale") {
		t.Fatalf("emoji InactiveIconClasses = %q", emoji.InactiveIconClasses())
	}
}

func TestXDataAndInputID(t *testing.T) {
	if got := (Config{Value: 2}).XData(); got != "{ currentVal: 2 }" {
		t.Fatalf("XData = %q", got)
	}
	if got := (Config{ID: "r"}).InputID(3); got != "r-3" {
		t.Fatalf("InputID = %q, want r-3", got)
	}
}

func TestValueLabel(t *testing.T) {
	star := Config{Style: StyleStars}
	if got := star.ValueLabel(1); got != "one star" {
		t.Fatalf("ValueLabel(1) = %q, want 'one star'", got)
	}
	if got := star.ValueLabel(3); got != "3 stars" {
		t.Fatalf("ValueLabel(3) = %q, want '3 stars'", got)
	}
	emoji := Config{Style: StyleEmoji}
	if got := emoji.ValueLabel(3); got != "neutral" {
		t.Fatalf("emoji ValueLabel(3) = %q, want 'neutral'", got)
	}
	// Emoji with value outside the default option set falls back to star wording.
	if got := emoji.ValueLabel(9); got != "9 stars" {
		t.Fatalf("emoji ValueLabel(9) = %q, want '9 stars'", got)
	}
	if got := emoji.ValueLabel(1); got != "very dissatisfied" {
		t.Fatalf("emoji ValueLabel(1) = %q", got)
	}
}

func TestEmojiIcon(t *testing.T) {
	cfg := Config{Style: StyleEmoji}
	if got := cfg.EmojiIcon(5); got != "😍" {
		t.Fatalf("EmojiIcon(5) = %q, want 😍", got)
	}
	if got := cfg.EmojiIcon(99); got != "🙂" {
		t.Fatalf("EmojiIcon(99) fallback = %q, want 🙂", got)
	}
}

func TestBindClass(t *testing.T) {
	emoji := (Config{Style: StyleEmoji}).BindClass(2)
	if !strings.Contains(emoji, "currentVal === 2") {
		t.Fatalf("emoji BindClass = %q", emoji)
	}
	star := (Config{Style: StyleStars}).BindClass(2)
	if !strings.Contains(star, "currentVal >= 2") {
		t.Fatalf("star BindClass = %q", star)
	}
}

func TestReadOnlyIconState(t *testing.T) {
	cfg := Config{Style: StyleStars, Value: 3}
	if got := readOnlyIconState(cfg, 2); got != cfg.ActiveIconClasses() {
		t.Fatalf("readOnlyIconState active = %q", got)
	}
	if got := readOnlyIconState(cfg, 4); got != cfg.InactiveIconClasses() {
		t.Fatalf("readOnlyIconState inactive = %q", got)
	}
}

func TestRenderShowLabel(t *testing.T) {
	html := renderRating(t, Config{Label: "Quality", ShowLabel: true, Value: 2})
	if !strings.Contains(html, ">Quality<") {
		t.Fatalf("ShowLabel should render visible label:\n%s", html)
	}
	hidden := renderRating(t, Config{Label: "Quality", Value: 2})
	if strings.Contains(hidden, `class="text-sm font-medium`) {
		t.Fatalf("label span should be absent when ShowLabel false:\n%s", hidden)
	}
}

func TestRenderDisabledInputs(t *testing.T) {
	html := renderRating(t, Config{Value: 2, Disabled: true})
	if !strings.Contains(html, "disabled") {
		t.Fatalf("disabled rating should render disabled inputs:\n%s", html)
	}
	if !strings.Contains(html, "opacity-60") {
		t.Fatalf("disabled rating icons should dim:\n%s", html)
	}
}

func TestRenderReadOnlyEmoji(t *testing.T) {
	html := renderRating(t, Config{Style: StyleEmoji, ReadOnly: true, Value: 3})
	if strings.Contains(html, `type="radio"`) {
		t.Fatalf("read-only emoji must not render radios:\n%s", html)
	}
	if !strings.Contains(html, `role="img"`) {
		t.Fatalf("read-only rating should render role=img:\n%s", html)
	}
	if !strings.Contains(html, `aria-label="neutral"`) {
		t.Fatalf("read-only emoji aria-label should use sentiment:\n%s", html)
	}
	if !strings.Contains(html, "😐") {
		t.Fatalf("read-only emoji should render emoji icons:\n%s", html)
	}
}

func TestRenderSizesAndAttrs(t *testing.T) {
	html := renderRating(t, Config{
		Size:  SizeLG,
		Value: 1,
		Attrs: templ.Attributes{"data-test": "rating-root"},
	})
	if !strings.Contains(html, "size-8") {
		t.Fatalf("large rating should use size-8 icons:\n%s", html)
	}
	if !strings.Contains(html, `data-test="rating-root"`) {
		t.Fatalf("Attrs escape hatch should apply to root:\n%s", html)
	}
}

func TestRenderCustomMax(t *testing.T) {
	html := renderRating(t, Config{ID: "r", Max: 3, Value: 2})
	if !strings.Contains(html, `id="r-3"`) {
		t.Fatalf("custom max should render option 3:\n%s", html)
	}
	if strings.Contains(html, `id="r-4"`) {
		t.Fatalf("custom max=3 should not render option 4:\n%s", html)
	}
}
