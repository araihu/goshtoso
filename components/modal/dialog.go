package modal

import (
	"context"
	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/components"
	"github.com/araihu/goshtoso/expressions"
	"io"
)

// DialogConfig configures an event-controlled native dialog with arbitrary content.
type DialogConfig struct {
	// ID identifies modal:open and modal:close events through their detail.id.
	ID string
	// Title labels the dialog for assistive technology.
	Title string
	// Content is the dialog body, including any application-owned controls.
	Content templ.Component
	// Compact renders a headerless, unpadded dialog near the top of the viewport.
	// Title remains available to assistive technology. Compact overrides FullscreenMobile.
	Compact bool
	// FullscreenMobile fills the viewport below the sm breakpoint.
	FullscreenMobile bool
}

// DialogInstance is a renderable content dialog.
type DialogInstance struct {
	expressionOverrides expressions.Modal
	cfg                 DialogConfig
}

// Dialog renders arbitrary content in a native dialog above other overlays.
// Dispatch modal:open or modal:close with detail.id matching Config.ID.
func Dialog(cfg DialogConfig) DialogInstance { return DialogInstance{cfg: cfg} }

// Kind identifies this component as a modal.
func (DialogInstance) Kind() components.Kind { return components.KindModal }

// Render writes the dialog markup.
func (i DialogInstance) Render(ctx context.Context, w io.Writer) error {
	ctx = expressions.With(ctx, expressions.Set{Modal: i.expressionOverrides})
	return dialogTemplate(i.cfg).Render(ctx, w)
}

var _ components.Component = DialogInstance{}

// WithExpressions returns a copy with component-scoped expression overrides.
func (i DialogInstance) WithExpressions(values expressions.Modal) DialogInstance {
	i.expressionOverrides = expressions.Merge(expressions.Set{Modal: i.expressionOverrides}, expressions.Set{Modal: values}).Modal
	return i
}
