package layout_test

import (
	"bytes"
	"errors"
	"fmt"
	"html"
	"math"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/components"
	"github.com/araihu/goshtoso/components/layout"
	"github.com/araihu/goshtoso/expressions"
)

func TestValidationWritesNothing(t *testing.T) {
	cases := []struct {
		name       string
		change     func(*layout.Config)
		path, code string
	}{
		{"missing id", func(c *layout.Config) { c.ID = "" }, "ID", "invalid_id"},
		{"whitespace id", func(c *layout.Config) { c.ID = "bad id" }, "ID", "invalid_id"},
		{"empty label", func(c *layout.Config) { c.Label = " " }, "Label", "invalid_label"},
		{"mode", func(c *layout.Config) { c.Mode = "dense" }, "Mode", "invalid_mode"},
		{"columns", func(c *layout.Config) { c.Columns = 5 }, "Columns", "invalid_columns"},
		{"width", func(c *layout.Config) { c.MinItemWidthRem = math.NaN() }, "MinItemWidthRem", "invalid_min_width"},
		{"gap", func(c *layout.Config) { c.Gap = "huge" }, "Gap", "invalid_gap"},
		{"duplicate item", func(c *layout.Config) { c.Items = append(c.Items, c.Items[0]) }, "Items[1].ID", "duplicate_id"},
		{"span", func(c *layout.Config) { c.Items[0].Span = "wide" }, "Items[0].Span", "invalid_span"},
		{"role", func(c *layout.Config) { c.RootAttrs = templ.Attributes{"ROLE": "grid"} }, `RootAttrs["ROLE"]`, "reserved_attribute"},
		{"duplicate attrs", func(c *layout.Config) { c.RootAttrs = templ.Attributes{"title": "a", "Title": "b"} }, `RootAttrs["title"]`, "duplicate_attribute"},
		{"generated id", func(c *layout.Config) { c.Items[0].RootAttrs = templ.Attributes{"id": "test--layout-instructions"} }, `Items[0].RootAttrs["id"]`, "duplicate_id"},
		{"nonstring attr", func(c *layout.Config) { c.RootAttrs = templ.Attributes{"class": 5} }, `RootAttrs["class"]`, "invalid_attribute_value"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := layout.Config{ID: "test", Label: "Test", Reorder: &layout.ReorderConfig{}, Items: []layout.Item{{ID: "item", Label: "Item"}}}
			tc.change(&c)
			var out bytes.Buffer
			err := layout.Container(c).Render(t.Context(), &out)
			var validation layout.ValidationError
			if !errors.As(err, &validation) || validation.Path != tc.path || validation.Code != tc.code || out.Len() != 0 {
				t.Fatalf("error=%v output=%q", err, out.String())
			}
		})
	}
}
func TestReactiveAttributeOwnership(t *testing.T) {
	for _, name := range []string{"x-bind:aria-label", ":aria-describedby", "X-BIND:ARIA-LABELLEDBY", "x-data", "data-layout-gap", "style", "onclick"} {
		var out bytes.Buffer
		err := layout.Container(layout.Config{ID: "safe", Label: "Safe", RootAttrs: templ.Attributes{name: "bad"}}).Render(t.Context(), &out)
		if err == nil || out.Len() != 0 {
			t.Fatalf("accepted %s", name)
		}
	}
	var out bytes.Buffer
	err := layout.Container(layout.Config{ID: "safe", Label: "Safe", RootAttrs: templ.Attributes{"x-bind:aria-current": "true", "x-bind:aria-busy": "false", "x-bind:aria-expanded": "true"}}).Render(t.Context(), &out)
	if err != nil {
		t.Fatal(err)
	}
}
func TestNaturalContentAndSafeMarkup(t *testing.T) {
	c := layout.Config{ID: `opaque:"id`, Label: `Collection <script>`, Columns: 2, Mode: layout.ModeMasonry, Gap: layout.GapSmall, Reorder: &layout.ReorderConfig{Group: "widgets"}, RootClass: "consumer", RootAttrs: templ.Attributes{"class": "extra", "data-dashboard-widget": "yes"}, Items: []layout.Item{{ID: "a", Label: `A <img>`}}}
	// Use a plain component to prove the public contract has no Card dependency.
	c.Items[0].Content = templ.Raw(`<p data-plain>Plain text</p>`)
	component := layout.Container(c)
	if component.Kind() != components.KindLayout {
		t.Fatal(component.Kind())
	}
	c.RootAttrs["class"] = "mutated"
	c.Items[0].Label = "mutated"
	c.Reorder.Disabled = true
	var out bytes.Buffer
	if err := component.Render(t.Context(), &out); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`role="list"`, `role="listitem"`, `data-plain`, `goshtoso-layout consumer extra`, `Collection &lt;script&gt;`, `A &lt;img&gt;`, `data-layout-reorder="true"`, `aria-describedby=`, `hidden`, `disabled`, `<noscript>`} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("missing %s", want)
		}
	}
	if strings.Contains(out.String(), "mutated") || strings.Contains(out.String(), "<script>") {
		t.Fatal(out.String())
	}
}
func TestExpressionSlotsAndPrecedence(t *testing.T) {
	custom := expressions.Layout{MoveAnnouncement: func(item, container, position string) string { return position + " / " + container + " / " + item }, HandleLabel: func(item string) string { return "Arrange " + item }}
	var out bytes.Buffer
	c := layout.Container(layout.Config{ID: "test", Label: "Test", Reorder: &layout.ReorderConfig{}, Items: []layout.Item{{ID: "a", Label: "Alpha"}}}).WithExpressions(custom)
	if err := c.Render(t.Context(), &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Arrange Alpha") || !strings.Contains(html.UnescapeString(out.String()), `"slot":"position"`) {
		t.Fatal(out.String())
	}
	out.Reset()
	err := c.WithExpressions(expressions.Layout{MoveAnnouncement: func(_, _, _ string) string { return "missing tokens" }}).Render(t.Context(), &out)
	if err == nil || out.Len() != 0 {
		t.Fatalf("err=%v bytes=%d", err, out.Len())
	}
}

func TestAttributeNamesAndIDAuthority(t *testing.T) {
	for _, key := range []string{"data-bad name", "x-on:click bad", "aria-label=bad", "data-bad\x00name"} {
		c := layout.Config{ID: "test", Label: "Test", RootAttrs: templ.Attributes{key: "value"}}
		var out bytes.Buffer
		err := layout.Container(c).Render(t.Context(), &out)
		var validation layout.ValidationError
		if !errors.As(err, &validation) || validation.Code != "unsupported_attribute" || out.Len() != 0 {
			t.Fatalf("%q: %v, output %q", key, err, out.String())
		}
	}
	for _, key := range []string{"id", "ID", "x-bind:id", "X-BIND:ID", ":Id"} {
		for _, item := range []bool{false, true} {
			if item && strings.EqualFold(key, "id") {
				continue
			}
			c := layout.Config{ID: "test", Label: "Test"}
			attrs := templ.Attributes{key: "replacement"}
			path := fmt.Sprintf("RootAttrs[%q]", key)
			if item {
				c.Items = []layout.Item{{ID: "a", Label: "Alpha", RootAttrs: attrs}}
				path = "Items[0]." + path
			} else {
				c.RootAttrs = attrs
			}
			var out bytes.Buffer
			err := layout.Container(c).Render(t.Context(), &out)
			var validation layout.ValidationError
			if !errors.As(err, &validation) || validation.Path != path || validation.Code != "reserved_attribute" || out.Len() != 0 {
				t.Fatalf("%s: %v, output %q", path, err, out.String())
			}
		}
	}
	var out bytes.Buffer
	err := layout.Container(layout.Config{ID: "test", Label: "Test", Items: []layout.Item{{ID: "a", Label: "Alpha", RootAttrs: templ.Attributes{"ID": "fragment-a"}}}}).Render(t.Context(), &out)
	if err != nil || !strings.Contains(out.String(), `id="fragment-a"`) {
		t.Fatalf("valid item ID: %v %s", err, out.String())
	}
}
