package carousel

import (
	"encoding/base64"
	"encoding/json"
	"slices"

	"github.com/a-h/templ"
)

// Slide represents a single carousel slide
type Slide struct {
	// ImgSrc is the image URL
	ImgSrc string
	// ImgAlt is the image alt text
	ImgAlt string
	// Title is the optional slide heading.
	Title string
	// Description is the optional slide body text.
	Description string
	// CTAHref is the call-to-action link.
	CTAHref string
	// CTALabel is the call-to-action button label.
	CTALabel string
}

// AutoplayConfig enables automatic slide rotation
type AutoplayConfig struct {
	// Interval in milliseconds between slides (default 4000)
	Interval int
}

// HTMXConfig enables lazy loading of carousel content via HTMX
type HTMXConfig struct {
	// Get is the URL to fetch carousel content from (hx-get)
	Get string
	// Trigger controls when the request fires (hx-trigger, default "load")
	Trigger string
	// Swap controls how the response is inserted (hx-swap, default "innerHTML")
	Swap string
	// Indicator is a CSS selector for a loading indicator element (hx-indicator)
	Indicator string
}

// Config holds configuration for the Carousel component
type Config struct {
	// ID is a unique identifier for the carousel instance
	ID string
	// Slides are the static slide data (ignored if HTMX is set)
	Slides []Slide
	// Autoplay enables rotation in static mode; ignored when HTMX is non-nil.
	Autoplay *AutoplayConfig
	// Touch enables swipe gestures in static mode; ignored when HTMX is non-nil.
	Touch bool
	// AspectRatio sets static-mode sizing; ignored when HTMX is non-nil.
	AspectRatio string
	// Height overrides static-mode slide height; ignored when HTMX is non-nil.
	Height string
	// RootClass allows additional CSS classes on the container.
	RootClass string
	// HTMX enables lazy loading of carousel content (nil = static mode)
	HTMX *HTMXConfig
}

// CardConfig holds configuration for the card-framed carousel.
type CardConfig struct {
	// ID is a unique identifier for the carousel instance.
	ID string
	// Slides are the static slide data.
	Slides []Slide
	// Touch enables swipe gesture support.
	Touch bool
	// Height overrides the slides container height.
	Height string
	// RootClass allows additional CSS classes on the article container.
	RootClass string
}

func slideHasOverlay(slide Slide) bool {
	return slide.Title != "" ||
		slide.Description != "" ||
		(slide.CTALabel != "" && slide.CTAHref != "")
}

func hasOverlay(cfg Config) bool {
	return slices.ContainsFunc(cfg.Slides, slideHasOverlay)
}

// transitionAttr returns the duration suffix for Alpine's x-transition modifier
// (e.g. "1000ms"). Use with transitionAttrs() to build the full attribute name —
// Alpine's `x-transition.opacity.duration.500ms` modifier syntax requires the
// duration as part of the attribute key, not as a value (a value like "1000ms"
// gets evaluated as a JS expression and breaks).
func transitionAttr(cfg Config) string {
	if cfg.Touch {
		return "700ms"
	}
	return "1000ms"
}

// transitionAttrs builds a templ attribute map encoding Alpine's
// `x-transition.opacity.duration.<suffix>` modifier. A `true` value tells templ
// to render the attribute as a boolean (key only).
func transitionAttrs(suffix string) templ.Attributes {
	return templ.Attributes{
		"x-transition.opacity.duration." + suffix: true,
	}
}

// containerClasses returns the outer carousel container CSS classes
func containerClasses(cfg Config) string {
	base := "relative w-full overflow-hidden"
	if cfg.RootClass != "" {
		base += " " + cfg.RootClass
	}
	return base
}

// slidesContainerClasses returns the inner slides wrapper CSS classes
func slidesContainerClasses(cfg Config) string {
	if cfg.Height != "" {
		return "relative " + cfg.Height + " w-full"
	}
	if cfg.AspectRatio != "" {
		return "aspect-" + cfg.AspectRatio + " relative w-full"
	}
	return "relative min-h-[50svh] w-full"
}

// navButtonClasses returns CSS for prev/next navigation buttons
func navButtonClasses() string {
	return "absolute top-1/2 z-20 flex rounded-full -translate-y-1/2 items-center justify-center bg-surface/40 p-2 text-on-surface transition hover:bg-surface/60 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary active:outline-offset-0 dark:bg-surface-dark/40 dark:text-on-surface-dark dark:hover:bg-surface-dark/60 dark:focus-visible:outline-primary-dark"
}

// indicatorContainerClasses returns CSS for the indicators wrapper
func indicatorContainerClasses(cfg Config) string {
	base := "absolute rounded-radius bottom-3 md:bottom-5 left-1/2 z-20 flex -translate-x-1/2 gap-4 md:gap-3 px-1.5 py-1 md:px-2"
	if hasOverlay(cfg) || cfg.Autoplay != nil {
		// No background when text overlay is present (indicators over dark gradient)
		return base
	}
	return base + " bg-surface/75 dark:bg-surface-dark/75"
}

// indicatorActiveClasses returns CSS for the active indicator dot
func indicatorActiveClasses(cfg Config) string {
	if hasOverlay(cfg) || cfg.Autoplay != nil {
		return "bg-on-surface-dark"
	}
	return "bg-on-surface dark:bg-on-surface-dark"
}

// indicatorInactiveClasses returns CSS for inactive indicator dots
func indicatorInactiveClasses(cfg Config) string {
	if hasOverlay(cfg) || cfg.Autoplay != nil {
		return "bg-on-surface-dark/50"
	}
	return "bg-on-surface/50 dark:bg-on-surface-dark/50"
}

type slideData struct {
	ImgSrc      string `json:"imgSrc"`
	ImgAlt      string `json:"imgAlt"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	CTAURL      string `json:"ctaUrl,omitempty"`
	CTAText     string `json:"ctaText,omitempty"`
}

// slidesJSON serializes non-executable carousel data for the external factory.
func slidesJSON(slides []Slide) string {
	data := make([]slideData, len(slides))
	for i, slide := range slides {
		data[i] = slideData{
			ImgSrc:      slide.ImgSrc,
			ImgAlt:      slide.ImgAlt,
			Title:       slide.Title,
			Description: slide.Description,
		}
		if slide.CTAHref != "" && slide.CTALabel != "" {
			data[i].CTAURL = string(templ.URL(slide.CTAHref))
			data[i].CTAText = slide.CTALabel
		}
	}
	encoded, err := json.Marshal(data)
	if err != nil {
		return "[]"
	}
	return string(encoded)
}

func slidesData(slides []Slide) string {
	return base64.StdEncoding.EncodeToString([]byte(slidesJSON(slides)))
}

func autoplayInterval(cfg Config) int {
	if cfg.Autoplay == nil || cfg.Autoplay.Interval <= 0 {
		return 4000
	}
	return cfg.Autoplay.Interval
}

// cardContainerClasses returns CSS for the card wrapper.
func cardContainerClasses(rootClass string) string {
	base := "group flex max-w-sm flex-col overflow-hidden rounded-radius border border-outline bg-surface-alt text-on-surface dark:border-outline-dark dark:bg-surface-dark-alt dark:text-on-surface-dark"
	if rootClass != "" {
		base += " " + rootClass
	}
	return base
}
