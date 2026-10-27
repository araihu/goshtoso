// Package themeplayground hosts isolated, versioned theme previews.
package themeplayground

import (
	"encoding/json"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/themes"
)

type Config struct {
	ID        string
	Src       string
	Title     string
	RootClass string
	RootAttrs templ.Attributes
}

// FrameConfig is server-owned. Preview messages carry definitions, never assets or markup.
type FrameConfig struct {
	ID             string
	Title          string
	Content        templ.Component
	Head           templ.Component
	Fonts          []themes.FontOption
	FontCatalogURL string
}

func (c FrameConfig) data() string {
	b, _ := json.Marshal(struct {
		ID             string              `json:"id"`
		Fonts          []themes.FontOption `json:"fonts"`
		FontCatalogURL string              `json:"fontCatalogURL"`
		Tokens         []themes.TokenSpec  `json:"tokens"`
	}{c.ID, c.Fonts, c.FontCatalogURL, themes.Tokens()})
	return string(b)
}
