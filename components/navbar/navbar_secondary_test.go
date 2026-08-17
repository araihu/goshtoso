package navbar

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

func renderNavbarResult(cfg Config) (string, error) {
	var buf bytes.Buffer
	err := Navbar(cfg).Render(context.Background(), &buf)
	return buf.String(), err
}

func renderSecondaryRowResult(cfg SecondaryConfig) (string, error) {
	var buf bytes.Buffer
	err := SecondaryRow(cfg).Render(context.Background(), &buf)
	return buf.String(), err
}

func TestSecondaryConfigValidateRejectsDeterministicErrors(t *testing.T) {
	t.Parallel()

	link := SecondaryLink{Label: "Overview", Href: "/overview"}
	action := templ.Raw(`<button type="button">Action</button>`)

	tests := []struct {
		name    string
		cfg     SecondaryConfig
		wantErr string
	}{
		{
			name: "content exclusive with links",
			cfg: SecondaryConfig{
				Links:   []SecondaryLink{link},
				Content: templ.Raw(`<div>content</div>`),
			},
			wantErr: `navbar: invalid Secondary.Content: cannot be combined with Links or Actions`,
		},
		{
			name: "nil action uses ascending index",
			cfg: SecondaryConfig{
				Actions: []templ.Component{action, nil},
			},
			wantErr: `navbar: invalid Secondary.Actions[1]: must not be nil`,
		},
		{
			name: "blank label trims whitespace",
			cfg: SecondaryConfig{
				Links: []SecondaryLink{
					link,
					{Label: " \n\t ", Href: "/details"},
				},
			},
			wantErr: `navbar: invalid Secondary.Links[1].Label: must not be blank after trimming whitespace`,
		},
		{
			name: "blank href trims whitespace",
			cfg: SecondaryConfig{
				Links: []SecondaryLink{
					{Label: "Overview", Href: " \t "},
				},
			},
			wantErr: `navbar: invalid Secondary.Links[0].Href: must not be blank after trimming whitespace`,
		},
		{
			name: "invalid current token",
			cfg: SecondaryConfig{
				Links: []SecondaryLink{
					{Label: "Overview", Href: "/overview", Current: SecondaryCurrent("step")},
				},
			},
			wantErr: `navbar: invalid Secondary.Links[0].Current: must be empty, page, or location`,
		},
		{
			name: "at most one current link",
			cfg: SecondaryConfig{
				Links: []SecondaryLink{
					{Label: "Overview", Href: "/overview", Current: SecondaryCurrentPage},
					{Label: "Details", Href: "/details", Current: SecondaryCurrentLocation},
				},
			},
			wantErr: `navbar: invalid Secondary.Links: at most one link may have Current`,
		},
		{
			name: "main navigation landmark collision",
			cfg: SecondaryConfig{
				Links:     []SecondaryLink{link},
				AriaLabel: "  MAIN \n navigation  ",
			},
			wantErr: `navbar: invalid Secondary.AriaLabel: must differ from main navigation`,
		},
		{
			name: "root attrs duplicate beats class type",
			cfg: SecondaryConfig{
				Links: []SecondaryLink{link},
				RootAttrs: templ.Attributes{
					"class": []string{"bad"},
					"Class": "dup",
				},
			},
			wantErr: `navbar: invalid Secondary.RootAttrs["class"]: duplicate case-insensitive attribute keys`,
		},
		{
			name: "root attrs reserved beats unsupported",
			cfg: SecondaryConfig{
				Links: []SecondaryLink{link},
				RootAttrs: templ.Attributes{
					"role":    "navigation",
					"z-index": "10",
				},
			},
			wantErr: `navbar: invalid Secondary.RootAttrs["role"]: reserved attribute`,
		},
		{
			name: "root attrs class must be string",
			cfg: SecondaryConfig{
				Links: []SecondaryLink{link},
				RootAttrs: templ.Attributes{
					"class": []string{"bad"},
				},
			},
			wantErr: `navbar: invalid Secondary.RootAttrs["class"]: class value must be a string`,
		},
		{
			name: "root attrs unsupported key",
			cfg: SecondaryConfig{
				Links: []SecondaryLink{link},
				RootAttrs: templ.Attributes{
					"title": "bad",
				},
			},
			wantErr: `navbar: invalid Secondary.RootAttrs["title"]: unsupported secondary-root attribute`,
		},
		{
			name: "link attrs duplicate beats action key",
			cfg: SecondaryConfig{
				Links: []SecondaryLink{
					{
						Label: "Overview",
						Href:  "/overview",
						LinkAttrs: templ.Attributes{
							"HX-POST": "/mutate",
							"hx-post": "/mutate-again",
						},
					},
				},
			},
			wantErr: `navbar: invalid Secondary.Links[0].LinkAttrs["hx-post"]: duplicate case-insensitive attribute keys`,
		},
		{
			name: "link attrs reserved attribute",
			cfg: SecondaryConfig{
				Links: []SecondaryLink{
					{
						Label: "Overview",
						Href:  "/overview",
						LinkAttrs: templ.Attributes{
							"href": "/override",
						},
					},
				},
			},
			wantErr: `navbar: invalid Secondary.Links[0].LinkAttrs["href"]: reserved attribute`,
		},
		{
			name: "link attrs class must be string",
			cfg: SecondaryConfig{
				Links: []SecondaryLink{
					{
						Label: "Overview",
						Href:  "/overview",
						LinkAttrs: templ.Attributes{
							"class": 99,
						},
					},
				},
			},
			wantErr: `navbar: invalid Secondary.Links[0].LinkAttrs["class"]: class value must be a string`,
		},
		{
			name: "link attrs action key beats unsupported",
			cfg: SecondaryConfig{
				Links: []SecondaryLink{
					{
						Label: "Overview",
						Href:  "/overview",
						LinkAttrs: templ.Attributes{
							"hx-post": "/mutate",
							"zzz":     "nope",
						},
					},
				},
			},
			wantErr: `navbar: invalid Secondary.Links[0].LinkAttrs["hx-post"]: action or mutation attribute must be placed in Actions`,
		},
		{
			name: "link attrs unsupported key",
			cfg: SecondaryConfig{
				Links: []SecondaryLink{
					{
						Label: "Overview",
						Href:  "/overview",
						LinkAttrs: templ.Attributes{
							"style": "display:none",
						},
					},
				},
			},
			wantErr: `navbar: invalid Secondary.Links[0].LinkAttrs["style"]: unsupported primitive-link attribute`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.cfg.Validate()
			require.EqualError(t, err, tt.wantErr)

			html, renderErr := renderSecondaryRowResult(tt.cfg)
			require.EqualError(t, renderErr, tt.wantErr)
			assert.Empty(t, html)
		})
	}
}

func TestConfigValidateRejectsPrimaryLandmarkCollisionsWhenSecondaryContentPresent(t *testing.T) {
	t.Parallel()

	baseSecondary := &SecondaryConfig{
		Links: []SecondaryLink{{Label: "Overview", Href: "/overview"}},
	}

	tests := []struct {
		name    string
		attrs   templ.Attributes
		wantErr string
	}{
		{
			name:    "aria label reserved",
			attrs:   templ.Attributes{"aria-label": "products"},
			wantErr: `navbar: invalid NavAttrs["aria-label"]: reserved when secondary content is present; primary label is main navigation`,
		},
		{
			name:    "aria labelledby reserved",
			attrs:   templ.Attributes{"aria-labelledby": "heading"},
			wantErr: `navbar: invalid NavAttrs["aria-labelledby"]: reserved when secondary content is present; primary label is main navigation`,
		},
		{
			name:    "role reserved",
			attrs:   templ.Attributes{"role": "banner"},
			wantErr: `navbar: invalid NavAttrs["role"]: reserved when secondary content is present; primary element is navigation`,
		},
		{
			name:    "aria hidden reserved",
			attrs:   templ.Attributes{"aria-hidden": "true"},
			wantErr: `navbar: invalid NavAttrs["aria-hidden"]: reserved when secondary content is present; primary landmark must remain exposed`,
		},
		{
			name:    "aria roledescription reserved",
			attrs:   templ.Attributes{"aria-roledescription": "tabs"},
			wantErr: `navbar: invalid NavAttrs["aria-roledescription"]: reserved when secondary content is present; primary landmark role is component-owned`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := Config{
				Brand:     templ.Raw(`<span>Acme</span>`),
				NavAttrs:  tt.attrs,
				Secondary: baseSecondary,
			}
			err := cfg.Validate()
			require.EqualError(t, err, tt.wantErr)

			html, renderErr := renderNavbarResult(cfg)
			require.EqualError(t, renderErr, tt.wantErr)
			assert.Empty(t, html)
		})
	}
}

func TestSecondaryRowAndNavbarSuppressEmptySecondaryMarkup(t *testing.T) {
	t.Parallel()

	baseHTML, err := renderNavbarResult(Config{
		Brand: templ.Raw(`<span>Acme</span>`),
		Links: []NavLink{{Label: "Home", Href: "/", Active: true}},
	})
	require.NoError(t, err)

	withAttributeOnly, err := renderNavbarResult(Config{
		Brand: templ.Raw(`<span>Acme</span>`),
		Links: []NavLink{{Label: "Home", Href: "/", Active: true}},
		Secondary: &SecondaryConfig{
			AriaLabel: "Custom secondary",
			RootClass: "consumer-secondary",
			RootAttrs: templ.Attributes{
				"id": "navbar-secondary-row",
			},
		},
	})
	require.NoError(t, err)

	assert.Equal(t, baseHTML, withAttributeOnly)
	assert.NotContains(t, withAttributeOnly, `data-navbar-shell="true"`)
	assert.NotContains(t, withAttributeOnly, `data-navbar-secondary="true"`)

	rowHTML, err := renderSecondaryRowResult(SecondaryConfig{
		AriaLabel: "Custom secondary",
		RootClass: "consumer-secondary",
		RootAttrs: templ.Attributes{
			"id": "navbar-secondary-row",
		},
	})
	require.NoError(t, err)
	assert.Empty(t, rowHTML)
}

func TestNavbarRendersSecondaryLinksActionsAndLinkState(t *testing.T) {
	t.Parallel()

	cfg := Config{
		Brand: templ.Raw(`<span>Acme</span>`),
		Secondary: &SecondaryConfig{
			Links: []SecondaryLink{
				{
					Label:   "Overview",
					Href:    "/overview",
					Current: SecondaryCurrentLocation,
					LinkAttrs: templ.Attributes{
						"id":               "nav-overview",
						"class":            "consumer-link",
						"data-track":       "overview",
						"aria-label":       "Overview label",
						"aria-describedby": "overview-help",
						"hx-get":           "/api/overview",
						"hx-target":        "#navbar-secondary-row",
						"hx-swap":          "outerHTML",
						"hx-push-url":      "/overview",
						"hx-select":        "#navbar-secondary-row",
						"hx-indicator":     "#loading",
						"hx-confirm":       "Continue?",
					},
				},
				{
					Label: "Details",
					Href:  "/details",
				},
			},
			Actions: []templ.Component{
				templ.Raw(`<button id="secondary-action" type="button">Filter</button>`),
			},
			AriaLabel:  " Product \n Sections ",
			Scrollable: true,
			RootClass:  "root-class",
			RootAttrs: templ.Attributes{
				"id":                "navbar-secondary-row",
				"class":             "root-attrs-class",
				"data-owner":        "consumer",
				"hx-get":            "/api/navbar/secondary",
				"x-data":            "{ open: false }",
				"aria-describedby":  "secondary-help",
				"aria-description":  "Secondary routes",
				"aria-details":      "secondary-details",
				"aria-errormessage": "secondary-errors",
				"aria-keyshortcuts": "Alt+Shift+N",
				"aria-live":         "polite",
				"aria-atomic":       "true",
				"aria-busy":         "false",
				"aria-relevant":     "additions text",
			},
		},
	}

	markup := renderNavbar(t, cfg)
	doc := parseTestHTML(t, markup)

	shell := mustFindElement(t, doc, "div", attrEquals("data-navbar-shell", "true"))
	children := childElements(shell)
	require.Len(t, children, 2)
	assert.Equal(t, "nav", children[0].Data)
	assert.Equal(t, "div", children[1].Data)
	assert.Equal(t, "main navigation", getAttr(children[0], "aria-label"))

	secondaryRoot := children[1]
	assert.Equal(t, "true", getAttr(secondaryRoot, "data-navbar-secondary"))
	assert.Equal(t, "navbar-secondary-row", getAttr(secondaryRoot, "id"))
	assert.Equal(t, "consumer", getAttr(secondaryRoot, "data-owner"))
	assert.Equal(t, "/api/navbar/secondary", getAttr(secondaryRoot, "hx-get"))
	assert.Equal(t, "{ open: false }", getAttr(secondaryRoot, "x-data"))
	assert.Equal(t, "secondary-help", getAttr(secondaryRoot, "aria-describedby"))
	assert.Equal(t, "Secondary routes", getAttr(secondaryRoot, "aria-description"))
	assert.Equal(t, "secondary-details", getAttr(secondaryRoot, "aria-details"))
	assert.Equal(t, "secondary-errors", getAttr(secondaryRoot, "aria-errormessage"))
	assert.Equal(t, "Alt+Shift+N", getAttr(secondaryRoot, "aria-keyshortcuts"))
	assert.Equal(t, "polite", getAttr(secondaryRoot, "aria-live"))
	assert.Equal(t, "true", getAttr(secondaryRoot, "aria-atomic"))
	assert.Equal(t, "false", getAttr(secondaryRoot, "aria-busy"))
	assert.Equal(t, "additions text", getAttr(secondaryRoot, "aria-relevant"))

	rootClass := getAttr(secondaryRoot, "class")
	assertClassTokensInOrder(t, rootClass, "min-w-0", "root-class", "root-attrs-class")

	secondaryChildren := childElements(secondaryRoot)
	require.Len(t, secondaryChildren, 2)
	assert.Equal(t, "nav", secondaryChildren[0].Data)
	assert.Equal(t, "div", secondaryChildren[1].Data)
	assert.Equal(t, "true", getAttr(secondaryChildren[1], "data-navbar-actions"))

	linksNav := secondaryChildren[0]
	assert.Equal(t, "Product Sections", getAttr(linksNav, "aria-label"))
	assert.NotEqual(t, "tablist", getAttr(linksNav, "role"))

	anchors := findElements(linksNav, "a", nil)
	require.Len(t, anchors, 2)
	assert.Equal(t, "Overview", textContent(anchors[0]))
	assert.Equal(t, "/overview", getAttr(anchors[0], "href"))
	assert.Equal(t, "location", getAttr(anchors[0], "aria-current"))
	assert.Equal(t, "nav-overview", getAttr(anchors[0], "id"))
	assert.Equal(t, "overview", getAttr(anchors[0], "data-track"))
	assert.Equal(t, "Overview label", getAttr(anchors[0], "aria-label"))
	assert.Equal(t, "overview-help", getAttr(anchors[0], "aria-describedby"))
	assert.Equal(t, "/api/overview", getAttr(anchors[0], "hx-get"))
	assert.Equal(t, "#navbar-secondary-row", getAttr(anchors[0], "hx-target"))
	assert.Equal(t, "outerHTML", getAttr(anchors[0], "hx-swap"))
	assert.Equal(t, "/overview", getAttr(anchors[0], "hx-push-url"))
	assert.Equal(t, "#navbar-secondary-row", getAttr(anchors[0], "hx-select"))
	assert.Equal(t, "#loading", getAttr(anchors[0], "hx-indicator"))
	assert.Equal(t, "Continue?", getAttr(anchors[0], "hx-confirm"))

	currentClass := getAttr(anchors[0], "class")
	assertClassTokens(t, currentClass,
		"inline-flex", "min-h-11", "min-w-11", "px-3", "py-2", "whitespace-nowrap",
		"border-b-2", "font-semibold", "text-on-surface-strong", "border-primary",
		"focus-visible:outline-2", "focus-visible:outline-offset-2", "consumer-link",
	)
	assertClassTokens(t, getAttr(anchors[1], "class"),
		"inline-flex", "min-h-11", "min-w-11", "px-3", "py-2", "whitespace-nowrap",
		"border-b-2", "border-transparent", "text-on-surface", "hover:border-outline-strong",
		"hover:text-on-surface-strong", "focus-visible:outline-2", "focus-visible:outline-offset-2",
	)

	actions := findElements(secondaryChildren[1], "button", attrEquals("id", "secondary-action"))
	require.Len(t, actions, 1)
	assert.Equal(t, "Filter", textContent(actions[0]))
	assert.Equal(t, 1, strings.Count(markup, `id="secondary-action"`))
	assert.NotContains(t, markup, `role="tab"`)
	assert.NotContains(t, markup, `aria-selected`)
	assert.NotContains(t, markup, `aria-controls`)
	assert.NotContains(t, markup, `tabpanel`)
}

func TestNavbarRendersSecondaryContentEscapeHatch(t *testing.T) {
	t.Parallel()

	markup := renderNavbar(t, Config{
		Brand: templ.Raw(`<span>Acme</span>`),
		Secondary: &SecondaryConfig{
			Content: templ.Raw(`<section id="secondary-content"><button type="button">Custom</button></section>`),
			RootAttrs: templ.Attributes{
				"id": "navbar-secondary-row",
			},
		},
	})
	doc := parseTestHTML(t, markup)
	secondaryRoot := mustFindElement(t, doc, "div", attrEquals("data-navbar-secondary", "true"))

	children := childElements(secondaryRoot)
	require.Len(t, children, 1)
	assert.Equal(t, "section", children[0].Data)
	assert.Equal(t, "secondary-content", getAttr(children[0], "id"))
	assert.Empty(t, findElements(secondaryRoot, "nav", nil))
	assert.Empty(t, findElements(secondaryRoot, "div", attrEquals("data-navbar-actions", "true")))
}

func TestSecondaryRowMatchesNavbarSecondaryMarkup(t *testing.T) {
	t.Parallel()

	secondary := SecondaryConfig{
		Links: []SecondaryLink{
			{Label: "Overview", Href: "/overview", Current: SecondaryCurrentPage},
			{Label: "Details", Href: "/details"},
		},
		Actions: []templ.Component{
			templ.Raw(`<button id="secondary-action" type="button">Filter</button>`),
		},
		RootClass: "root-class",
		RootAttrs: templ.Attributes{
			"id": "navbar-secondary-row",
		},
	}

	navbarMarkup := renderNavbar(t, Config{
		Brand:     templ.Raw(`<span>Acme</span>`),
		Secondary: &secondary,
	})
	rowMarkup := renderSecondaryRow(t, secondary)

	navbarDoc := parseTestHTML(t, navbarMarkup)
	rowDoc := parseTestHTML(t, rowMarkup)

	navbarRoot := mustFindElement(t, navbarDoc, "div", attrEquals("data-navbar-secondary", "true"))
	rowRoot := mustFindElement(t, rowDoc, "div", attrEquals("data-navbar-secondary", "true"))

	assert.Equal(t, renderNodeHTML(t, navbarRoot), renderNodeHTML(t, rowRoot))
}

func renderSecondaryRow(t *testing.T, cfg SecondaryConfig) string {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, SecondaryRow(cfg).Render(context.Background(), &buf))
	return buf.String()
}

func parseTestHTML(t *testing.T, markup string) *html.Node {
	t.Helper()
	doc, err := html.Parse(strings.NewReader("<!doctype html><html><body>" + markup + "</body></html>"))
	require.NoError(t, err)
	return doc
}

func renderNodeHTML(t *testing.T, node *html.Node) string {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, html.Render(&buf, node))
	return buf.String()
}

func mustFindElement(t *testing.T, root *html.Node, tag string, match func(*html.Node) bool) *html.Node {
	t.Helper()
	nodes := findElements(root, tag, match)
	require.NotEmpty(t, nodes, "missing <%s> element", tag)
	require.Len(t, nodes, 1, "expected one <%s> element", tag)
	return nodes[0]
}

func findElements(root *html.Node, tag string, match func(*html.Node) bool) []*html.Node {
	var out []*html.Node
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode && (tag == "" || node.Data == tag) {
			if match == nil || match(node) {
				out = append(out, node)
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)
	return out
}

func childElements(node *html.Node) []*html.Node {
	var children []*html.Node
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode {
			children = append(children, child)
		}
	}
	return children
}

func attrEquals(name, value string) func(*html.Node) bool {
	return func(node *html.Node) bool {
		return getAttr(node, name) == value
	}
}

func getAttr(node *html.Node, name string) string {
	for _, attr := range node.Attr {
		if strings.EqualFold(attr.Key, name) {
			return attr.Val
		}
	}
	return ""
}

func textContent(node *html.Node) string {
	var parts []string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			parts = append(parts, n.Data)
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(node)
	return strings.Join(strings.Fields(strings.Join(parts, " ")), " ")
}

func assertClassTokens(t *testing.T, classValue string, want ...string) {
	t.Helper()
	got := strings.Fields(classValue)
	for _, token := range want {
		assert.Containsf(t, got, token, "missing class token %q in %q", token, classValue)
	}
}

func assertClassTokensInOrder(t *testing.T, classValue string, want ...string) {
	t.Helper()
	got := strings.Fields(classValue)
	indexes := make([]int, 0, len(want))
	for _, token := range want {
		index := slicesIndex(got, token)
		require.NotEqualf(t, -1, index, "missing class token %q in %q", token, classValue)
		indexes = append(indexes, index)
	}
	require.True(t, sort.IntsAreSorted(indexes), "class tokens not in expected order: %v in %q", want, classValue)
}

func slicesIndex(values []string, want string) int {
	for i, value := range values {
		if value == want {
			return i
		}
	}
	return -1
}

func TestValidationErrorError(t *testing.T) {
	t.Parallel()

	err := ValidationError{Path: `Secondary.RootAttrs["role"]`, Reason: "reserved attribute"}
	assert.Equal(t, `navbar: invalid Secondary.RootAttrs["role"]: reserved attribute`, err.Error())
}

func TestSecondaryConfigValidateAcceptsAttributeOnlyAndNormalizesNilSecondaryInConfigValidate(t *testing.T) {
	t.Parallel()

	require.NoError(t, SecondaryConfig{
		RootClass: "secondary-row",
		RootAttrs: templ.Attributes{"id": "navbar-secondary-row"},
	}.Validate())
	require.NoError(t, Config{
		Brand:     templ.Raw(`<span>Acme</span>`),
		Secondary: &SecondaryConfig{RootAttrs: templ.Attributes{"id": "navbar-secondary-row"}},
	}.Validate())
}

func TestValidationErrorPathsStayStableForRepresentativeLaterIndexes(t *testing.T) {
	t.Parallel()

	cfg := SecondaryConfig{
		Links: []SecondaryLink{
			{Label: "One", Href: "/one"},
			{Label: "Two", Href: "/two"},
			{Label: " \t ", Href: "/three"},
		},
	}

	err := cfg.Validate()
	require.EqualError(t, err, `navbar: invalid Secondary.Links[2].Label: must not be blank after trimming whitespace`)
}

func TestSecondaryConfigValidateRejectsDuplicateCaseInsensitiveKeyByCanonicalOrder(t *testing.T) {
	t.Parallel()

	cfg := SecondaryConfig{
		Links: []SecondaryLink{{Label: "Overview", Href: "/overview"}},
		RootAttrs: templ.Attributes{
			"data-z": "ok",
			"Class":  "dup-one",
			"class":  "dup-two",
		},
	}

	err := cfg.Validate()
	require.EqualError(t, err, `navbar: invalid Secondary.RootAttrs["class"]: duplicate case-insensitive attribute keys`)
}

func TestSecondaryConfigValidateRejectsLinkAttributeDuplicateByCanonicalLowercaseKey(t *testing.T) {
	t.Parallel()

	cfg := SecondaryConfig{
		Links: []SecondaryLink{
			{
				Label: "Overview",
				Href:  "/overview",
				LinkAttrs: templ.Attributes{
					"DATA-track": "one",
					"data-track": "two",
				},
			},
		},
	}

	err := cfg.Validate()
	require.EqualError(t, err, `navbar: invalid Secondary.Links[0].LinkAttrs["data-track"]: duplicate case-insensitive attribute keys`)
}

func TestSecondaryRenderErrorsProduceNoBytesInHelpers(t *testing.T) {
	t.Parallel()

	markup, err := renderSecondaryRowResult(SecondaryConfig{
		Links:   []SecondaryLink{{Label: "Overview", Href: "/overview"}},
		Content: templ.Raw(`<div>content</div>`),
	})
	require.EqualError(t, err, `navbar: invalid Secondary.Content: cannot be combined with Links or Actions`)
	assert.Empty(t, markup)
}

func TestNavbarRenderErrorsProduceNoBytesForSecondaryValidation(t *testing.T) {
	t.Parallel()

	markup, err := renderNavbarResult(Config{
		Brand: templ.Raw(`<span>Acme</span>`),
		Secondary: &SecondaryConfig{
			Links: []SecondaryLink{{Label: "Overview", Href: "/overview", LinkAttrs: templ.Attributes{"onclick": "bad()"}}},
		},
	})
	require.EqualError(t, err, `navbar: invalid Secondary.Links[0].LinkAttrs["onclick"]: action or mutation attribute must be placed in Actions`)
	assert.Empty(t, markup)
}

func TestSecondaryConfigValidateErrorFormatting(t *testing.T) {
	t.Parallel()

	cases := []ValidationError{
		{Path: "Secondary.Content", Reason: "cannot be combined with Links or Actions"},
		{Path: `Secondary.Links[0].LinkAttrs["href"]`, Reason: "reserved attribute"},
	}
	for _, tc := range cases {
		assert.Equal(t, fmt.Sprintf("navbar: invalid %s: %s", tc.Path, tc.Reason), tc.Error())
	}
}
