package table

import (
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/expressions"
	"golang.org/x/net/html"
)

func TestMobileFieldsAndFragments(t *testing.T) {
	cfg := Config{
		ID: "hosts", Mobile: &MobileConfig{PrimaryColumn: "name", Expandable: true},
		Columns: []Column{{Key: "pin", Label: "Pinned"}, {Key: "name", Label: "Host", MobileSecondary: true}, {Key: "arch", Label: "Architecture", MobileSecondary: true}},
		Rows:    []Row{{ID: "worker/one", Cells: map[string]Cell{"name": {Text: "Worker"}, "arch": {Text: "arm64"}, "pin": {Component: templ.Raw(`<input id="pin-worker" aria-label="Pinned" type="checkbox">`)}}}},
	}
	for _, component := range []templ.Component{Table(cfg), TableRows(cfg), TableRow(cfg, cfg.Rows[0])} {
		html := renderT(t, component)
		mustContainAll(t, html, `data-table-mobile-row="worker/one"`, `data-table-mobile-primary`, `headers="hosts-column-name"`, `id="hosts-mobile-worker%2Fone-arch"`, `aria-controls="hosts-mobile-worker%2Fone-arch"`, `arm64`, `Additional details`)
		if count := strings.Count(html, `id="pin-worker"`); count != 1 {
			t.Fatalf("each interactive cell must render once, got %d", count)
		}
		if strings.Contains(html, `id="hosts-mobile-worker%2Fone-name"`) {
			t.Fatal("the primary field must not collapse even when marked secondary")
		}
	}
	html := renderT(t, Table(cfg).WithExpressions(expressions.Table{DetailsLabel: "More information"}))
	mustContainAll(t, html, `mobileRows: {}`, `More information`, `id="hosts-column-arch"`)
}

func TestMobileExpansionIsOptionalAndNeedsStableRows(t *testing.T) {
	for _, tc := range []struct {
		name      string
		mobile    *MobileConfig
		rowID     string
		secondary bool
	}{
		{"default", nil, "one", true},
		{"all fields", &MobileConfig{PrimaryColumn: "name"}, "one", true},
		{"anonymous row", &MobileConfig{Expandable: true}, "", true},
		{"no secondary fields", &MobileConfig{Expandable: true}, "one", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := Config{Mobile: tc.mobile, Columns: []Column{{Key: "name", Label: "Name"}, {Key: "arch", Label: "Architecture", MobileSecondary: tc.secondary}}, Rows: []Row{{ID: tc.rowID, Cells: map[string]Cell{"name": {Text: "Worker"}, "arch": {Text: "arm64"}}}}}
			html := renderT(t, Table(cfg))
			mustContainAll(t, html, `Worker`, `arm64`)
			if strings.Contains(html, `data-table-mobile-row=`) || strings.Contains(html, `aria-controls=`) {
				t.Fatal("non-expandable rows must expose every field without a toggle")
			}
			if tc.mobile == nil && strings.Contains(html, "goshtoso-table-responsive") {
				t.Fatal("the default table must retain its scroll layout")
			}
		})
	}
}

func TestMobileRowActionsKeepInteractiveCellsSeparate(t *testing.T) {
	for _, row := range []Row{
		{Link: "/record", LinkMode: LinkFull},
		{OnClick: "window.openRecord()"},
		{HTMX: &RowHTMXConfig{Get: "/record", Target: "#detail"}},
	} {
		row.ID = "__proto__"
		row.Cells = map[string]Cell{"name": {Component: templ.Raw(`<a href="#nested">Nested link</a>`)}, "arch": {Text: "arm64"}}
		cfg := Config{ID: "records", Mobile: &MobileConfig{PrimaryColumn: "name", Expandable: true}, Columns: []Column{{Key: "name", Label: "Name"}, {Key: "arch", Label: "Architecture", MobileSecondary: true}}}
		html := renderT(t, TableRow(cfg, row))
		mustContainAll(t, html, `role="row"`, `href="#nested"`, `row:`, `Additional details`)
		opening, _, _ := strings.Cut(html, ">")
		for _, invalid := range []string{`role="link"`, `role="button"`, `tabindex="0"`, `cursor-pointer`} {
			if strings.Contains(opening, invalid) {
				t.Fatalf("stacked row must keep native table semantics, found %s", invalid)
			}
		}
	}
}

func TestMobileAuxiliaryRowsKeepTableRoles(t *testing.T) {
	base := Config{ID: "hosts", Mobile: &MobileConfig{}, Columns: []Column{{Key: "name", Label: "Host"}}, Rows: []Row{{ID: "one", Expandable: true, Detail: templ.Raw(`<p>Details</p>`), Actions: templ.Raw(`<button>Action</button>`), Cells: map[string]Cell{"name": {Text: "Worker"}}}}}
	lazy, infinite := base, base
	lazy.HTMX, lazy.LazyLoad = &HTMXConfig{Endpoint: "/rows"}, true
	infinite.HTMX, infinite.InfiniteScroll = &HTMXConfig{Endpoint: "/rows"}, &InfiniteScrollConfig{HasMore: true, NextPage: 2}
	for name, cfg := range map[string]Config{"actions and details": base, "loading": lazy, "sentinel": infinite} {
		t.Run(name, func(t *testing.T) {
			document, err := html.Parse(strings.NewReader(renderT(t, Table(cfg))))
			if err != nil {
				t.Fatal(err)
			}
			var visit func(*html.Node)
			visit = func(node *html.Node) {
				if expected, ok := map[string]string{"tr": "row", "td": "cell"}[node.Data]; ok && node.Type == html.ElementNode {
					role := ""
					for _, attr := range node.Attr {
						if attr.Key == "role" {
							role = attr.Val
						}
					}
					if role != expected {
						t.Errorf("%s has role %q, want %q", node.Data, role, expected)
					}
				}
				for child := node.FirstChild; child != nil; child = child.NextSibling {
					visit(child)
				}
			}
			visit(document)
		})
	}
}
