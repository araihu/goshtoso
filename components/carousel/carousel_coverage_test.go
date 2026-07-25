package carousel

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestCoverageCarouselConfigHelpers(t *testing.T) {
	if !hasOverlay(Config{Slides: []Slide{{Title: "Title"}}}) ||
		!hasOverlay(Config{Slides: []Slide{{Description: "Description"}}}) ||
		!hasOverlay(Config{Slides: []Slide{{CTALabel: "Open", CTAHref: "/open"}}}) {
		t.Fatal("slide content should render overlays")
	}
	if hasOverlay(Config{}) || hasOverlay(Config{Slides: []Slide{{CTALabel: "Missing href"}}}) {
		t.Fatal("empty slides and incomplete CTAs should not render overlays")
	}

	helperCases := []struct {
		name string
		got  string
		want string
	}{
		{"default transition", transitionAttr(Config{}), "1000ms"},
		{"touch transition", transitionAttr(Config{Touch: true}), "700ms"},
		{"root class", containerClasses(Config{RootClass: "custom-root"}), "relative w-full overflow-hidden custom-root"},
		{"default slides classes", slidesContainerClasses(Config{}), "relative min-h-[50svh] w-full"},
		{"height slides classes", slidesContainerClasses(Config{Height: "h-32"}), "relative h-32 w-full"},
		{"aspect slides classes", slidesContainerClasses(Config{AspectRatio: "3/1"}), "aspect-3/1 relative w-full"},
		{"nav button classes", navButtonClasses(), "absolute top-1/2 z-20 flex rounded-full -translate-y-1/2 items-center justify-center bg-surface/40 p-2 text-on-surface transition hover:bg-surface/60 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary active:outline-offset-0 dark:bg-surface-dark/40 dark:text-on-surface-dark dark:hover:bg-surface-dark/60 dark:focus-visible:outline-primary-dark"},
		{"default indicators", indicatorContainerClasses(Config{}), "absolute rounded-radius bottom-3 md:bottom-5 left-1/2 z-20 flex -translate-x-1/2 gap-4 md:gap-3 px-1.5 py-1 md:px-2 bg-surface/75 dark:bg-surface-dark/75"},
		{"overlay indicators", indicatorContainerClasses(Config{Slides: []Slide{{Title: "Title"}}}), "absolute rounded-radius bottom-3 md:bottom-5 left-1/2 z-20 flex -translate-x-1/2 gap-4 md:gap-3 px-1.5 py-1 md:px-2"},
		{"default active indicator", indicatorActiveClasses(Config{}), "bg-on-surface dark:bg-on-surface-dark"},
		{"autoplay active indicator", indicatorActiveClasses(Config{Autoplay: &AutoplayConfig{}}), "bg-on-surface-dark"},
		{"default inactive indicator", indicatorInactiveClasses(Config{}), "bg-on-surface/50 dark:bg-on-surface-dark/50"},
		{"overlay inactive indicator", indicatorInactiveClasses(Config{Slides: []Slide{{Description: "Description"}}}), "bg-on-surface-dark/50"},
		{"card container", cardContainerClasses(""), "group flex max-w-sm flex-col overflow-hidden rounded-radius border border-outline bg-surface-alt text-on-surface dark:border-outline-dark dark:bg-surface-dark-alt dark:text-on-surface-dark"},
		{"card root class", cardContainerClasses("custom-card"), "group flex max-w-sm flex-col overflow-hidden rounded-radius border border-outline bg-surface-alt text-on-surface dark:border-outline-dark dark:bg-surface-dark-alt dark:text-on-surface-dark custom-card"},
	}
	for _, tc := range helperCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Fatalf("got %q, want %q", tc.got, tc.want)
			}
		})
	}

	attrs := transitionAttrs("700ms")
	if got := attrs["x-transition.opacity.duration.700ms"]; got != true {
		t.Fatalf("transitionAttrs missing boolean Alpine transition attribute: %#v", attrs)
	}
}

func TestCoverageGenerateAlpineDataIncludesOptionalBehavior(t *testing.T) {
	data := generateAlpineData(Config{
		Slides: []Slide{{ImgSrc: "/one.webp", ImgAlt: "One"}},
		Autoplay: &AutoplayConfig{
			Interval: 0,
		},
		Touch: true,
	})

	for _, want := range []string{
		"slides:[{imgSrc:'/one.webp',imgAlt:'One'}]",
		"currentSlideIndex:1",
		"previous:function()",
		"next:function()",
		"autoplayIntervalTime:4000",
		"isPaused:false",
		"setAutoplayInterval:function(t)",
		"touchStartX:null",
		"handleTouchStart:function(e)",
		"handleTouchEnd:function()",
	} {
		if !strings.Contains(data, want) {
			t.Fatalf("Alpine data missing %q in:\n%s", want, data)
		}
	}
}

func TestCoverageRenderDefaultCarousel(t *testing.T) {
	html := renderCoverageCarousel(t, Config{
		ID: "coverage-default",
		Slides: []Slide{
			{ImgSrc: "/assets/one.webp", ImgAlt: "Slide one"},
			{ImgSrc: "/assets/two.webp", ImgAlt: "Slide two"},
		},
		Touch:       true,
		AspectRatio: "3/1",
		RootClass:   "coverage-root",
	})

	for _, want := range []string{
		`id="coverage-default"`,
		`x-data=`,
		`coverage-root`,
		`aspect-3/1 relative w-full`,
		`x-on:touchstart="handleTouchStart($event)"`,
		`aria-label="previous slide"`,
		`aria-label="next slide"`,
		`role="group"`,
		`aria-label="slides"`,
		`x-transition.opacity.duration.700ms`,
		`x-bind:class="[currentSlideIndex === index + 1 ? &#39;bg-on-surface dark:bg-on-surface-dark&#39; : &#39;bg-on-surface/50 dark:bg-on-surface-dark/50&#39;]"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("default carousel render missing %q in:\n%s", want, html)
		}
	}
}

func TestCoverageRenderInferredSlideOverlays(t *testing.T) {
	textHTML := renderCoverageCarousel(t, Config{
		ID:     "coverage-text",
		Slides: coverageSlides(),
		Height: "h-64",
	})
	for _, want := range []string{
		`id="coverage-text"`,
		`relative h-64 w-full`,
		`x-transition.opacity.duration.1000ms`,
		`x-text="slide.title"`,
		`x-text="slide.description"`,
		`bg-linear-to-t from-surface-dark/85`,
		`x-bind:class="[currentSlideIndex === index + 1 ? &#39;bg-on-surface-dark&#39; : &#39;bg-on-surface-dark/50&#39;]"`,
	} {
		if !strings.Contains(textHTML, want) {
			t.Fatalf("text carousel render missing %q in:\n%s", want, textHTML)
		}
	}
	for _, want := range []string{
		`x-bind:href="slide.ctaUrl"`,
		`x-text="slide.ctaText"`,
		`x-if="slide.ctaText && slide.ctaUrl"`,
	} {
		if !strings.Contains(textHTML, want) {
			t.Fatalf("CTA carousel render missing %q in:\n%s", want, textHTML)
		}
	}
}

func TestCoverageRenderAutoplayCardAndHTMXCarousels(t *testing.T) {
	autoplayHTML := renderCoverageCarousel(t, Config{
		ID:       "coverage-autoplay",
		Slides:   coverageSlides(),
		Autoplay: &AutoplayConfig{Interval: 2500},
	})
	for _, want := range []string{
		`id="coverage-autoplay"`,
		`x-init="autoplay"`,
		`autoplayIntervalTime:2500`,
		`aria-label="pause carousel"`,
		`x-bind:aria-pressed="isPaused"`,
		`setAutoplayInterval(autoplayIntervalTime)`,
	} {
		if !strings.Contains(autoplayHTML, want) {
			t.Fatalf("autoplay carousel render missing %q in:\n%s", want, autoplayHTML)
		}
	}

	cardHTML := renderStructuralCarousel(t, CardCarousel(CardConfig{
		ID:        "coverage-card",
		Slides:    coverageSlides(),
		RootClass: "coverage-card-root",
		Touch:     true,
	}))
	for _, want := range []string{
		`<article`,
		`id="coverage-card"`,
		`style="-webkit-mask-image: -webkit-radial-gradient(white, black)"`,
		`relative h-48 lg:h-64 w-full`,
		`x-transition.opacity.duration.300ms`,
		`coverage-card-root`,
		`x-on:touchstart="handleTouchStart($event)"`,
		`x-text="slides[currentSlideIndex - 1]?.title || ''"`,
		`x-text="slides[currentSlideIndex - 1]?.description || ''"`,
	} {
		if !strings.Contains(cardHTML, want) {
			t.Fatalf("card carousel render missing %q in:\n%s", want, cardHTML)
		}
	}

	htmxHTML := renderCoverageCarousel(t, Config{
		ID:        "coverage-htmx",
		RootClass: "coverage-loader",
		HTMX: &HTMXConfig{
			Get:       "/api/carousel",
			Indicator: "#loading",
		},
	})
	for _, want := range []string{
		`id="coverage-htmx"`,
		`coverage-loader min-h-[50svh] flex items-center justify-center`,
		`hx-get="/api/carousel"`,
		`hx-trigger="load"`,
		`hx-swap="innerHTML"`,
		`hx-indicator="#loading"`,
		`Loading carousel...`,
	} {
		if !strings.Contains(htmxHTML, want) {
			t.Fatalf("HTMX carousel render missing %q in:\n%s", want, htmxHTML)
		}
	}
}

func renderCoverageCarousel(t *testing.T, cfg Config) string {
	t.Helper()

	var buf bytes.Buffer
	if err := Carousel(cfg).Render(context.Background(), &buf); err != nil {
		t.Fatalf("render carousel: %v", err)
	}
	return buf.String()
}

func coverageSlides() []Slide {
	return []Slide{
		{
			ImgSrc:      "/assets/one.webp",
			ImgAlt:      "Slide one",
			Title:       "First slide",
			Description: "First description",
			CTAHref:     "/first",
			CTALabel:    "Open first",
		},
		{
			ImgSrc:      "/assets/two.webp",
			ImgAlt:      "Slide two",
			Title:       "Second slide",
			Description: "Second description",
			CTAHref:     "/second",
			CTALabel:    "Open second",
		},
	}
}
