//go:build e2e && full

package e2e

import (
	"fmt"
	"testing"

	"github.com/araihu/goshtoso/themes"
	"github.com/stretchr/testify/require"
)

// The browser must resolve OKLCH values into distinct colors and useful ratios.
func TestThemeContrastResolvesColors(t *testing.T) {
	page := newIsolatedPage(t)
	gotoThemePage(t, page)
	definition := editorDefinition(t, page)
	colors := map[string]bool{}
	for _, key := range []string{"surface", "primary", "secondary"} {
		text, err := page.Locator("[data-theme-contrast-pair=" + key + "] [data-theme-contrast-ratio]").TextContent()
		require.NoError(t, err)
		var ratio float64
		_, err = fmt.Sscanf(text, "%f:1", &ratio)
		require.NoError(t, err)
		require.InDelta(t, themes.Contrast(definition.Tokens["on-"+key], definition.Tokens[key]), ratio, .01)
		color, err := page.Locator("[data-theme-contrast-pair="+key+"] [data-theme-contrast-sample]").Evaluate(`el=>getComputedStyle(el).backgroundColor`, nil)
		require.NoError(t, err)
		colors[color.(string)] = true
	}
	require.Len(t, colors, 3)
}

func TestThemeContrastMinimalDarkSmallCopyMeetsAA(t *testing.T) {
	page := newIsolatedPage(t)
	gotoThemePage(t, page)
	choosePreviewPreset(t, page, "minimal")
	_, err := page.Evaluate(`()=>{const root=document.querySelector('[data-theme-editor]');root.dispatchEvent(new CustomEvent('themeeditor:load',{detail:{id:'site-theme-editor',previewMode:'dark'}}));}`, nil)
	require.NoError(t, err)
	_, err = page.WaitForFunction(`()=>document.querySelector('#theme-playground-frame').contentDocument?.documentElement.classList.contains('dark')`, nil)
	require.NoError(t, err)
	resolved, err := page.Evaluate(`()=>{
  const doc=document.querySelector('#theme-playground-frame').contentDocument,style=getComputedStyle(doc.documentElement),canvas=doc.createElement('canvas'),ctx=canvas.getContext('2d');
  return ['on-surface-dark','surface-dark','surface-dark-alt'].map(key=>{ctx.clearRect(0,0,1,1);ctx.fillStyle=style.getPropertyValue('--color-'+key);ctx.fillRect(0,0,1,1);return '#'+[...ctx.getImageData(0,0,1,1).data].slice(0,3).map(v=>v.toString(16).padStart(2,'0')).join('');});
 }`, nil)
	require.NoError(t, err)
	values := resolved.([]any)
	foreground, err := themes.ParseColor(values[0].(string))
	require.NoError(t, err)
	for _, value := range values[1:] {
		background, err := themes.ParseColor(value.(string))
		require.NoError(t, err)
		require.GreaterOrEqual(t, themes.Contrast(foreground, background), 4.5, "Minimal dark small copy must meet AA on both surfaces")
	}
}
