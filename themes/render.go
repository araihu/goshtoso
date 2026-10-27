package themes

import (
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

var fontWeight = regexp.MustCompile(`^(normal|[1-9][0-9]{0,3}(\.[0-9]+)?( [1-9][0-9]{0,3}(\.[0-9]+)?)?)$`)

func validFontWeight(value string) bool {
	if value == "normal" {
		return true
	}
	if !fontWeight.MatchString(value) {
		return false
	}
	weights := strings.Fields(value)
	last := 0.0
	for _, weight := range weights {
		n, e := strconv.ParseFloat(weight, 64)
		if e != nil || !finite(n) || n < 1 || n > 1000 || n < last {
			return false
		}
		last = n
	}
	return true
}

func safeFontURL(raw string) bool {
	if raw == "" || strings.ContainsAny(raw, "<>\"'\\\r\n\x00") {
		return false
	}
	u, e := url.Parse(raw)
	if e != nil || u.User != nil || u.Fragment != "" {
		return false
	}
	return (u.Scheme == "https" && u.Host != "") || (u.Scheme == "" && u.Host == "" && strings.HasPrefix(u.Path, "/") && !strings.HasPrefix(u.Path, "//"))
}
func familyCSS(names []string) (string, error) {
	if len(names) == 0 || len(names) > 20 {
		return "", ErrInvalid
	}
	out := []string{}
	for _, name := range names {
		if name == "" || len(name) > 200 || strings.ContainsAny(name, "<>\x00\r\n") {
			return "", ErrInvalid
		}
		switch name {
		case "serif", "sans-serif", "monospace", "system-ui", "ui-serif", "ui-sans-serif", "ui-monospace", "cursive", "fantasy":
			out = append(out, name)
		default:
			out = append(out, strconv.Quote(name))
		}
	}
	return strings.Join(out, ","), nil
}

// RenderFontCSS returns validated font declarations and its CSS family expression.
func RenderFontCSS(ref FontRef, resolve FontResolver) (string, string, error) {
	if !ValidFontRef(ref) {
		return "", "", ErrInvalid
	}
	if resolve == nil {
		resolve = DefaultFontResolver
	}
	r, err := resolve(ref)
	if err != nil {
		return "", "", err
	}
	family := ""
	if r.URL == "" {
		var e error
		family, e = familyCSS(r.Families)
		if e != nil {
			return "", "", e
		}
	}
	var before strings.Builder
	if r.Stylesheet != "" {
		if !safeFontURL(r.Stylesheet) {
			return "", "", ErrInvalid
		}
		before.WriteString("@import url(" + strconv.Quote(r.Stylesheet) + ");\n")
	}
	if r.URL != "" {
		if !safeFontURL(r.URL) || !validFontWeight(r.Weight) || !slicesContain([]string{"normal", "italic", "oblique"}, r.Style) || !slicesContain([]string{"woff2", "woff", "truetype", "opentype"}, r.Format) {
			return "", "", ErrInvalid
		}
		alias := "goshtoso-font-" + strings.NewReplacer(":", "-", "/", "-", ".", "-").Replace(ref.ID)
		family = strconv.Quote(alias) + ",system-ui,sans-serif"
		fmt.Fprintf(&before, "@font-face{font-family:%s;src:url(%s) format(%s);font-weight:%s;font-style:%s;font-display:swap;}\n", strconv.Quote(alias), strconv.Quote(r.URL), strconv.Quote(r.Format), r.Weight, r.Style)
	}
	return before.String(), family, nil
}
func slicesContain(values []string, value string) bool {
	return slices.Contains(values, value)
}

// RenderCSS emits a safe literal stylesheet under the fixed custom root scope.
// A nil resolver uses the library catalog. External identities require a
// consumer resolver; definitions never provide font URLs or CSS selectors.
func RenderCSS(d Definition, resolve FontResolver) (string, error) {
	if err := Validate(d); err != nil {
		return "", err
	}
	if resolve == nil {
		resolve = DefaultFontResolver
	}
	titleCSS, title, err := RenderFontCSS(d.Typography.Title, resolve)
	if err != nil {
		return "", err
	}
	bodyCSS, body, err := RenderFontCSS(d.Typography.Body, resolve)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	// Imports must precede font-face/style rules. Each role is resolved separately.
	for _, css := range []string{titleCSS, bodyCSS} {
		for line := range strings.SplitSeq(css, "\n") {
			if strings.HasPrefix(line, "@import") {
				b.WriteString(line + "\n")
			}
		}
	}
	for _, css := range []string{titleCSS, bodyCSS} {
		for line := range strings.SplitSeq(css, "\n") {
			if strings.HasPrefix(line, "@font-face") {
				b.WriteString(line + "\n")
			}
		}
	}
	b.WriteString(":root[data-theme=\"custom\"]{\n")
	for _, spec := range Tokens() {
		fmt.Fprintf(&b, "--color-%s:%s;\n", spec.Key, d.Tokens[spec.Key].CSS())
	}
	fmt.Fprintf(&b, "--font-title:%s;--font-body:%s;--font-paragraph:%s;--radius-radius:%grem;color-scheme:light;\n}\n", title, body, body, d.RadiusRem)
	b.WriteString(":root[data-theme=\"custom\"].dark{color-scheme:dark;}\n")
	return b.String(), nil
}
