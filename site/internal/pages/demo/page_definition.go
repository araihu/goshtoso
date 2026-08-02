package demo

import "github.com/a-h/templ"

// PageDefinition describes one routable demo-site page independently of the
// package that renders it.
type PageDefinition struct {
	Key         string
	Title       string
	Active      string
	Description string
	Type        string
	Content     func() templ.Component
}
