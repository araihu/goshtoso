// Package scrollregioncompat keeps the site independently buildable against
// its released Goshtoso dependency while the source checkout demonstrates a
// newer ScrollRegion naming API. It is site-only compatibility rendering, not
// a public component API.
package scrollregioncompat

import (
	"context"
	"fmt"
	"html"
	"io"
	"strings"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/components/scrollregion"
)

// Named renders the released ScrollRegion markup with the same focusable
// viewport landmark contract used by the current source demo. It deliberately
// calls only the released four-field Config API.
func Named(cfg scrollregion.Config, label string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, writer io.Writer) error {
		if strings.TrimSpace(label) == "" {
			label = "Scrollable content"
		}
		if !hasHorizontalAxis(cfg.ViewportClass) {
			cfg.ViewportClass = strings.TrimSpace(cfg.ViewportClass + " overflow-x-auto")
		}
		cfg.ViewportClass = strings.TrimSpace(cfg.ViewportClass + " rounded-radius focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary dark:focus-visible:outline-primary-dark")
		var rendered strings.Builder
		if err := scrollregion.ScrollRegion(cfg).Render(ctx, &rendered); err != nil {
			return err
		}
		const marker = `data-goshtoso-scroll-viewport tabindex="0"`
		if strings.Count(rendered.String(), marker) != 1 {
			return fmt.Errorf("released ScrollRegion viewport marker changed")
		}
		markup := strings.Replace(rendered.String(), marker, marker+` role="region" aria-label="`+html.EscapeString(label)+`"`, 1)
		_, err := io.WriteString(writer, markup)
		return err
	})
}

func hasHorizontalAxis(classes string) bool {
	for class := range strings.FieldsSeq(classes) {
		if strings.HasPrefix(class, "overflow-x-") {
			return true
		}
	}
	return false
}
