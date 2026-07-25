package rating

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable rating component.
type Instance struct {
	cfg Config
}

// Rating returns a renderable rating component.
func Rating(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as a rating.
func (Instance) Kind() components.Kind {
	return components.KindRating
}

// Render writes the rating markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	return ratingTemplate(i.cfg).Render(ctx, w)
}

// DisplayInstance is a renderable rating display component.
type DisplayInstance struct {
	cfg DisplayConfig
}

// RatingDisplay returns a renderable rating display component.
func RatingDisplay(cfg DisplayConfig) DisplayInstance {
	return DisplayInstance{cfg: cfg}
}

// Kind identifies the component as a rating display.
func (DisplayInstance) Kind() components.Kind {
	return components.KindRatingDisplay
}

// Render writes the rating display markup.
func (i DisplayInstance) Render(ctx context.Context, w io.Writer) error {
	return ratingDisplayTemplate(i.cfg).Render(ctx, w)
}

var (
	_ components.Component = Instance{}
	_ components.Component = DisplayInstance{}
)
