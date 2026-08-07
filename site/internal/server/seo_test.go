package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	startpages "github.com/araihu/goshtoso/site/internal/pages/demo/contentpages/start"
	"github.com/stretchr/testify/require"
)

func TestComponentPageRendersSEOMetadata(t *testing.T) {
	s := &Server{}
	req := httptest.NewRequest(http.MethodGet, "/components/accordion", nil)
	rec := httptest.NewRecorder()

	s.renderDemo(rec, req, "components/accordion")

	require.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	require.Contains(t, body, "<title>Accordion Component - Goshtoso UI Library for Go</title>")
	require.Contains(t, body, `<meta name="description" content="Build accessible accordion interfaces in Go`)
	require.Contains(t, body, `<link rel="canonical" href="https://goshtoso.araihu.com/components/accordion">`)
	require.Contains(t, body, `<meta property="og:title" content="Accordion Component - Goshtoso UI Library for Go">`)
	require.Contains(t, body, `<meta name="twitter:card" content="summary_large_image">`)
	require.Contains(t, body, `<meta property="og:image:type" content="image/png">`)
	require.Contains(t, body, `<meta property="og:image:width" content="1200">`)
	require.Contains(t, body, `<meta property="og:image:height" content="630">`)
	require.Contains(t, body, `<meta property="og:image:alt" content="Accordion Component - Goshtoso UI Library for Go — Goshtoso Go UI component library preview">`)
	require.Contains(t, body, `<meta name="twitter:image:alt" content="Accordion Component - Goshtoso UI Library for Go — Goshtoso Go UI component library preview">`)
	require.Contains(t, body, `"@type":"TechArticle"`)
}

func TestIconComponentPageRendersCompleteSEOMetadata(t *testing.T) {
	s := &Server{}
	req := httptest.NewRequest(http.MethodGet, "/components/icon", nil)
	rec := httptest.NewRecorder()

	s.renderDemo(rec, req, "components/icon")

	require.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	require.Contains(t, body, "<title>Icon Component - Goshtoso UI Library for Go</title>")
	require.Contains(t, body, `<meta name="description" content="Render accessible SVG sprite symbols and generate release-verified consumer-local icon packs with typed Go bindings and attribution.">`)
	require.Contains(t, body, `<link rel="canonical" href="https://goshtoso.araihu.com/components/icon">`)
	require.Contains(t, body, `<meta property="og:url" content="https://goshtoso.araihu.com/components/icon">`)
	require.Contains(t, body, `<meta property="og:image:type" content="image/png">`)
	require.Contains(t, body, `<meta property="og:image:width" content="1200">`)
	require.Contains(t, body, `<meta property="og:image:height" content="630">`)
	require.Contains(t, body, `<meta property="og:image:alt" content="Icon Component - Goshtoso UI Library for Go — Goshtoso Go UI component library preview">`)
	require.Contains(t, body, `<meta name="twitter:image:alt" content="Icon Component - Goshtoso UI Library for Go — Goshtoso Go UI component library preview">`)
	require.Equal(t, 1, strings.Count(body, `property="og:url"`))
	require.Equal(t, 1, strings.Count(body, `name="twitter:image"`))
}

func TestLandingPageRendersSEOMetadata(t *testing.T) {
	rec := httptest.NewRecorder()

	err := startpages.LandingPage().Render(context.Background(), rec)

	require.NoError(t, err)
	body := rec.Body.String()
	require.Contains(t, body, "<title>Go UI components for server-rendered apps | Goshtoso</title>")
	require.Contains(t, body, `<meta name="description" content="Build server-rendered Go interfaces with pre-generated templ components`)
	require.Contains(t, body, `<link rel="canonical" href="https://goshtoso.araihu.com/">`)
	require.Contains(t, body, `"@type":"SoftwareApplication"`)
}

func TestLandingThemePlaygroundRouteRendersIsolatedDocument(t *testing.T) {
	s := &Server{}
	req := httptest.NewRequest(http.MethodGet, "/playground/theme", nil)
	rec := httptest.NewRecorder()

	s.handleLandingThemePlayground(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "text/html; charset=utf-8", rec.Header().Get("Content-Type"))
	require.Contains(t, rec.Body.String(), `<html lang="en" data-theme="araihu" data-landing-playground`)
	require.Contains(t, rec.Body.String(), `id="home-theme-picker"`)
}

func TestAgentsPageRendersSkillInstallGuidance(t *testing.T) {
	s := &Server{}
	req := httptest.NewRequest(http.MethodGet, "/docs/agents", nil)
	rec := httptest.NewRecorder()

	s.renderDemo(rec, req, "docs/agents")

	require.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	require.Contains(t, body, "<title>Using Goshtoso With AI Agents</title>")
	require.Contains(t, body, `npx skills add araihu/goshtoso --skill using-goshtoso`)
	require.Contains(t, body, `npx skills use araihu/goshtoso --skill using-goshtoso`)
	require.Contains(t, body, "This skill is for consumer agents.")
	require.Contains(t, body, `<link rel="canonical" href="https://goshtoso.araihu.com/docs/agents">`)
}

func TestApplicationPatternsPageRendersSEOMetadata(t *testing.T) {
	s := &Server{}
	req := httptest.NewRequest(http.MethodGet, "/docs/application-patterns", nil)
	rec := httptest.NewRecorder()

	s.renderDemo(rec, req, "docs/application-patterns")

	require.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	require.Contains(t, body, "<title>Application Patterns for Goshtoso</title>")
	require.Contains(t, body, `<meta name="description" content="Compose App Shell, Operations List, Detail Workspace, and Multi-step Workflow`)
	require.Contains(t, body, `<link rel="canonical" href="https://goshtoso.araihu.com/docs/application-patterns">`)
	require.Contains(t, body, `"@type":"TechArticle"`)
}

func TestRobotsAndSitemapExposePublicPages(t *testing.T) {
	s := &Server{}

	robotsReq := httptest.NewRequest(http.MethodGet, "/robots.txt", nil)
	robotsRec := httptest.NewRecorder()
	s.handleRobots(robotsRec, robotsReq)

	require.Equal(t, http.StatusOK, robotsRec.Code)
	require.Equal(t, "text/plain; charset=utf-8", robotsRec.Header().Get("Content-Type"))
	require.Equal(t, "User-agent: *\nAllow: /\nSitemap: https://goshtoso.araihu.com/sitemap.xml\n", robotsRec.Body.String())

	sitemapReq := httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil)
	sitemapRec := httptest.NewRecorder()
	s.handleSitemap(sitemapRec, sitemapReq)

	require.Equal(t, http.StatusOK, sitemapRec.Code)
	require.Equal(t, "application/xml; charset=utf-8", sitemapRec.Header().Get("Content-Type"))
	body := sitemapRec.Body.String()
	require.True(t, strings.HasPrefix(body, `<?xml version="1.0" encoding="UTF-8"?>`))
	require.Contains(t, body, `<loc>https://goshtoso.araihu.com/</loc>`)
	require.Contains(t, body, `<loc>https://goshtoso.araihu.com/components/accordion</loc>`)
	require.Contains(t, body, `<loc>https://goshtoso.araihu.com/docs/agents</loc>`)
	require.Contains(t, body, `<loc>https://goshtoso.araihu.com/docs/application-patterns</loc>`)
	require.Contains(t, body, `<loc>https://goshtoso.araihu.com/examples/chat</loc>`)
	require.NotContains(t, body, `/api/`)
}
