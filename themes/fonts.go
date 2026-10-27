package themes

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"slices"
	"strings"
)

// FontOption is a consumer-visible font choice; each upload is a distinct item.
type FontOption struct {
	Source   string   `json:"source"`
	ID       string   `json:"id"`
	Label    string   `json:"label"`
	Weight   string   `json:"weight"`
	Style    string   `json:"style"`
	Variable bool     `json:"variable"`
	Families []string `json:"families,omitempty"`
	URL      string   `json:"url,omitempty"`
	Format   string   `json:"format,omitempty"`
}

// Ref returns the durable identity represented by this option.
func (f FontOption) Ref() FontRef { return FontRef{Source: f.Source, ID: f.ID} }

// ResolvedFont contains trusted renderer inputs supplied by a font provider.
type ResolvedFont struct {
	Families   []string
	URL        string
	Format     string
	Weight     string
	Style      string
	Stylesheet string
}

// FontResolver maps an identity to trusted font assets and family names.
type FontResolver func(FontRef) (ResolvedFont, error)

func stackRef(stack string) FontRef {
	canonical := families(stack)
	for i := range canonical {
		canonical[i] = strings.ToLower(canonical[i])
	}
	sum := sha256.Sum256([]byte(strings.Join(canonical, "\x00")))
	return FontRef{Source: "system", ID: "stack-" + hex.EncodeToString(sum[:12])}
}
func families(stack string) []string {
	out := []string{}
	for name := range strings.SplitSeq(stack, ",") {
		name = strings.Trim(strings.TrimSpace(name), "\"'")
		if name != "" {
			out = append(out, name)
		}
	}
	return out
}

// DefaultFonts returns the built-in family stacks and curated Google choices.
// The result and every family slice are caller-owned. Additional providers are
// injected by consumers; this function never fetches remote font catalogs.
func DefaultFonts() []FontOption {
	byID := map[string]FontOption{}
	add := func(stack string) {
		if stack == "" {
			return
		}
		fs := families(stack)
		ref := stackRef(stack)
		byID[ref.ID] = FontOption{Source: ref.Source, ID: ref.ID, Label: fs[0] + " · system", Weight: "default", Style: "normal", Families: fs}
	}
	add("system-ui, sans-serif")
	add("ui-serif, serif")
	add("ui-monospace, monospace")
	add(source.Defaults["--font-body"])
	add(source.Defaults["--font-title"])
	for _, values := range source.Overrides {
		add(values["--font-body"])
		add(values["--font-title"])
	}
	out := make([]FontOption, 0, len(byID)+19)
	for _, f := range byID {
		out = append(out, f)
	}
	for _, name := range []string{"Jost", "Inter", "Roboto", "Lato", "Montserrat", "Instrument Sans", "Open Sans", "Oswald", "Poppins", "Raleway", "Source Code Pro", "Playfair Display", "Merriweather", "Denk One", "Concert One", "Inconsolata", "Space Mono", "Playpen Sans", "Fira Code", "JetBrains Mono"} {
		id := strings.ToLower(strings.ReplaceAll(name, " ", "-"))
		out = append(out, FontOption{Source: "google", ID: id, Label: name + " · Google Fonts", Weight: "default", Style: "normal", Families: []string{name, "sans-serif"}})
	}
	out = append(out, FontOption{Source: "bundled", ID: "go-regular", Label: "Go Regular · hosted", Families: []string{"Go"}, URL: "/assets/fonts/go-regular.197d9f3703b4.ttf", Format: "truetype", Weight: "400", Style: "normal"})
	slices.SortFunc(out, func(a, b FontOption) int {
		if n := strings.Compare(a.Source, b.Source); n != 0 {
			return n
		}
		if n := strings.Compare(a.Label, b.Label); n != 0 {
			return n
		}
		return strings.Compare(a.ID, b.ID)
	})
	return out
}

// ResolveFont resolves a source/id from a supplied catalog, rejecting unknown IDs.
func ResolveFont(catalog []FontOption, ref FontRef) (ResolvedFont, error) {
	for _, f := range catalog {
		if f.Ref() != ref {
			continue
		}
		r := ResolvedFont{Families: slices.Clone(f.Families), URL: f.URL, Format: f.Format, Weight: f.Weight, Style: f.Style}
		if r.Weight == "unknown" {
			r.Weight = "normal"
		}
		if r.Style == "unknown" {
			r.Style = "normal"
		}
		if ref.Source == "google" {
			if _, err := familyCSS(f.Families); err != nil {
				return ResolvedFont{}, fmt.Errorf("%w: Google font family", ErrInvalid)
			}
			q := url.Values{"family": {f.Families[0]}, "display": {"swap"}}
			r.Stylesheet = "https://fonts.googleapis.com/css2?" + q.Encode()
		}
		return r, nil
	}
	return ResolvedFont{}, fmt.Errorf("%w: font %s/%s", ErrInvalid, ref.Source, ref.ID)
}

// DefaultFontResolver resolves only the library's own immutable catalog.
func DefaultFontResolver(ref FontRef) (ResolvedFont, error) { return ResolveFont(DefaultFonts(), ref) }
