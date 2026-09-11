//go:build e2e && full && goshtoso_current_source

package e2e

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/assets"
	"github.com/araihu/goshtoso/components/carousel"
	"github.com/araihu/goshtoso/components/codeblock"
	"github.com/araihu/goshtoso/components/combobox"
	"github.com/araihu/goshtoso/components/fileinput"
	"github.com/araihu/goshtoso/components/head"
	"github.com/araihu/goshtoso/components/navbar"
	"github.com/araihu/goshtoso/components/pagination"
	"github.com/araihu/goshtoso/components/schemaform"
	"github.com/araihu/goshtoso/expressions"
	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func expressionFixture(t *testing.T) *httptest.Server {
	t.Helper()
	renderer := expressions.NewRenderer(expressions.Set{CodeBlock: expressions.CodeBlock{CopyLabel: "Copiar", CopiedLabel: `Copiado 'sim' "agora"`, ErrorText: "Falhou"}})
	mux := http.NewServeMux()
	mux.Handle("GET /assets/", assets.Handler())
	mux.HandleFunc("GET /fragment", func(w http.ResponseWriter, r *http.Request) {
		ctx := expressions.With(r.Context(), expressions.Set{Pagination: expressions.Pagination{NextText: "Continuar", NextLabel: "Próxima página"}})
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = renderer.Render(ctx, w, pagination.Pagination(pagination.Config{Mode: pagination.ModeSimple, CurrentPage: 1, TotalPages: 2}))
	})
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		ctx := expressions.With(r.Context(), expressions.Set{
			Combobox:   expressions.Combobox{Placeholder: "Escolha", ClearLabel: "Limpar", SelectedLabel: func(n int) string { return fmt.Sprintf("Escolhidos: %d 'itens'", n) }},
			Carousel:   expressions.Carousel{PauseLabel: "Pausar", PlayLabel: "Reproduzir", SlideLabel: func(n int) string { return fmt.Sprintf("Imagem %d", n) }},
			Navbar:     expressions.Navbar{OpenMenuLabel: `Abrir "menu"`, CloseMenuLabel: `Fechar 'menu'`},
			SchemaForm: expressions.SchemaForm{RequiredLabel: "Obrigatório"},
			FileInput:  expressions.FileInput{EmptyText: `Nenhum "arquivo"`},
		})
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = io.WriteString(w, `<!DOCTYPE html><html><head>`)
		_ = head.Dependencies(head.WithLocalRuntime()).Render(ctx, w)
		_, _ = io.WriteString(w, `</head><body>`)
		for _, component := range expressionFixtureComponents() {
			_ = renderer.Render(ctx, w, component)
		}
		_, _ = io.WriteString(w, `<button id="fragment-trigger" hx-get="/fragment" hx-target="#fragment">Carregar</button><div id="fragment"></div></body></html>`)
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server
}

func expressionFixtureComponents() []templ.Component {
	cfg := combobox.Config{ID: "choices", Name: "choices", Mode: combobox.ModeMultiple, DisablePersistence: false, EnableClearAll: true, Source: combobox.Source{Static: []combobox.Option{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}, {Value: "c", Label: "Gamma"}}}}
	return []templ.Component{
		schemaform.Fields(schemaform.FieldsConfig{Fields: []schemaform.Field{{Path: "required-name", Name: "required-name", Label: "Required name", Kind: schemaform.KindString, Required: true}}}),
		codeblock.CodeBlock(codeblock.Config{Code: "hello", Language: "text"}),
		combobox.Combobox(cfg, cfg.InitialState()),
		fileinput.FileInput(fileinput.Config{ID: "upload", Name: "upload", Appearance: fileinput.AppearanceUpload}),
		navbar.Navbar(navbar.Config{Links: []navbar.NavLink{{Label: "Home", Href: "/"}}}),
		carousel.Carousel(carousel.Config{ID: "slides", Autoplay: &carousel.AutoplayConfig{Interval: 60000}, Slides: []carousel.Slide{{ImgSrc: "/assets/images/goshtoso-logo.svg", ImgAlt: "One"}, {ImgSrc: "/assets/images/goshtoso-logo.svg", ImgAlt: "Two"}}}),
	}
}

func TestExpressions_CurrentSourceBrowser(t *testing.T) {
	server := expressionFixture(t)
	for _, theme := range []string{"goshtoso", "minimal"} {
		for _, dark := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/dark=%t", theme, dark), func(t *testing.T) {
				page := newPage(t, sharedBrowser, playwright.BrowserNewPageOptions{Viewport: &playwright.Size{Width: 390, Height: 844}})
				failures := watchPageFailures(page)
				_, err := page.Goto(server.URL)
				require.NoError(t, err)
				_, err = page.WaitForFunction("() => window.Alpine && window.goshtosoComboboxClient", nil)
				require.NoError(t, err)
				_, err = page.Evaluate(`async ([theme,dark]) => { document.documentElement.dataset.theme=theme; document.documentElement.classList.toggle('dark',dark); Object.defineProperty(navigator,'clipboard',{configurable:true,value:{writeText:async()=>{}}}); }`, []any{theme, dark})
				require.NoError(t, err)
				assertExpressionCopy(t, page)
				assertExpressionSelection(t, page)
				require.NoError(t, page.Locator("[data-open-label]").Click())
				label, err := page.Locator("[data-open-label]").Evaluate("el => el.getAttribute('aria-label')", nil)
				require.NoError(t, err)
				require.Equal(t, "Fechar 'menu'", label)
				require.NoError(t, page.Keyboard().Press("Escape"))
				label, err = page.Locator("[data-open-label]").Evaluate("el => el.getAttribute('aria-label')", nil)
				require.NoError(t, err)
				require.Equal(t, `Abrir "menu"`, label)
				text, err := page.Locator("[data-empty-text]").TextContent()
				require.NoError(t, err)
				require.Equal(t, `Nenhum "arquivo"`, text)
				_, err = page.WaitForFunction(`() => !!document.querySelector('#slides [aria-label="Imagem 2"]')`, nil)
				require.NoError(t, err)
				require.NoError(t, page.Locator("#fragment-trigger").Click())
				require.NoError(t, page.Locator(`#fragment [aria-label="Próxima página"]`).WaitFor())
				text, err = page.Locator(`#fragment [aria-label="Próxima página"]`).TextContent()
				require.NoError(t, err)
				require.Contains(t, text, "Continuar")
				require.NoError(t, page.Locator("[data-pause-label]").Click())
				_, err = page.WaitForFunction(`() => document.querySelector('[data-pause-label]').getAttribute('aria-label') === 'Reproduzir'`, nil)
				require.NoError(t, err)
				requiredState, err := page.Evaluate(`() => { const input=document.querySelector('input[name$="required-name"]'); return input.required && input.validity.valueMissing && !input.checkValidity(); }`)
				require.NoError(t, err)
				require.Equal(t, true, requiredState)
				failures.RequireEmpty(t)
			})
		}
	}
}

func assertExpressionCopy(t *testing.T, page playwright.Page) {
	t.Helper()
	button := page.Locator("[data-code-block-copy]")
	require.NoError(t, button.Click())
	_, err := page.WaitForFunction(`() => document.querySelector('[data-code-block-copy-status]').textContent === 'Copiado \'sim\' "agora"'`, nil)
	require.NoError(t, err)
	_, err = page.Evaluate(`() => { navigator.clipboard.writeText = async () => { throw new Error('fixture failure') }; }`)
	require.NoError(t, err)
	require.NoError(t, button.Click())
	_, err = page.WaitForFunction(`() => document.querySelector('[data-code-block-copy-status]').textContent === 'Falhou'`, nil)
	require.NoError(t, err)
}

func assertExpressionSelection(t *testing.T, page playwright.Page) {
	t.Helper()
	require.NoError(t, page.Locator("#choices-trigger").Click())
	require.NoError(t, page.Locator(`#choices [data-value="a"]`).Click())
	require.NoError(t, page.Locator(`#choices [data-value="b"]`).Click())
	_, err := page.WaitForFunction(`() => document.querySelector('#choices-trigger-label').textContent === "Escolhidos: 2 'itens'"`, nil)
	require.NoError(t, err)
	require.NoError(t, page.Keyboard().Press("Escape"))
	_, err = page.Evaluate(`() => { sessionStorage.setItem('goshtoso:combobox:choices:selected', JSON.stringify(['a','b','a','missing','gone','old'])); Alpine.$data(document.querySelector('#choices')).restoreSelection(); }`)
	require.NoError(t, err)
	_, err = page.WaitForFunction(`() => document.querySelector('#choices-trigger-label').textContent === "Escolhidos: 2 'itens'"`, nil)
	require.NoError(t, err)
}
