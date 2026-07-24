package table

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable table component.
type Instance struct {
	cfg Config
}

// Table returns a renderable table component.
func Table(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as a table.
func (Instance) Kind() components.Kind {
	return components.KindTable
}

// Render writes the table markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	return tableTemplate(i.cfg).Render(ctx, w)
}

// TableHeadContentInstance renders table head row content without its wrapper.
type TableHeadContentInstance struct {
	cfg Config
}

// TableHeadContent returns renderable table head row content.
func TableHeadContent(cfg Config) TableHeadContentInstance {
	return TableHeadContentInstance{cfg: cfg}
}

// Kind identifies the component as table head content.
func (TableHeadContentInstance) Kind() components.Kind {
	return components.KindTableHeadContent
}

// Render writes the table head row content.
func (i TableHeadContentInstance) Render(ctx context.Context, w io.Writer) error {
	return tableHeadContentTemplate(i.cfg).Render(ctx, w)
}

// TableRowsInstance renders table rows without a tbody wrapper.
type TableRowsInstance struct {
	cfg Config
}

// TableRows returns renderable table rows.
func TableRows(cfg Config) TableRowsInstance {
	return TableRowsInstance{cfg: cfg}
}

// Kind identifies the component as table rows.
func (TableRowsInstance) Kind() components.Kind {
	return components.KindTableRows
}

// Render writes the table rows.
func (i TableRowsInstance) Render(ctx context.Context, w io.Writer) error {
	return tableRowsTemplate(i.cfg).Render(ctx, w)
}

// TableRowInstance is a renderable table row.
type TableRowInstance struct {
	cfg Config
	row Row
}

// TableRow returns a renderable table row.
func TableRow(cfg Config, row Row) TableRowInstance {
	return TableRowInstance{cfg: cfg, row: row}
}

// Kind identifies the component as a table row.
func (TableRowInstance) Kind() components.Kind {
	return components.KindTableRow
}

// Render writes the table row markup.
func (i TableRowInstance) Render(ctx context.Context, w io.Writer) error {
	return tableRowTemplate(i.cfg, i.row).Render(ctx, w)
}

// TablePaginationNavInstance is a renderable table pagination nav.
type TablePaginationNavInstance struct {
	cfg Config
}

// TablePaginationNav returns a renderable table pagination nav.
func TablePaginationNav(cfg Config) TablePaginationNavInstance {
	return TablePaginationNavInstance{cfg: cfg}
}

// Kind identifies the component as a table pagination nav.
func (TablePaginationNavInstance) Kind() components.Kind {
	return components.KindTablePaginationNav
}

// Render writes the table pagination nav markup.
func (i TablePaginationNavInstance) Render(ctx context.Context, w io.Writer) error {
	return tablePaginationNavTemplate(i.cfg).Render(ctx, w)
}

// ImageCellInstance is a renderable table image cell.
type ImageCellInstance struct {
	imageURL string
	label    string
	detail   string
}

// ImageCell returns a renderable table image cell.
func ImageCell(imageURL string, label string, detail string) ImageCellInstance {
	return ImageCellInstance{
		imageURL: imageURL,
		label:    label,
		detail:   detail,
	}
}

// Kind identifies the component as a table image cell.
func (ImageCellInstance) Kind() components.Kind {
	return components.KindTableImageCell
}

// Render writes the table image cell markup.
func (i ImageCellInstance) Render(ctx context.Context, w io.Writer) error {
	return imageCellTemplate(i.imageURL, i.label, i.detail).Render(ctx, w)
}

var (
	_ components.Component = Instance{}
	_ components.Component = TableHeadContentInstance{}
	_ components.Component = TableRowsInstance{}
	_ components.Component = TableRowInstance{}
	_ components.Component = TablePaginationNavInstance{}
	_ components.Component = ImageCellInstance{}
)
