package combobox

import (
	"bytes"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestOptionsSupportProgrammaticKeyboardFocus(t *testing.T) {
	for _, server := range []bool{false, true} {
		name := "client"
		if server {
			name = "server"
		}
		t.Run(name, func(t *testing.T) {
			options := []Option{{Value: "enabled", Label: "Enabled"}, {Value: "disabled", Label: "Disabled", Disabled: true}}
			cfg := Config{ID: "keyboard", Name: "value", Mode: ModeMultiple, Source: Source{Static: options}}
			if server {
				cfg.Source = Source{LazyEndpoint: "/options"}
				cfg.ToggleEndpoint = "/toggle"
				cfg.OptionsEndpoint = "/options"
				cfg.ClearEndpoint = "/clear"
			}
			var buf bytes.Buffer
			if err := Combobox(cfg, State{Options: options}).Render(t.Context(), &buf); err != nil {
				t.Fatal(err)
			}
			document, err := html.Parse(strings.NewReader(buf.String()))
			if err != nil {
				t.Fatal(err)
			}
			count := 0
			var visit func(*html.Node)
			visit = func(node *html.Node) {
				attrs := map[string]string{}
				for _, attr := range node.Attr {
					attrs[attr.Key] = attr.Val
				}
				if attrs["role"] == "option" {
					count++
					if attrs["tabindex"] != "-1" {
						t.Errorf("option %q is not programmatically focusable", attrs["data-value"])
					}
					if attrs["data-value"] == "disabled" && attrs["aria-disabled"] != "true" {
						t.Error("disabled option lost its navigation exclusion")
					}
				}
				for child := node.FirstChild; child != nil; child = child.NextSibling {
					visit(child)
				}
			}
			visit(document)
			if count != 2 {
				t.Fatalf("got %d options", count)
			}
		})
	}
}
