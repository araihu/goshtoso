package layout

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/a-h/templ"
)

func validID(s string) bool {
	return s != "" && len(s) <= 256 && utf8.ValidString(s) && !strings.ContainsFunc(s, unicode.IsSpace)
}
func validLabel(s string) bool {
	return utf8.ValidString(s) && strings.TrimSpace(s) != "" && utf8.RuneCountInString(s) <= 256
}
func invalid(path, code string) error { return ValidationError{path, code} }
func (c Config) validate() error {
	if !validID(c.ID) {
		return invalid("ID", "invalid_id")
	}
	if !validLabel(c.Label) {
		return invalid("Label", "invalid_label")
	}
	if c.Mode != ModeStack && c.Mode != ModeGrid && c.Mode != ModeMasonry {
		return invalid("Mode", "invalid_mode")
	}
	if c.Columns < 0 || c.Columns > 4 {
		return invalid("Columns", "invalid_columns")
	}
	if c.MinItemWidthRem != 0 && (math.IsNaN(c.MinItemWidthRem) || math.IsInf(c.MinItemWidthRem, 0) || c.MinItemWidthRem < 8 || c.MinItemWidthRem > 40) {
		return invalid("MinItemWidthRem", "invalid_min_width")
	}
	if c.Gap != GapMedium && c.Gap != GapSmall && c.Gap != GapLarge {
		return invalid("Gap", "invalid_gap")
	}
	if err := validateAttrs(c.RootAttrs, "RootAttrs", false); err != nil {
		return err
	}
	return c.validateItems()
}

func (c Config) validateItems() error {
	ids := map[string]bool{}
	dom := map[string]bool{c.ID: true}
	if c.enabled() {
		dom[c.ID+"--layout-instructions"] = true
	}
	for n, item := range c.Items {
		path := fmt.Sprintf("Items[%d]", n)
		if !validID(item.ID) {
			return invalid(path+".ID", "invalid_id")
		}
		if ids[item.ID] {
			return invalid(path+".ID", "duplicate_id")
		}
		ids[item.ID] = true
		if !validLabel(item.Label) {
			return invalid(path+".Label", "invalid_label")
		}
		if item.Span != SpanAuto && item.Span != SpanFullWidth {
			return invalid(path+".Span", "invalid_span")
		}
		if err := validateAttrs(item.RootAttrs, path+".RootAttrs", true); err != nil {
			return err
		}
		for key, value := range item.RootAttrs {
			if strings.EqualFold(key, "id") {
				id := value.(string)
				if dom[id] {
					return invalid(path+".RootAttrs[\""+key+"\"]", "duplicate_id")
				}
				dom[id] = true
			}
		}
	}
	return nil
}

func validateAttrs(attrs templ.Attributes, path string, item bool) error {
	keys := make([]string, 0, len(attrs))
	for key := range attrs {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	seen := map[string]bool{}
	for _, key := range keys {
		name := strings.ToLower(key)
		target := path + "[" + fmt.Sprintf("%q", key) + "]"
		if seen[name] {
			return invalid(target, "duplicate_attribute")
		}
		seen[name] = true
		if reserved(name) {
			return invalid(target, "reserved_attribute")
		}
		if !allowed(name, item) {
			return invalid(target, "unsupported_attribute")
		}
		text, ok := attrs[key].(string)
		if !ok || (name == "id" && !validID(text)) {
			return invalid(target, "invalid_attribute_value")
		}
	}
	return nil
}
func reserved(name string) bool {
	if after, ok := strings.CutPrefix(name, "x-bind:"); ok {
		return reserved(after)
	}
	if after, ok := strings.CutPrefix(name, ":"); ok {
		return reserved(after)
	}
	for _, prefix := range []string{"data-goshtoso-layout", "data-layout-"} {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return slices.Contains([]string{"role", "tabindex", "aria-label", "aria-labelledby", "aria-describedby", "hidden", "draggable", "x-layout", "x-data", "x-init", "style"}, name)
}
func allowed(name string, item bool) bool {
	if name == "id" {
		return item
	}
	if strings.HasPrefix(name, "x-bind:") {
		return slices.Contains([]string{"x-bind:aria-current", "x-bind:aria-expanded", "x-bind:aria-busy"}, name)
	}
	return slices.Contains([]string{"class", "title", "lang", "dir"}, name) || strings.HasPrefix(name, "aria-") || strings.HasPrefix(name, "data-") || strings.HasPrefix(name, "hx-") || strings.HasPrefix(name, "x-on:")
}
func normalizedAttrs(attrs templ.Attributes) templ.Attributes {
	out := templ.Attributes{}
	for k, v := range attrs {
		if !strings.EqualFold(k, "class") {
			out[strings.ToLower(k)] = v
		}
	}
	return out
}
func classes(base, extra string, attrs templ.Attributes) string {
	parts := []string{base, extra}
	for k, v := range attrs {
		if strings.EqualFold(k, "class") {
			parts = append(parts, v.(string))
		}
	}
	return strings.TrimSpace(strings.Join(parts, " "))
}
