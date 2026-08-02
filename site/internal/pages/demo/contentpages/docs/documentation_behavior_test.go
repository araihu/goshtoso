package docspages

import (
	"context"
	"strings"
	"testing"

	"github.com/araihu/goshtoso/components/combobox"
	"github.com/araihu/goshtoso/components/form"
	linkcomponent "github.com/araihu/goshtoso/components/link"
	"github.com/araihu/goshtoso/components/pagination"
	"github.com/araihu/goshtoso/components/schemaform"
	"github.com/araihu/goshtoso/components/search"
	"github.com/araihu/goshtoso/components/sidebar"
	"github.com/araihu/goshtoso/components/spinner"
	"github.com/araihu/goshtoso/components/structuredinput"
	"github.com/araihu/goshtoso/components/toast"
	"github.com/araihu/goshtoso/components/tooltip"
	"github.com/araihu/goshtoso/site/internal/pages/catalog"
	"github.com/stretchr/testify/require"
)

func TestRepresentativeDocumentedBehaviorMatchesRendering(t *testing.T) {
	var sidebarOut strings.Builder
	require.NoError(t, sidebar.Sidebar(sidebar.Config{LogoText: "Docs"}).Render(context.Background(), &sidebarOut))
	require.Contains(t, sidebarOut.String(), `href="/"`)

	var tooltipOut strings.Builder
	require.NoError(t, tooltip.Tooltip("audit-tooltip", "Audit label").Render(context.Background(), &tooltipOut))
	require.Contains(t, tooltipOut.String(), ">Hover Me</button>")
	require.Contains(t, tooltipOut.String(), `aria-describedby="audit-tooltip"`)

	var spinnerOut strings.Builder
	require.NoError(t, spinner.Spinner(spinner.Config{}).Render(context.Background(), &spinnerOut))
	require.Contains(t, spinnerOut.String(), `aria-hidden="true"`)
	require.NotContains(t, spinnerOut.String(), "aria-busy")
	require.NotContains(t, spinnerOut.String(), `role="status"`)

	var toastOut strings.Builder
	require.NoError(t, toast.ToastContainer(toast.ContainerConfig{}).Render(context.Background(), &toastOut))
	require.Contains(t, toastOut.String(), "md:bottom-0")
	require.Contains(t, toastOut.String(), "md:right-0")

	var linkOut strings.Builder
	require.NoError(t, linkcomponent.Link("/docs", linkcomponent.WithTarget("_blank")).Render(context.Background(), &linkOut))
	require.Contains(t, linkOut.String(), `rel="noopener noreferrer"`)
	require.Equal(t, "/items?filter=open&page=3", pagination.Config{BaseURL: "/items?filter=open"}.PageURL(3))
}

func TestComplexInputDocumentedBehaviorMatchesRendering(t *testing.T) {
	staticConfig := combobox.Config{
		ID:     "required-audit",
		Name:   "required-audit",
		Source: combobox.Source{Static: []combobox.Option{{Value: "a", Label: "A"}}},
	}
	var optionalCombobox strings.Builder
	require.NoError(t, combobox.Combobox(staticConfig, staticConfig.InitialState()).Render(context.Background(), &optionalCombobox))
	staticConfig.Required = true
	var requiredCombobox strings.Builder
	require.NoError(t, combobox.Combobox(staticConfig, staticConfig.InitialState()).Render(context.Background(), &requiredCombobox))
	require.NotContains(t, optionalCombobox.String(), `aria-required="true"`)
	require.Contains(t, requiredCombobox.String(), `aria-required="true"`)

	lazyConfig := combobox.Config{
		ID:              "lazy-audit",
		Name:            "lazy-audit",
		EnableSearch:    true,
		ToggleEndpoint:  "/toggle",
		OptionsEndpoint: "/options",
		ClearEndpoint:   "/clear",
		Source:          combobox.Source{LazyEndpoint: "/lazy-source"},
	}
	var lazyCombobox strings.Builder
	require.NoError(t, combobox.Combobox(lazyConfig, combobox.State{}).Render(context.Background(), &lazyCombobox))
	require.Contains(t, lazyCombobox.String(), `hx-get="/options"`)
	require.NotContains(t, lazyCombobox.String(), "/lazy-source")

	var defaultForm strings.Builder
	require.NoError(t, form.Form(form.Config{}).Render(context.Background(), &defaultForm))
	require.Contains(t, defaultForm.String(), "@keydown.enter")
	allowEnter := false
	var enterForm strings.Builder
	require.NoError(t, form.Form(form.Config{PreventEnterSubmit: &allowEnter}).Render(context.Background(), &enterForm))
	require.NotContains(t, enterForm.String(), "@keydown.enter")

	var schemaFields strings.Builder
	require.NoError(t, schemaform.Fields(schemaform.FieldsConfig{
		Fields: []schemaform.Field{{Path: "replicas", Label: "Replicas", Kind: schemaform.KindInteger}},
	}).Render(context.Background(), &schemaFields))
	require.Contains(t, schemaFields.String(), `name="values.replicas"`)
	require.Contains(t, schemaFields.String(), `step="1"`)

	var searchField strings.Builder
	require.NoError(t, search.SearchField(search.Config{}).Render(context.Background(), &searchField))
	require.Contains(t, searchField.String(), `id="search"`)
	require.Contains(t, searchField.String(), "⌘ K")

	var structured strings.Builder
	require.NoError(t, structuredinput.StructuredInput(structuredinput.Config{
		Name: "rules",
		Columns: []structuredinput.Column{{
			Key:     "priority",
			Type:    structuredinput.ColumnSelect,
			Options: []structuredinput.Option{{Value: "high"}},
		}},
	}).Render(context.Background(), &structured))
	require.Contains(t, structured.String(), "Add row")
	require.Contains(t, structured.String(), "high")
	require.Contains(t, structured.String(), "structuredInput")
}

func TestCorrectedCatalogDescriptionsMatchRenderedBehavior(t *testing.T) {
	descriptions := make(map[string]string)
	for _, page := range catalog.ComponentPages() {
		descriptions[page.Key] = strings.ToLower(page.Description)
	}

	require.Contains(t, descriptions["components/fileinput"], "native accept hints")
	require.NotContains(t, descriptions["components/fileinput"], "validation states")
	require.Contains(t, descriptions["components/range"], "generated or custom ticks")
	require.Contains(t, descriptions["components/tags-list"], "duplicate-preserving")
	require.NotContains(t, descriptions["components/textarea"], "counter")
	require.Contains(t, descriptions["components/structured-input"], "repeatable")
	require.Contains(t, descriptions["components/structured-input"], "typed columns")
	require.NotContains(t, descriptions["components/toast"], "position")
	require.Contains(t, descriptions["components/spinner"], "decorative")
	require.NotContains(t, descriptions["components/spinner"], "label")
}
