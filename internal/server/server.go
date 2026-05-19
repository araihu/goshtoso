package server

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/araihu/goshtoso/components/carousel"
	combobox "github.com/araihu/goshtoso/components/combobox/v2"
	"github.com/araihu/goshtoso/internal/pages/demo/components"
)

// Server handles HTTP requests for both original PenguinUI and Goshtoso components
type Server struct {
	projectRoot string
	mux         *http.ServeMux
}

// New creates a new server instance
func New(projectRoot string) *Server {
	s := &Server{
		projectRoot: projectRoot,
		mux:         http.NewServeMux(),
	}
	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	// Compiled assets (CSS, JS)
	assetsDir := filepath.Join(s.projectRoot, "assets")
	assetsHandler := http.StripPrefix("/assets/", http.FileServer(http.Dir(assetsDir)))
	s.mux.Handle("/assets/", assetsHandler)

	// Original PenguinUI static files (for direct access if needed)
	originalDir := filepath.Join(s.projectRoot)
	originalHandler := http.StripPrefix("/original/", http.FileServer(http.Dir(originalDir)))
	s.mux.Handle("/original/", originalHandler)

	// Component comparison pages
	s.mux.HandleFunc("/components/", s.handleComponent)

	// API endpoints for HTMX demos
	s.mux.HandleFunc("/api/hello", s.handleAPIHello)
	s.mux.HandleFunc("/api/components/button", s.handleButtonFragment)
	s.mux.HandleFunc("/api/components/accordion-content/", s.handleAccordionContent)
	s.mux.HandleFunc("/api/components/tab-content/", s.handleTabContent)
	s.mux.HandleFunc("/api/components/table/rows", s.handleTableRows)
	s.mux.HandleFunc("/api/components/toast", s.handleToastOOB)
	s.mux.HandleFunc("/api/components/carousel/slides", s.handleCarouselSlides)
	s.mux.HandleFunc("/api/components/form-validation", s.handleFormValidation)
	s.mux.HandleFunc("/api/components/steps/demo", s.handleStepsDemo)

	// HTMX SSR combobox (v2) — users demo runs server-mode lazy search.
	usersHandler := combobox.Handler(components.UsersCfg, usersProvider)
	s.mux.Handle("/api/components/combobox-new/users/options", usersHandler)
	s.mux.Handle("/api/components/combobox-new/users/toggle", usersHandler)
	s.mux.Handle("/api/components/combobox-new/users/clear", usersHandler)

	// Docs pages
	s.mux.HandleFunc("/docs/theme", s.handleThemePage)
	s.mux.HandleFunc("/getting-started", s.handleGettingStarted)

	// Landing page
	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			components.LandingPage().Render(r.Context(), w)
			return
		}
		http.NotFound(w, r)
	})
}

func (s *Server) handleComponent(w http.ResponseWriter, r *http.Request) {
	// Extract component name from URL
	path := strings.TrimPrefix(r.URL.Path, "/components/")
	if path == "" {
		http.Redirect(w, r, "/components/button", http.StatusMovedPermanently)
		return
	}

	componentName := strings.Split(path, "/")[0]

	switch componentName {
	case "button":
		components.ButtonDemoPage().Render(r.Context(), w)
	case "codeblock":
		components.CodeBlockDemoPage().Render(r.Context(), w)
	case "accordion":
		components.AccordionDemoPage().Render(r.Context(), w)
	case "sidebar":
		components.SidebarDemoPage().Render(r.Context(), w)
	case "avatar":
		components.AvatarDemoPage().Render(r.Context(), w)
	case "badge":
		components.BadgeDemoPage().Render(r.Context(), w)
	case "banner":
		components.BannerDemoPage().Render(r.Context(), w)
	case "card":
		components.CardDemoPage().Render(r.Context(), w)
	case "combobox":
		components.ComboboxDemoPage().Render(r.Context(), w)
	case "combobox-new":
		components.ComboboxNewPage().Render(r.Context(), w)
	case "alert":
		components.AlertDemoPage().Render(r.Context(), w)
	case "modal":
		components.ModalDemoPage().Render(r.Context(), w)
	case "tabs":
		components.TabsDemoPage().Render(r.Context(), w)
	case "table":
		components.TableDemoPage().Render(r.Context(), w)
	case "toggle":
		components.ToggleDemoPage().Render(r.Context(), w)
	case "pagination":
		components.PaginationDemoPage().Render(r.Context(), w)
	case "checkbox":
		components.CheckboxDemoPage().Render(r.Context(), w)
	case "dropdown":
		components.DropdownDemoPage().Render(r.Context(), w)
	case "form":
		components.FormDemoPage().Render(r.Context(), w)
	case "select":
		components.SelectDemoPage().Render(r.Context(), w)
	case "spinner":
		components.SpinnerDemoPage().Render(r.Context(), w)
	case "steps":
		components.StepsDemoPage().Render(r.Context(), w)
	case "text-input":
		components.TextInputDemoPage().Render(r.Context(), w)
	case "textarea":
		components.TextareaDemoPage().Render(r.Context(), w)
	case "toast":
		components.ToastDemoPage().Render(r.Context(), w)
	case "tooltip":
		components.TooltipDemoPage().Render(r.Context(), w)
	case "breadcrumbs":
		components.BreadcrumbsDemoPage().Render(r.Context(), w)
	case "carousel":
		components.CarouselDemoPage().Render(r.Context(), w)
	case "navbar":
		components.NavbarDemoPage().Render(r.Context(), w)
	case "tags-list":
		components.TagsListDemoPage().Render(r.Context(), w)
	case "key-value":
		components.KeyValueDemoPage().Render(r.Context(), w)
	case "triplet":
		components.TripletDemoPage().Render(r.Context(), w)
	case "fileinput":
		components.FileInputDemoPage().Render(r.Context(), w)
	case "form-validation":
		components.FormValidationDemoPage().Render(r.Context(), w)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) handleAPIHello(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `<p class="text-green-600">Hello from HTMX! Request received at %s %s</p>`, r.Method, r.URL.Path)
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

	components.StepsHTMXFlow(current).Render(r.Context(), w)
}

func (s *Server) handleButtonFragment(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	// Get disabled state from query params
	disabled := r.URL.Query().Get("disabled") == "true"

	// Render just the button grid fragment
	components.ButtonFragment(disabled).Render(r.Context(), w)
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
		fmt.Fprintf(w, `<div class="space-y-2">
			<h5 class="font-medium text-on-surface-strong dark:text-on-surface-dark-strong">Server Response A</h5>
			<p class="text-sm text-on-surface dark:text-on-surface-dark">This content was loaded from the server at <strong>%s</strong>.</p>
			<p class="text-sm text-on-surface dark:text-on-surface-dark">You can use this pattern to load large datasets, forms, or any dynamic content on demand.</p>
			<div class="flex gap-2 mt-3">
				<span class="px-2 py-1 text-xs bg-primary/10 text-primary dark:bg-primary-dark/10 dark:text-primary-dark rounded">Loaded via HTMX</span>
				<span class="px-2 py-1 text-xs bg-success/10 text-success dark:bg-success/10 dark:text-success rounded">On Demand</span>
			</div>
		</div>`, time.Now().Format("15:04:05"))
	case "lazy-content-b":
		fmt.Fprintf(w, `<div class="space-y-2">
			<h5 class="font-medium text-on-surface-strong dark:text-on-surface-dark-strong">Server Response B</h5>
			<p class="text-sm text-on-surface dark:text-on-surface-dark">This is another example of lazy-loaded content fetched at <strong>%s</strong>.</p>
			<div class="p-3 bg-surface-alt dark:bg-surface-dark-alt rounded text-sm">
				<code class="text-xs">Request: GET %s</code>
			</div>
			<p class="text-xs text-on-surface/60 dark:text-on-surface-dark/60 mt-2">Perfect for performance optimization - content only loads when needed!</p>
		</div>`, time.Now().Format("15:04:05"), r.URL.Path)
	case "lazy-content-1":
		fmt.Fprintf(w, `<div class="space-y-2">
			<h5 class="font-medium text-on-surface-strong dark:text-on-surface-dark-strong">Dynamic Content Loaded!</h5>
			<p class="text-sm text-on-surface dark:text-on-surface-dark">Loaded at %s</p>
			<p class="text-sm text-on-surface dark:text-on-surface-dark">This demonstrates how you can defer loading heavy content until the user actually needs it.</p>
		</div>`, time.Now().Format("15:04:05"))
	default:
		fmt.Fprintf(w, `<div class="text-sm text-on-surface dark:text-on-surface-dark">Unknown content ID: %s</div>`, contentID)
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
		fmt.Fprintf(w, `<div class="space-y-2">
			<h5 class="font-medium text-on-surface-strong dark:text-on-surface-dark-strong">Details (Lazy Loaded)</h5>
			<p class="text-sm text-on-surface dark:text-on-surface-dark">This content was fetched from the server at <strong>%s</strong> via HTMX.</p>
			<p class="text-sm text-on-surface dark:text-on-surface-dark">The panel only made the request when the tab was first selected, saving bandwidth and server load.</p>
			<div class="flex gap-2 mt-3">
				<span class="px-2 py-1 text-xs bg-primary/10 text-primary dark:bg-primary-dark/10 dark:text-primary-dark rounded">hx-get</span>
				<span class="px-2 py-1 text-xs bg-success/10 text-success dark:bg-success/10 dark:text-success rounded">Loaded Once</span>
			</div>
		</div>`, time.Now().Format("15:04:05"))
	case "activity":
		fmt.Fprintf(w, `<div class="space-y-2">
			<h5 class="font-medium text-on-surface-strong dark:text-on-surface-dark-strong">Recent Activity</h5>
			<p class="text-sm text-on-surface dark:text-on-surface-dark">Fetched at <strong>%s</strong>.</p>
			<ul class="text-sm text-on-surface dark:text-on-surface-dark list-disc list-inside space-y-1 mt-2">
				<li>User joined the group <em>Go Developers</em></li>
				<li>New comment on <em>HTMX Patterns</em></li>
				<li>Badge earned: <strong>Early Adopter</strong></li>
			</ul>
		</div>`, time.Now().Format("15:04:05"))
	default:
		fmt.Fprintf(w, `<div class="text-sm text-on-surface dark:text-on-surface-dark">Unknown tab content: %s</div>`, tabID)
	}
}

func (s *Server) handleThemePage(w http.ResponseWriter, r *http.Request) {
	components.ThemeDemoPage().Render(r.Context(), w)
}

func (s *Server) handleGettingStarted(w http.ResponseWriter, r *http.Request) {
	components.GettingStartedPage().Render(r.Context(), w)
}

func (s *Server) handleCarouselSlides(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	// Simulate server processing delay
	time.Sleep(500 * time.Millisecond)

	// Return a rendered static carousel with sample slides
	cfg := carousel.Config{
		Variant: carousel.WithText,
		Slides: []carousel.Slide{
			{
				ImgSrc:      "https://penguinui.s3.amazonaws.com/component-assets/carousel/default-slide-1.webp",
				ImgAlt:      "Vibrant abstract painting with swirling blue and light pink hues on a canvas.",
				Title:       "Loaded via HTMX",
				Description: fmt.Sprintf("This carousel was fetched from the server at %s.", time.Now().Format("15:04:05")),
			},
			{
				ImgSrc:      "https://penguinui.s3.amazonaws.com/component-assets/carousel/default-slide-2.webp",
				ImgAlt:      "Vibrant abstract painting with swirling red, yellow, and pink hues on a canvas.",
				Title:       "Dynamic Content",
				Description: "Slides can come from a database, API, or any backend source.",
			},
			{
				ImgSrc:      "https://penguinui.s3.amazonaws.com/component-assets/carousel/default-slide-3.webp",
				ImgAlt:      "Vibrant abstract painting with swirling blue and purple hues on a canvas.",
				Title:       "HTMX + Alpine.js",
				Description: "Server delivers HTML fragments, Alpine handles interactivity.",
			},
		},
	}
	carousel.Carousel(cfg).Render(r.Context(), w)
}

// usersProvider is an OptionsProvider for the combobox-new users demo.
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

// ServeHTTP implements http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}
