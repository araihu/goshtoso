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
						}) && [...tree.querySelectorAll('details[open]')].every(branch => {
							const guide = getComputedStyle(branch, '::before');
							const marker = getComputedStyle(branch.querySelector(':scope > summary'), '::before');
							const gap = parseFloat(guide.top) - parseFloat(getComputedStyle(branch).paddingTop) - parseFloat(marker.top) - parseFloat(marker.height);
							return guide.borderInlineStartWidth === '1px' && Math.abs(gap - 22) <= 1 &&
								getComputedStyle(branch.querySelector(':scope > .gs-schema-tree-children')).borderInlineStartWidth === '0px';
						});
					}`)
					require.NoError(t, err)
					require.Equal(t, true, rails, "only vertical nesting guides should remain")
					codeOwnership, err := page.Evaluate(`() => {
						const host = document.querySelector('#schema-tree-response');
						host.classList.add('schema-prose-host');
						const style = document.createElement('style');
						style.textContent = '.schema-prose-host code { border: 2px solid red; border-radius: 8px; }';
						document.head.append(style);
						const owned = [...host.querySelectorAll('.gs-schema-tree-name, .gs-schema-tree-constraints code')];
						const description = host.querySelector('.gs-schema-tree-description code');
						const valid = owned.length > 0 && !!description && owned.every(node => {
							const computed = getComputedStyle(node);
							return computed.borderTopWidth === '0px' && computed.borderTopLeftRadius === '0px';
						}) && getComputedStyle(description).borderTopWidth === '2px';
						style.remove();
						host.classList.remove('schema-prose-host');
						return valid;
					}`)
					require.NoError(t, err)
					require.Equal(t, true, codeOwnership, "tree-owned code stays borderless while descriptions retain host prose styling")
					if width == 1440 && theme == "goshtoso" {
						_, err = page.Locator("#schema-tree-response").Screenshot(playwright.LocatorScreenshotOptions{Path: playwright.String(fmt.Sprintf("/tmp/schema-tree-dark-%t.png", dark))})
						require.NoError(t, err)
					}
				})
			}
		}
	}
}
