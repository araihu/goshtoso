package layout

import (
	"context"
	"encoding/json"
	"github.com/a-h/templ"
	"io"
	"maps"
	"slices"
	"strings"

	"github.com/araihu/goshtoso/components"
	"github.com/araihu/goshtoso/expressions"
)

// Instance is an immutable renderable layout configuration.
type Instance struct {
	cfg                 Config
	expressionOverrides expressions.Layout
	contentDrag         bool
}

// Container creates a natural-height collection without card chrome.
func Container(c Config) Instance {
	c.Items = slices.Clone(c.Items)
	c.RootAttrs = cloneAttrs(c.RootAttrs)
	for n := range c.Items {
		c.Items[n].RootAttrs = cloneAttrs(c.Items[n].RootAttrs)
	}
	if c.Reorder != nil {
		reorder := *c.Reorder
		c.Reorder = &reorder
	}
	return Instance{cfg: c}
}

// Kind identifies the component type, independently of instance IDs.
func (Instance) Kind() components.Kind { return components.KindLayout }

// WithExpressions returns a copy with per-field text overrides.
func (i Instance) WithExpressions(e expressions.Layout) Instance {
	i.expressionOverrides = expressions.Merge(expressions.Set{Layout: i.expressionOverrides}, expressions.Set{Layout: e}).Layout
	return i
}

// WithContentDrag returns a copy that permits mouse or pen pickup from non-interactive content.
// When enabled, content text dragging moves the item instead of selecting text.
// Links, controls and editable regions keep their actions. Touch uses the handle.
func (i Instance) WithContentDrag(enabled bool) Instance {
	i.contentDrag = enabled
	return i
}

// Render validates the entire configuration before writing any markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	if err := i.cfg.validate(); err != nil {
		return err
	}
	ctx = expressions.With(ctx, expressions.Set{Layout: i.expressionOverrides})
	e := expressions.From(ctx).Layout
	announcement, err := moveParts(e.MoveAnnouncement)
	if err != nil {
		return err
	}
	c := i.cfg
	if c.Columns == 0 {
		c.Columns = 1
	}
	if c.MinItemWidthRem == 0 {
		c.MinItemWidthRem = 12
	}
	return containerTemplate(c, e, announcement, c.enabled() && i.contentDrag).Render(ctx, w)
}

var _ components.Component = Instance{}

// Slots are parsed before consumer labels enter the browser representation.
func moveParts(fn func(string, string, string) string) (string, error) {
	tokens := []string{"\uE000goshtoso-item\uE001", "\uE000goshtoso-container\uE001", "\uE000goshtoso-position\uE001"}
	text := fn(tokens[0], tokens[1], tokens[2])
	for _, token := range tokens {
		if !strings.Contains(text, token) {
			return "", invalid("Expressions.MoveAnnouncement", "invalid_attribute_value")
		}
	}
	parts := []map[string]string{}
	for text != "" {
		first, slot := len(text), -1
		for n, token := range tokens {
			if idx := strings.Index(text, token); idx >= 0 && idx < first {
				first = idx
				slot = n
			}
		}
		if first > 0 {
			parts = append(parts, map[string]string{"text": text[:first]})
		}
		if slot < 0 {
			break
		}
		parts = append(parts, map[string]string{"slot": []string{"item", "container", "position"}[slot]})
		text = text[first+len(tokens[slot]):]
	}
	data, err := json.Marshal(parts)
	return string(data), err
}

func cloneAttrs(attrs templ.Attributes) templ.Attributes { return maps.Clone(attrs) }
