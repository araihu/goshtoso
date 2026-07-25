package carousel

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable carousel component.
type Instance struct {
	cfg Config
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
	return carouselTemplate(i.cfg).Render(ctx, w)
}

// CardCarouselInstance is a renderable card carousel component.
type CardCarouselInstance struct {
	cfg CardConfig
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
	return cardCarouselTemplate(i.cfg).Render(ctx, w)
}

var (
	_ components.Component = Instance{}
	_ components.Component = CardCarouselInstance{}
)
