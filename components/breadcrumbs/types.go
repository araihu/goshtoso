package breadcrumbs

import "github.com/a-h/templ"

// SeparatorStyle controls the separator between breadcrumb items
type SeparatorStyle string

const (
	// Chevron renders a ">" chevron SVG between items
	Chevron SeparatorStyle = "chevron"
	// Slash renders a "/" text separator between items
	Slash SeparatorStyle = "slash"
)

// Item represents a single breadcrumb link
type Item struct {
	// Label is the display text
	Label string
	// Href is the link URL
	Href string
	// Icon is an optional icon rendered before the label
	Icon templ.Component
	// Tooltip is an optional tooltip on the icon
	Tooltip string
	// LinkAttrs are extra HTML attributes on the <a> tag
	LinkAttrs templ.Attributes
}

// Config holds configuration for the breadcrumb component
type Config struct {
	// Items are the intermediate breadcrumb links (not the current page)
	Items []Item
	// Current is the label for the current page (rendered as bold text, no link)
	Current string
	// Separator controls the separator style (default: Chevron)
	Separator SeparatorStyle
	// NavClass allows additional CSS classes on the outer <nav>.
	NavClass string
	// NavAttrs are extra HTML attributes on the <nav> element
	NavAttrs templ.Attributes
}
