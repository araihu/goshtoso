//go:build e2e && (full || popover)

package e2e

import (
	"fmt"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

// Record visible geometry before opening so the first painted transition frame
// must satisfy the same bounds as the settled panel.
func observePopoverPaints(t *testing.T, page playwright.Page, id string) {
	t.Helper()
	_, err := page.Evaluate(`id => {
    window.popoverPaints = {active:true, frames:[]};
    const sample = () => {
      if (!popoverPaints.active) return;
      const panel=document.querySelector('#'+id+' [data-popover-panel]');
      const style=getComputedStyle(panel);
      if (style.display!=='none' && style.visibility!=='hidden' && Number(style.opacity)>0) {
        const r=panel.getBoundingClientRect();
        popoverPaints.frames.push({left:r.left,top:r.top,right:r.right,bottom:r.bottom});
      }
      requestAnimationFrame(sample);
    };
    requestAnimationFrame(sample);
  }`, id)
	require.NoError(t, err)
}

func assertPopoverFits(t *testing.T, page playwright.Page, id string) {
	t.Helper()
	_, err := page.WaitForFunction(`id => {
   const panel=document.querySelector('#'+id+' [data-popover-panel]');
   if (!panel || getComputedStyle(panel).display==='none') return false;
   const r=panel.getBoundingClientRect();
   return r.width>0 && r.height>0 && r.left>=7 && r.top>=7 && r.right<=innerWidth-7 && r.bottom<=innerHeight-7;
 }`, id)
	if err != nil {
		diagnostic, _ := page.Evaluate(`id=>{const p=document.querySelector('#'+id+' [data-popover-panel]'); const r=document.querySelector('#'+id); return {html:p.outerHTML,rect:p.getBoundingClientRect().toJSON(),trigger:r.querySelector('button').getBoundingClientRect().toJSON(),root:r.outerHTML.slice(0,400),active:document.activeElement.outerHTML.slice(0,250)};}`, id)
		t.Logf("position diagnostic %#v", diagnostic)
	}
	require.NoError(t, err)
	_, err = page.Evaluate(`async id => {
   const panel=document.querySelector('#'+id+' [data-popover-panel]');
   await Promise.all(panel.getAnimations().map(animation=>animation.finished.catch(()=>{})));
   await new Promise(resolve=>requestAnimationFrame(()=>requestAnimationFrame(resolve)));
 }`, id)
	require.NoError(t, err)
	fits, err := page.Evaluate(`id => {
   const r=document.querySelector('#'+id+' [data-popover-panel]').getBoundingClientRect();
   return r.left>=7 && r.top>=7 && r.right<=innerWidth-7 && r.bottom<=innerHeight-7;
 }`, id)
	require.NoError(t, err)
	require.Equal(t, true, fits, "settled panel must remain in the viewport")
	painted, err := page.Evaluate(`()=>{
    if (!window.popoverPaints?.active) return true;
    popoverPaints.active=false;
    return popoverPaints.frames.length>0 && popoverPaints.frames.every(r=>
      r.left>=7 && r.top>=7 && r.right<=innerWidth-7 && r.bottom<=innerHeight-7);
  }`)
	require.NoError(t, err)
	require.Equal(t, true, painted, "every visible opening frame must remain in the viewport")
}

func TestPopover_ViewportCollisions(t *testing.T) {
	for _, theme := range []string{"goshtoso", "minimal"} {
		for _, dark := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/dark=%t", theme, dark), func(t *testing.T) {
				page := newPage(t, sharedBrowser)
				require.NoError(t, page.SetViewportSize(390, 500))
				_, err := page.Goto(baseURL + "/components/popover")
				require.NoError(t, err)
				require.NoError(t, waitForAlpine(page))
				_, err = page.Evaluate(`config=>{
      document.documentElement.dataset.theme=config.theme;
      document.documentElement.classList.toggle('dark',config.dark);
      const box=document.querySelector('#popover-collision-scroll');
      box.style.cssText='position:fixed;left:8px;top:calc(100vh - 100px);width:calc(100vw - 16px);height:80px;overflow:auto;transform:translateZ(0);filter:drop-shadow(0 1px 1px #0002);z-index:100;padding:4px;';
      box.scrollTop=box.scrollHeight;
      document.querySelector('#popover-collision-start-panel').style.minHeight='350px';
    }`, map[string]any{"theme": theme, "dark": dark})
				require.NoError(t, err)
				trigger := page.Locator("#popover-collision-start [data-popover-trigger] button")
				observePopoverPaints(t, page, "popover-collision-start")
				require.NoError(t, trigger.Focus())
				require.NoError(t, page.Keyboard().Press("ArrowDown"))
				assertPopoverFits(t, page, "popover-collision-start")
				portaled, err := page.Locator("#popover-collision-start-panel").Evaluate(`p=>p.matches(':popover-open')`, nil)
				require.NoError(t, err)
				require.Equal(t, true, portaled, "transformed scroll container must not clip the panel")
				_, err = page.WaitForFunction(`()=>document.activeElement.textContent.trim()==='Action 1'`, nil)
				require.NoError(t, err)
				require.NoError(t, page.SetViewportSize(280, 240))
				assertPopoverFits(t, page, "popover-collision-start")
				scrollable, err := page.Locator("#popover-collision-start-panel").Evaluate(`p=>p.scrollHeight>p.clientHeight && getComputedStyle(p).overflowY==='auto'`, nil)
				require.NoError(t, err)
				require.Equal(t, true, scrollable, "oversized panel must scroll")
				retained, err := page.Locator("#popover-collision-start-panel").Evaluate(`async p=>{
                    p.scrollTop=100;
                    const expected=p.scrollTop;
                    p.dispatchEvent(new Event('scroll',{bubbles:true}));
                    await new Promise(resolve=>requestAnimationFrame(()=>requestAnimationFrame(resolve)));
                    return expected>0 && p.scrollTop===expected;
                }`, nil)
				require.NoError(t, err)
				require.Equal(t, true, retained, "repositioning must preserve scrolling through oversized content")
				_, err = page.Evaluate(`()=>{document.querySelector('#popover-collision-scroll').scrollTop=0;}`)
				require.NoError(t, err)
				assertPopoverFits(t, page, "popover-collision-start")
				_, err = page.Evaluate(`()=>{const box=document.querySelector('#popover-collision-scroll');box.scrollTop=box.scrollHeight;}`)
				require.NoError(t, err)
				require.NoError(t, page.Keyboard().Press("Escape"))
				require.NoError(t, page.Locator("#popover-collision-start-panel").WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateHidden}))
				_, err = page.WaitForFunction(`()=>document.activeElement===document.querySelector('#popover-collision-start [data-popover-trigger] button')`, nil)
				require.NoError(t, err)
				// Exercise the shared Dropdown runtime near the opposite horizontal edge.
				observePopoverPaints(t, page, "dropdown-collision-end")
				require.NoError(t, page.Locator("#dropdown-collision-end [data-popover-trigger] button").Click())
				assertPopoverFits(t, page, "dropdown-collision-end")
				require.NoError(t, page.Keyboard().Press("Escape"))
				// Ordinary panels retain absolute positioning and configured alignment.
				_, err = page.Evaluate(`()=>{
      const root=document.querySelector('#filters-popover');
      document.body.append(root);
      root.style.cssText='position:fixed;right:8px;bottom:8px;z-index:101';
    }`)
				require.NoError(t, err)
				observePopoverPaints(t, page, "filters-popover")
				require.NoError(t, page.Locator("#filters-popover [data-popover-trigger] button").Click())
				assertPopoverFits(t, page, "filters-popover")
				position, err := page.Locator("#filters-popover-panel").Evaluate(`p=>getComputedStyle(p).position`, nil)
				require.NoError(t, err)
				require.Equal(t, "absolute", position)
			})
		}
	}
}
