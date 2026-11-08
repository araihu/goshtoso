package combobox

import (
	"bytes"
	"strings"
	"testing"

	"github.com/araihu/goshtoso/expressions"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

func TestSearchInputNamesOnlyServerQueriesAndUsesLocalizedLabel(t *testing.T) {
	for _, server := range []bool{false, true} {
		for _, label := range []string{"Search…", "Pesquisar opções"} {
			name := "client/" + label
			if server {
				name = "server/" + label
			}
			t.Run(name, func(t *testing.T) {
				cfg := Config{ID: "accounts", Name: "account", EnableSearch: true, Source: Source{Static: []Option{{Value: "alice", Label: "Alice"}}}}
				if server {
					cfg.Source = Source{LazyEndpoint: "/options"}
					cfg.OptionsEndpoint = "/options"
				}
				ctx := expressions.With(t.Context(), expressions.Set{Combobox: expressions.Combobox{SearchPlaceholder: label, EmptyText: "Nenhuma opção"}})
				var out bytes.Buffer
				require.NoError(t, body(cfg, State{Options: []Option{{Value: "alice", Label: "Alice"}}, Selected: []string{"alice"}}).Render(ctx, &out))
				doc, err := html.Parse(strings.NewReader(out.String()))
				require.NoError(t, err)
				search := searchInputAttrs(doc)
				require.NotNil(t, search)
				require.Equal(t, label, search["aria-label"])
				require.Equal(t, label, search["placeholder"])
				require.Contains(t, out.String(), `aria-labelledby="accounts-trigger"`)
				require.Equal(t, "accounts-options", cfg.triggerAttributes()["aria-controls"])
				if server {
					require.Equal(t, "q", search["name"])
					require.Equal(t, "/options", search["hx-get"])
				} else {
					_, present := search["name"]
					require.False(t, present, "internal search must not become a successful form control")
				}
				if !server {
					require.Contains(t, out.String(), `data-combobox-empty`)
					require.Contains(t, out.String(), "Nenhuma opção")
				}
				require.Contains(t, out.String(), `type="hidden" name="account" value="alice"`, "selection retains the public form field")
			})
		}
	}
}

func searchInputAttrs(node *html.Node) map[string]string {
	attrs := map[string]string{}
	for _, attr := range node.Attr {
		attrs[attr.Key] = attr.Val
	}
	if _, ok := attrs["data-combobox-search"]; ok {
		return attrs
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if found := searchInputAttrs(child); found != nil {
			return found
		}
	}
	return nil
}

func TestSelectedTriggerLabelKeepsReadableTextAndAccentBorder(t *testing.T) {
	state := State{Selected: []string{"account"}}
	label := triggerLabelClass(state)
	require.Contains(t, label, "text-on-surface-strong")
	require.Contains(t, label, "dark:text-on-surface-dark-strong")
	require.NotContains(t, label, "text-secondary")
	require.Contains(t, triggerBtnClass(state), "border-secondary")
}
