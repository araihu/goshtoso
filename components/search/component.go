package search

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable search component.
type Instance struct {
	cfg Config
}

// Search returns a renderable search component.
func Search(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as a search.
func (Instance) Kind() components.Kind {
	return components.KindSearch
}

// Render writes the search markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	return searchTemplate(i.cfg).Render(ctx, w)
}

// FieldInstance is a renderable search field component.
type FieldInstance struct {
	cfg Config
}

// SearchField returns a renderable search field component.
func SearchField(cfg Config) FieldInstance {
	return FieldInstance{cfg: cfg}
}

// Kind identifies the component as a search field.
func (FieldInstance) Kind() components.Kind {
	return components.KindSearchField
}

// Render writes the search field markup.
func (i FieldInstance) Render(ctx context.Context, w io.Writer) error {
	return searchFieldTemplate(i.cfg).Render(ctx, w)
}

// ModalInstance is a renderable search modal component.
type ModalInstance struct {
	cfg Config
}

// SearchModal returns a renderable search modal component.
func SearchModal(cfg Config) ModalInstance {
	return ModalInstance{cfg: cfg}
}

// Kind identifies the component as a search modal.
func (ModalInstance) Kind() components.Kind {
	return components.KindSearchModal
}

// Render writes the search modal markup.
func (i ModalInstance) Render(ctx context.Context, w io.Writer) error {
	return searchModalTemplate(i.cfg).Render(ctx, w)
}

var (
	_ components.Component = Instance{}
	_ components.Component = FieldInstance{}
	_ components.Component = ModalInstance{}
)
