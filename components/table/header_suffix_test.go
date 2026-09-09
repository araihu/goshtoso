package table

import (
	"github.com/a-h/templ"
	"testing"
)

func TestHeaderSuffixSurvivesHeaderFragments(t *testing.T) {
	for _, sortable := range []bool{false, true} {
		cfg := Config{ID: "services", HTMX: &HTMXConfig{Endpoint: "/services"}, Columns: []Column{{Key: "pinned", Label: "Pinned", Sortable: sortable, HeaderSuffix: templ.Raw(`<button aria-label="Pinning help">?</button>`)}}}
		for _, component := range []templ.Component{Table(cfg), TableHeadContent(cfg)} {
			html := renderT(t, component)
			mustContainAll(t, html, `scope="col"`, `Pinned`, `aria-label="Pinning help"`)
			if sortable {
				mustContainAll(t, html, `x-on:click.stop`, `hx-disinherit="*"`)
			}
		}
	}
}
