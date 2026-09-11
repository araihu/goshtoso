package rating

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/expressions"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable rating component.
type Instance struct {
	expressionOverrides expressions.Rating
	cfg                 Config
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
	ctx = expressions.With(ctx, expressions.Set{Rating: i.expressionOverrides})
	return ratingTemplate(i.cfg).Render(ctx, w)
}

// DisplayInstance is a renderable rating display component.
type DisplayInstance struct {
	expressionOverrides expressions.Rating
	cfg                 DisplayConfig
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
	ctx = expressions.With(ctx, expressions.Set{Rating: i.expressionOverrides})
	return ratingDisplayTemplate(i.cfg).Render(ctx, w)
}

var (
	_ components.Component = Instance{}
	_ components.Component = DisplayInstance{}
)

// WithExpressions returns a copy with component-scoped text overrides. Existing
// explicit config labels take precedence. Empty fields inherit render defaults.
func (i Instance) WithExpressions(values expressions.Rating) Instance {
	i.expressionOverrides = expressions.Merge(expressions.Set{Rating: i.expressionOverrides}, expressions.Set{Rating: values}).Rating
	return i
}

// WithExpressions returns a copy with component-scoped text overrides. Existing
// explicit config labels take precedence. Empty fields inherit render defaults.
func (i DisplayInstance) WithExpressions(values expressions.Rating) DisplayInstance {
	i.expressionOverrides = expressions.Merge(expressions.Set{Rating: i.expressionOverrides}, expressions.Set{Rating: values}).Rating
	return i
}
