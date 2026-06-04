package pagination

import "net/url"

// Variant represents pagination style variants
type Variant string

const (
	// WithEllipsis shows page numbers with ellipsis for gaps
	WithEllipsis Variant = "ellipsis"
	// Simple shows only Previous and Next buttons
	Simple Variant = "simple"
)

// PageItem represents a single item in the pagination list
type PageItem struct {
	// Page is the page number (0 means ellipsis)
	Page int
	// IsEllipsis indicates this is an ellipsis placeholder
	IsEllipsis bool
	// IsCurrent indicates this is the active page
	IsCurrent bool
}

// Config holds configuration for the pagination component
type Config struct {
	// ID is the pagination element ID
	ID string
	// Variant determines the pagination style
	Variant Variant
	// CurrentPage is the 1-indexed current page number
	CurrentPage int
	// TotalPages is the total number of pages
	TotalPages int
	// BaseURL is the base URL for page links (appends ?page=N)
	BaseURL string
	// Class allows additional CSS classes on the nav element
	NavClass string

	// --- HTMX Integration ---
	// HTMXTarget is the HTMX swap target selector
	HTMXTarget string
	// HTMXSwap is the HTMX swap strategy (default: innerHTML)
	HTMXSwap string
}

// HasPrevious returns true if there is a previous page
func (cfg Config) HasPrevious() bool {
	return cfg.CurrentPage > 1
}

// HasNext returns true if there is a next page
func (cfg Config) HasNext() bool {
	return cfg.CurrentPage < cfg.TotalPages
}

// PreviousPage returns the previous page number
func (cfg Config) PreviousPage() int {
	if cfg.CurrentPage > 1 {
		return cfg.CurrentPage - 1
	}
	return 1
}

// NextPage returns the next page number
func (cfg Config) NextPage() int {
	if cfg.CurrentPage < cfg.TotalPages {
		return cfg.CurrentPage + 1
	}
	return cfg.TotalPages
}

// PageURL builds the URL for a specific page
func (cfg Config) PageURL(page int) string {
	if cfg.BaseURL == "" {
		return "#"
	}
	u, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return "#"
	}
	q := u.Query()
	q.Set("page", itoa(page))
	u.RawQuery = q.Encode()
	return u.String()
}

// SwapStrategy returns the HTMX swap strategy, defaulting to innerHTML
func (cfg Config) SwapStrategy() string {
	if cfg.HTMXSwap != "" {
		return cfg.HTMXSwap
	}
	return "innerHTML"
}

// NavClasses returns CSS classes for the nav element
func (cfg Config) NavClasses() string {
	base := "flex items-center"
	if cfg.NavClass != "" {
		base += " " + cfg.NavClass
	}
	return base
}

// ListClasses returns CSS classes for the ul element
func (cfg Config) ListClasses() string {
	return "flex shrink-0 items-center gap-2 text-sm font-medium"
}

// PrevNextClasses returns CSS classes for previous/next links
func (cfg Config) PrevNextClasses(enabled bool) string {
	if enabled {
		return "flex items-center rounded-radius p-1 text-on-surface hover:text-primary dark:text-on-surface-dark dark:hover:text-primary-dark"
	}
	return "flex items-center rounded-radius p-1 text-on-surface/40 dark:text-on-surface-dark/40 cursor-not-allowed"
}

// PageClasses returns CSS classes for a page number link
func (cfg Config) PageClasses(isCurrent bool) string {
	if isCurrent {
		return "flex size-6 items-center justify-center rounded-radius bg-primary p-1 font-bold text-on-primary dark:bg-primary-dark dark:text-on-primary-dark"
	}
	return "flex size-6 items-center justify-center rounded-radius p-1 text-on-surface hover:text-primary dark:text-on-surface-dark dark:hover:text-primary-dark"
}

// EllipsisClasses returns CSS classes for the ellipsis indicator
func (cfg Config) EllipsisClasses() string {
	return "flex size-6 items-center justify-center rounded-radius p-1 text-on-surface hover:text-primary dark:text-on-surface-dark dark:hover:text-primary-dark"
}

// Pages returns the list of page items to render for the ellipsis variant
func (cfg Config) Pages() []PageItem {
	if cfg.TotalPages <= 0 {
		return nil
	}

	// If total pages <= 7, show all pages
	if cfg.TotalPages <= 7 {
		items := make([]PageItem, cfg.TotalPages)
		for i := 1; i <= cfg.TotalPages; i++ {
			items[i-1] = PageItem{
				Page:      i,
				IsCurrent: i == cfg.CurrentPage,
			}
		}
		return items
	}

	// For larger page counts, use ellipsis
	var items []PageItem

	// Always show first page
	items = append(items, PageItem{Page: 1, IsCurrent: cfg.CurrentPage == 1})

	// Determine the range around current page
	start := cfg.CurrentPage - 1
	end := cfg.CurrentPage + 1

	// Adjust if near the beginning
	if start <= 2 {
		start = 2
		end = 4
	}

	// Adjust if near the end
	if end >= cfg.TotalPages-1 {
		end = cfg.TotalPages - 1
		start = cfg.TotalPages - 3
	}

	// Clamp
	if start < 2 {
		start = 2
	}
	if end > cfg.TotalPages-1 {
		end = cfg.TotalPages - 1
	}

	// Add ellipsis before range if needed
	if start > 2 {
		items = append(items, PageItem{IsEllipsis: true})
	}

	// Add pages in range
	for i := start; i <= end; i++ {
		items = append(items, PageItem{Page: i, IsCurrent: i == cfg.CurrentPage})
	}

	// Add ellipsis after range if needed
	if end < cfg.TotalPages-1 {
		items = append(items, PageItem{IsEllipsis: true})
	}

	// Always show last page
	items = append(items, PageItem{Page: cfg.TotalPages, IsCurrent: cfg.CurrentPage == cfg.TotalPages})

	return items
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
