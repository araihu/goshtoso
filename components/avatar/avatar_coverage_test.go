package avatar

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

// renderAvatar renders cfg and returns the HTML, failing the test on error.
func renderAvatar(t *testing.T, cfg Config) string {
	t.Helper()
	var buf bytes.Buffer
	if err := Avatar(cfg).Render(context.Background(), &buf); err != nil {
		t.Fatalf("render avatar: %v", err)
	}
	return buf.String()
}

func TestGetInitials(t *testing.T) {
	cases := []struct {
		name  string
		email string
		want  string
	}{
		{"John Doe", "", "JD"},          // two words
		{"dev-ops", "", "DO"},           // hyphen split
		{"data_eng", "", "DE"},          // underscore split
		{"Engineering", "", "EN"},       // single long word -> first two letters
		{"x", "", "X"},                  // single one-char word
		{"  Ada  Lovelace  ", "", "AL"}, // extra whitespace
		{"", "alice@x.com", "AL"},       // fallback to email
		{"", "z", "Z"},                  // single-char email
		{"", "", "?"},                   // nothing
		{"123 abc", "", "1A"},           // non-letter first char passes through toUpper
	}
	for _, c := range cases {
		if got := GetInitials(c.name, c.email); got != c.want {
			t.Errorf("GetInitials(%q, %q) = %q, want %q", c.name, c.email, got, c.want)
		}
	}
}

func TestToUpper(t *testing.T) {
	cases := []struct {
		in   byte
		want string
	}{
		{'a', "A"},
		{'z', "Z"},
		{'A', "A"}, // already upper, untouched
		{'5', "5"}, // digit untouched
		{'@', "@"}, // symbol untouched
	}
	for _, c := range cases {
		if got := toUpper(c.in); got != c.want {
			t.Errorf("toUpper(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestSplitWords(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"John Doe", []string{"John", "Doe"}},
		{"dev-ops", []string{"dev", "ops"}},
		{"a_b_c", []string{"a", "b", "c"}},
		{"  leading", []string{"leading"}},
		{"trailing  ", []string{"trailing"}},
		{"", nil},
	}
	for _, c := range cases {
		got := splitWords(c.in)
		if len(got) != len(c.want) {
			t.Errorf("splitWords(%q) = %v, want %v", c.in, got, c.want)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("splitWords(%q)[%d] = %q, want %q", c.in, i, got[i], c.want[i])
			}
		}
	}
}

func TestJSEscapeSingle(t *testing.T) {
	cases := map[string]string{
		`plain`:       `plain`,
		`a'b`:         `a\'b`,
		`a\b`:         `a\\b`, // backslash escaped
		"line\nbreak": `line\nbreak`,
		"car\rriage":  `car\rriage`,
		"tab\there":   `tab\there`,
	}
	for in, want := range cases {
		if got := jsEscapeSingle(in); got != want {
			t.Errorf("jsEscapeSingle(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestResolvedInitials(t *testing.T) {
	cases := []struct {
		cfg  Config
		want string
	}{
		{Config{Initials: "JS"}, "JS"},            // explicit wins
		{Config{Name: "Grace Hopper"}, "GH"},      // derived from name
		{Config{}, "?"},                           // empty fallback
		{Config{Initials: "X", Name: "Ada"}, "X"}, // explicit beats name
	}
	for _, c := range cases {
		if got := c.cfg.resolvedInitials(); got != c.want {
			t.Errorf("resolvedInitials(%+v) = %q, want %q", c.cfg, got, c.want)
		}
	}
}

func TestHasInitialsAndHasImage(t *testing.T) {
	if !(Config{Initials: "JS"}).hasInitials() {
		t.Error("hasInitials should be true when Initials set")
	}
	if (Config{Name: "John"}).hasInitials() {
		t.Error("hasInitials should be false when only Name set (Initials empty)")
	}
	if !(Config{Src: "/a.png"}).hasImage() {
		t.Error("hasImage should be true when Src set")
	}
	if !(Config{SrcExpr: "avatarSrc"}).hasImage() {
		t.Error("hasImage should be true when SrcExpr set")
	}
	if (Config{}).hasImage() {
		t.Error("hasImage should be false with no Src/SrcExpr")
	}
}

func TestSizeClasses(t *testing.T) {
	cases := map[Size]string{
		SizeXS:   "size-8 text-xs",
		SizeSM:   "size-10 text-sm",
		SizeMD:   "size-14 text-2xl",
		SizeLG:   "size-20 text-3xl",
		SizeXL:   "size-24 text-4xl",
		Size2XL:  "size-32 text-5xl",
		Size(""): "size-14 text-2xl", // default
	}
	for size, want := range cases {
		if got := (Config{Size: size}).sizeClasses(); got != want {
			t.Errorf("sizeClasses(%q) = %q, want %q", size, got, want)
		}
	}
}

func TestStatusSizeClasses(t *testing.T) {
	cases := map[Size]string{
		SizeXS:   "size-2",
		SizeSM:   "size-2.5",
		SizeLG:   "size-5",
		SizeXL:   "size-6",
		Size2XL:  "size-7",
		SizeMD:   "size-4", // default branch
		Size(""): "size-4",
	}
	for size, want := range cases {
		if got := (Config{Size: size}).statusSizeClasses(); got != want {
			t.Errorf("statusSizeClasses(%q) = %q, want %q", size, got, want)
		}
	}
}

func TestSpinnerSizeClasses(t *testing.T) {
	cases := map[Size]string{
		SizeXS:   "size-4",
		SizeSM:   "size-5",
		SizeLG:   "size-8",
		SizeXL:   "size-10",
		Size2XL:  "size-12",
		SizeMD:   "size-6", // default
		Size(""): "size-6",
	}
	for size, want := range cases {
		if got := (Config{Size: size}).spinnerSizeClasses(); got != want {
			t.Errorf("spinnerSizeClasses(%q) = %q, want %q", size, got, want)
		}
	}
}

func TestStatusClasses(t *testing.T) {
	cases := map[Status]string{
		StatusOffline: "bg-outline dark:bg-outline-dark",
		StatusInfo:    "bg-info",
		StatusSuccess: "bg-success",
		StatusWarning: "bg-warning",
		StatusDanger:  "bg-danger",
		Status(""):    "", // default branch
	}
	for status, want := range cases {
		if got := (Config{Status: status}).statusClasses(); got != want {
			t.Errorf("statusClasses(%q) = %q, want %q", status, got, want)
		}
	}
}

func TestToneClasses(t *testing.T) {
	for _, v := range []Tone{ToneDefault, ToneInverse, TonePrimary, ToneSecondary, ToneInfo, ToneSuccess, ToneWarning, ToneDanger, Tone("")} {
		if got := (Config{Tone: v}).toneClasses(); got == "" {
			t.Errorf("toneClasses(%q) returned empty", v)
		}
	}
	// ToneDefault and unknown variant produce the same outline classes.
	if (Config{Tone: ToneDefault}).toneClasses() != (Config{Tone: Tone("nope")}).toneClasses() {
		t.Error("unknown variant should match ToneDefault")
	}
}

func TestVariantFillClasses(t *testing.T) {
	cases := map[Tone]string{
		ToneInverse:   "bg-surface-dark-alt",
		TonePrimary:   "bg-primary",
		ToneSecondary: "bg-secondary",
		ToneInfo:      "bg-info",
		ToneSuccess:   "bg-success",
		ToneWarning:   "bg-warning",
		ToneDanger:    "bg-danger",
		ToneDefault:   "bg-surface-alt",
	}
	for v, want := range cases {
		got := (Config{Tone: v}).toneFillClasses()
		if !strings.Contains(got, want) {
			t.Errorf("toneFillClasses(%q) = %q, want substring %q", v, got, want)
		}
	}
}

func TestRadiusClasses(t *testing.T) {
	cases := map[Radius]string{
		RadiusNone:    "rounded-none",
		RadiusXS:      "rounded-xs",
		RadiusSM:      "rounded-sm",
		RadiusLG:      "rounded-lg",
		RadiusMD:      "rounded-md", // explicit md hits default
		RadiusDefault: "rounded-md", // empty hits default
	}
	for r, want := range cases {
		if got := (Config{Radius: r}).radiusClasses(); got != want {
			t.Errorf("radiusClasses(%q) = %q, want %q", r, got, want)
		}
	}
}

func TestShapeClasses(t *testing.T) {
	if got := (Config{Shape: ShapeCircle}).shapeClasses(); got != "rounded-full" {
		t.Errorf("circle shapeClasses = %q, want rounded-full", got)
	}
	if got := (Config{Shape: ShapeSquare, Radius: RadiusSM}).shapeClasses(); got != "rounded-sm" {
		t.Errorf("square shapeClasses = %q, want rounded-sm", got)
	}
	// Default shape (empty) falls through to circle.
	if got := (Config{}).shapeClasses(); got != "rounded-full" {
		t.Errorf("default shapeClasses = %q, want rounded-full", got)
	}
}

func TestBorderClasses(t *testing.T) {
	if got := (Config{Border: false}).borderClasses(); got != "" {
		t.Errorf("borderClasses with Border=false = %q, want empty", got)
	}
	// Explicit color is used verbatim.
	if got := (Config{Border: true, BorderColor: "border-pink"}).borderClasses(); !strings.Contains(got, "border-pink") {
		t.Errorf("borderClasses should use explicit color, got %q", got)
	}
	// Tone-derived default colors.
	cases := map[Tone]string{
		ToneInfo:    "border-info",
		ToneSuccess: "border-success",
		ToneWarning: "border-warning",
		ToneDanger:  "border-danger",
		ToneDefault: "border-primary", // default branch
		TonePrimary: "border-primary",
	}
	for v, want := range cases {
		got := (Config{Border: true, Tone: v}).borderClasses()
		if !strings.Contains(got, want) {
			t.Errorf("borderClasses(variant=%q) = %q, want substring %q", v, got, want)
		}
		if !strings.Contains(got, "border-2") || !strings.Contains(got, "p-0.5") {
			t.Errorf("borderClasses(variant=%q) missing border-2/p-0.5: %q", v, got)
		}
	}
}

// TestRenderStatusIndicator covers the avatarWithStatus templ path, which is
// only reached when Status is non-empty.
func TestRenderStatusIndicator(t *testing.T) {
	html := renderAvatar(t, Config{
		Initials: "JS",
		Status:   StatusSuccess,
		Size:     SizeLG,
	})
	for _, want := range []string{
		"bg-success",         // statusClasses
		"size-5",             // statusSizeClasses for LG
		"bottom-0.5 end-0",   // position class
		`aria-hidden="true"`, // status dot is decorative
		"absolute rounded-full border-2",
	} {
		if !strings.Contains(html, want) {
			t.Errorf("status avatar missing %q in:\n%s", want, html)
		}
	}
}

// TestRenderReactiveStatusIndicator covers the reactive branch of avatarWithStatus
// where the status-dot size class is deferred to the Alpine scope.
func TestRenderReactiveStatusIndicator(t *testing.T) {
	html := renderAvatar(t, Config{
		Initials: "JS",
		Status:   StatusDanger,
		Reactive: true,
	})
	if !strings.Contains(html, `x-bind:class="avatarStatusSizeClass"`) {
		t.Errorf("reactive status avatar should bind avatarStatusSizeClass:\n%s", html)
	}
	if !strings.Contains(html, "bg-danger") {
		t.Errorf("reactive status avatar missing bg-danger:\n%s", html)
	}
}

// TestRenderWithIcon covers the Icon branch of layerInitials and userIcon.
func TestRenderWithIcon(t *testing.T) {
	html := renderAvatar(t, Config{
		Icon: userIcon(),
		Tone: TonePrimary,
	})
	for _, want := range []string{
		"<svg",           // icon SVG rendered
		"M7.5 6a4.5 4.5", // userIcon path fragment
		"bg-primary",     // variant applied
	} {
		if !strings.Contains(html, want) {
			t.Errorf("icon avatar missing %q in:\n%s", want, html)
		}
	}
	// Icon replaces initials text — no "?" fallback rendered.
	if strings.Contains(html, ">?<") {
		t.Errorf("icon avatar should not render initials fallback:\n%s", html)
	}
}

// TestRenderSrcExprMode covers the SrcExpr branches across layerImage,
// layerInitials and layerLoading.
func TestRenderSrcExprMode(t *testing.T) {
	html := renderAvatar(t, Config{
		SrcExpr:  "avatarSrc",
		Initials: "JS",
	})
	for _, want := range []string{
		`x-bind:src="avatarSrc"`,
		"!(avatarSrc) || !imgLoaded || imgError", // layerInitials x-show
		// templ HTML-escapes `&&` to `&amp;&amp;` inside the attribute value.
		"(avatarSrc) &amp;&amp; !imgLoaded &amp;&amp; !imgError", // layerLoading x-show
		"if (avatarSrc) { imgError = false; }",                   // layerImage x-effect
	} {
		if !strings.Contains(html, want) {
			t.Errorf("SrcExpr avatar missing %q in:\n%s", want, html)
		}
	}
}

// TestRenderStaticImage covers the static-Src branch of layerImage and the
// non-overlay x-show on layerInitials/layerLoading.
func TestRenderStaticImage(t *testing.T) {
	html := renderAvatar(t, Config{
		Src:      "/photo.webp",
		Alt:      "A person",
		Initials: "JS",
	})
	for _, want := range []string{
		`src="/photo.webp"`,
		`alt="A person"`,
		`x-show="!imgLoaded || imgError"`,  // layerInitials static branch
		`x-show="!imgLoaded && !imgError"`, // layerLoading static branch
		"x-data=\"{ imgLoaded: false, imgError: false }\"",
	} {
		if !strings.Contains(html, want) {
			t.Errorf("static image avatar missing %q in:\n%s", want, html)
		}
	}
}

// TestRenderReactiveSizeOnly covers the Reactive-but-not-ReactiveRadius branch
// of avatarLayers (bindClass == "avatarSizeClass").
func TestRenderReactiveSizeOnly(t *testing.T) {
	html := renderAvatar(t, Config{Initials: "JS", Reactive: true})
	if !strings.Contains(html, `x-bind:class="avatarSizeClass"`) {
		t.Errorf("reactive-size avatar should bind avatarSizeClass:\n%s", html)
	}
	// Size class must be omitted from the baked container class string.
	if strings.Contains(html, "size-14 text-2xl") {
		t.Errorf("reactive avatar should omit baked size class:\n%s", html)
	}
}

// TestRenderReactiveRadiusOnly covers the ReactiveRadius-but-not-Reactive branch.
func TestRenderReactiveRadiusOnly(t *testing.T) {
	html := renderAvatar(t, Config{
		Initials:       "JS",
		Shape:          ShapeSquare,
		ReactiveRadius: true,
	})
	if !strings.Contains(html, `x-bind:class="avatarRadiusClass"`) {
		t.Errorf("reactive-radius avatar should bind avatarRadiusClass:\n%s", html)
	}
}

// TestRenderBorderUsesFillLayer covers the Border branch in avatarLayers where
// the root skin is cleared and toneFillClasses lands on the inner layer.
func TestRenderBorderUsesFillLayer(t *testing.T) {
	html := renderAvatar(t, Config{
		Initials: "JS",
		Tone:     TonePrimary,
		Border:   true,
	})
	if !strings.Contains(html, "bg-primary") {
		t.Errorf("bordered avatar should render fill class on layer:\n%s", html)
	}
	if !strings.Contains(html, "border-2") {
		t.Errorf("bordered avatar missing border-2:\n%s", html)
	}
}

// TestAvatarStackReactivePropagates verifies StackConfig.Reactive flows to items
// and the default label is applied when Label is empty.
func TestAvatarStackReactivePropagates(t *testing.T) {
	var buf bytes.Buffer
	err := AvatarStack(StackConfig{
		Reactive: true,
		Items:    []Config{{Initials: "AA"}, {Initials: "BB"}},
	}).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("render stack: %v", err)
	}
	html := buf.String()
	if !strings.Contains(html, `aria-label="Avatar group"`) {
		t.Errorf("empty-label stack should default to 'Avatar group':\n%s", html)
	}
	if !strings.Contains(html, `x-bind:class="avatarSizeClass"`) {
		t.Errorf("reactive stack should propagate reactive size binding to items:\n%s", html)
	}
}
