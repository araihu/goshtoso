package table

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/a-h/templ"
)

// Variant represents table style variants
type Variant string

const (
	Default      Variant = "default"
	Striped      Variant = "striped"
	WithCheckbox Variant = "checkbox"
)

// SortDir represents the sort direction
type SortDir string

const (
	SortAsc  SortDir = "asc"
	SortDesc SortDir = "desc"
	SortNone SortDir = ""
)

// Column defines a table column header
type Column struct {
	// Key is the column identifier used to look up cell values in Row.Cells
	Key string
	// Label is the display text for the column header
	Label string
	// Sortable marks this column as sortable (renders clickable header)
	Sortable bool
	// Width is an optional Tailwind width class (e.g. "w-32", "min-w-[200px]")
	Width string
	// Align is "left" (default), "center", or "right"
	Align string
}

// Cell holds the content for a single table cell
type Cell struct {
	// Text is plain text content
	Text string
	// Component is a templ component to render (overrides all other fields)
	Component templ.Component
	// Description renders a muted secondary line below Text (stacked layout for name+id pairs)
	Description string
	// BadgeColor wraps Text in a semi-solid badge. Accepts: "success", "danger", "warning", "info", "neutral", "primary", "secondary"
	BadgeColor string
	// Code renders Text in monospace style
	Code bool
}

// LinkMode controls how Row.Link navigates when clicked.
type LinkMode string

const (
	// LinkSPA swaps only #main-content-area via HTMX (default, fast, but HTMX history
	// cache may break Alpine.js state on back-button — see USAGE.md pitfalls).
	LinkSPA LinkMode = ""
	// LinkBoost swaps the full <body> via HTMX hx-select="body" (no white flash,
	// Alpine re-initializes cleanly, back-button re-fetches from server).
	LinkBoost LinkMode = "boost"
	// LinkFull uses a plain <a href> — full page navigation with no HTMX.
	// Safest for pages with complex Alpine state that breaks on history restore.
	LinkFull LinkMode = "full"
)

// Row represents a single table row
type Row struct {
	// ID is a unique identifier for the row (used for checkbox IDs)
	ID string
	// Cells maps column keys to cell content
	Cells map[string]Cell
	// Link makes the row clickable — navigates when clicked.
	// The navigation strategy is controlled by LinkMode.
	Link string
	// LinkMode controls how Link navigates. Default (empty) = SPA swap of #main-content-area.
	// Use LinkBoost for full-body swap, or LinkFull for plain navigation.
	LinkMode LinkMode
	// OnClick is a JS/Alpine expression executed on row click.
	// Use for opening modals, toggling state, etc.
	// Ignored when Link is set (Link takes precedence).
	OnClick string
	// HXGet triggers an HTMX GET request when the row is clicked.
	// Ignored when Link or OnClick is set.
	HXGet string
	// HXPost triggers an HTMX POST request when the row is clicked.
	// Ignored when Link, OnClick, or HXGet is set.
	HXPost string
	// HXTarget is the CSS selector for the HTMX swap target (used with HXGet/HXPost).
	HXTarget string
	// HXSwap is the HTMX swap strategy (used with HXGet/HXPost). Defaults to "innerHTML".
	HXSwap string
	// HXPushURL pushes the request URL into browser history when HXGet/HXPost
	// navigates. Off by default. Link-mode rows always push and ignore this.
	HXPushURL bool
	// AlpineAttrs is a pass-through for per-row Alpine directives (e.g.
	// `x-show`, `x-data`, `x-bind:class`). Keys become HTML attribute names
	// verbatim; values go through templ's attribute escaping. Use for
	// client-side row filters (x-show) that sit on top of a gotable.
	AlpineAttrs map[string]string
	// Expandable shows a chevron toggle and an expandable detail section below the row
	Expandable bool
	// Detail is rendered in the expanded panel below the row when Expandable is true
	Detail templ.Component
	// Actions is rendered in a trailing actions column (e.g., edit/delete buttons)
	Actions templ.Component
}

// IsActionable returns true if the row has any interactive behavior
// (link, click handler, HTMX action, expandable, or custom actions).
func (r Row) IsActionable() bool {
	return r.Link != "" || r.OnClick != "" || r.HXGet != "" || r.HXPost != "" || r.Expandable || r.Actions != nil
}

// ClickableRole returns the ARIA role to advertise for a clickable row, or
// "" if the row is not keyboard-activatable as a single unit. Rows whose
// only interactive surfaces are nested (Actions or Expandable-only without a
// row-level target) keep the default role and let the inner controls own
// focus. See table.templ for the tabindex/onkeydown pairing.
func (r Row) ClickableRole() string {
	if r.Link != "" {
		return "link"
	}
	if r.OnClick != "" || r.HXGet != "" || r.HXPost != "" {
		return "button"
	}
	return ""
}

// HasLinkedRows returns true if any row has a Link
func (cfg Config) HasLinkedRows() bool {
	for _, r := range cfg.Rows {
		if r.Link != "" {
			return true
		}
	}
	return false
}

// HasActionableRows returns true if any row is actionable (link, click, HTMX, expandable, or has actions).
func (cfg Config) HasActionableRows() bool {
	for _, r := range cfg.Rows {
		if r.IsActionable() {
			return true
		}
	}
	return false
}

// HasExpandableRows returns true if any row is expandable
func (cfg Config) HasExpandableRows() bool {
	for _, r := range cfg.Rows {
		if r.Expandable {
			return true
		}
	}
	return false
}

// HasActions returns true if any row has an Actions component
func (cfg Config) HasActions() bool {
	for _, r := range cfg.Rows {
		if r.Actions != nil {
			return true
		}
	}
	return false
}

// ColCount returns the total number of visible columns (columns + optional checkbox + optional actions/expand)
func (cfg Config) ColCount() int {
	n := len(cfg.Columns)
	if cfg.ShowCheckbox {
		n++
	}
	if cfg.HasActions() || cfg.HasExpandableRows() {
		n++
	}
	return n
}

// PaginationMode determines pagination behavior
type PaginationMode string

const (
	// PaginationTraditional renders page numbers below the table (default)
	PaginationTraditional PaginationMode = ""
	// PaginationInfiniteScroll appends rows on scroll. The sentinel's
	// IntersectionObserver attaches to the nearest ancestor scroller (usually
	// the page's main content area) by default — the table itself does NOT
	// create its own inner scroller. Set `PaginationConfig.ContainerHeight`
	// only when the host page genuinely needs a bounded, table-local scroll
	// (e.g. a dashboard with multiple independently-scrollable widgets).
	// Default = page-scroll (one scrollbar, idiomatic for full-page tables).
	PaginationInfiniteScroll PaginationMode = "infinite"
)

// PaginationConfig holds pagination state
type PaginationConfig struct {
	// Mode is the pagination style: traditional (default) or infinite scroll
	Mode PaginationMode
	// CurrentPage is the 1-indexed current page number
	CurrentPage int
	// TotalPages is the total number of pages
	TotalPages int
	// PerPage is the number of items per page
	PerPage int
	// HasMore indicates if more rows are available (used by infinite scroll)
	HasMore bool
	// ContainerHeight opts into Pattern A for infinite scroll: the table wraps
	// itself in a max-height + overflow-y-auto container and scrolls
	// internally. Leave empty (the default) for Pattern B, where the sentinel
	// row reveals against the nearest ancestor scroller. Values are raw CSS
	// (e.g. "400px", "60vh"). Only consulted when Mode is
	// PaginationInfiniteScroll.
	ContainerHeight string
}

// IsInfiniteScroll returns true if this pagination uses infinite scroll mode
func (p *PaginationConfig) IsInfiniteScroll() bool {
	return p != nil && p.Mode == PaginationInfiniteScroll
}

// IsContained returns true when the infinite-scroll table should render its
// own capped scroll container (Pattern A). False means the sentinel reveals
// against the nearest ancestor scroller (Pattern B — the default). Used by
// the template to decide whether to emit `max-height + overflow-y: auto`.
func (p *PaginationConfig) IsContained() bool {
	return p != nil && p.Mode == PaginationInfiniteScroll && p.ContainerHeight != ""
}

// NextPage returns CurrentPage + 1
func (p *PaginationConfig) NextPage() int {
	if p == nil {
		return 2
	}
	return p.CurrentPage + 1
}

// GetContainerHeight returns the container height, defaulting to "400px"
func (p *PaginationConfig) GetContainerHeight() string {
	if p == nil || p.ContainerHeight == "" {
		return "400px"
	}
	return p.ContainerHeight
}

// InfiniteScrollConfig holds infinite scroll state (deprecated: use PaginationConfig with Mode)
type InfiniteScrollConfig struct {
	// NextPage is the next page number to load
	NextPage int
	// HasMore indicates if more rows are available
	HasMore bool
}

// FilterType represents the kind of filter control
type FilterType string

const (
	FilterSearch FilterType = "search" // free-text input
	FilterSelect FilterType = "select" // dropdown with options
	FilterToggle FilterType = "toggle" // boolean switch
)

// FilterOption represents a single option in a select filter
type FilterOption struct {
	Value string
	Label string
}

// Filter defines a single filter control in the filter bar
type Filter struct {
	// Key is the query parameter name sent to the server
	Key string
	// Label is the display text
	Label string
	// Type is the control type (search, select, toggle)
	Type FilterType
	// Placeholder for search/select inputs
	Placeholder string
	// Options for select-type filters (static)
	Options []FilterOption
	// HTMXOptionsURL loads select options dynamically via HTMX on load
	HTMXOptionsURL string
	// DefaultValue is the initial value
	DefaultValue string
}

// FilterVariant switches the filter bar layout. FilterVariantBar (the
// default, empty string) is the full-width bordered block with an optional
// collapsible header — suitable for primary page tables. FilterVariantInline
// drops the border, collapsible toggle, and padding wrapper so the filter
// controls render as a plain flex row; suitable for modals and tight chrome
// where the bar's own chrome would compete with the host's.
type FilterVariant string

const (
	FilterVariantBar    FilterVariant = ""
	FilterVariantInline FilterVariant = "inline"
)

// FilterConfig holds the filter bar configuration
type FilterConfig struct {
	// Filters is the list of filter controls
	Filters []Filter
	// Collapsible enables a toggle to show/hide the filter bar.
	// Ignored when Variant is FilterVariantInline.
	Collapsible bool
	// InitiallyExpanded controls whether filters start visible (default: true).
	// Ignored when Variant is FilterVariantInline.
	InitiallyExpanded bool
	// Variant selects the layout (bar vs inline). See FilterVariant.
	Variant FilterVariant
	// HxTarget overrides the CSS selector that filter changes swap into.
	// Default (empty) resolves to "#{tbody-id}". Use when the table sits
	// inside a larger fragment that should be re-rendered as a unit — for
	// example, a modal body that includes both the table and surrounding
	// header text, or a cluster-picker card that needs pager state reset.
	HxTarget string
	// HxSwap overrides the htmx swap strategy used when filters apply.
	// Default (empty) resolves to "innerHTML". Use "outerHTML" when the
	// HxTarget is itself the wrapper that the server re-renders (e.g. a
	// catalog grid whose empty-state lives on the wrapper element).
	HxSwap string
}

// ResolvedHxTarget returns the CSS selector that filter changes should swap
// into. Caller-provided HxTarget wins; falls back to "#{tbody-id}".
func (c *FilterConfig) ResolvedHxTarget(cfg Config) string {
	if c != nil && c.HxTarget != "" {
		return c.HxTarget
	}
	return "#" + cfg.TbodyID()
}

// ResolvedHxSwap returns the htmx swap strategy filter requests should use.
// Caller-provided HxSwap wins; falls back to "innerHTML".
func (c *FilterConfig) ResolvedHxSwap() string {
	if c != nil && c.HxSwap != "" {
		return c.HxSwap
	}
	return "innerHTML"
}

// Config holds configuration for the table component
type Config struct {
	// ID is the table element ID
	ID string
	// Caption is a screen-reader-only description of the table (rendered as <caption class="sr-only">).
	Caption string
	// Columns defines the table headers
	Columns []Column
	// Rows holds the table data
	Rows []Row
	// Variant determines the table style
	Variant Variant
	// ShowCheckbox adds a select-all checkbox column
	ShowCheckbox bool
	// Class allows additional CSS classes on the container
	Class string

	// --- Sorting ---
	// SortBy is the currently sorted column key
	SortBy string
	// SortDir is the current sort direction ("asc" or "desc")
	SortDir SortDir

	// --- HTMX Integration ---
	// HTMXEndpoint is the base URL for HTMX requests (sorting, pagination, lazy load)
	HTMXEndpoint string
	// HTMXTarget overrides the default HTMX swap target (defaults to tbody ID)
	HTMXTarget string

	// --- Lazy Loading ---
	// LazyLoad loads the table body via HTMX.
	LazyLoad bool
	// LazyTrigger is the htmx trigger for LazyLoad. Defaults to "load".
	// Use values like "click from:#load-table" for explicit user-triggered loading.
	LazyTrigger string

	// --- Pagination ---
	// Pagination enables traditional pagination below the table
	Pagination *PaginationConfig

	// --- Infinite Scroll ---
	// InfiniteScroll enables loading more rows on scroll
	InfiniteScroll *InfiniteScrollConfig

	// --- Filters ---
	// Filters enables a filter bar above the table
	Filters *FilterConfig

	// --- Extra Query Params ---
	// ExtraQueryParams are appended to all auto-generated HTMX URLs (sort, pagination, infinite scroll).
	// Use for filter state that must persist across requests. Format: "&key=value&key2=value2"
	ExtraQueryParams string
}

// GetID returns the table ID, defaulting to "table"
func (cfg Config) GetID() string {
	if cfg.ID != "" {
		return cfg.ID
	}
	return "table"
}

// TbodyID returns the ID for the tbody element
func (cfg Config) TbodyID() string {
	return cfg.GetID() + "-tbody"
}

// LazyLoadTrigger returns the htmx trigger used by lazy-loaded tbody requests.
func (cfg Config) LazyLoadTrigger() string {
	if cfg.LazyTrigger != "" {
		return cfg.LazyTrigger
	}
	return "load"
}

// TheadID returns the ID for the thead element
func (cfg Config) TheadID() string {
	return cfg.GetID() + "-thead"
}

// PaginationID returns the ID for the pagination container
func (cfg Config) PaginationID() string {
	return cfg.GetID() + "-pagination"
}

// PaginationBaseURL returns the base URL for pagination links with per_page and sort params
func (cfg Config) PaginationBaseURL() string {
	params := map[string]string{}
	if cfg.Pagination != nil && cfg.Pagination.PerPage > 0 {
		params["per_page"] = strconv.Itoa(cfg.Pagination.PerPage)
	}
	if cfg.SortBy != "" && cfg.SortDir != SortNone {
		params["order_by"] = cfg.SortBy
		params["order_dir"] = string(cfg.SortDir)
	}
	return tableURL(cfg.HTMXEndpoint, params, cfg.ExtraQueryParams)
}

// HasSortableColumns returns true if any column is sortable
func (cfg Config) HasSortableColumns() bool {
	for _, col := range cfg.Columns {
		if col.Sortable {
			return true
		}
	}
	return false
}

// IsSortedBy returns true if the table is currently sorted by the given column
func (cfg Config) IsSortedBy(key string) bool {
	return cfg.SortBy == key
}

// NextSortDir returns the next sort direction when clicking a column header.
// Cycles: neutral → asc → desc → neutral
func (cfg Config) NextSortDir(key string) SortDir {
	if cfg.SortBy != key || cfg.SortDir == SortNone {
		return SortAsc
	}
	if cfg.SortDir == SortAsc {
		return SortDesc
	}
	return SortNone
}

// SortURL builds the HTMX URL for sorting by a given column.
// When direction cycles back to SortNone, omits sort params to reset to natural order.
func (cfg Config) SortURL(key string) string {
	dir := cfg.NextSortDir(key)
	params := map[string]string{"table_id": cfg.GetID()}
	if cfg.Pagination != nil && cfg.Pagination.PerPage > 0 {
		params["per_page"] = strconv.Itoa(cfg.Pagination.PerPage)
	}
	if dir != SortNone {
		params["order_by"] = key
		params["order_dir"] = string(dir)
	}
	return tableURL(cfg.HTMXEndpoint, params, cfg.ExtraQueryParams)
}

// PageURL builds the HTMX URL for a specific page
func (cfg Config) PageURL(page int) string {
	params := map[string]string{"page": strconv.Itoa(page)}
	if cfg.Pagination != nil && cfg.Pagination.PerPage > 0 {
		params["per_page"] = strconv.Itoa(cfg.Pagination.PerPage)
	}
	if cfg.SortBy != "" && cfg.SortDir != SortNone {
		params["order_by"] = cfg.SortBy
		params["order_dir"] = string(cfg.SortDir)
	}
	return tableURL(cfg.HTMXEndpoint, params, cfg.ExtraQueryParams)
}

// NextPageURL builds the HTMX URL for infinite scroll
func (cfg Config) NextPageURL() string {
	// Support new PaginationConfig infinite scroll mode
	if cfg.Pagination != nil && cfg.Pagination.IsInfiniteScroll() {
		params := map[string]string{
			"page":    strconv.Itoa(cfg.Pagination.NextPage()),
			"variant": "infinite",
		}
		if cfg.Pagination.PerPage > 0 {
			params["per_page"] = strconv.Itoa(cfg.Pagination.PerPage)
		}
		if cfg.SortBy != "" && cfg.SortDir != SortNone {
			params["order_by"] = cfg.SortBy
			params["order_dir"] = string(cfg.SortDir)
		}
		return tableURL(cfg.HTMXEndpoint, params, cfg.ExtraQueryParams)
	}
	// Legacy InfiniteScrollConfig support
	if cfg.InfiniteScroll == nil {
		return ""
	}
	params := map[string]string{"page": strconv.Itoa(cfg.InfiniteScroll.NextPage)}
	if cfg.SortBy != "" && cfg.SortDir != SortNone {
		params["order_by"] = cfg.SortBy
		params["order_dir"] = string(cfg.SortDir)
	}
	return tableURL(cfg.HTMXEndpoint, params, cfg.ExtraQueryParams)
}

func tableURL(endpoint string, params map[string]string, extra string) string {
	u, err := url.Parse(endpoint)
	if err != nil {
		u = &url.URL{Path: endpoint}
	}
	q := u.Query()
	for key, value := range parseExtraQuery(extra) {
		q.Set(key, value)
	}
	for key, value := range params {
		if value == "" {
			q.Del(key)
			continue
		}
		q.Set(key, value)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func parseExtraQuery(extra string) map[string]string {
	out := map[string]string{}
	extra = strings.TrimPrefix(extra, "?")
	extra = strings.TrimPrefix(extra, "&")
	if extra == "" {
		return out
	}
	values, err := url.ParseQuery(extra)
	if err != nil {
		return out
	}
	for key, vals := range values {
		if len(vals) > 0 {
			out[key] = vals[len(vals)-1]
		}
	}
	return out
}

// ContainerClasses returns the outer wrapper CSS classes.
// overflow-x-auto lets the table scroll horizontally on narrow screens
// without widening the parent. overflow-y-clip is paired explicitly:
// CSS coerces a `visible` axis to `auto` when the other is non-visible,
// so without it the wrapper becomes a Y scroll container too. During the
// htmx-row-in append keyframe the briefly translateY-ed rows make
// scrollHeight > clientHeight for ~100ms, painting a transient vertical
// scrollbar that vanishes on settle — the layout-shift flicker reported
// on infinite-scroll tables.
func (cfg Config) ContainerClasses() string {
	base := "overflow-x-auto overflow-y-clip w-full rounded-radius border border-outline dark:border-outline-dark"
	if cfg.Class != "" {
		base += " " + cfg.Class
	}
	return base
}

// TableClasses returns the <table> element CSS classes
func (cfg Config) TableClasses() string {
	return "w-full text-left text-sm text-on-surface dark:text-on-surface-dark"
}

// TheadClasses returns the <thead> CSS classes
func (cfg Config) TheadClasses() string {
	return "border-b border-outline bg-surface-alt text-sm text-on-surface-strong dark:border-outline-dark dark:bg-surface-dark-alt dark:text-on-surface-dark-strong"
}

// TbodyClasses returns the <tbody> CSS classes
func (cfg Config) TbodyClasses() string {
	return "divide-y divide-outline dark:divide-outline-dark"
}

// RowClasses returns CSS classes for a table row
func (cfg Config) RowClasses() string {
	if cfg.Variant == Striped {
		return "even:bg-primary/5 dark:even:bg-primary-dark/10"
	}
	return ""
}

// CellClasses returns CSS classes for a table cell
func (cfg Config) CellClasses() string {
	return "p-4"
}

// HeaderCellClasses returns CSS classes for a non-sortable header cell
func (cfg Config) HeaderCellClasses() string {
	return "p-4"
}

// ColumnCellClasses returns CSS classes for a cell in a specific column (applies width + alignment)
func ColumnCellClasses(col Column) string {
	cls := "p-4"
	if col.Width != "" {
		cls += " " + col.Width
	}
	switch col.Align {
	case "center":
		cls += " text-center"
	case "right":
		cls += " text-right"
	}
	return cls
}

// ColumnHeaderClasses returns CSS classes for a header cell in a specific column
func ColumnHeaderClasses(col Column) string {
	cls := "p-4"
	if col.Width != "" {
		cls += " " + col.Width
	}
	switch col.Align {
	case "center":
		cls += " text-center"
	case "right":
		cls += " text-right"
	}
	return cls
}

// BadgeCellClasses returns CSS classes for a badge based on color name.
// Solid fill with on-X foreground guarantees WCAG AA contrast across every
// theme. The previous "border-X bg-X/10 text-X" recipe collapsed to ratios
// as low as 1.12:1 because the /10 background mixes with the surface to
// produce a near-X color, killing text-X contrast.
func BadgeCellClasses(color string) string {
	base := "inline-flex overflow-hidden rounded-radius px-2 py-0.5 text-xs font-medium"
	switch color {
	case "success":
		return base + " bg-success text-on-success"
	case "danger":
		return base + " bg-danger text-on-danger"
	case "warning":
		return base + " bg-warning text-on-warning"
	case "info":
		return base + " bg-info text-on-info"
	case "primary":
		return base + " bg-primary text-on-primary dark:bg-primary-dark dark:text-on-primary-dark"
	case "secondary":
		return base + " bg-secondary text-on-secondary dark:bg-secondary-dark dark:text-on-secondary-dark"
	case "neutral":
		return base + " bg-surface-alt text-on-surface dark:bg-surface-dark-alt dark:text-on-surface-dark"
	default:
		return base + " border border-outline dark:border-outline-dark"
	}
}

// SortableHeaderClasses returns CSS classes for a sortable header cell
func (cfg Config) SortableHeaderClasses(key string) string {
	base := "p-4 cursor-pointer select-none hover:bg-surface dark:hover:bg-surface-dark transition-colors"
	if cfg.IsSortedBy(key) {
		base += " text-primary dark:text-primary-dark"
	}
	return base
}

// CheckboxClasses returns CSS classes for checkboxes
func (cfg Config) CheckboxClasses() string {
	return "before:content[''] peer relative size-4 appearance-none overflow-hidden rounded border border-outline bg-surface before:absolute before:inset-0 checked:border-primary checked:before:bg-primary focus:outline-2 focus:outline-offset-2 focus:outline-outline-strong checked:focus:outline-primary active:outline-offset-0 dark:border-outline-dark dark:bg-surface-dark-alt dark:checked:border-primary-dark dark:checked:before:bg-primary-dark dark:focus:outline-outline-dark-strong dark:checked:focus:outline-primary-dark"
}

// PaginationPages returns the page numbers to display
func (p *PaginationConfig) PaginationPages() []int {
	if p == nil || p.TotalPages <= 0 {
		return nil
	}
	pages := make([]int, 0, p.TotalPages)
	for i := 1; i <= p.TotalPages; i++ {
		pages = append(pages, i)
	}
	return pages
}

// ActionButtonClasses returns CSS classes for action buttons in table cells
func ActionButtonClasses() string {
	return "cursor-pointer whitespace-nowrap rounded-radius bg-transparent p-0.5 font-semibold text-primary outline-primary hover:opacity-75 focus-visible:outline-2 focus-visible:outline-offset-2 active:opacity-100 active:outline-offset-0 dark:text-primary-dark dark:outline-primary-dark"
}

// StatusBadgeClasses returns CSS classes for status badges. Solid bg-X / text-on-X
// for guaranteed AA contrast across themes (see BadgeCellClasses for rationale).
func StatusBadgeClasses(status string) string {
	base := "inline-flex overflow-hidden rounded-radius px-1 py-0.5 text-xs font-medium"
	switch status {
	case "active", "success":
		return base + " bg-success text-on-success"
	case "canceled", "danger":
		return base + " bg-danger text-on-danger"
	default:
		return base
	}
}

// HasFilters returns true if filter config has at least one filter
func (cfg Config) HasFilters() bool {
	return cfg.Filters != nil && len(cfg.Filters.Filters) > 0
}

// FilterBarID returns the ID for the filter bar container
func (cfg Config) FilterBarID() string {
	return cfg.GetID() + "-filters"
}

// filterControlID returns a stable id for a single filter input. Stability
// matters: hx-preserve relies on a matching id between the existing DOM and
// the swap response to know which element to keep across renders.
func filterControlID(cfg Config, filter Filter) string {
	return cfg.FilterBarID() + "-" + filter.Key
}

// filterAlpineInit generates a name for the Alpine.data registration.
// Converts hyphens to camelCase since Alpine evaluates x-data as JS expressions.
func filterAlpineInit(cfg Config) string {
	return hyphenToCamel(cfg.GetID()) + "Filters"
}

// hyphenToCamel converts a hyphenated string to camelCase (e.g. "filtered-table" → "filteredTable").
func hyphenToCamel(s string) string {
	result := make([]byte, 0, len(s))
	upper := false
	for i := 0; i < len(s); i++ {
		if s[i] == '-' {
			upper = true
			continue
		}
		if upper {
			if s[i] >= 'a' && s[i] <= 'z' {
				result = append(result, s[i]-32)
			} else {
				result = append(result, s[i])
			}
			upper = false
		} else {
			result = append(result, s[i])
		}
	}
	return string(result)
}

// filterScriptData generates a JS script block that registers an Alpine.data component.
// This avoids templ's HTML attribute escaping that breaks & and quotes.
func filterScriptData(cfg Config) string {
	var filters strings.Builder
	filters.WriteString("{")
	for i, f := range cfg.Filters.Filters {
		if i > 0 {
			filters.WriteString(", ")
		}
		filters.WriteString(f.Key + ": '" + jsEscape(f.DefaultValue) + "'")
	}
	filters.WriteString("}")

	expanded := "true"
	if cfg.Filters.Collapsible && !cfg.Filters.InitiallyExpanded {
		expanded = "false"
	}

	endpoint := cfg.HTMXEndpoint
	perPage := ""
	if cfg.Pagination != nil && cfg.Pagination.PerPage > 0 {
		perPage = "&per_page=" + itoa(cfg.Pagination.PerPage)
	}
	// ExtraQueryParams already starts with '&' (or is empty) — appended raw
	// after the auto `?_filter=1` marker so static query state (modal context
	// like ?addon_name=X) survives every filter request.
	extra := cfg.ExtraQueryParams
	hxTarget := cfg.Filters.ResolvedHxTarget(cfg)
	hxSwap := cfg.Filters.ResolvedHxSwap()
	name := filterAlpineInit(cfg)

	// Register eagerly when Alpine is already up (modal-swapped table); fall
	// back to alpine:init for the initial page load. The alpine:init listener
	// only fires ONCE per page; HTMX-swapped scripts arrive after init and
	// would otherwise never register, leaving x-data="<name>Filters" silent.
	//
	// After late registration we MUST re-initialize the wrapper subtree —
	// Alpine resolves x-data="<name>" at mount time, so a div that mounted
	// before the script ran is bound to an empty {} until we tell Alpine to
	// rebind. initTree(wrapper) is idempotent for already-bound trees.
	return fmt.Sprintf(`(() => {
		const name = '%s';
		const register = () => {
			Alpine.data(name, () => ({
				filtersExpanded: %s,
				filters: %s,
				buildFilterURL() {
					let url = '%s?_filter=1%s%s';
					for (const [k, v] of Object.entries(this.filters)) {
						if (v !== '' && v !== 'false') {
							url += '&' + encodeURIComponent(k) + '=' + encodeURIComponent(v);
						}
					}
					return url;
				},
				applyFilters() {
					htmx.ajax('GET', this.buildFilterURL(), {target: '%s', swap: '%s'});
				}
			}));
			document.querySelectorAll('[x-data="' + name + '"]').forEach((el) => {
				if (typeof Alpine.initTree === 'function') Alpine.initTree(el);
			});
		};
		if (window.Alpine && window.Alpine.version) {
			register();
		} else {
			document.addEventListener('alpine:init', register);
		}
	})();
	// Intercept all HTMX requests from this table to append filter params
	document.addEventListener('htmx:configRequest', (evt) => {
		var el = evt.detail.elt;
		var wrapper = el.closest('[x-data="%s"]');
		if (!wrapper) return;
		var comp = Alpine.$data(wrapper);
		if (!comp || !comp.filters) return;
		for (const [k, v] of Object.entries(comp.filters)) {
			if (v !== '' && v !== 'false') {
				evt.detail.parameters[k] = v;
			}
		}
	});`, name, expanded, filters.String(), endpoint, perPage, extra, hxTarget, hxSwap, name)
}

// jsEscape escapes a string for safe embedding in single-quoted JS literals
func jsEscape(s string) string {
	result := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '\'':
			result = append(result, '\\', '\'')
		case '\\':
			result = append(result, '\\', '\\')
		case '\n':
			result = append(result, '\\', 'n')
		case '\r':
			result = append(result, '\\', 'r')
		default:
			result = append(result, s[i])
		}
	}
	return string(result)
}

// itoa converts an int to string without importing strconv
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	digits := make([]byte, 0, 10)
	for n > 0 {
		digits = append(digits, byte('0'+n%10))
		n /= 10
	}
	if neg {
		digits = append(digits, '-')
	}
	// reverse
	for i, j := 0, len(digits)-1; i < j; i, j = i+1, j-1 {
		digits[i], digits[j] = digits[j], digits[i]
	}
	return string(digits)
}
