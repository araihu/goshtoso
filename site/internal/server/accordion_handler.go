package server

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

func (s *Server) handleAccordionContent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	// Extract the content ID from the URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/components/accordion-content/")
	contentID, _, _ := strings.Cut(path, "/")

	// Simulate server processing delay
	time.Sleep(500 * time.Millisecond)

	// Return different content based on the ID
	switch contentID {
	case "lazy-content-a":
		_, _ = fmt.Fprintf(w, `<div class="space-y-2">
			<h5 class="font-medium text-on-surface-strong dark:text-on-surface-dark-strong">Server Response A</h5>
			<p class="text-sm text-on-surface dark:text-on-surface-dark">This content was loaded from the server at <strong>%s</strong>.</p>
			<p class="text-sm text-on-surface dark:text-on-surface-dark">You can use this pattern to load large datasets, forms, or any dynamic content on demand.</p>
			<div class="flex gap-2 mt-3">
				<span class="px-2 py-1 text-xs bg-primary/10 text-primary dark:bg-primary-dark/10 dark:text-primary-dark rounded">Loaded via HTMX</span>
				<span class="px-2 py-1 text-xs bg-success/10 text-success dark:bg-success/10 dark:text-success rounded">On Demand</span>
			</div>
		</div>`, time.Now().Format("15:04:05"))
	case "lazy-content-b":
		_, _ = fmt.Fprintf(w, `<div class="space-y-2">
			<h5 class="font-medium text-on-surface-strong dark:text-on-surface-dark-strong">Server Response B</h5>
			<p class="text-sm text-on-surface dark:text-on-surface-dark">This is another example of lazy-loaded content fetched at <strong>%s</strong>.</p>
			<div class="p-3 bg-surface-alt dark:bg-surface-dark-alt rounded text-sm">
				<code class="text-xs">Request: GET %s</code>
			</div>
			<p class="text-xs text-on-surface/60 dark:text-on-surface-dark/60 mt-2">Perfect for performance optimization - content only loads when needed!</p>
		</div>`, time.Now().Format("15:04:05"), r.URL.Path)
	case "lazy-content-1":
		_, _ = fmt.Fprintf(w, `<div class="space-y-2">
			<h5 class="font-medium text-on-surface-strong dark:text-on-surface-dark-strong">Dynamic Content Loaded!</h5>
			<p class="text-sm text-on-surface dark:text-on-surface-dark">Loaded at %s</p>
			<p class="text-sm text-on-surface dark:text-on-surface-dark">This demonstrates how you can defer loading heavy content until the user actually needs it.</p>
		</div>`, time.Now().Format("15:04:05"))
	default:
		_, _ = fmt.Fprintf(w, `<div class="text-sm text-on-surface dark:text-on-surface-dark">Unknown content ID: %s</div>`, contentID)
	}
}
