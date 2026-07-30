package demo

import (
	"encoding/json"
	"html"
	"strings"
)

const (
	SiteName        = "Goshtoso"
	SiteBaseURL     = "https://goshtoso.araihu.com"
	OGImagePath     = "/assets/images/goshtoso-art.png"
	HomeOGImagePath = "/assets/images/goshtoso-social-card.png"
)

// PageMeta describes crawler and social-preview metadata for one public page.
type PageMeta struct {
	Title       string
	Description string
	Path        string
	Type        string
	ImagePath   string
}

func DefaultMeta(title string) PageMeta {
	return PageMeta{
		Title:       title,
		Description: "Build server-rendered Go interfaces with typed templ components, bundled assets, HTMX, Alpine.js, and Tailwind CSS.",
		Path:        "/",
		Type:        "TechArticle",
	}
}

func HomeMeta() PageMeta {
	return PageMeta{
		Title:       "Go UI components for server-rendered apps | Goshtoso",
		Description: "Build server-rendered Go interfaces with pre-generated templ components, bundled assets, HTMX, Alpine.js, and Tailwind CSS.",
		Path:        "/",
		Type:        "SoftwareApplication",
		ImagePath:   HomeOGImagePath,
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
	imagePath := strings.TrimSpace(m.ImagePath)
	if imagePath == "" {
		imagePath = OGImagePath
	}
	return SiteBaseURL + imagePath
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
