//go:build e2e && (full || modal)

package e2e

import (
	"github.com/stretchr/testify/require"
	"testing"
)

// A native dialog can be retired by an HTMX swap after its open event queues
// Alpine's effect, but before that effect runs. Retiring it must not reopen a
// disconnected node or leave an asynchronous browser error behind.
func TestContentDialogQueuedOpenSurvivesRemoval(t *testing.T) {
	for _, theme := range []string{"goshtoso", "minimal"} {
		for _, dark := range []bool{false, true} {
			t.Run(theme+map[bool]string{false: "/light", true: "/dark"}[dark], func(t *testing.T) {
				page := newPage(t, sharedBrowser)
				failures := watchPageFailures(page)
				_, err := page.Goto(baseURL + "/components/modal")
				require.NoError(t, err)
				_, err = page.WaitForFunction("()=>window.Alpine && document.getElementById('compactNote')", nil)
				require.NoError(t, err)
				_, err = page.Evaluate(`async ({theme,dark})=>{
   document.documentElement.dataset.theme=theme;document.documentElement.classList.toggle('dark',dark);
   const dialog=document.getElementById('compactNote');
   window.dispatchEvent(new CustomEvent('modal:open',{detail:{id:dialog.id}}));
   Alpine.mutateDom(()=>dialog.remove());
   await new Promise(resolve=>requestAnimationFrame(()=>requestAnimationFrame(()=>setTimeout(resolve,75))));
   return dialog.isConnected;
  }`, map[string]any{"theme": theme, "dark": dark})
				require.NoError(t, err)
				failures.RequireEmpty(t)
			})
		}
	}
}
