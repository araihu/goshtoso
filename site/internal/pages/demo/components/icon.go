package components

import (
	"strconv"
	"strings"

	"github.com/araihu/goshtoso/components/icon"
	"github.com/araihu/goshtoso/components/icon/heroicons"
)

// iconExample returns a copy-ready Go fragment for one bundled glyph. Zero
// options stay omitted because Icon defaults already express external mode and
// medium sizing; explicit values are retained when they change that default.
func iconExample(glyph heroicons.Glyph, cfg icon.Config) string {
	var fields []string
	fields = append(fields,
		"SpriteURL: heroicons.SpriteURL",
		"Symbol:    heroicons."+glyph.GoName,
	)
	if cfg.Size != "" {
		fields = append(fields, "Size:      icon.Size"+strings.ToUpper(string(cfg.Size)))
	}
	if label := strings.TrimSpace(cfg.Label); label != "" && !cfg.Decorative {
		fields = append(fields, "Label:     "+strconv.Quote(label))
	}
	if cfg.Decorative {
		fields = append(fields, "Decorative: true")
	}
	if rootClass := strings.TrimSpace(cfg.RootClass); rootClass != "" {
		fields = append(fields, "RootClass: "+strconv.Quote(rootClass))
	}

	var b strings.Builder
	b.WriteString("@icon.Icon(icon.Config{\n")
	for _, field := range fields {
		b.WriteString("    ")
		b.WriteString(field)
		b.WriteString(",\n")
	}
	b.WriteString("})")
	return b.String()
}
