//go:build e2e && (full || schematree)

package e2e

import (
	"fmt"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestSchemaTreeThemesResponsiveAndNativeDisclosure(t *testing.T) {
	_, browser, _ := setupPlaywright(t)
	for _, width := range []int{390, 1440} {
		for _, theme := range []string{"goshtoso", "minimal"} {
			for _, dark := range []bool{false, true} {
				t.Run(fmt.Sprintf("%d/%s/dark=%t", width, theme, dark), func(t *testing.T) {
					page := newPage(t, browser, playwright.BrowserNewPageOptions{Viewport: &playwright.Size{Width: width, Height: 1000}, JavaScriptEnabled: playwright.Bool(false)})
					_, err := page.Goto(baseURL + "/components/schema-tree")
					require.NoError(t, err)
					_, err = page.Evaluate(`([theme, dark]) => { document.documentElement.dataset.theme = theme; document.documentElement.classList.toggle('dark', dark); }`, []any{theme, dark})
					require.NoError(t, err)
					branch := page.Locator("#schema-tree-response details").First()
					optionalCount, err := page.Locator(`#schema-tree-response [data-required="false"]`).Count()
					require.NoError(t, err)
					require.Zero(t, optionalCount, "default tree should only label required properties")
					dangerColor, err := branch.Evaluate(`el => {
						const required = el.querySelector('[data-required="true"]');
						const probe = document.createElement('span');
						probe.style.color = document.documentElement.classList.contains('dark')
							? 'var(--color-danger-action-dark)' : 'var(--color-danger-action)';
						el.append(probe);
						const matches = getComputedStyle(required).color === getComputedStyle(probe).color;
						probe.remove();
						return matches;
					}`, nil)
					require.NoError(t, err)
					require.Equal(t, true, dangerColor, "required labels should use the theme's readable danger tone")
					typography, err := branch.Evaluate(`el => {
						const type = getComputedStyle(el.querySelector('.gs-schema-tree-type'));
						const name = getComputedStyle(el.querySelector('.gs-schema-tree-name'));
						const state = getComputedStyle(el.querySelector('.gs-schema-tree-state'));
						return type.fontStyle === 'italic' && type.fontWeight === '400' &&
							type.fontFamily === name.fontFamily && name.fontStyle === 'normal' &&
							Number(name.fontWeight) >= 600 && state.fontStyle === 'normal';
					}`, nil)
					require.NoError(t, err)
					require.Equal(t, true, typography, "types should be italic monospace without changing field names or status labels")
					require.NoError(t, branch.Locator("summary").First().Focus())
					require.NoError(t, page.Keyboard().Press("Enter"))
					open, err := branch.Evaluate("el => el.open", nil)
					require.NoError(t, err)
					require.Equal(t, false, open)
					require.NoError(t, page.Keyboard().Press("Enter"))
					require.NoError(t, page.Locator("#schema-tree-response .gs-schema-tree-description a").Click())
					open, err = branch.Evaluate("el => el.open", nil)
					require.NoError(t, err)
					require.Equal(t, true, open, "description link toggled branch")
					layout, err := page.Evaluate(`() => [...document.querySelectorAll('#schema-tree-fragment .gs-schema-tree')].every(tree => tree.scrollWidth <= tree.clientWidth + 1 && [...tree.querySelectorAll('.gs-schema-tree-state')].every(state => { const s = state.getBoundingClientRect(), t = tree.getBoundingClientRect(); return !s.width || s.right <= t.right + 1; }))`)
					require.NoError(t, err)
					require.Equal(t, true, layout, "tree or field status overflow")
					rails, err := page.Evaluate(`() => {
						const tree = document.querySelector('#schema-tree-response .gs-schema-tree');
						return [...tree.querySelectorAll('.gs-schema-tree-item')].every(item => {
							const style = getComputedStyle(item);
							return style.borderBlockStartWidth === '0px' && style.borderBlockEndWidth === '0px';
						}) && getComputedStyle(tree.querySelector('.gs-schema-tree-children')).borderInlineStartWidth === '1px';
					}`)
					require.NoError(t, err)
					require.Equal(t, true, rails, "only vertical nesting guides should remain")
					if width == 1440 && theme == "goshtoso" {
						_, err = page.Locator("#schema-tree-response").Screenshot(playwright.LocatorScreenshotOptions{Path: playwright.String(fmt.Sprintf("/tmp/schema-tree-dark-%t.png", dark))})
						require.NoError(t, err)
					}
				})
			}
		}
	}
}
