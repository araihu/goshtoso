package server

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	shellassets "github.com/araihu/goshtoso-app-shells/componentdocshell/assets"
	chartassets "github.com/araihu/goshtoso-charts/assets"
	"github.com/araihu/goshtoso/components/carousel"
	combobox "github.com/araihu/goshtoso/components/combobox"
	"github.com/araihu/goshtoso/components/toast"
	siteassets "github.com/araihu/goshtoso/site/assets"
	"github.com/araihu/goshtoso/site/internal/examples/ticker"
	"github.com/araihu/goshtoso/site/internal/pages/demo"
	buttonpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/button"
	comboboxpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/combobox"
	stepspage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/steps"
	"github.com/araihu/goshtoso/site/internal/pages/demo/components"
)

// Server handles HTTP requests for Goshtoso components
type Server struct {
	projectRoot  string
	mux          *http.ServeMux
	tickerBroker *ticker.Broker
}

// New creates a new server instance
func New(projectRoot string) *Server {
	broker := ticker.NewBroker(ticker.NewSimulator(1), tickerInterval())
	// Process-lifetime stream shared by all viewers; never cancelled.
	go broker.Run(context.Background())

	s := &Server{
		projectRoot:  projectRoot,
		mux:          http.NewServeMux(),
		tickerBroker: broker,
	}
	s.setupRoutes()
	return s
}

// tickerInterval is the simulator tick rate, overridable via GOSHTOSO_TICKER_MS
// (milliseconds) for fast, deterministic E2E. Defaults to 1s.
func tickerInterval() time.Duration {
	if v := os.Getenv("GOSHTOSO_TICKER_MS"); v != "" {
		if ms, err := strconv.Atoi(v); err == nil && ms > 0 {
			return time.Duration(ms) * time.Millisecond
		}
	}
	return time.Second
}

func (s *Server) setupRoutes() {
	// Compiled assets (CSS, JS)
	assetsDir := filepath.Join(s.projectRoot, "assets")
	assetsHandler := http.StripPrefix("/assets/", http.FileServer(http.Dir(assetsDir)))
	s.mux.Handle("/assets/", assetsHandler)
	s.mux.Handle("/site-assets/", siteassets.Handler())
	s.mux.Handle("/componentdocshell/assets/", shellassets.Handler())
	s.mux.Handle("GET "+chartassets.Prefix, withImmutableCache(chartassets.Handler()))

	// Favicons are referenced at root paths from <head>, so they are served from
	// the root mux, but the files live alongside the other assets on disk.
	// Exact-match patterns; no path traversal possible.
	for _, name := range []string{
		"favicon.ico", "favicon.svg", "favicon-96x96.png", "apple-touch-icon.png",
		"site.webmanifest", "web-app-manifest-192x192.png", "web-app-manifest-512x512.png",
	} {
		s.mux.HandleFunc("/"+name, func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, ".webmanifest") {
				w.Header().Set("Content-Type", "application/manifest+json")
			}
			http.ServeFile(w, r, filepath.Join(assetsDir, r.URL.Path[1:]))
		})
	}

	// Component comparison pages
	s.mux.HandleFunc("/robots.txt", s.handleRobots)
	s.mux.HandleFunc("/sitemap.xml", s.handleSitemap)
	s.mux.HandleFunc("/form", s.handleFormPage)
	s.mux.HandleFunc("/components/", s.handleComponent)

	// Examples pages
	s.mux.HandleFunc("/examples", s.handleExample)
	s.mux.HandleFunc("/examples/", s.handleExample)
	s.registerTodoRoutes()
	s.registerExpenseRoutes()
	s.registerChatRoutes()
	s.registerLogsRoutes()
	s.registerProfileRoutes()
	s.registerTickerRoutes()
	s.registerWizardRoutes()

	// API endpoints for HTMX demos
	s.mux.HandleFunc("/api/hello", s.handleAPIHello)
	s.mux.HandleFunc("/api/components/button", s.handleButtonFragment)
	s.mux.HandleFunc("/api/components/accordion-content/", s.handleAccordionContent)
	s.mux.HandleFunc("/api/components/tab-content/", s.handleTabContent)
	s.mux.HandleFunc("/api/components/table/rows", s.handleTableRows)
	s.mux.HandleFunc("/api/components/toast", s.handleToastOOB)
	s.mux.HandleFunc("/api/components/carousel/slides", s.handleCarouselSlides)
	s.mux.HandleFunc("/api/components/form/external-submit", s.handleFormExternalSubmit)
	s.mux.HandleFunc("/api/components/form-validation", s.handleFormValidation)
	s.mux.HandleFunc("/api/components/steps/demo", s.handleStepsDemo)
	s.mux.HandleFunc("/api/components/radio/echo", s.handleRadioEcho)
	s.mux.HandleFunc("/api/components/banner/action", s.handleBannerAction)
	s.mux.HandleFunc("/api/components/dropdown/action", s.handleDropdownAction)
	s.mux.HandleFunc("/api/components/search/items", s.handleSearchItems)
	s.registerGettingStartedRoutes()

	// Combobox users demo runs server-mode lazy search.
	usersHandler := combobox.Handler(comboboxpage.UsersCfg, usersProvider)
	s.mux.Handle("/api/components/combobox/users/options", usersHandler)
	s.mux.Handle("/api/components/combobox/users/toggle", usersHandler)
	s.mux.Handle("/api/components/combobox/users/clear", usersHandler)
	clustersHandler := combobox.Handler(comboboxpage.ClusterCfg, clustersProvider)
	s.mux.Handle("/api/components/combobox/clusters/options", clustersHandler)
	s.mux.Handle("/api/components/combobox/clusters/toggle", clustersHandler)
	s.mux.Handle("/api/components/combobox/clusters/clear", clustersHandler)

	// Docs pages
	s.mux.HandleFunc("/docs/agents", s.handleAgentsPage)
	s.mux.HandleFunc("/docs/application-patterns", s.handleApplicationPatternsPage)
	s.mux.HandleFunc("/docs/component-model", s.handleComponentModelPage)
	s.mux.HandleFunc("/docs/theme", s.handleThemePage)
	s.mux.HandleFunc("/modules/charts", s.handleChartsModulePage)
	s.mux.HandleFunc("/modules/app-shells", s.handleAppShellsModulePage)
	s.mux.HandleFunc("/getting-started", s.handleGettingStarted)
	s.mux.HandleFunc("/attributions", s.handleAttributions)
	s.mux.HandleFunc("/license", s.handleLicense)
	s.mux.HandleFunc("/privacy", s.handlePrivacy)

	// Landing page
	s.mux.HandleFunc("/playground/theme", s.handleLandingThemePlayground)
	s.mux.HandleFunc("/playground/extensions/charts", s.handleChartsShowcase)
	s.mux.HandleFunc("/playground/extensions/charts/frame", s.handleChartsShowcaseFrame)
	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			_ = components.LandingPage().Render(r.Context(), w)
			return
		}
		http.NotFound(w, r)
	})
}

func (s *Server) handleChartsShowcase(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600, stale-while-revalidate=86400")
	_ = components.ChartsShowcasePageForQuery(r.URL.Query().Get("variant")).Render(r.Context(), w)
}

func (s *Server) handleChartsShowcaseFrame(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600, stale-while-revalidate=86400")
	_ = components.ChartsShowcaseFrameForQuery(r.URL.Query().Get("variant")).Render(r.Context(), w)
}

func withImmutableCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleLandingThemePlayground(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = components.LandingPlaygroundPage().Render(r.Context(), w)
}

func (s *Server) handleRobots(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/robots.txt" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = fmt.Fprintf(w, "User-agent: *\nAllow: /\nSitemap: %s/sitemap.xml\n", demo.SiteBaseURL)
}

type sitemapURLSet struct {
	XMLName xml.Name     `xml:"urlset"`
	Xmlns   string       `xml:"xmlns,attr"`
	URLs    []sitemapURL `xml:"url"`
}

type sitemapURL struct {
	Loc        string `xml:"loc"`
	LastMod    string `xml:"lastmod,omitempty"`
	ChangeFreq string `xml:"changefreq,omitempty"`
	Priority   string `xml:"priority,omitempty"`
}

func (s *Server) handleSitemap(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/sitemap.xml" {
		http.NotFound(w, r)
		return
	}
	pages := components.AllPublicMeta()
	lastMod := time.Now().UTC().Format("2006-01-02")
	urls := make([]sitemapURL, 0, len(pages))
	for _, page := range pages {
		priority := "0.7"
		if page.Path == "/" {
			priority = "1.0"
		} else if strings.HasPrefix(page.Path, "/components/") {
			priority = "0.8"
		}
		urls = append(urls, sitemapURL{
			Loc:        page.CanonicalURL(),
			LastMod:    lastMod,
			ChangeFreq: "weekly",
			Priority:   priority,
		})
	}
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	_, _ = w.Write([]byte(xml.Header))
	_ = xml.NewEncoder(w).Encode(sitemapURLSet{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  urls,
	})
}

func (s *Server) handleComponent(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/components/")
	if path == "" {
		http.Redirect(w, r, "/components/accordion", http.StatusMovedPermanently)
		return
	}
	componentName := strings.Split(path, "/")[0]
	if componentName == "form-validation" {
		http.Redirect(w, r, "/form#server-side-validation", http.StatusMovedPermanently)
		return
	}
	s.renderDemo(w, r, "components/"+componentName)
}

func (s *Server) handleFormPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/form" {
		http.NotFound(w, r)
		return
	}
	s.renderDemo(w, r, "components/form")
}

func (s *Server) handleExample(w http.ResponseWriter, r *http.Request) {
	// Strip leading "/examples" then any leading slash → canonical name.
	sub := strings.TrimPrefix(r.URL.Path, "/examples")
	sub = strings.TrimPrefix(sub, "/")

	switch sub {
	case "", "index":
		s.renderDemo(w, r, "examples")
	case "todo":
		s.renderTodoPage(w, r)
	case "expense":
		s.renderExpensePage(w, r)
	case "chat":
		s.renderChatPage(w, r)
	case "logs":
		s.renderDemo(w, r, "examples/logs")
	case "profile":
		s.renderProfilePage(w, r)
	case "ticker":
		s.renderDemo(w, r, "examples/ticker")
	case "wizard":
		s.renderWizardPage(w, r)
	default:
		http.NotFound(w, r)
	}
}

// renderDemo picks the Layout (full document) or Fragment (HTMX swap) renderer
// based on the HX-Request header, then looks the page up in the demo registry.
// All sidebar-navigable pages flow through here so direct loads and fragment
// nav stay in lock-step.
func (s *Server) renderDemo(w http.ResponseWriter, r *http.Request, key string) {
	entry, ok := components.LookupDemo(key)
	if !ok {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	content := entry.Content()
	meta := components.DemoMeta(key, entry)
	if r.Header.Get("HX-Request") == "true" && r.Header.Get("HX-Boosted") != "true" {
		_ = demo.ComponentDocsFragment(meta, entry.Active, content, storageAllowed(r)).Render(r.Context(), w)
		return
	}
	_ = demo.ComponentDocsLayout(meta, entry.Active, content, storageAllowed(r)).Render(r.Context(), w)
}

// handleRadioEcho returns a small HTML fragment for the radio demo's HTMX showcase.
// It echoes back the radio value selected by the user.
func (s *Server) handleRadioEcho(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	value := r.URL.Query().Get("value")
	if value == "" {
		value = "(empty)"
	}
	_, _ = fmt.Fprintf(w, `Server: you picked <span class="font-mono font-semibold">%s</span> at %s.`,
		htmlEscape(value), time.Now().Format("15:04:05.000"))
}

// htmlEscape applies a minimal HTML escape suitable for an attribute-free text fragment.
func htmlEscape(s string) string {
	r := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		"\"", "&quot;",
		"'", "&#39;",
	)
	return r.Replace(s)
}

func (s *Server) handleAPIHello(w http.ResponseWriter, r *http.Request) {
	if delay, err := strconv.Atoi(r.URL.Query().Get("delay_ms")); err == nil && delay > 0 && delay <= 2_000 {
		time.Sleep(time.Duration(delay) * time.Millisecond)
	}
	w.Header().Set("Content-Type", "text/html")
	_, _ = fmt.Fprintf(w, `<p class="text-green-600">Hello from HTMX! Request received at %s %s</p>`, r.Method, r.URL.Path)
}

func (s *Server) handleBannerAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = fmt.Fprint(w, `<p class="text-sm font-medium text-success">Banner action received</p>`)
}

func (s *Server) handleDropdownAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = fmt.Fprint(w, `<p class="font-medium text-success">Dropdown action received</p>`)
}

func (s *Server) handleFormExternalSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	version := r.FormValue("version")
	if version == "" {
		version = "selected version"
	} else {
		version = "v" + version
	}
	_ = toast.OOBToast(toast.Config{
		Tone:    toast.ToneSuccess,
		Title:   "Upgrade request submitted",
		Message: fmt.Sprintf("Target version: %s", version),
	}).Render(r.Context(), w)
}

func (s *Server) handleStepsDemo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	current := 2
	if raw := r.URL.Query().Get("step"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			current = parsed
		}
	}

	if current < 1 {
		current = 1
	}
	if current > 4 {
		current = 4
	}

	_ = stepspage.StepsHTMXFlow(current).Render(r.Context(), w)
}

func (s *Server) handleButtonFragment(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	// Get disabled state from query params
	disabled := r.URL.Query().Get("disabled") == "true"

	// Render just the button grid fragment
	_ = buttonpage.ButtonFragment(disabled).Render(r.Context(), w)
}

func (s *Server) handleAccordionContent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	// Extract the content ID from the URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/components/accordion-content/")
	contentID := strings.Split(path, "/")[0]

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

func (s *Server) handleAgentsPage(w http.ResponseWriter, r *http.Request) {
	s.renderDemo(w, r, "docs/agents")
}

func (s *Server) handleApplicationPatternsPage(w http.ResponseWriter, r *http.Request) {
	s.renderDemo(w, r, "docs/application-patterns")
}

func (s *Server) handleComponentModelPage(w http.ResponseWriter, r *http.Request) {
	s.renderDemo(w, r, "docs/component-model")
}

func (s *Server) handleThemePage(w http.ResponseWriter, r *http.Request) {
	s.renderDemo(w, r, "docs/theme")
}

func (s *Server) handleChartsModulePage(w http.ResponseWriter, r *http.Request) {
	s.renderDemo(w, r, "modules/charts")
}

func (s *Server) handleAppShellsModulePage(w http.ResponseWriter, r *http.Request) {
	s.renderDemo(w, r, "modules/app-shells")
}

func (s *Server) handleGettingStarted(w http.ResponseWriter, r *http.Request) {
	s.renderDemo(w, r, "getting-started")
}

func (s *Server) handleAttributions(w http.ResponseWriter, r *http.Request) {
	s.renderDemo(w, r, "attributions")
}

func (s *Server) handleLicense(w http.ResponseWriter, r *http.Request) {
	s.renderDemo(w, r, "license")
}

func (s *Server) handlePrivacy(w http.ResponseWriter, r *http.Request) {
	s.renderDemo(w, r, "privacy")
}

func (s *Server) handleCarouselSlides(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	// Simulate server processing delay
	time.Sleep(500 * time.Millisecond)

	// Return a rendered static carousel with sample slides
	cfg := carousel.Config{
		Slides: []carousel.Slide{
			{
				ImgSrc:      "/assets/images/carousel/slide-1.webp",
				ImgAlt:      "Vibrant abstract painting with swirling blue and light pink hues on a canvas.",
				Title:       "Loaded via HTMX",
				Description: fmt.Sprintf("This carousel was fetched from the server at %s.", time.Now().Format("15:04:05")),
			},
			{
				ImgSrc:      "/assets/images/carousel/slide-2.webp",
				ImgAlt:      "Vibrant abstract painting with swirling red, yellow, and pink hues on a canvas.",
				Title:       "Dynamic Content",
				Description: "Slides can come from a database, API, or any backend source.",
			},
			{
				ImgSrc:      "/assets/images/carousel/slide-3.webp",
				ImgAlt:      "Vibrant abstract painting with swirling blue and purple hues on a canvas.",
				Title:       "HTMX + Alpine.js",
				Description: "Server delivers HTML fragments, Alpine handles interactivity.",
			},
		},
	}
	_ = carousel.Carousel(cfg).Render(r.Context(), w)
}

// usersProvider is an OptionsProvider for the combobox users demo.
// Filters a static seed list by substring match on the search query — good
// enough to exercise the lazy/search path without touching a real backend.
func usersProvider(_ context.Context, search string, _ map[string]string) ([]combobox.Option, error) {
	seed := []combobox.Option{
		{Value: "alice", Label: "Alice"},
		{Value: "bob", Label: "Bob"},
		{Value: "albert", Label: "Albert"},
		{Value: "carol", Label: "Carol"},
		{Value: "dave", Label: "Dave"},
		{Value: "eve", Label: "Eve"},
	}
	if search == "" {
		return seed, nil
	}
	q := strings.ToLower(search)
	out := make([]combobox.Option, 0, len(seed))
	for _, o := range seed {
		if strings.Contains(strings.ToLower(o.Label), q) {
			out = append(out, o)
		}
	}
	return out, nil
}

// clustersProvider is an OptionsProvider for the combobox cascading-dependency demo.
func clustersProvider(_ context.Context, search string, deps map[string]string) ([]combobox.Option, error) {
	seed := map[string][]combobox.Option{
		"aws": {
			{Value: "prod-use1", Label: "prod-use1"},
			{Value: "staging-use1", Label: "staging-use1"},
			{Value: "archive-usw2", Label: "archive-usw2", Disabled: true},
		},
		"gcp": {
			{Value: "prod-us-central1", Label: "prod-us-central1"},
			{Value: "analytics-eu", Label: "analytics-eu"},
		},
		"azure": {
			{Value: "prod-eastus", Label: "prod-eastus"},
			{Value: "ml-westeurope", Label: "ml-westeurope"},
		},
	}
	provider := deps["provider"]
	if provider == "" {
		provider = "aws"
	}
	options := seed[provider]
	if search == "" {
		return options, nil
	}
	q := strings.ToLower(search)
	out := make([]combobox.Option, 0, len(options))
	for _, o := range options {
		if strings.Contains(strings.ToLower(o.Label), q) {
			out = append(out, o)
		}
	}
	return out, nil
}

// ServeHTTP implements http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}
