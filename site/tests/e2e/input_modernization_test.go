//go:build e2e && full

package e2e

import (
	"bytes"
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/components/combobox"
	"github.com/araihu/goshtoso/components/search"
	selectfield "github.com/araihu/goshtoso/components/select"
	"github.com/araihu/goshtoso/components/structuredinput"
	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func inputModernizationMarkup(t *testing.T, component templ.Component) string {
	t.Helper()
	var b bytes.Buffer
	require.NoError(t, component.Render(t.Context(), &b))
	return b.String()
}

func inputModernizationPage(t *testing.T) playwright.Page {
	t.Helper()
	page := newPage(t, sharedBrowser)
	_, err := page.Goto(baseURL + "/components/button")
	require.NoError(t, err)
	_, err = page.WaitForFunction(`() => window.Alpine && window.htmx`, nil)
	require.NoError(t, err)
	_, err = page.Evaluate(`() => { window.inputModernizationErrors=[]; window.addEventListener('error', e => inputModernizationErrors.push(e.message)); const host=document.createElement('div'); host.id='input-modernization'; document.body.append(host); }`)
	require.NoError(t, err)
	t.Cleanup(func() {
		errs, err := page.Evaluate(`() => inputModernizationErrors`)
		require.NoError(t, err)
		require.Empty(t, errs)
	})
	return page
}

func inputModernizationMount(t *testing.T, page playwright.Page, markup string) {
	t.Helper()
	_, err := page.Evaluate(`async html => { const host=document.querySelector('#input-modernization'); await htmx.swap({sourceElement:host,target:host,text:html,swap:'innerHTML'}); await Alpine.nextTick(); }`, markup)
	require.NoError(t, err)
}

func TestSelectBindingDefaultsAndRepeatedMounts(t *testing.T) {
	page := inputModernizationPage(t)
	cfg := selectfield.Config{ID: "binding-probe", Name: "choice", Options: []selectfield.Option{{Value: "a", Label: "Alpha", Selected: true}, {Value: "b", Label: "Beta"}}, Alpine: &selectfield.AlpineConfig{Model: "draft.value"}}
	markup := inputModernizationMarkup(t, selectfield.Select(cfg))
	for _, tc := range []struct{ parent, wantParent, wantSelected string }{{"''", "", "a"}, {"undefined", "undefined", "a"}, {"'missing'", "", ""}, {"'b'", "b", "b"}} {
		inputModernizationMount(t, page, `<section id="binding-parent" x-data="{draft:{value:`+tc.parent+`}}">`+markup+`</section>`)
		_, err := page.WaitForFunction(`expected => {const root=document.querySelector('#binding-probe-trigger').closest('[x-data]'); return String(Alpine.$data(document.querySelector('#binding-parent')).draft.value)===expected[0] && (Alpine.$data(root).selectedOption?.value || '')===expected[1]}`, []string{tc.wantParent, tc.wantSelected})
		require.NoError(t, err)
		_, err = page.Evaluate(`async () => {window.removedSelect=Alpine.$data(document.querySelector('#binding-probe-trigger').closest('[x-data]')); document.querySelector('#binding-probe-trigger').closest('[x-data]').remove(); await Alpine.nextTick(); Alpine.$data(document.querySelector('#binding-parent')).draft.value='b'; await Alpine.nextTick();}`)
		require.NoError(t, err)
		selected, err := page.Evaluate(`() => removedSelect.selectedOption?.value || ''`)
		require.NoError(t, err)
		require.Equal(t, tc.wantSelected, selected, "removed root's watchers must no longer run")
	}
}

func TestComboboxPersistenceOnFragmentEntry(t *testing.T) {
	page := inputModernizationPage(t)
	_, err := page.Evaluate(`() => sessionStorage.setItem('goshtoso:combobox:industry:selected','["tech"]')`)
	require.NoError(t, err)
	for range 2 {
		_, err = page.Evaluate(`async () => { await htmx.ajax('GET','/components/combobox',{target:'#main-content',swap:'innerHTML'}); await Alpine.nextTick(); }`)
		require.NoError(t, err)
		_, err = page.WaitForFunction(`() => document.querySelector('#industry input[type=hidden][value=tech]') && document.querySelector('#industry-trigger-label').textContent==='Technology'`, nil)
		require.NoError(t, err)
		_, err = page.Evaluate(`async () => {await htmx.ajax('GET','/components/button',{target:'#main-content',swap:'innerHTML'}); await Alpine.nextTick();}`)
		require.NoError(t, err)
	}
}

func TestComboboxPersistenceMountIsolation(t *testing.T) {
	page := inputModernizationPage(t)
	var markup string
	for _, id := range []string{"one", "two", "malformed", "disabled"} {
		cfg := combobox.Config{ID: id, Name: id, DisablePersistence: id == "disabled", Source: combobox.Source{Static: []combobox.Option{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}}}}
		markup += inputModernizationMarkup(t, combobox.Combobox(cfg, combobox.State{Options: cfg.Source.Static, Selected: []string{"a"}}))
	}
	_, err := page.Evaluate(`() => {for (const id of ['one','two','disabled']) sessionStorage.setItem('goshtoso:combobox:'+id+':selected','["b"]'); sessionStorage.setItem('goshtoso:combobox:malformed:selected','{bad');}`)
	require.NoError(t, err)
	inputModernizationMount(t, page, markup)
	values, err := page.Evaluate(`() => ['one','two','malformed','disabled'].map(id=>document.querySelector('#'+id+' input[type=hidden]').value)`)
	require.NoError(t, err)
	require.Equal(t, []any{"b", "b", "a", "a"}, values)
	_, err = page.Evaluate(`async () => {sessionStorage.setItem('goshtoso:combobox:one:selected','["a"]'); window.dispatchEvent(new PageTransitionEvent('pageshow',{persisted:false})); await Alpine.nextTick();}`)
	require.NoError(t, err)
	value, err := page.Locator("#one input[type=hidden]").InputValue()
	require.NoError(t, err)
	require.Equal(t, "b", value, "ordinary pageshow must not overwrite live edits")
	_, err = page.Evaluate(`async () => {window.dispatchEvent(new PageTransitionEvent('pageshow',{persisted:true})); await Alpine.nextTick();}`)
	require.NoError(t, err)
	value, err = page.Locator("#one input[type=hidden]").InputValue()
	require.NoError(t, err)
	require.Equal(t, "a", value)
	_, err = page.Evaluate(`() => {Storage.prototype.getItem=function(){throw new Error('storage unavailable')};}`)
	require.NoError(t, err)
	inputModernizationMount(t, page, markup)
	value, err = page.Locator("#one input[type=hidden]").InputValue()
	require.NoError(t, err)
	require.Equal(t, "a", value)
}

func TestComboboxServerMorphReconcilesDataAndPreservesDraftFocus(t *testing.T) {
	page := inputModernizationPage(t)
	cfg := combobox.Config{ID: "server-morph", Name: "pick", EnableSearch: true, EnableClearAll: true, Mode: combobox.ModeMultiple, Source: combobox.Source{LazyEndpoint: "/unused"}, OptionsEndpoint: "/unused", ToggleEndpoint: "/unused", ClearEndpoint: "/unused"}
	old := combobox.State{Options: []combobox.Option{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}}, Selected: []string{"a"}, Search: "original"}
	inputModernizationMount(t, page, inputModernizationMarkup(t, combobox.Combobox(cfg, old)))
	_, err := page.Evaluate(`async () => {window.originalCombo=document.querySelector('#server-morph'); const data=Alpine.$data(originalCombo);data.isOpen=true;await Alpine.nextTick();window.originalSearch=originalCombo.querySelector('[data-combobox-search]');originalSearch.focus();originalSearch.value='dirty query';originalSearch.setSelectionRange(3,3);}`)
	require.NoError(t, err)
	for _, selected := range [][]string{{"b"}, nil, {"b"}} {
		state := combobox.State{Options: []combobox.Option{{Value: "b", Label: "Changed Beta"}}, Selected: selected, Search: "server reset"}
		_, err = page.Evaluate(`async html => {await htmx.swap({sourceElement:originalCombo,target:originalCombo,text:html,swap:'outerMorph'}); await Alpine.nextTick();}`, inputModernizationMarkup(t, combobox.Combobox(cfg, state)))
		require.NoError(t, err)
		result, err := page.Evaluate(`() => ({same:document.querySelector('#server-morph')===originalCombo,open:Alpine.$data(originalCombo).isOpen,focus:document.activeElement===originalSearch,value:originalSearch.value,caret:originalSearch.selectionStart,selected:[...originalCombo.querySelectorAll('input[type=hidden]')].map(e=>e.value),aria:originalCombo.querySelector('[role=option]').getAttribute('aria-selected'),options:originalCombo.querySelectorAll('[role=option]').length,label:originalCombo.querySelector('[data-combobox-trigger-label-outer]').textContent})`)
		require.NoError(t, err)
		wantValues := []any{}
		label := "Select…"
		aria := "false"
		if len(selected) > 0 {
			wantValues = []any{"b"}
			label = "Changed Beta"
			aria = "true"
		}
		require.Equal(t, map[string]any{"same": true, "open": true, "focus": true, "value": "dirty query", "caret": 3, "selected": wantValues, "aria": aria, "options": 1, "label": label}, result)
	}
}

func TestInputReplacementResetsServerConfiguration(t *testing.T) {
	page := inputModernizationPage(t)
	render := func(updated bool) string {
		opts := []selectfield.Option{{Value: "a", Label: "Alpha", Selected: true}, {Value: "b", Label: "Beta"}}
		rows := []structuredinput.Entry{{"key": "initial"}}
		columns := []structuredinput.Column{{Key: "key", Label: "Key"}}
		url := "/old-source"
		if updated {
			opts = []selectfield.Option{{Value: "b", Label: "Updated Beta", Selected: true}}
			rows = []structuredinput.Entry{{"other": "reset"}}
			columns = []structuredinput.Column{{Key: "other", Label: "Other", Default: "new default"}}
			url = "/new-source"
		}
		return inputModernizationMarkup(t, selectfield.Select(selectfield.Config{ID: "replace-select", Name: "select", Options: opts})) + inputModernizationMarkup(t, structuredinput.StructuredInput(structuredinput.Config{ID: "replace-rows", Name: "rows", Entries: rows, Columns: columns})) + inputModernizationMarkup(t, search.Search(search.Config{ID: "replace-search", ItemsURL: url}))
	}
	inputModernizationMount(t, page, render(false))
	_, err := page.Evaluate(`() => {Alpine.$data(document.querySelector('#replace-rows')).entries=[['dirty']];const state=Alpine.$data(document.querySelector('#replace-search-dialog'));state.clientItems=[{title:'cached'}];state.clientItemsLoaded=true;state.query='dirty';}`)
	require.NoError(t, err)
	inputModernizationMount(t, page, render(true))
	result, err := page.Evaluate(`() => {const rows=Alpine.$data(document.querySelector('#replace-rows'));const search=Alpine.$data(document.querySelector('#replace-search-dialog'));const select=Alpine.$data(document.querySelector('#replace-select-trigger').closest('[x-data]'));return {rows:rows.entries,default:rows.newRow,selection:select.selectedOption.label,query:search.query,loaded:search.clientItemsLoaded,items:search.clientItems,source:document.querySelector('#replace-search-dialog').dataset.searchSourceUrl}}`)
	require.NoError(t, err)
	require.Equal(t, map[string]any{"rows": []any{[]any{"reset"}}, "default": []any{"new default"}, "selection": "Updated Beta", "query": "", "loaded": false, "items": []any{}, "source": "/new-source"}, result)
}
