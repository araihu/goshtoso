package demo

import (
	"encoding/json"
	"html"
	"strings"
)

const (
	SiteName    = "Goshtoso"
	SiteBaseURL = "https://goshtoso.araihu.com"
	OGImagePath = "/assets/images/goshtoso-art.png"
)

// PageMeta describes crawler and social-preview metadata for one public page.
type PageMeta struct {
	Title       string
	Description string
	Path        string
	Type        string
}

func DefaultMeta(title string) PageMeta {
	return PageMeta{
		Title:       title,
		Description: "Goshtoso is a server-rendered Go UI component library built with templ, HTMX, Alpine.js, and Tailwind CSS.",
		Path:        "/",
		Type:        "TechArticle",
	}
}

func HomeMeta() PageMeta {
	return PageMeta{
		Title:       "Goshtoso - Go HTMX Component Library",
		Description: "Build interactive, server-rendered Go UIs with templ, HTMX, Alpine.js, Tailwind CSS, and copy-pasteable Goshtoso components.",
		Path:        "/",
		Type:        "SoftwareApplication",
	}
}

func (m PageMeta) TitleText() string {
	title := strings.TrimSpace(m.Title)
	if title == "" {
		return SiteName
	}
	if strings.Contains(title, SiteName) {
		return title
	}
	return title + " - " + SiteName
}

func (m PageMeta) CanonicalURL() string {
	path := strings.TrimSpace(m.Path)
	if path == "" {
		path = "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return SiteBaseURL + path
}

func (m PageMeta) OGImageURL() string {
	return SiteBaseURL + OGImagePath
}

func (m PageMeta) SchemaType() string {
	if m.Type != "" {
		return m.Type
	}
	return "TechArticle"
}

func (m PageMeta) JSONLD() string {
	body := map[string]any{
		"@context":    "https://schema.org",
		"@type":       m.SchemaType(),
		"name":        m.TitleText(),
		"description": m.Description,
		"url":         m.CanonicalURL(),
		"publisher": map[string]string{
			"@type": "Organization",
			"name":  SiteName,
			"url":   SiteBaseURL,
		},
	}
	if m.SchemaType() == "SoftwareApplication" {
		body["applicationCategory"] = "DeveloperApplication"
		body["operatingSystem"] = "Any"
	}
	b, err := json.Marshal(body)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func SafeJSONLD(m PageMeta) string {
	return strings.ReplaceAll(m.JSONLD(), "</", "<\\/")
}

func EscapeAttr(s string) string {
	return html.EscapeString(s)
}
