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
	libraryassets "github.com/araihu/goshtoso/assets"
	combobox "github.com/araihu/goshtoso/components/combobox"
	siteassets "github.com/araihu/goshtoso/site/assets"
	"github.com/araihu/goshtoso/site/internal/examples/ticker"
	"github.com/araihu/goshtoso/site/internal/pages/demo"
	comboboxpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/combobox"
	modulespages "github.com/araihu/goshtoso/site/internal/pages/demo/contentpages/modules"
	startpages "github.com/araihu/goshtoso/site/internal/pages/demo/contentpages/start"
	demoregistry "github.com/araihu/goshtoso/site/internal/pages/demo/registry"
)

const (
	scrollRegionBFullThemeQuery    = "t-gs-011-theme"
	scrollRegionBFullConsumerQuery = "t-gs-011-consumer"
	scrollRegionBFullThemeSource   = "server-routed-html"
	// ScrollRegionBFullConsumerRouteToken is transport only. B-FULL cells keep
	// their distinct consumer-scrollregion identity in receipts and traces.
	ScrollRegionBFullConsumerRouteToken = "scrollregion"
)

type scrollRegionBFullRoutedAppearance struct {
	Theme    string
	Consumer bool
}

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
	s.setupAssetRoutes()

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
	s.mux.HandleFunc("/docs/iconpack", s.handleIconpackPage)
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
			_ = startpages.LandingPage().Render(r.Context(), w)
			return
		}
		http.NotFound(w, r)
	})
}

func (s *Server) setupAssetRoutes() {
	// Compiled assets (CSS, JS)
	assetsDir := filepath.Join(s.projectRoot, "assets")
	assetsHandler := http.StripPrefix("/assets/", http.FileServer(http.Dir(assetsDir)))
	s.mux.Handle("/assets/", libraryassets.WithCacheControl(assetsHandler))
	bootstrapSprite := filepath.Join(s.projectRoot, "site", "internal", "demoicons", "bootstrapicons", "sprite.svg")
	s.mux.HandleFunc("GET /assets/icons/bootstrapicons/sprite.svg", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/svg+xml")
		http.ServeFile(w, r, bootstrapSprite)
	})
	s.mux.Handle("/site-assets/", libraryassets.WithCacheControl(siteassets.Handler()))
	s.mux.Handle("/componentdocshell/assets/", libraryassets.WithCacheControl(shellassets.Handler()))
	s.mux.Handle("GET "+chartassets.Prefix, withImmutableCache(chartassets.Handler()))

	// Favicons are referenced at root paths from <head>, so they are served from
	// the root mux, but the files live alongside the other assets on disk.
	// Exact-match patterns; no path traversal possible.
	for _, name := range []string{
		"favicon.ico", "favicon.svg", "favicon-96x96.png", "apple-touch-icon.png",
		"site.webmanifest", "web-app-manifest-192x192.png", "web-app-manifest-512x512.png",
	} {
		routeName := name
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(routeName, ".webmanifest") {
				w.Header().Set("Content-Type", "application/manifest+json")
			}
			http.ServeFile(w, r, filepath.Join(assetsDir, routeName))
		})
		s.mux.Handle("/"+routeName, libraryassets.WithCacheControl(handler))
	}
}

func (s *Server) handleChartsShowcase(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600, stale-while-revalidate=86400")
	_ = modulespages.ChartsShowcasePageForQuery(r.URL.Query().Get("variant")).Render(r.Context(), w)
}

func (s *Server) handleChartsShowcaseFrame(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600, stale-while-revalidate=86400")
	_ = modulespages.ChartsShowcaseFrameForQuery(r.URL.Query().Get("variant")).Render(r.Context(), w)
}

func withImmutableCache(next http.Handler) http.Handler {
	return libraryassets.WithCacheControl(next)
}

func (s *Server) handleLandingThemePlayground(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = startpages.LandingPlaygroundPage().Render(r.Context(), w)
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
	pages := demoregistry.AllPublicMeta()
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
	entry, ok := demoregistry.Lookup(key)
	if !ok {
		http.NotFound(w, r)
		return
	}
	appearance, themed, err := scrollRegionBFullRoutedAppearanceForRequest(r, key)
	if err != nil {
		http.Error(w, "T-GS-011 routed theme: "+err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	content := entry.Content()
	meta := demoregistry.MetaForKey(key)
	if r.Header.Get("HX-Request") == "true" && r.Header.Get("HX-Boosted") != "true" {
		if themed {
			http.Error(w, "T-GS-011 routed theme: full-document response required", http.StatusBadRequest)
			return
		}
		_ = demo.ComponentDocsFragment(meta, entry.Active, content, storageAllowed(r)).Render(r.Context(), w)
		return
	}
	if themed {
		var rendered strings.Builder
		if err := demo.ComponentDocsLayoutWithInitialTheme(meta, entry.Active, content, storageAllowed(r), appearance.Theme).Render(r.Context(), &rendered); err != nil {
			http.Error(w, "T-GS-011 routed theme: render failed", http.StatusInternalServerError)
			return
		}
		bound, err := bindScrollRegionBFullInitialAppearance(rendered.String(), appearance)
		if err != nil {
			http.Error(w, "T-GS-011 routed theme: "+err.Error(), http.StatusInternalServerError)
			return
		}
		_, _ = w.Write([]byte(bound))
		return
	}
	_ = demo.ComponentDocsLayout(meta, entry.Active, content, storageAllowed(r)).Render(r.Context(), w)
}

// scrollRegionBFullRoutedAppearanceForRequest intentionally grants no generic
// theme switcher. It recognizes only the evidence route's literal theme axis,
// so normal public component docs remain locked to Arai Hu.
func scrollRegionBFullRoutedAppearanceForRequest(r *http.Request, key string) (scrollRegionBFullRoutedAppearance, bool, error) {
	query := r.URL.Query()
	theme, hasTheme := query[scrollRegionBFullThemeQuery]
	consumer, hasConsumer := query[scrollRegionBFullConsumerQuery]
	if !hasTheme && !hasConsumer {
		return scrollRegionBFullRoutedAppearance{}, false, nil
	}
	if key != "components/scroll-region" || r.URL.Path != "/components/scroll-region" {
		return scrollRegionBFullRoutedAppearance{}, false, fmt.Errorf("query is limited to /components/scroll-region")
	}
	if len(theme) != 1 {
		return scrollRegionBFullRoutedAppearance{}, false, fmt.Errorf("requires exactly one %s value", scrollRegionBFullThemeQuery)
	}
	appearance := scrollRegionBFullRoutedAppearance{Theme: theme[0]}
	switch appearance.Theme {
	case "araihu", "goshtoso", "minimal":
	default:
		return scrollRegionBFullRoutedAppearance{}, false, fmt.Errorf("unsupported %s %q", scrollRegionBFullThemeQuery, appearance.Theme)
	}
	if hasConsumer {
		if len(consumer) != 1 || consumer[0] != ScrollRegionBFullConsumerRouteToken {
			return scrollRegionBFullRoutedAppearance{}, false, fmt.Errorf("unsupported %s", scrollRegionBFullConsumerQuery)
		}
		appearance.Consumer = true
	}
	return appearance, true, nil
}

func bindScrollRegionBFullInitialAppearance(body string, appearance scrollRegionBFullRoutedAppearance) (string, error) {
	lower := strings.ToLower(body)
	start := strings.Index(lower, "<html")
	if start < 0 {
		return "", fmt.Errorf("full document lacks html element")
	}
	endRelative := strings.Index(lower[start:], ">")
	if endRelative < 0 {
		return "", fmt.Errorf("html start tag is incomplete")
	}
	end := start + endRelative
	attributes := ` data-theme="` + appearance.Theme + `" data-goshtoso-theme-initial-source="` + scrollRegionBFullThemeSource + `"`
	if appearance.Consumer {
		attributes += ` data-goshtoso-scrollregion-consumer-theme="t-gs-011"`
	}
	return body[:end] + attributes + body[end:], nil
}

func (s *Server) handleAPIHello(w http.ResponseWriter, r *http.Request) {
	if delay, err := strconv.Atoi(r.URL.Query().Get("delay_ms")); err == nil && delay > 0 && delay <= 2_000 {
		time.Sleep(time.Duration(delay) * time.Millisecond)
	}
	w.Header().Set("Content-Type", "text/html")
	_, _ = fmt.Fprintf(w, `<p class="text-green-600">Hello from HTMX! Request received at %s %s</p>`, r.Method, r.URL.Path)
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

func (s *Server) handleIconpackPage(w http.ResponseWriter, r *http.Request) {
	s.renderDemo(w, r, "docs/iconpack")
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
