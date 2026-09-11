package expressions

import (
	"context"
	"io"

	"github.com/a-h/templ"
)

type contextKey struct{}

// With overlays expressions on the current render context. Empty fields inherit
// earlier context values, then renderer defaults, then English. The set is copied;
// applications remain responsible for concurrency-safe captured function state.
func With(ctx context.Context, set Set) context.Context {
	return context.WithValue(ctx, contextKey{}, Merge(overrides(ctx), set))
}

func overrides(ctx context.Context) Set {
	set, _ := ctx.Value(contextKey{}).(Set)
	return set
}

// From returns the context's expressions with English fallbacks filled in.
func From(ctx context.Context) Set {
	return Merge(English(), overrides(ctx))
}

// Text selects an explicit component value before its inherited expression.
func Text(explicit, inherited string) string {
	if explicit != "" {
		return explicit
	}
	return inherited
}

// Renderer supplies application defaults beneath request and component overrides.
// A renderer can be shared by concurrent requests if its message functions are
// safe for concurrent use. Updating configuration means constructing a new renderer.
type Renderer struct {
	defaults Set
}

// NewRenderer captures application defaults by value. The zero Renderer uses English.
func NewRenderer(defaults Set) Renderer {
	return Renderer{defaults: defaults}
}

// Render writes component using request expressions over application defaults.
// Cancellation, deadlines, and other values in ctx are preserved.
func (r Renderer) Render(ctx context.Context, w io.Writer, component templ.Component) error {
	ctx = context.WithValue(ctx, contextKey{}, Merge(r.defaults, overrides(ctx)))
	return component.Render(ctx, w)
}
