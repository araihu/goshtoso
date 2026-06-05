package card

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

func render(t *testing.T, c templ.Component) string {
	t.Helper()
	var buf bytes.Buffer
	if err := c.Render(context.Background(), &buf); err != nil {
		t.Fatalf("render: %v", err)
	}
	return buf.String()
}

func TestContainerClasses(t *testing.T) {
	tests := []struct {
		name     string
		cfg      Config
		contains []string
		absent   []string
	}{
		{
			name:     "default vertical",
			cfg:      Config{},
			contains: []string{"flex-col", "max-w-sm", "border-outline", "bg-surface-alt"},
			absent:   []string{"border-primary", "md:grid-cols-8"},
		},
		{
			name:     "primary variant",
			cfg:      Config{Variant: Primary},
			contains: []string{"border-2 border-primary", "dark:border-primary-dark"},
			absent:   []string{" border-outline "},
		},
		{
			name:     "horizontal layout",
			cfg:      Config{Layout: LayoutHorizontal},
			contains: []string{"max-w-2xl", "grid", "md:grid-cols-8"},
			absent:   []string{"max-w-sm", "flex-col"},
		},
		{
			name:     "root class appended",
			cfg:      Config{RootClass: "custom-card"},
			contains: []string{"custom-card"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.cfg.ContainerClasses()
			for _, want := range tc.contains {
				if !strings.Contains(got, want) {
					t.Errorf("ContainerClasses() = %q, want substring %q", got, want)
				}
			}
			for _, no := range tc.absent {
				if strings.Contains(got, no) {
					t.Errorf("ContainerClasses() = %q, should not contain %q", got, no)
				}
			}
		})
	}
}

func TestLayoutDependentClasses(t *testing.T) {
	vertical := Config{Layout: LayoutVertical}
	horizontal := Config{Layout: LayoutHorizontal}

	if got := vertical.ImageContainerClasses(); !strings.Contains(got, "h-44") {
		t.Errorf("vertical ImageContainerClasses() = %q, want h-44", got)
	}
	if got := horizontal.ImageContainerClasses(); !strings.Contains(got, "col-span-3") {
		t.Errorf("horizontal ImageContainerClasses() = %q, want col-span-3", got)
	}

	if got := vertical.ImageClasses(); strings.Contains(got, "h-52") {
		t.Errorf("vertical ImageClasses() = %q, should not contain h-52", got)
	}
	if got := horizontal.ImageClasses(); !strings.Contains(got, "h-52") {
		t.Errorf("horizontal ImageClasses() = %q, want h-52", got)
	}

	if got := vertical.ContentClasses(); !strings.Contains(got, "gap-4") {
		t.Errorf("vertical ContentClasses() = %q, want gap-4", got)
	}
	if got := horizontal.ContentClasses(); !strings.Contains(got, "col-span-5") {
		t.Errorf("horizontal ContentClasses() = %q, want col-span-5", got)
	}
}

func TestStaticClasses(t *testing.T) {
	var cfg Config
	if got := cfg.TagClasses(); !strings.Contains(got, "font-medium") {
		t.Errorf("TagClasses() = %q", got)
	}
	if got := cfg.TitleClasses(); !strings.Contains(got, "font-bold") {
		t.Errorf("TitleClasses() = %q", got)
	}
	if got := cfg.DescriptionClasses(); !strings.Contains(got, "text-pretty") {
		t.Errorf("DescriptionClasses() = %q", got)
	}
}

func TestPredicates(t *testing.T) {
	if (Config{}).HasImage() {
		t.Error("empty Image should be HasImage()=false")
	}
	if !(Config{Image: "x.png"}).HasImage() {
		t.Error("set Image should be HasImage()=true")
	}
	if (Config{}).HasRating() {
		t.Error("zero Rating should be HasRating()=false")
	}
	if !(Config{Rating: 1}).HasRating() {
		t.Error("positive Rating should be HasRating()=true")
	}
	if (Config{Rating: -1}).HasRating() {
		t.Error("negative Rating should be HasRating()=false")
	}
	if (Config{}).HasPrice() {
		t.Error("empty Price should be HasPrice()=false")
	}
	if !(Config{Price: "$1"}).HasPrice() {
		t.Error("set Price should be HasPrice()=true")
	}
}

func TestCardRenderDefault(t *testing.T) {
	html := render(t, Card(Config{Title: "Hello", Description: "World"}))
	if !strings.Contains(html, "<article") {
		t.Errorf("render missing <article>: %s", html)
	}
	if !strings.Contains(html, "Hello") || !strings.Contains(html, "World") {
		t.Errorf("render missing title/description: %s", html)
	}
	// aria-describedby links title to description id
	if !strings.Contains(html, `aria-describedby="Hello-desc"`) {
		t.Errorf("render missing aria-describedby link: %s", html)
	}
	if !strings.Contains(html, `id="Hello-desc"`) {
		t.Errorf("render missing description id: %s", html)
	}
	// no image when unset
	if strings.Contains(html, "<img") {
		t.Errorf("render should not include img when Image unset: %s", html)
	}
}

func TestCardRenderImageAndTag(t *testing.T) {
	html := render(t, Card(Config{
		Image:    "pic.png",
		ImageAlt: "alt text",
		Tag:      "News",
		Title:    "T",
	}))
	if !strings.Contains(html, `src="pic.png"`) {
		t.Errorf("missing img src: %s", html)
	}
	if !strings.Contains(html, `alt="alt text"`) {
		t.Errorf("missing img alt: %s", html)
	}
	if !strings.Contains(html, ">News<") {
		t.Errorf("missing tag text: %s", html)
	}
}

func TestCardRenderNoDescriptionNoTag(t *testing.T) {
	html := render(t, Card(Config{Title: "OnlyTitle"}))
	if strings.Contains(html, "<p ") {
		t.Errorf("should not render description paragraph: %s", html)
	}
	if strings.Contains(html, "<span") {
		t.Errorf("should not render tag span: %s", html)
	}
}

func TestCardRenderFooter(t *testing.T) {
	footer := templ.Raw(`<button>Buy</button>`)
	html := render(t, Card(Config{Title: "T", Footer: footer}))
	if !strings.Contains(html, "<button>Buy</button>") {
		t.Errorf("footer not rendered: %s", html)
	}
}

func TestCardRenderEscaping(t *testing.T) {
	html := render(t, Card(Config{Title: "<script>", Description: `"quote"`}))
	if strings.Contains(html, "<script>") {
		t.Errorf("title not escaped: %s", html)
	}
}

func TestStarRating(t *testing.T) {
	tests := []struct {
		rating     int
		wantFilled int
	}{
		{0, 0},
		{3, 3},
		{5, 5},
		{7, 5}, // clamps visually at 5 stars rendered
	}
	for _, tc := range tests {
		html := render(t, StarRating(tc.rating))
		filled := strings.Count(html, "text-amber-500")
		if filled != tc.wantFilled {
			t.Errorf("StarRating(%d) filled=%d, want %d", tc.rating, filled, tc.wantFilled)
		}
		total := strings.Count(html, "<svg")
		if total != 5 {
			t.Errorf("StarRating(%d) rendered %d stars, want 5", tc.rating, total)
		}
		if !strings.Contains(html, "Rated") {
			t.Errorf("StarRating(%d) missing sr-only label: %s", tc.rating, html)
		}
	}
}
