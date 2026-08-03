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
			name:     "primary appearance",
			cfg:      Config{Appearance: AppearancePrimary},
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
		{
			name:     "pressed interaction",
			cfg:      Config{Interaction: InteractionPressed},
			contains: []string{"hover:translate-y-1.5", "active:translate-y-2", "motion-reduce:transform-none"},
			absent:   []string{"hover:-translate-y"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.cfg.containerClasses()
			for _, want := range tc.contains {
				if !strings.Contains(got, want) {
					t.Errorf("containerClasses() = %q, want substring %q", got, want)
				}
			}
			for _, no := range tc.absent {
				if strings.Contains(got, no) {
					t.Errorf("containerClasses() = %q, should not contain %q", got, no)
				}
			}
		})
	}
}

func TestCardRenderCustomMediaBeforeContent(t *testing.T) {
	media := templ.Raw(`<div data-card-media>Interactive project preview</div>`)
	html := render(t, Card(Config{
		Media:      media,
		MediaClass: "custom-media",
		Title:      "Project",
	}))

	mediaIndex := strings.Index(html, "data-card-media")
	titleIndex := strings.Index(html, "Project")
	if mediaIndex < 0 || titleIndex <= mediaIndex {
		t.Fatalf("custom media must render before content: %s", html)
	}
	if !strings.Contains(html, "custom-media") {
		t.Fatalf("custom media class missing: %s", html)
	}
}

func TestCardCustomMediaTakesPrecedenceOverImage(t *testing.T) {
	html := render(t, Card(Config{
		Image: "fallback.png",
		Media: templ.Raw(`<div data-card-media></div>`),
		Title: "Project",
	}))

	if strings.Contains(html, "fallback.png") {
		t.Fatalf("fallback image rendered with custom media: %s", html)
	}
}

func TestLayoutDependentClasses(t *testing.T) {
	vertical := Config{Layout: LayoutVertical}
	horizontal := Config{Layout: LayoutHorizontal}

	if got := vertical.imageContainerClasses(); !strings.Contains(got, "h-44") {
		t.Errorf("vertical imageContainerClasses() = %q, want h-44", got)
	}
	if got := horizontal.imageContainerClasses(); !strings.Contains(got, "col-span-3") {
		t.Errorf("horizontal imageContainerClasses() = %q, want col-span-3", got)
	}

	if got := vertical.imageClasses(); strings.Contains(got, "h-52") {
		t.Errorf("vertical imageClasses() = %q, should not contain h-52", got)
	}
	if got := horizontal.imageClasses(); !strings.Contains(got, "h-52") {
		t.Errorf("horizontal imageClasses() = %q, want h-52", got)
	}

	if got := vertical.contentClasses(); !strings.Contains(got, "gap-4") {
		t.Errorf("vertical contentClasses() = %q, want gap-4", got)
	}
	if got := horizontal.contentClasses(); !strings.Contains(got, "col-span-5") {
		t.Errorf("horizontal contentClasses() = %q, want col-span-5", got)
	}
}

func TestStaticClasses(t *testing.T) {
	var cfg Config
	if got := cfg.tagClasses(); !strings.Contains(got, "font-medium") {
		t.Errorf("tagClasses() = %q", got)
	}
	if got := cfg.titleClasses(); !strings.Contains(got, "font-bold") {
		t.Errorf("titleClasses() = %q", got)
	}
	if got := cfg.descriptionClasses(); !strings.Contains(got, "text-pretty") {
		t.Errorf("descriptionClasses() = %q", got)
	}
}

func TestPredicates(t *testing.T) {
	if (Config{}).hasImage() {
		t.Error("empty Image should be hasImage()=false")
	}
	if !(Config{Image: "x.png"}).hasImage() {
		t.Error("set Image should be hasImage()=true")
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

func TestCardRenderBodyBetweenDescriptionAndFooter(t *testing.T) {
	body := templ.Raw(`<ul data-card-body><li>First detail</li></ul>`)
	footer := templ.Raw(`<button data-card-footer>Continue</button>`)
	html := render(t, Card(Config{
		Title:       "T",
		Description: "Summary",
		Body:        body,
		Footer:      footer,
	}))

	descriptionIndex := strings.Index(html, "Summary")
	bodyIndex := strings.Index(html, "data-card-body")
	footerIndex := strings.Index(html, "data-card-footer")
	if descriptionIndex < 0 || bodyIndex <= descriptionIndex || footerIndex <= bodyIndex {
		t.Fatalf("card content order = description:%d body:%d footer:%d, html: %s", descriptionIndex, bodyIndex, footerIndex, html)
	}
}

func TestCardRenderEscaping(t *testing.T) {
	html := render(t, Card(Config{Title: "<script>", Description: `"quote"`}))
	if strings.Contains(html, "<script>") {
		t.Errorf("title not escaped: %s", html)
	}
}
