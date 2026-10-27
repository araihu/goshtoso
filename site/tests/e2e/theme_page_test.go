//go:build e2e && full

package e2e

import (
	"fmt"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestThemePage_ControlledPreviewIsIsolated(t *testing.T) {
	page := newIsolatedPage(t)
	gotoThemePage(t, page)
	failures := watchPageFailures(page)
	before, err := page.Locator("html").GetAttribute("data-theme")
	require.NoError(t, err)
	choosePreviewPreset(t, page, "dracula")
	_, err = page.WaitForFunction(`() => document.querySelector('#theme-playground-frame')?.contentDocument?.documentElement.dataset.theme==='custom'`, nil)
	require.NoError(t, err)
	after, err := page.Locator("html").GetAttribute("data-theme")
	require.NoError(t, err)
	require.Equal(t, before, after)
	stored, err := page.Evaluate(`() => ['themeOverrides','themeTitleFont','themeBodyFont','themeRadius'].map(k=>localStorage.getItem(k))`, nil)
	require.NoError(t, err)
	require.Equal(t, []any{nil, nil, nil, nil}, stored)
	failures.RequireEmpty(t)
}

func TestThemePage_FontComboboxSearchAndTypeface(t *testing.T) {
	page := newIsolatedPage(t)
	gotoThemePage(t, page)
	require.NoError(t, page.Locator("#site-theme-editor-title-trigger").Click())
	picker := page.Locator("#site-theme-editor-title")
	require.NoError(t, picker.Locator("[data-combobox-search]").Fill("Go Regular"))
	option := picker.GetByRole("option", playwright.LocatorGetByRoleOptions{Name: "Go Regular", Exact: new(false)})
	require.NoError(t, option.Click())
	_, err := page.WaitForFunction(`() => JSON.parse(document.querySelector('[data-theme-definition]').value).typography.title.id==='go-regular'`, nil)
	require.NoError(t, err)
	style, err := page.Locator("#site-theme-editor-title-trigger").GetAttribute("style")
	require.NoError(t, err)
	require.Contains(t, style, "goshtoso-font-go-regular")
}

func TestThemePage_LongSelectedFontFitsMobile(t *testing.T) {
	page := newIsolatedPage(t)
	require.NoError(t, page.SetViewportSize(390, 900))
	gotoThemePage(t, page)
	_, err := page.Evaluate(`()=>{
 const root=document.querySelector('[data-theme-editor]'),config=JSON.parse(root.dataset.themeEditor),definition=JSON.parse(root.querySelector('[data-theme-definition]').value),font=config.fonts.find(f=>f.source==='bundled'&&f.id==='go-regular');
 font.label='Uploaded title font with a long descriptive name and immutable reference '.repeat(3);
 definition.typography.title={source:font.source,id:font.id};
 root.dispatchEvent(new CustomEvent('themeeditor:load',{detail:{id:config.id,definition,fonts:config.fonts}}));
}`, nil)
	require.NoError(t, err)
	_, err = page.WaitForFunction(`()=>document.querySelector('#site-theme-editor-title-trigger').textContent.includes('Uploaded title font')`, nil)
	require.NoError(t, err)
	for _, open := range []bool{false, true} {
		if open {
			require.NoError(t, page.Locator("#site-theme-editor-title-trigger").Click())
		}
		fits, err := page.Evaluate(`()=>{const r=document.querySelector('#site-theme-editor-title-trigger').getBoundingClientRect();return document.documentElement.scrollWidth<=innerWidth+1&&r.right<=innerWidth+1}`, nil)
		require.NoError(t, err)
		require.Equal(t, true, fits)
	}
}

func TestThemePage_ValidationAndResetModal(t *testing.T) {
	page := newIsolatedPage(t)
	gotoThemePage(t, page)
	initial, err := page.Locator("[data-theme-definition]").InputValue()
	require.NoError(t, err)
	require.NoError(t, page.Locator("#site-theme-editor-primary-picker-trigger").Click())
	color := page.Locator("[data-theme-color=primary]")
	require.NoError(t, color.Fill("#165d9b"))
	require.NoError(t, color.Press("Tab"))
	_, err = page.WaitForFunction(`initial => document.querySelector('[data-theme-definition]').value!==initial`, initial)
	require.NoError(t, err)
	changed, err := page.Locator("[data-theme-definition]").InputValue()
	require.NoError(t, err)
	require.NoError(t, color.Fill("invalid"))
	require.NoError(t, color.Press("Tab"))
	invalid, err := color.GetAttribute("aria-invalid")
	require.NoError(t, err)
	require.Equal(t, "true", invalid)
	require.NoError(t, color.Press("Escape"))
	require.NoError(t, page.Locator("[data-theme-reset]").Click())
	require.NoError(t, page.Locator("#site-theme-editor-confirm [data-theme-confirm-cancel]").Click())
	raw, err := color.InputValue()
	require.NoError(t, err)
	require.Equal(t, "invalid", raw)
	value, err := page.Locator("[data-theme-definition]").InputValue()
	require.NoError(t, err)
	require.Equal(t, changed, value)
	require.NoError(t, page.Locator("[data-theme-reset]").Click())
	require.NoError(t, page.Locator("#site-theme-editor-confirm [data-theme-confirm-accept]").Click())
	_, err = page.WaitForFunction(`initial => document.querySelector('[data-theme-definition]').value===initial`, initial)
	require.NoError(t, err)
}

func TestThemePage_DraftDoesNotPersistOnReload(t *testing.T) {
	page := newIsolatedPage(t)
	gotoThemePage(t, page)
	choosePreviewPreset(t, page, "minimal")
	_, err := page.Reload()
	require.NoError(t, err)
	waitForThemeEditor(t, page)
	value, err := page.Locator("[data-theme-definition]").InputValue()
	require.NoError(t, err)
	require.Contains(t, value, `"presetKey":"araihu"`)
}

func TestThemePage_FragmentNavBootstrapsData(t *testing.T) {
	page := newIsolatedPage(t)
	dismissCookieBanner(t, page)
	_, err := page.Goto(baseURL + "/getting-started")
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))
	failures := watchPageFailures(page)
	require.NoError(t, page.Locator("a[href='/docs/theme']").First().Click())
	waitForThemeEditor(t, page)
	choosePreviewPreset(t, page, "arctic")
	failures.RequireEmpty(t)
}

func TestThemePage_ExportCurrentDraft(t *testing.T) {
	page := newIsolatedPage(t)
	gotoThemePage(t, page)
	download, err := page.ExpectDownload(func() error { return page.Locator("[data-theme-export]").Click() })
	require.NoError(t, err)
	require.Equal(t, "theme.css", download.SuggestedFilename())
}

func TestThemePage_ResponsiveThemeMatrix(t *testing.T) {
	page := newIsolatedPage(t)
	gotoThemePage(t, page)
	for _, width := range []int{390, 1440} {
		for _, theme := range []string{"goshtoso", "minimal"} {
			for _, dark := range []bool{false, true} {
				t.Run(fmt.Sprintf("%d/%s/dark=%t", width, theme, dark), func(t *testing.T) {
					require.NoError(t, page.SetViewportSize(width, 900))
					_, err := page.Evaluate(`([theme,dark])=>{document.documentElement.dataset.theme=theme;document.documentElement.classList.toggle('dark',dark);}`, []any{theme, dark})
					require.NoError(t, err)
					fits, err := page.Evaluate(`()=>document.documentElement.scrollWidth<=innerWidth+1 && ['#site-theme-editor','#theme-playground-frame','#site-theme-editor-title-trigger'].every(selector=>{const r=document.querySelector(selector).getBoundingClientRect();return r.width>0 && r.left>=0 && r.right<=innerWidth+1;})`, nil)
					require.NoError(t, err)
					require.Equal(t, true, fits)
					require.NoError(t, page.Locator("#site-theme-editor-title-trigger").Click())
					popupFits, err := page.Locator("#site-theme-editor-title-options").Evaluate(`el=>{const r=el.getBoundingClientRect();return r.width>0&&r.left>=0&&r.right<=innerWidth+1}`, nil)
					require.NoError(t, err)
					require.Equal(t, true, popupFits)
					require.NoError(t, page.Locator("#site-theme-editor-title-trigger").Press("Escape"))
					picker := page.Locator("#site-theme-editor-primary-picker-trigger")
					require.NoError(t, picker.Click())
					require.NoError(t, page.Locator("#site-theme-editor-primary-palette [data-cls=white]").WaitFor())
					paletteFits, err := page.Locator("#site-theme-editor-primary-palette").Evaluate(`el=>{const r=el.getBoundingClientRect();return document.documentElement.scrollWidth<=innerWidth+1&&r.width>0&&r.left>=0&&r.right<=innerWidth+1}`, nil)
					require.NoError(t, err)
					require.Equal(t, true, paletteFits)
					require.NoError(t, picker.Press("Escape"))
				})
			}
		}
	}
}

func TestThemePage_PreviewModeAndRichPresetOptions(t *testing.T) {
	page := newIsolatedPage(t)
	gotoThemePage(t, page)
	initial, err := page.Locator("[data-theme-definition]").InputValue()
	require.NoError(t, err)
	parent, err := page.Locator("html").GetAttribute("class")
	require.NoError(t, err)
	dark := page.Locator("[data-theme-mode][value=dark]")
	require.NoError(t, dark.Check())
	_, err = page.WaitForFunction(`()=>document.querySelector('#theme-playground-frame').contentDocument.documentElement.classList.contains('dark')`, nil)
	require.NoError(t, err)
	require.Equal(t, 1, mustCount(t, page.Locator("[data-theme-mode]:checked")))
	require.NoError(t, dark.Press("ArrowLeft"))
	_, err = page.WaitForFunction(`()=>!document.querySelector('#theme-playground-frame').contentDocument.documentElement.classList.contains('dark')`, nil)
	require.NoError(t, err)
	current, err := page.Locator("[data-theme-definition]").InputValue()
	require.NoError(t, err)
	require.Equal(t, initial, current)
	after, err := page.Locator("html").GetAttribute("class")
	require.NoError(t, err)
	require.Equal(t, parent, after)
	require.NoError(t, page.Locator("#site-theme-editor-preset-trigger").Click())
	list := page.Locator("#site-theme-editor-preset-listbox")
	require.GreaterOrEqual(t, mustCount(t, list.Locator("[data-theme-swatches]")), 16)
	option := list.GetByRole("option", playwright.LocatorGetByRoleOptions{Name: "Arctic", Exact: new(true)})
	require.NoError(t, option.WaitFor())
	require.Equal(t, 3, mustCount(t, option.Locator("[data-theme-swatch]")))
	font, err := list.GetByRole("option", playwright.LocatorGetByRoleOptions{Name: "Arctic", Exact: new(true)}).Locator("[data-theme-font-option]").GetAttribute("style")
	require.NoError(t, err)
	require.Contains(t, font, "font-family:")
}

func TestThemePage_PaletteShellMaterializesColorsAndResets(t *testing.T) {
	page := newIsolatedPage(t)
	gotoThemePage(t, page)
	baseline, err := page.Evaluate(`()=>JSON.stringify(JSON.parse(document.querySelector('[data-theme-definition]').value).tokens.primary)`, nil)
	require.NoError(t, err)
	trigger := page.Locator("#site-theme-editor-primary-picker-trigger")
	palette := page.Locator("#site-theme-editor-primary-palette")
	require.NoError(t, trigger.Click())
	require.NoError(t, palette.Locator(`[data-cls="blue-700"]`).Click())
	_, err = page.WaitForFunction(`()=>{const c=JSON.parse(document.querySelector('[data-theme-definition]').value).tokens.primary;return c.space==='oklch'&&Math.abs(c.channels[0]-.488)<.00001}`, nil)
	require.NoError(t, err)
	_, err = page.WaitForFunction(`()=>document.activeElement.id==='site-theme-editor-primary-picker-trigger'`, nil)
	require.NoError(t, err)
	_, err = page.WaitForFunction(`()=>{const d=document.querySelector('#theme-playground-frame').contentDocument;return getComputedStyle(d.documentElement).getPropertyValue('--color-primary').trim()==='oklch(0.488 0.243 264.376 / 1)'}`, nil)
	require.NoError(t, err)
	require.NoError(t, trigger.Press("ArrowDown"))
	_, err = page.WaitForFunction(`()=>document.querySelector('#site-theme-editor-primary-palette').contains(document.activeElement)`, nil)
	require.NoError(t, err)
	require.NoError(t, palette.Locator(`[data-cls="white"]`).Press("Enter"))
	_, err = page.WaitForFunction(`()=>{const c=JSON.parse(document.querySelector('[data-theme-definition]').value).tokens.primary;return c.space==='srgb'&&c.channels.every(n=>n===1)}`, nil)
	require.NoError(t, err)
	require.NoError(t, trigger.Click())
	require.NoError(t, palette.GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Reset", Exact: new(true)}).Click())
	_, err = page.WaitForFunction(`before=>JSON.stringify(JSON.parse(document.querySelector('[data-theme-definition]').value).tokens.primary)===before`, baseline)
	require.NoError(t, err)
	require.NoError(t, trigger.Click())
	require.NoError(t, page.Locator("[data-theme-color=primary]").Press("Escape"))

	_, err = page.WaitForFunction(`()=>document.activeElement.id==='site-theme-editor-primary-picker-trigger'`, nil)

	require.NoError(t, err)
	value, err := page.Locator("[data-theme-definition]").InputValue()
	require.NoError(t, err)
	require.NotContains(t, value, "var(--")
}

func TestThemePage_ColorRolesMatchPreviewMode(t *testing.T) {
	page := newIsolatedPage(t)
	gotoThemePage(t, page)
	for _, dark := range []bool{false, true} {
		if dark {
			require.NoError(t, page.Locator("[data-theme-mode][value=dark]").Check())
		}
		for _, key := range []string{"surface-alt", "on-surface-strong", "outline-strong"} {
			visible, err := page.Locator("[data-theme-color-row=" + key + "]").IsVisible()
			require.NoError(t, err)
			require.Equal(t, !dark, visible, "role %s, dark=%t", key, dark)
		}
		key := "surface"
		color, expected := "white", "color(srgb 1 1 1)"
		if dark {
			key = "surface-dark"
			color, expected = "black", "color(srgb 0 0 0)"
		}
		require.NoError(t, page.Locator("#site-theme-editor-"+key+"-picker-trigger").Click())
		require.NoError(t, page.Locator("#site-theme-editor-"+key+"-palette [data-cls="+color+"]").Click())
		_, err := page.WaitForFunction(`([dark,expected])=>{const d=document.querySelector('#theme-playground-frame').contentDocument;return d.documentElement.classList.contains('dark')===dark&&getComputedStyle(d.body).backgroundColor===expected}`, []any{dark, expected})
		require.NoError(t, err)
	}
}

func TestThemePage_VisualContrastAdvisory(t *testing.T) {
	page := newIsolatedPage(t)
	gotoThemePage(t, page)
	for _, width := range []int{390, 1440} {
		require.NoError(t, page.SetViewportSize(width, 900))
		for _, mode := range []string{"light", "dark"} {
			for _, sample := range []struct {
				grade       string
				gray, alpha float64
			}{{"aaa", 0, 1}, {"aa", .4, 1}, {"aa-large", .5, 1}, {"fail", 1, 1}, {"unrated", 0, .5}} {
				t.Run(fmt.Sprintf("%d/%s/%s", width, mode, sample.grade), func(t *testing.T) {
					_, err := page.Evaluate(`([mode,gray,alpha])=>{
      const root=document.querySelector('[data-theme-editor]'),definition=JSON.parse(root.querySelector('[data-theme-definition]').value),suffix=mode==='dark'?'-dark':'';
      definition.tokens['primary'+suffix]={space:'srgb',channels:[gray,gray,gray],alpha};
      definition.tokens['on-primary'+suffix]={space:'srgb',channels:[1,1,1],alpha:1};
      root.dispatchEvent(new CustomEvent('themeeditor:load',{detail:{id:'site-theme-editor',definition,previewMode:mode}}));
     }`, []any{mode, sample.gray, sample.alpha})
					require.NoError(t, err)
					_, err = page.WaitForFunction(`([grade,mode])=>document.querySelector('[data-theme-contrast-pair=primary]').dataset.themeContrastRating===grade&&document.querySelector('[data-theme-mode]:checked').value===mode`, []any{sample.grade, mode})
					require.NoError(t, err)
					require.Equal(t, 1, mustCount(t, page.Locator("[data-theme-contrast-pair=primary] [data-theme-contrast-grade]:not([hidden])")))
					if sample.grade == "aaa" {
						ratio, err := page.Locator("[data-theme-contrast-pair=primary] [data-theme-contrast-ratio]").TextContent()
						require.NoError(t, err)
						require.Equal(t, "21.00:1", ratio)
					}
					valid, err := page.Evaluate(`()=>{
      const p=document.querySelector('[data-theme-contrast-pair=primary]'),s=p.querySelector('[data-theme-contrast-sample]'),r=p.querySelector('[data-theme-contrast-reverse]'),bounds=p.getBoundingClientRect();
      return getComputedStyle(s).color===getComputedStyle(r).backgroundColor&&getComputedStyle(r).color===getComputedStyle(s).backgroundColor&&bounds.width>0&&bounds.left>=0&&bounds.right<=innerWidth+1&&document.documentElement.scrollWidth<=innerWidth+1;
     }`, nil)
					require.NoError(t, err)
					require.Equal(t, true, valid)
				})
			}
		}
	}
}

func TestThemePage_RadiusUpdatesWhileDragging(t *testing.T) {
	page := newIsolatedPage(t)
	gotoThemePage(t, page)
	slider := page.Locator("[data-theme-radius]")
	require.NoError(t, slider.Focus())
	require.NoError(t, slider.Press("Home"))
	require.NoError(t, slider.Press("ArrowRight"))
	require.InDelta(t, .05, editorDefinition(t, page).RadiusRem, .001)
	box, err := slider.BoundingBox()
	require.NoError(t, err)
	require.NotNil(t, box)
	require.NoError(t, page.Mouse().Move(box.X+box.Width*.25, box.Y+box.Height/2))
	require.NoError(t, page.Mouse().Down())
	defer page.Mouse().Up()
	require.NoError(t, page.Mouse().Move(box.X+box.Width*.75, box.Y+box.Height/2))
	_, err = page.WaitForFunction(`()=>{
  const input=document.querySelector('[data-theme-radius]'),definition=JSON.parse(document.querySelector('[data-theme-definition]').value),frame=document.querySelector('#theme-playground-frame').contentDocument;
  return input.valueAsNumber>2&&definition.radiusRem===input.valueAsNumber&&parseFloat(getComputedStyle(frame.documentElement).getPropertyValue('--radius-radius'))===input.valueAsNumber;
 }`, nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err)
}
