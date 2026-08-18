package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/araihu/goshtoso/site/internal/pages/demo"
	demoregistry "github.com/araihu/goshtoso/site/internal/pages/demo/registry"
)

func (s *Server) handleLLMs(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/llms.txt" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = fmt.Fprint(w, generatedLLMsText(demoregistry.AllPublicMeta()))
}

// generatedLLMsText builds the LLM-facing site map from the same public page
// registry used by the HTML routes and sitemap. Keeping this index derived
// from PageMeta prevents documentation links from drifting between artifacts.
func generatedLLMsText(pages []demo.PageMeta) string {
	var content strings.Builder
	home := demo.HomeMeta()

	fmt.Fprintf(&content, "# %s\n\n", demo.SiteName)
	fmt.Fprintf(&content, "> %s\n\n", llmsSingleLine(home.Description))
	content.WriteString("Goshtoso is a Go UI component library and documentation site for server-rendered interfaces. Public documentation covers typed templ components, accessible interaction patterns, bundled assets, HTMX, Alpine.js, themes, charts, app shells, and runnable examples.\n\n")
	content.WriteString("## Public pages\n\n")

	for _, page := range pages {
		label := strings.TrimSpace(page.Title)
		if label == "" {
			label = page.TitleText()
		}
		fmt.Fprintf(&content, "- [%s](%s): %s\n", label, page.CanonicalURL(), llmsSingleLine(page.Description))
	}

	content.WriteString("\n## Machine-readable resources\n\n")
	fmt.Fprintf(&content, "- [Sitemap XML](%s/sitemap.xml): The generated XML index of every public Goshtoso page.\n", demo.SiteBaseURL)
	fmt.Fprintf(&content, "- [Robots.txt](%s/robots.txt): Crawl guidance and the sitemap location.\n", demo.SiteBaseURL)

	return content.String()
}

func llmsSingleLine(value string) string {
	return strings.Join(strings.Fields(value), " ")
}
