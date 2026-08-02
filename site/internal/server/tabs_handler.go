package server

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

func (s *Server) handleTabContent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	path := strings.TrimPrefix(r.URL.Path, "/api/components/tab-content/")
	tabID := strings.Split(path, "/")[0]

	// Simulate server processing delay
	time.Sleep(500 * time.Millisecond)

	switch tabID {
	case "details":
		_, _ = fmt.Fprintf(w, `<div class="space-y-2">
			<h5 class="font-medium text-on-surface-strong dark:text-on-surface-dark-strong">Details (Lazy Loaded)</h5>
			<p class="text-sm text-on-surface dark:text-on-surface-dark">This content was fetched from the server at <strong>%s</strong> via HTMX.</p>
			<p class="text-sm text-on-surface dark:text-on-surface-dark">The panel only made the request when the tab was first selected, saving bandwidth and server load.</p>
			<div class="flex gap-2 mt-3">
				<span class="px-2 py-1 text-xs bg-primary/10 text-primary dark:bg-primary-dark/10 dark:text-primary-dark rounded">hx-get</span>
				<span class="px-2 py-1 text-xs bg-success/10 text-success dark:bg-success/10 dark:text-success rounded">Loaded Once</span>
			</div>
		</div>`, time.Now().Format("15:04:05"))
	case "activity":
		_, _ = fmt.Fprintf(w, `<div class="space-y-2">
			<h5 class="font-medium text-on-surface-strong dark:text-on-surface-dark-strong">Recent Activity</h5>
			<p class="text-sm text-on-surface dark:text-on-surface-dark">Fetched at <strong>%s</strong>.</p>
			<ul class="text-sm text-on-surface dark:text-on-surface-dark list-disc list-inside space-y-1 mt-2">
				<li>User joined the group <em>Go Developers</em></li>
				<li>New comment on <em>HTMX Patterns</em></li>
				<li>Badge earned: <strong>Early Adopter</strong></li>
			</ul>
		</div>`, time.Now().Format("15:04:05"))
	default:
		_, _ = fmt.Fprintf(w, `<div class="text-sm text-on-surface dark:text-on-surface-dark">Unknown tab content: %s</div>`, tabID)
	}
}
