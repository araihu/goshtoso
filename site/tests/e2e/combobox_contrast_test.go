//go:build e2e && (full || combobox)

package e2e

import (
	"fmt"
	"os"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

// Supply a local, pinned axe-core script to keep accessibility checks offline.
func TestCombobox_SelectedTriggerAxe(t *testing.T) {
	script := os.Getenv("GOSHTOSO_AXE_SCRIPT")
	if script == "" {
		t.Skip("set GOSHTOSO_AXE_SCRIPT to a local axe-core/axe.min.js")
	}
	page := newPage(t, sharedBrowser)
	_, err := page.Goto(baseURL + "/components/combobox")
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))
	// Region is selected on the server, covering initial and OOB label classes.
	require.NoError(t, page.Locator("#region-trigger-label").WaitFor())
	require.NoError(t, page.Locator("#industry-trigger").Click())
	require.NoError(t, page.Locator("#industry [data-combobox-search]").Fill("no-such-industry"))
	require.NoError(t, page.Locator("#industry [data-combobox-empty]").WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}))
	_, err = page.AddScriptTag(playwright.PageAddScriptTagOptions{Path: playwright.String(script)})
	require.NoError(t, err)
	for _, theme := range []string{"goshtoso", "minimal"} {
		for _, dark := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/dark=%t", theme, dark), func(t *testing.T) {
				_, err := page.Evaluate(`async config=>{
      document.documentElement.dataset.theme=config.theme;
      document.documentElement.classList.toggle('dark',config.dark);
      getComputedStyle(document.querySelector('#region-trigger-label')).color;
      await new Promise(resolve=>requestAnimationFrame(resolve));
      await Promise.all(document.getAnimations().filter(a=>a.effect.getComputedTiming().iterations!==Infinity).map(a=>a.finished.catch(()=>{})));
    }`, map[string]any{"theme": theme, "dark": dark})
				require.NoError(t, err)
				violations, err := page.Evaluate(`async()=>{const result=await window.axe.run(['#region','#industry'],{runOnly:{type:'tag',values:['wcag2a','wcag2aa','wcag21aa']}});return result.violations.map(v=>({id:v.id,nodes:v.nodes.map(n=>({target:n.target,summary:n.failureSummary,html:n.html})),label:document.querySelector('#industry-trigger-label').className,color:getComputedStyle(document.querySelector('#industry-trigger-label')).color}))}`)
				require.NoError(t, err)
				require.Empty(t, violations, "selected text and open empty feedback must remain accessible")
			})
		}
	}
}
