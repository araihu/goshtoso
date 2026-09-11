package carousel

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/expressions"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable carousel component.
type Instance struct {
	expressionOverrides expressions.Carousel
	cfg                 Config
}

// Carousel returns a renderable carousel component.
func Carousel(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as a carousel.
func (Instance) Kind() components.Kind {
	return components.KindCarousel
}

// Render writes the carousel markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	ctx = expressions.With(ctx, expressions.Set{Carousel: i.expressionOverrides})
	return carouselTemplate(i.cfg).Render(ctx, w)
}

// CardCarouselInstance is a renderable card carousel component.
type CardCarouselInstance struct {
	expressionOverrides expressions.Carousel
	cfg                 CardConfig
}

// CardCarousel returns a renderable card carousel component.
func CardCarousel(cfg CardConfig) CardCarouselInstance {
	return CardCarouselInstance{cfg: cfg}
}

// Kind identifies the component as a card carousel.
func (CardCarouselInstance) Kind() components.Kind {
	return components.KindCardCarousel
}

// Render writes the card carousel markup.
func (i CardCarouselInstance) Render(ctx context.Context, w io.Writer) error {
	ctx = expressions.With(ctx, expressions.Set{Carousel: i.expressionOverrides})
	return cardCarouselTemplate(i.cfg).Render(ctx, w)
}

var (
	_ components.Component = Instance{}
	_ components.Component = CardCarouselInstance{}
)

// WithExpressions returns a copy with component-scoped text overrides. Existing
// explicit config labels take precedence. Empty fields inherit render defaults.
func (i Instance) WithExpressions(values expressions.Carousel) Instance {
	i.expressionOverrides = expressions.Merge(expressions.Set{Carousel: i.expressionOverrides}, expressions.Set{Carousel: values}).Carousel
	return i
}

// WithExpressions returns a copy with component-scoped text overrides. Existing
// explicit config labels take precedence. Empty fields inherit render defaults.
func (i CardCarouselInstance) WithExpressions(values expressions.Carousel) CardCarouselInstance {
	i.expressionOverrides = expressions.Merge(expressions.Set{Carousel: i.expressionOverrides}, expressions.Set{Carousel: values}).Carousel
	return i
}
