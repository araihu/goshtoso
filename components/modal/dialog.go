package modal

import (
	"context"
	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/components"
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
	// FullscreenMobile fills the viewport below the sm breakpoint.
	FullscreenMobile bool
}

// DialogInstance is a renderable content dialog.
type DialogInstance struct{ cfg DialogConfig }

// Dialog renders arbitrary content in a native dialog above other overlays.
// Dispatch modal:open or modal:close with detail.id matching Config.ID.
func Dialog(cfg DialogConfig) DialogInstance { return DialogInstance{cfg: cfg} }

// Kind identifies this component as a modal.
func (DialogInstance) Kind() components.Kind { return components.KindModal }

// Render writes the dialog markup.
func (i DialogInstance) Render(ctx context.Context, w io.Writer) error {
	return dialogTemplate(i.cfg).Render(ctx, w)
}

var _ components.Component = DialogInstance{}
