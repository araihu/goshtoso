package carousel

import (
	"fmt"
	"strings"

	"github.com/a-h/templ"
)

// Variant represents carousel visual variants
type Variant string

const (
	// Default shows images only with prev/next buttons and indicators
	Default Variant = "default"
	// WithText adds title + description overlay with gradient background
	WithText Variant = "with-text"
	// WithCTA adds title + description + call-to-action button
	WithCTA Variant = "with-cta"
	// OnCard wraps the carousel in an article card with product info below
	OnCard Variant = "on-card"
)

// Slide represents a single carousel slide
type Slide struct {
	// ImgSrc is the image URL
	ImgSrc string
	// ImgAlt is the image alt text
	ImgAlt string
	// Title is the slide heading (used by WithText, WithCTA)
	Title string
	// Description is the slide body text (used by WithText, WithCTA)
	Description string
	// CTAUrl is the call-to-action link (used by WithCTA)
	CTAUrl string
	// CTAText is the call-to-action button label (used by WithCTA)
	CTAText string
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
	// Variant determines the visual style
	Variant Variant
	// Autoplay enables automatic slide rotation (nil = disabled)
	Autoplay *AutoplayConfig
	// Touch enables swipe gesture support
	Touch bool
	// AspectRatio sets a fixed aspect ratio (e.g. "3/1"), empty = min-h-[50svh]
	AspectRatio string
	// Height overrides the slides container height (e.g. "h-48 lg:h-64" for card variant)
	Height string
	// Class allows additional CSS classes on the container
	Class string
	// HTMX enables lazy loading of carousel content (nil = static mode)
	HTMX *HTMXConfig
}

// hasOverlay returns true if the variant shows text/CTA over the slides
func hasOverlay(v Variant) bool {
	return v == WithText || v == WithCTA
}

// transitionAttr returns the duration suffix for Alpine's x-transition modifier
// (e.g. "1000ms"). Use with transitionAttrs() to build the full attribute name —
// Alpine's `x-transition.opacity.duration.500ms` modifier syntax requires the
// duration as part of the attribute key, not as a value (a value like "1000ms"
// gets evaluated as a JS expression and breaks).
func transitionAttr(cfg Config) string {
	switch cfg.Variant {
	case OnCard:
		return "300ms"
	default:
		if cfg.Touch {
			return "700ms"
		}
		return "1000ms"
	}
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
	if cfg.Class != "" {
		base += " " + cfg.Class
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
	if hasOverlay(cfg.Variant) || cfg.Autoplay != nil {
		// No background when text overlay is present (indicators over dark gradient)
		return base
	}
	return base + " bg-surface/75 dark:bg-surface-dark/75"
}

// indicatorActiveClasses returns CSS for the active indicator dot
func indicatorActiveClasses(cfg Config) string {
	if hasOverlay(cfg.Variant) || cfg.Autoplay != nil {
		return "bg-on-surface-dark"
	}
	return "bg-on-surface dark:bg-on-surface-dark"
}

// indicatorInactiveClasses returns CSS for inactive indicator dots
func indicatorInactiveClasses(cfg Config) string {
	if hasOverlay(cfg.Variant) || cfg.Autoplay != nil {
		return "bg-on-surface-dark/50"
	}
	return "bg-on-surface/50 dark:bg-on-surface-dark/50"
}

// generateAlpineData creates the Alpine.js x-data object from config
func generateAlpineData(cfg Config) string {
	var b strings.Builder
	b.WriteString("{slides:")
	b.WriteString(slidesToJSON(cfg.Slides))
	b.WriteString(",currentSlideIndex:1,")

	// Navigation methods
	b.WriteString("previous:function(){if(this.currentSlideIndex>1){this.currentSlideIndex=this.currentSlideIndex-1}else{this.currentSlideIndex=this.slides.length}},")
	b.WriteString("next:function(){if(this.currentSlideIndex<this.slides.length){this.currentSlideIndex=this.currentSlideIndex+1}else{this.currentSlideIndex=1}}")

	// Autoplay state and methods
	if cfg.Autoplay != nil {
		interval := cfg.Autoplay.Interval
		if interval <= 0 {
			interval = 4000
		}
		fmt.Fprintf(&b, ",autoplayIntervalTime:%d,isPaused:false,autoplayInterval:null,", interval)
		b.WriteString("autoplay:function(){this.autoplayInterval=setInterval(()=>{if(!this.isPaused){this.next()}},this.autoplayIntervalTime)},")
		b.WriteString("setAutoplayInterval:function(t){clearInterval(this.autoplayInterval);this.autoplayIntervalTime=t;this.autoplay()}")
	}

	// Touch state and methods
	if cfg.Touch {
		b.WriteString(",touchStartX:null,touchEndX:null,swipeThreshold:50,")
		b.WriteString("handleTouchStart:function(e){this.touchStartX=e.touches[0].clientX},")
		b.WriteString("handleTouchMove:function(e){this.touchEndX=e.touches[0].clientX},")
		b.WriteString("handleTouchEnd:function(){if(this.touchEndX){if(this.touchStartX-this.touchEndX>this.swipeThreshold){this.next()}if(this.touchStartX-this.touchEndX<-this.swipeThreshold){this.previous()}this.touchStartX=null;this.touchEndX=null}}")
	}

	b.WriteString("}")
	return b.String()
}

// slidesToJSON serializes slides to a JavaScript array literal
func slidesToJSON(slides []Slide) string {
	var b strings.Builder
	b.WriteString("[")
	for i, s := range slides {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString("{")
		fmt.Fprintf(&b, "imgSrc:'%s',imgAlt:'%s'", jsEscape(s.ImgSrc), jsEscape(s.ImgAlt))
		if s.Title != "" {
			fmt.Fprintf(&b, ",title:'%s'", jsEscape(s.Title))
		}
		if s.Description != "" {
			fmt.Fprintf(&b, ",description:'%s'", jsEscape(s.Description))
		}
		if s.CTAUrl != "" {
			fmt.Fprintf(&b, ",ctaUrl:'%s'", jsEscape(s.CTAUrl))
		}
		if s.CTAText != "" {
			fmt.Fprintf(&b, ",ctaText:'%s'", jsEscape(s.CTAText))
		}
		b.WriteString("}")
	}
	b.WriteString("]")
	return b.String()
}

// jsEscape escapes a string for use inside JS single quotes
func jsEscape(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `'`, `\'`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	return s
}

// cardContainerClasses returns CSS for the card wrapper (OnCard variant)
func cardContainerClasses() string {
	return "group flex max-w-sm flex-col overflow-hidden rounded-radius border border-outline bg-surface-alt text-on-surface dark:border-outline-dark dark:bg-surface-dark-alt dark:text-on-surface-dark"
}
