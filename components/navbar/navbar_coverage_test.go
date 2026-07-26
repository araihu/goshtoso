package navbar

import (
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/stretchr/testify/assert"
)

func TestCoverageActionPartitioning(t *testing.T) {
	leftDefault := ActionItem{Content: templ.Raw(`<span data-action="left-default"></span>`)}
	leftExplicit := ActionItem{Content: templ.Raw(`<span data-action="left-explicit"></span>`), Position: ActionLeft}
	right := ActionItem{Content: templ.Raw(`<span data-action="right"></span>`), Position: ActionRight}

	cfg := Config{Actions: []ActionItem{leftDefault, leftExplicit, right}}

	leftActions := cfg.leftActions()
	assert.Len(t, leftActions, 2)
	assert.Equal(t, ActionPosition(""), leftActions[0].Position)
	assert.Equal(t, ActionLeft, leftActions[1].Position)

	rightActions := cfg.rightActions()
	assert.Len(t, rightActions, 1)
	assert.Equal(t, ActionRight, rightActions[0].Position)
}

func TestCoverageNavbarClassHelpers(t *testing.T) {
	cfg := Config{NavClass: "sticky top-0"}
	assert.Contains(t, cfg.navClasses(), "border-b border-outline")
	assert.Contains(t, cfg.navClasses(), "sticky top-0")

	activeLink := linkClasses(true)
	assert.Contains(t, activeLink, "font-bold")
	assert.NotContains(t, activeLink, "text-on-surface underline-offset")

	inactiveLink := linkClasses(false)
	assert.Contains(t, inactiveLink, "font-medium")
	assert.Contains(t, inactiveLink, "text-on-surface")
	assert.NotContains(t, inactiveLink, "font-bold")

	dangerItem := menuItemClasses(true)
	assert.Contains(t, dangerItem, "text-danger")
	assert.Contains(t, dangerItem, "hover:bg-danger/5")

	standardItem := menuItemClasses(false)
	assert.Contains(t, standardItem, "text-on-surface")
	assert.Contains(t, standardItem, "hover:text-on-surface-strong")
	assert.NotContains(t, standardItem, "text-danger")
}

func TestCoverageRenderFullNavbarBranches(t *testing.T) {
	html := renderNavbar(t, Config{
		Brand:     templ.Raw(`<span>Acme</span>`),
		BrandHref: "/dashboard",
		Links: []NavLink{
			{Label: "Home", Href: "/home", Active: true, LinkAttrs: templ.Attributes{"data-nav": "home"}},
			{Label: "Docs", Href: "/docs"},
		},
		Actions: []ActionItem{
			{Content: templ.Raw(`<button data-action="left">Left</button>`)},
			{Content: templ.Raw(`<button data-action="right">Right</button>`), Position: ActionRight},
		},
		User: &UserProfile{
			Name:  "Ada Lovelace",
			Email: "ada@example.test",
		},
		UserMenu: []UserMenuItem{
			{Label: "Profile", Href: "/profile", Icon: templ.Raw(`<svg data-icon="profile"></svg>`), LinkAttrs: templ.Attributes{"data-menu": "profile"}},
			{Label: "Sign out", Href: "/logout", Danger: true},
		},
		NavAttrs: templ.Attributes{"data-navbar": "coverage"},
	})

	for _, want := range []string{
		`href="/dashboard"`,
		`data-navbar="coverage"`,
		`data-action="left"`,
		`data-action="right"`,
		`data-nav="home"`,
		`aria-current="page"`,
		`aria-label="user menu"`,
		`x-bind:aria-expanded="userDropDownIsOpen"`,
		`role="menu"`,
		`data-menu="profile"`,
		`data-icon="profile"`,
		`Ada Lovelace`,
		`ada@example.test`,
		`Sign out`,
		`text-danger`,
		`x-on:keydown.escape.window="mobileMenuIsOpen = false"`,
		`x-bind:aria-label="mobileMenuIsOpen ? 'Close mobile menu' : 'Open mobile menu'"`,
		`x-show="mobileMenuIsOpen"`,
	} {
		assert.Contains(t, html, want)
	}

	assert.Equal(t, 2, strings.Count(html, `href="/profile"`), "user menu item renders in desktop and mobile menus")
	assert.Contains(t, html, `class="size-10 rounded-full`)
	assert.Contains(t, html, `aria-hidden="true"`)
}

func TestCoverageRenderDefaultsAndSuppressesEmptyOptionalRegions(t *testing.T) {
	html := renderNavbar(t, Config{
		Brand: templ.Raw(`<span>Acme</span>`),
	})

	assert.Contains(t, html, `href="/"`)
	assert.NotContains(t, html, `x-bind:aria-label="mobileMenuIsOpen`)
	assert.NotContains(t, html, `aria-label="user menu"`)
}
