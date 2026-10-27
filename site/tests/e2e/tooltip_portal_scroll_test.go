//go:build e2e && (full || tooltip)

package e2e

import (
	"bytes"
	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/components/button"
	"github.com/araihu/goshtoso/components/tooltip"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTooltip_PortalTracksNestedScroll(t *testing.T) {
	page := newIsolatedPage(t)
	_, err := page.Goto(baseURL + "/components/tooltip")
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))
	var markup bytes.Buffer
	trigger := button.Button(button.WithType("button"), button.WithAttrs(templ.Attributes{"id": "portal-scroll-trigger", "aria-label": "Reset preview", "style": "width:40px;height:40px"}))
	require.NoError(t, tooltip.Tooltip("portal-scroll-tip", "Reset preview", tooltip.WithPosition(tooltip.PositionTop), tooltip.WithPortal(true), tooltip.WithTrigger(trigger)).Render(t.Context(), &markup))
	_, err = page.Evaluate(`markup=>{
  const scroll=document.createElement('div'),content=document.createElement('div');
  scroll.id='portal-scroll-fixture';scroll.style.cssText='position:fixed;top:200px;left:300px;width:300px;height:200px;overflow:auto;';
  content.style.cssText='padding-top:250px;height:700px;';content.innerHTML=markup;
  scroll.append(content);document.body.append(scroll);
 }`, markup.String())
	require.NoError(t, err)
	_, err = page.WaitForFunction(`()=>document.querySelector('[data-tooltip-portal-ready] #portal-scroll-tip')`, nil)
	require.NoError(t, err)
	require.NoError(t, page.Locator("#portal-scroll-trigger").Focus())
	_, err = page.Locator("#portal-scroll-fixture").Evaluate(`el=>el.scrollTop=200`, nil)
	require.NoError(t, err)
	_, err = page.WaitForFunction(`()=>{
  const b=document.querySelector('#portal-scroll-trigger').getBoundingClientRect(),tip=document.querySelector('#portal-scroll-tip');
  return tip.matches(':popover-open')&&Math.abs(b.top-tip.getBoundingClientRect().bottom-8)<2;
 }`, nil)
	require.NoError(t, err)
	attached, err := page.Evaluate(`async()=>{
  document.querySelector('#portal-scroll-fixture').scrollTop+=48;
  await new Promise(r=>requestAnimationFrame(()=>requestAnimationFrame(r)));
  const b=document.querySelector('#portal-scroll-trigger').getBoundingClientRect(),t=document.querySelector('#portal-scroll-tip').getBoundingClientRect();
  return Math.abs(b.top-t.bottom-8)<2;
 }`, nil)
	require.NoError(t, err)
	require.Equal(t, true, attached)
}
