// Package registry validates and indexes demo page definitions.
package registry

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/araihu/goshtoso/site/internal/pages/catalog"
	"github.com/araihu/goshtoso/site/internal/pages/demo"
)

// Registry is an immutable index of demo page definitions.
type Registry struct {
	definitions map[string]demo.PageDefinition
	keys        []string
}

// New validates definitions against the public component catalog and copies
// all input state into an immutable registry.
func New(definitions []demo.PageDefinition, componentCatalog []catalog.Entry) (*Registry, error) {
	catalogByKey := make(map[string]catalog.Entry, len(componentCatalog))
	for _, entry := range componentCatalog {
		if _, exists := catalogByKey[entry.Key]; exists {
			return nil, fmt.Errorf("duplicate component catalog key %q", entry.Key)
		}
		catalogByKey[entry.Key] = entry
	}

	byKey := make(map[string]demo.PageDefinition, len(definitions))
	for _, definition := range definitions {
		if err := validateDefinition(definition); err != nil {
			return nil, err
		}
		if _, exists := byKey[definition.Key]; exists {
			return nil, fmt.Errorf("duplicate key %q", definition.Key)
		}

		if strings.HasPrefix(definition.Key, "components/") {
			entry, exists := catalogByKey[definition.Key]
			if !exists {
				return nil, fmt.Errorf("component definition %q is not present in component catalog", definition.Key)
			}
			definition.Description = entry.Description
		}
		byKey[definition.Key] = definition
	}

	for _, entry := range componentCatalog {
		if _, exists := byKey[entry.Key]; !exists {
			return nil, fmt.Errorf("missing component definition for catalog key %q", entry.Key)
		}
	}

	keys := make([]string, 0, len(byKey))
	for key := range byKey {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	return &Registry{definitions: byKey, keys: keys}, nil
}

func validateDefinition(definition demo.PageDefinition) error {
	if definition.Key == "" {
		return fmt.Errorf("demo page has empty key")
	}
	if definition.Key != strings.TrimSpace(definition.Key) ||
		strings.HasPrefix(definition.Key, "/") ||
		strings.Contains(definition.Key, "\\") ||
		path.Clean(definition.Key) != definition.Key ||
		definition.Key == "." ||
		strings.ContainsAny(definition.Key, "?#") {
		return fmt.Errorf("demo page key %q is not canonical", definition.Key)
	}
	if definition.Content == nil {
		return fmt.Errorf("demo page %q has nil content factory", definition.Key)
	}
	return nil
}

// Lookup returns a copied page definition for a canonical key.
func (r *Registry) Lookup(key string) (demo.PageDefinition, bool) {
	definition, ok := r.definitions[key]
	return definition, ok
}

// MetaForKey returns crawler metadata for a registered page.
func (r *Registry) MetaForKey(key string) demo.PageMeta {
	definition, ok := r.Lookup(key)
	if !ok {
		return demo.DefaultMeta("Goshtoso")
	}
	return MetaForDefinition(definition)
}

// AllPublicMeta returns home metadata followed by every registered page in
// canonical key order.
func (r *Registry) AllPublicMeta() []demo.PageMeta {
	pages := make([]demo.PageMeta, 0, len(r.keys)+1)
	pages = append(pages, demo.HomeMeta())
	for _, key := range r.keys {
		pages = append(pages, MetaForDefinition(r.definitions[key]))
	}
	return pages
}

// MetaForDefinition derives crawler metadata from a page definition.
func MetaForDefinition(definition demo.PageDefinition) demo.PageMeta {
	title := definition.Title
	metaType := definition.Type
	if metaType == "" {
		metaType = "TechArticle"
	}

	switch {
	case strings.HasPrefix(definition.Key, "components/"):
		title = definition.Title + " Component - Goshtoso UI Library for Go"
	case strings.HasPrefix(definition.Key, "examples/"):
		title = definition.Title + " Example - Goshtoso Go UI Components"
		metaType = "SoftwareSourceCode"
	case definition.Key == "getting-started":
		title = "Getting Started with Goshtoso Go UI Components"
	case definition.Key == "docs/agents":
		title = "Using Goshtoso With AI Agents"
	case definition.Key == "docs/application-patterns":
		title = "Application Patterns for Goshtoso"
	case definition.Key == "docs/component-model":
		title = "Goshtoso Component Model"
	case definition.Key == "docs/icon-catalog":
		title = "Icon Catalog - Goshtoso UI Library for Go"
	case definition.Key == "docs/iconpack":
		title = "Icon Packs - Goshtoso UI Library for Go"
	case definition.Key == "docs/theme":
		title = "Themes - Goshtoso UI Library for Go"
	case definition.Key == "modules/charts":
		title = "Goshtoso Charts Module"
	case definition.Key == "modules/app-shells":
		title = "Goshtoso App Shells Module"
	case strings.HasPrefix(definition.Key, "modules/app-shells/"):
		title = definition.Title + " - Goshtoso App Shells"
	}

	description := definition.Description
	if description == "" {
		description = definition.Title + " documentation for building interactive server-rendered Go interfaces with Goshtoso, templ, HTMX, Alpine.js, and Tailwind CSS."
	}

	return demo.PageMeta{
		Title:       title,
		Description: description,
		Path:        "/" + definition.Key,
		Type:        metaType,
	}
}
