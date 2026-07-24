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
			if got := tt.cfg.resolvedID(); got != tt.want {
				t.Fatalf("resolvedID() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolvedName(t *testing.T) {
	if got := (Config{Name: "field"}).resolvedName(); got != "field" {
		t.Fatalf("resolvedName with Name = %q, want %q", got, "field")
	}
	// No name: falls back to resolvedID which falls back to "rating".
	if got := (Config{ID: "rid"}).resolvedName(); got != "rid" {
		t.Fatalf("resolvedName fallback to ID = %q, want %q", got, "rid")
	}
	if got := (Config{}).resolvedName(); got != "rating" {
		t.Fatalf("resolvedName default = %q, want %q", got, "rating")
	}
}

func TestResolvedMax(t *testing.T) {
	if got := (Config{Max: 10}).resolvedMax(); got != 10 {
		t.Fatalf("resolvedMax(10) = %d, want 10", got)
	}
	if got := (Config{}).resolvedMax(); got != 5 {
		t.Fatalf("resolvedMax default = %d, want 5", got)
	}
	if got := (Config{Max: -2}).resolvedMax(); got != 5 {
		t.Fatalf("resolvedMax(-2) = %d, want default 5", got)
	}
}

func TestResolvedValueClamping(t *testing.T) {
	if got := (Config{Value: -5}).resolvedValue(); got != 0 {
		t.Fatalf("resolvedValue(-5) = %d, want 0", got)
	}
	if got := (Config{Value: 99, Max: 5}).resolvedValue(); got != 5 {
		t.Fatalf("resolvedValue(99) = %d, want 5", got)
	}
	if got := (Config{Value: 3}).resolvedValue(); got != 3 {
		t.Fatalf("resolvedValue(3) = %d, want 3", got)
	}
}

func TestResolvedLabel(t *testing.T) {
	if got := (Config{Label: "Stars"}).resolvedLabel(); got != "Stars" {
		t.Fatalf("resolvedLabel(Stars) = %q", got)
	}
	if got := (Config{}).resolvedLabel(); got != "Rating" {
		t.Fatalf("resolvedLabel default = %q, want Rating", got)
	}
}

func TestRootClasses(t *testing.T) {
	base := (Config{}).rootClasses()
	if !strings.Contains(base, "inline-flex flex-col gap-2") {
		t.Fatalf("rootClasses missing base: %q", base)
	}
	if strings.Contains(base, "custom") {
		t.Fatalf("rootClasses should not contain custom when Class empty: %q", base)
	}
	withClass := (Config{RootClass: "custom-class"}).rootClasses()
	if !strings.Contains(withClass, "custom-class") {
		t.Fatalf("rootClasses missing appended Class: %q", withClass)
	}
}

func TestControlClasses(t *testing.T) {
	interactive := (Config{}).controlClasses()
	if !strings.Contains(interactive, "focus-within:outline-2") {
		t.Fatalf("interactive controlClasses should add focus outline: %q", interactive)
	}
	display := (DisplayConfig{}).controlClasses()
	if strings.Contains(display, "focus-within:outline-2") {
		t.Fatalf("display controlClasses should not add focus outline: %q", display)
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
			got := (Config{Size: tt.size}).iconClasses()
			if !strings.Contains(got, tt.want) {
				t.Fatalf("iconClasses(%q) = %q, want substring %q", tt.size, got, tt.want)
			}
			if strings.Contains(got, "opacity-60") {
				t.Fatalf("iconClasses should not dim when enabled: %q", got)
			}
		})
	}
	disabled := (Config{Disabled: true}).iconClasses()
	if !strings.Contains(disabled, "opacity-60") {
		t.Fatalf("disabled iconClasses should dim: %q", disabled)
	}
}

func TestActiveInactiveIconClasses(t *testing.T) {
	star := Config{Appearance: AppearanceStars}
	if got := star.activeIconClasses(); got != "text-warning" {
		t.Fatalf("star activeIconClasses = %q", got)
	}
	if !strings.Contains(star.inactiveIconClasses(), "text-on-surface-muted") {
		t.Fatalf("star inactiveIconClasses = %q", star.inactiveIconClasses())
	}
	emoji := Config{Appearance: AppearanceEmoji}
	if !strings.Contains(emoji.activeIconClasses(), "scale-110") {
		t.Fatalf("emoji activeIconClasses = %q", emoji.activeIconClasses())
	}
	if !strings.Contains(emoji.inactiveIconClasses(), "grayscale") {
		t.Fatalf("emoji inactiveIconClasses = %q", emoji.inactiveIconClasses())
	}
}

func TestXDataAndInputID(t *testing.T) {
	if got := (Config{Value: 2}).xData(); got != "{ currentVal: 2 }" {
		t.Fatalf("xData = %q", got)
	}
	if got := (Config{ID: "r"}).inputID(3); got != "r-3" {
		t.Fatalf("inputID = %q, want r-3", got)
	}
}

func TestValueLabel(t *testing.T) {
	star := Config{Appearance: AppearanceStars}
	if got := star.valueLabel(1); got != "one star" {
		t.Fatalf("valueLabel(1) = %q, want 'one star'", got)
	}
	if got := star.valueLabel(3); got != "3 stars" {
		t.Fatalf("valueLabel(3) = %q, want '3 stars'", got)
	}
	emoji := Config{Appearance: AppearanceEmoji}
	if got := emoji.valueLabel(3); got != "neutral" {
		t.Fatalf("emoji valueLabel(3) = %q, want 'neutral'", got)
	}
	// Emoji with value outside the default option set falls back to star wording.
	if got := emoji.valueLabel(9); got != "9 stars" {
		t.Fatalf("emoji valueLabel(9) = %q, want '9 stars'", got)
	}
	if got := emoji.valueLabel(1); got != "very dissatisfied" {
		t.Fatalf("emoji valueLabel(1) = %q", got)
	}
}

func TestEmojiIcon(t *testing.T) {
	cfg := Config{Appearance: AppearanceEmoji}
	if got := cfg.emojiIcon(5); got != "😍" {
		t.Fatalf("emojiIcon(5) = %q, want 😍", got)
	}
	if got := cfg.emojiIcon(99); got != "🙂" {
		t.Fatalf("emojiIcon(99) fallback = %q, want 🙂", got)
	}
}

func TestBindClass(t *testing.T) {
	emoji := (Config{Appearance: AppearanceEmoji}).bindClass(2)
	if !strings.Contains(emoji, "currentVal === 2") {
		t.Fatalf("emoji bindClass = %q", emoji)
	}
	star := (Config{Appearance: AppearanceStars}).bindClass(2)
	if !strings.Contains(star, "currentVal >= 2") {
		t.Fatalf("star bindClass = %q", star)
	}
}

func TestDisplayIconState(t *testing.T) {
	cfg := DisplayConfig{Appearance: AppearanceStars, Value: 3}
	if got := displayIconState(cfg, 2); got != cfg.activeIconClasses() {
		t.Fatalf("displayIconState active = %q", got)
	}
	if got := displayIconState(cfg, 4); got != cfg.inactiveIconClasses() {
		t.Fatalf("displayIconState inactive = %q", got)
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

func TestRenderEmojiDisplay(t *testing.T) {
	html := renderStructuralRating(t, RatingDisplay(DisplayConfig{Appearance: AppearanceEmoji, Value: 3}))
	if strings.Contains(html, `type="radio"`) {
		t.Fatalf("emoji display must not render radios:\n%s", html)
	}
	if !strings.Contains(html, `role="img"`) {
		t.Fatalf("rating display should render role=img:\n%s", html)
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
		Size:      SizeLG,
		Value:     1,
		RootAttrs: templ.Attributes{"data-test": "rating-root"},
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
