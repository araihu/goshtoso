package components

import (
	"context"
	"reflect"
	"strings"
	"testing"

	rootcomponents "github.com/araihu/goshtoso/components"
	"github.com/araihu/goshtoso/components/accordion"
	"github.com/araihu/goshtoso/components/alert"
	"github.com/araihu/goshtoso/components/avatar"
	"github.com/araihu/goshtoso/components/badge"
	"github.com/araihu/goshtoso/components/banner"
	"github.com/araihu/goshtoso/components/breadcrumbs"
	"github.com/araihu/goshtoso/components/button"
	"github.com/araihu/goshtoso/components/card"
	"github.com/araihu/goshtoso/components/carousel"
	"github.com/araihu/goshtoso/components/chatbubble"
	"github.com/araihu/goshtoso/components/checkbox"
	"github.com/araihu/goshtoso/components/codeblock"
	"github.com/araihu/goshtoso/components/combobox"
	"github.com/araihu/goshtoso/components/drawer"
	"github.com/araihu/goshtoso/components/dropdown"
	"github.com/araihu/goshtoso/components/fileinput"
	"github.com/araihu/goshtoso/components/form"
	linkcomponent "github.com/araihu/goshtoso/components/link"
	"github.com/araihu/goshtoso/components/modal"
	"github.com/araihu/goshtoso/components/navbar"
	"github.com/araihu/goshtoso/components/pagination"
	"github.com/araihu/goshtoso/components/palette"
	"github.com/araihu/goshtoso/components/radio"
	rangeinput "github.com/araihu/goshtoso/components/range"
	"github.com/araihu/goshtoso/components/rating"
	"github.com/araihu/goshtoso/components/schemaform"
	"github.com/araihu/goshtoso/components/search"
	selectfield "github.com/araihu/goshtoso/components/select"
	"github.com/araihu/goshtoso/components/sidebar"
	"github.com/araihu/goshtoso/components/spinner"
	"github.com/araihu/goshtoso/components/steps"
	"github.com/araihu/goshtoso/components/structuredinput"
	"github.com/araihu/goshtoso/components/table"
	"github.com/araihu/goshtoso/components/tabs"
	"github.com/araihu/goshtoso/components/tagslist"
	"github.com/araihu/goshtoso/components/textarea"
	"github.com/araihu/goshtoso/components/textinput"
	"github.com/araihu/goshtoso/components/toast"
	"github.com/araihu/goshtoso/components/toggle"
	"github.com/araihu/goshtoso/components/tooltip"
	"github.com/araihu/goshtoso/site/internal/pages/catalog"
	"github.com/araihu/goshtoso/site/internal/pages/demo"
	"github.com/stretchr/testify/require"
)

func TestFeedbackNavigationAPIMetadataRegistered(t *testing.T) {
	expected := map[string][]rootcomponents.Kind{
		"components/alert":       {rootcomponents.KindAlert},
		"components/toast":       {rootcomponents.KindToastContainer, rootcomponents.KindToast, rootcomponents.KindMessageToast, rootcomponents.KindOOBToast, rootcomponents.KindOOBMessageToast},
		"components/modal":       {rootcomponents.KindModal, rootcomponents.KindAlertDialog},
		"components/drawer":      {rootcomponents.KindDrawer},
		"components/spinner":     {rootcomponents.KindSpinner},
		"components/steps":       {rootcomponents.KindSteps},
		"components/tooltip":     {rootcomponents.KindTooltip},
		"components/breadcrumbs": {rootcomponents.KindBreadcrumbs},
		"components/dropdown":    {rootcomponents.KindDropdown},
		"components/link":        {rootcomponents.KindLink},
		"components/navbar":      {rootcomponents.KindNavbar},
		"components/pagination":  {rootcomponents.KindPagination},
		"components/sidebar":     {rootcomponents.KindSidebar, rootcomponents.KindSidebarOverlay},
		"components/tabs":        {rootcomponents.KindTabs},
	}

	require.Len(t, expected, 14)
	for key, wantKinds := range expected {
		entry, ok := Demos[key]
		require.Truef(t, ok, "missing Feedback/Navigation registry entry %q", key)
		require.NotEmptyf(t, entry.API, "%s must register structured API metadata", key)

		var catalogKinds []rootcomponents.Kind
		for _, page := range catalog.ComponentPages() {
			if page.Key == key {
				catalogKinds = page.Kinds
				break
			}
		}
		require.Equalf(t, wantKinds, catalogKinds, "%s catalog Kinds", key)

		seenIDs := make(map[string]struct{}, len(entry.API))
		gotKinds := make([]rootcomponents.Kind, 0, len(entry.API))
		for _, section := range entry.API {
			require.NotEmptyf(t, section.ID, "%s contains an empty API section ID", key)
			_, duplicate := seenIDs[section.ID]
			require.Falsef(t, duplicate, "%s contains duplicate API section ID %q", key, section.ID)
			seenIDs[section.ID] = struct{}{}
			if section.Kind != "" {
				gotKinds = append(gotKinds, section.Kind)
			}
		}
		require.ElementsMatchf(t, wantKinds, gotKinds, "%s API section Kinds", key)
	}
}

func TestFeedbackNavigationAPIRegistryUsesPageSectionSlices(t *testing.T) {
	expected := map[string][]demo.APISection{
		"components/alert":       alertAPISections,
		"components/toast":       toastAPISections,
		"components/modal":       modalAPISections,
		"components/drawer":      drawerAPISections,
		"components/spinner":     spinnerAPISections,
		"components/steps":       stepsAPISections,
		"components/tooltip":     tooltipAPISections,
		"components/breadcrumbs": breadcrumbsAPISections,
		"components/dropdown":    dropdownAPISections,
		"components/link":        linkAPISections,
		"components/navbar":      navbarAPISections,
		"components/pagination":  paginationAPISections,
		"components/sidebar":     sidebarAPISections,
		"components/tabs":        tabsAPISections,
	}

	require.Len(t, expected, 14)
	for key, pageSections := range expected {
		entry := Demos[key]
		require.NotEmpty(t, entry.API, key)
		require.NotEmpty(t, pageSections, key)
		require.Samef(
			t,
			&pageSections[0],
			&entry.API[0],
			"%s registry API must use the page's named metadata slice",
			key,
		)

		var rendered strings.Builder
		require.NoError(t, entry.Content().Render(context.Background(), &rendered))
		html := rendered.String()
		require.Equalf(t, 1, strings.Count(html, "data-api-reference"), "%s must render one structured API reference", key)
		for _, section := range pageSections {
			require.Equalf(
				t,
				1,
				strings.Count(html, `data-api-section="`+section.ID+`"`),
				"%s must render its %q API section exactly once",
				key,
				section.ID,
			)
		}
	}
}

func TestFeedbackNavigationStructAPIsDocumentEveryExportedFieldExactlyOnce(t *testing.T) {
	expectedTypes := []reflect.Type{
		reflect.TypeFor[alert.Config](),
		reflect.TypeFor[alert.LinkConfig](),
		reflect.TypeFor[alert.HTMXConfig](),
		reflect.TypeFor[alert.ActionConfig](),
		reflect.TypeFor[drawer.Config](),
		reflect.TypeFor[modal.Config](),
		reflect.TypeFor[modal.AlertDialogConfig](),
		reflect.TypeFor[modal.ButtonAction](),
		reflect.TypeFor[modal.HTMXConfig](),
		reflect.TypeFor[spinner.Config](),
		reflect.TypeFor[steps.Config](),
		reflect.TypeFor[steps.Step](),
		reflect.TypeFor[toast.Config](),
		reflect.TypeFor[toast.MessageConfig](),
		reflect.TypeFor[toast.Sender](),
		reflect.TypeFor[toast.HTMXConfig](),
		reflect.TypeFor[toast.ContainerConfig](),
		reflect.TypeFor[breadcrumbs.Config](),
		reflect.TypeFor[breadcrumbs.Item](),
		reflect.TypeFor[dropdown.Config](),
		reflect.TypeFor[dropdown.Item](),
		reflect.TypeFor[dropdown.Section](),
		reflect.TypeFor[navbar.Config](),
		reflect.TypeFor[navbar.NavLink](),
		reflect.TypeFor[navbar.UserProfile](),
		reflect.TypeFor[navbar.UserMenuItem](),
		reflect.TypeFor[navbar.ActionItem](),
		reflect.TypeFor[pagination.Config](),
		reflect.TypeFor[pagination.HTMXConfig](),
		reflect.TypeFor[pagination.PageItem](),
		reflect.TypeFor[sidebar.Config](),
		reflect.TypeFor[sidebar.Item](),
		reflect.TypeFor[sidebar.Section](),
		reflect.TypeFor[sidebar.OverlayConfig](),
		reflect.TypeFor[tabs.Config](),
		reflect.TypeFor[tabs.Tab](),
		reflect.TypeFor[tabs.TabHTMX](),
	}

	sectionsByType := make(map[reflect.Type][]demo.APISection, len(expectedTypes))
	for _, sections := range feedbackNavigationAPISectionSlices() {
		for _, section := range sections {
			require.NotEmpty(t, section.ID)
			require.NotEmpty(t, section.Title)
			require.NotEmptyf(t, section.Description, "%s section description", section.Title)

			seenProps := make(map[string]struct{}, len(section.Props))
			for _, prop := range section.Props {
				require.NotContainsf(t, seenProps, prop.Name, "%s.%s documented twice", section.Title, prop.Name)
				seenProps[prop.Name] = struct{}{}
				require.NotEmptyf(t, prop.Default, "%s.%s default", section.Title, prop.Name)
				require.NotEmptyf(t, prop.Description, "%s.%s description", section.Title, prop.Name)
			}
			if section.StructType != nil {
				sectionsByType[section.StructType] = append(sectionsByType[section.StructType], section)
			}
		}
	}
	require.Len(t, sectionsByType, len(expectedTypes))

	for _, typ := range expectedTypes {
		sections := sectionsByType[typ]
		require.Lenf(t, sections, 1, "%s must have exactly one StructAPI section", typ)
		section := sections[0]

		documented := make(map[string]demo.APIPropDoc, len(section.Props))
		for _, prop := range section.Props {
			documented[prop.Name] = prop
		}
		exportedCount := 0
		for index := range typ.NumField() {
			field := typ.Field(index)
			if !field.IsExported() {
				continue
			}
			exportedCount++
			prop, ok := documented[field.Name]
			require.Truef(t, ok, "%s.%s must be documented", typ, field.Name)
			if isComponentNamedScalar(field.Type) {
				require.NotEmptyf(t, prop.Allowed, "%s.%s must list allowed values", typ, field.Name)
			}
		}
		require.Lenf(t, documented, exportedCount, "%s must document every exported field exactly once", typ)
	}
}

func TestFeedbackNavigationPublicSignaturesAreExactAndUnique(t *testing.T) {
	expected := map[string]string{
		"Tooltip options.id":               "id string",
		"Tooltip options.label":            "label string",
		"Tooltip options.WithDescription":  "func WithDescription(description string) Option",
		"Tooltip options.WithPosition":     "func WithPosition(position Position) Option",
		"Tooltip options.WithActivation":   "func WithActivation(activation Activation) Option",
		"Tooltip options.WithTriggerLabel": "func WithTriggerLabel(label string) Option",
		"Tooltip options.WithTrigger":      "func WithTrigger(trigger templ.Component) Option",
		"Link options.href":                "href string",
		"Link options.WithTarget":          "func WithTarget(target string) Option",
		"Link options.WithRel":             "func WithRel(rel string) Option",
		"Link options.WithRole":            "func WithRole(role string) Option",
		"Link options.WithID":              "func WithID(id string) Option",
		"Link options.WithAppearance":      "func WithAppearance(appearance Appearance) Option",
		"Link options.WithSize":            "func WithSize(size Size) Option",
		"Link options.WithIcon":            "func WithIcon(icon templ.Component) Option",
		"Link options.WithIconPosition":    "func WithIconPosition(position IconPosition) Option",
		"Link options.WithRootClass":       "func WithRootClass(class string) Option",
		"Link options.WithAttrs":           "func WithAttrs(attrs templ.Attributes) Option",
		"Config helpers.HasPrevious":       "func (cfg Config) HasPrevious() bool",
		"Config helpers.HasNext":           "func (cfg Config) HasNext() bool",
		"Config helpers.PreviousPage":      "func (cfg Config) PreviousPage() int",
		"Config helpers.NextPage":          "func (cfg Config) NextPage() int",
		"Config helpers.PageURL":           "func (cfg Config) PageURL(page int) string",
		"Config helpers.Pages":             "func (cfg Config) Pages() []PageItem",
	}

	got := make(map[string]string, len(expected))
	for _, sections := range feedbackNavigationAPISectionSlices() {
		for _, section := range sections {
			for _, prop := range section.Props {
				if prop.Signature == "" {
					continue
				}
				key := section.Title + "." + prop.Name
				require.NotContainsf(t, got, key, "%s documented more than once", key)
				got[key] = prop.Signature
			}
		}
	}
	require.Equal(t, expected, got)
}

func TestFeedbackNavigationConstructorsAreDocumentedExactlyOnce(t *testing.T) {
	expected := map[string]rootcomponents.Kind{
		"alert.Alert(cfg Config) Instance":                                     rootcomponents.KindAlert,
		"drawer.Drawer(cfg Config) Instance":                                   rootcomponents.KindDrawer,
		"modal.Modal(cfg Config) Instance":                                     rootcomponents.KindModal,
		"modal.AlertDialog(cfg AlertDialogConfig) AlertDialogInstance":         rootcomponents.KindAlertDialog,
		"spinner.Spinner(cfg Config) Instance":                                 rootcomponents.KindSpinner,
		"steps.Steps(cfg Config) Instance":                                     rootcomponents.KindSteps,
		"toast.ToastContainer(cfg ContainerConfig) ContainerInstance":          rootcomponents.KindToastContainer,
		"toast.Toast(cfg Config) Instance":                                     rootcomponents.KindToast,
		"toast.MessageToast(cfg MessageConfig) MessageInstance":                rootcomponents.KindMessageToast,
		"toast.OOBToast(cfg Config) OOBInstance":                               rootcomponents.KindOOBToast,
		"toast.OOBMessageToast(cfg MessageConfig) OOBMessageInstance":          rootcomponents.KindOOBMessageToast,
		"tooltip.Tooltip(id string, label string, options ...Option) Instance": rootcomponents.KindTooltip,
		"breadcrumbs.Breadcrumbs(cfg Config) Instance":                         rootcomponents.KindBreadcrumbs,
		"dropdown.Dropdown(cfg Config) Instance":                               rootcomponents.KindDropdown,
		"link.Link(href string, options ...Option) Instance":                   rootcomponents.KindLink,
		"navbar.Navbar(cfg Config) Instance":                                   rootcomponents.KindNavbar,
		"pagination.Pagination(cfg Config) Instance":                           rootcomponents.KindPagination,
		"sidebar.Sidebar(cfg Config) Instance":                                 rootcomponents.KindSidebar,
		"sidebar.Overlay(cfg OverlayConfig) OverlayInstance":                   rootcomponents.KindSidebarOverlay,
		"tabs.Tabs(cfg Config) Instance":                                       rootcomponents.KindTabs,
	}

	got := make(map[string]rootcomponents.Kind, len(expected))
	for _, sections := range feedbackNavigationAPISectionSlices() {
		for _, section := range sections {
			if section.Constructor == "" {
				continue
			}
			require.NotContainsf(t, got, section.Constructor, "%s documented more than once", section.Constructor)
			got[section.Constructor] = section.Kind
		}
	}
	require.Equal(t, expected, got)
}

func TestFeedbackNavigationMetadataMatchesRepresentativeRenderBranches(t *testing.T) {
	var sidebarOut strings.Builder
	require.NoError(t, sidebar.Sidebar(sidebar.Config{LogoText: "Docs"}).Render(context.Background(), &sidebarOut))
	require.Contains(t, sidebarOut.String(), `href="/"`)
	require.Equal(t, `"/"`, apiProp(t, sidebarAPISections, "Config", "LogoHref").Default)

	var tooltipOut strings.Builder
	require.NoError(t, tooltip.Tooltip("audit-tooltip", "Audit label").Render(context.Background(), &tooltipOut))
	require.Contains(t, tooltipOut.String(), ">Hover Me</button>")
	require.Contains(t, tooltipOut.String(), `aria-describedby="audit-tooltip"`)
	require.True(t, apiProp(t, tooltipAPISections, "Tooltip options", "id").Required)
	require.True(t, apiProp(t, tooltipAPISections, "Tooltip options", "label").Required)
	require.Equal(t, `"Hover Me"`, apiProp(t, tooltipAPISections, "Tooltip options", "WithTriggerLabel").Default)

	var spinnerOut strings.Builder
	require.NoError(t, spinner.Spinner(spinner.Config{}).Render(context.Background(), &spinnerOut))
	spinnerHTML := spinnerOut.String()
	require.Contains(t, spinnerHTML, `aria-hidden="true"`)
	require.NotContains(t, spinnerHTML, "aria-busy")
	require.NotContains(t, spinnerHTML, `role="status"`)
	require.Contains(t, spinnerAPISections[0].Description, "does not render a label")
	require.Contains(t, spinnerAPISections[0].Description, "busy-state wrapper")

	var toastOut strings.Builder
	require.NoError(t, toast.ToastContainer(toast.ContainerConfig{}).Render(context.Background(), &toastOut))
	require.Contains(t, toastOut.String(), "md:bottom-0")
	require.Contains(t, toastOut.String(), "md:right-0")
	require.NotContains(t, strings.ToLower(toastAPISections[0].Description), "configurable position")

	var linkOut strings.Builder
	require.NoError(t, linkcomponent.Link("/docs", linkcomponent.WithTarget("_blank")).Render(context.Background(), &linkOut))
	require.Contains(t, linkOut.String(), `rel="noopener noreferrer"`)

	require.Equal(t, "/items?filter=open&page=3", pagination.Config{BaseURL: "/items?filter=open"}.PageURL(3))
	require.Equal(t, `"innerHTML" when Target is non-empty`, apiProp(t, paginationAPISections, "HTMXConfig", "Swap").Default)
}

func TestCorrectedCatalogDescriptionsMatchRenderedBehavior(t *testing.T) {
	descriptions := make(map[string]string)
	for _, page := range catalog.ComponentPages() {
		descriptions[page.Key] = strings.ToLower(page.Description)
	}

	require.Contains(t, descriptions["components/fileinput"], "native accept hints")
	require.Contains(t, descriptions["components/range"], "generated or custom ticks")
	require.Contains(t, descriptions["components/tags-list"], "duplicate-preserving")
	require.NotContains(t, descriptions["components/textarea"], "counter")
	require.Contains(t, descriptions["components/structured-input"], "repeatable")
	require.Contains(t, descriptions["components/structured-input"], "typed columns")
	require.NotContains(t, descriptions["components/toast"], "position")
	require.Contains(t, descriptions["components/spinner"], "decorative")
	require.NotContains(t, descriptions["components/spinner"], "label")
}

func TestEveryComponentDemoRegistersStructuredAPI(t *testing.T) {
	for _, page := range catalog.ComponentPages() {
		entry, ok := Demos[page.Key]
		require.Truef(t, ok, "catalog component %s must be registered", page.Key)
		require.NotEmptyf(t, entry.API, "%s must register structured API metadata", page.Key)
	}
}

func TestDisplayAPIMetadataRegistered(t *testing.T) {
	expected := map[string][]rootcomponents.Kind{
		"components/accordion": {
			rootcomponents.KindAccordion,
		},
		"components/avatar": {
			rootcomponents.KindAvatar,
			rootcomponents.KindAvatarStack,
		},
		"components/badge": {
			rootcomponents.KindBadge,
			rootcomponents.KindNotificationBadge,
			rootcomponents.KindNotificationDot,
			rootcomponents.KindAnimatingDot,
		},
		"components/banner": {
			rootcomponents.KindBanner,
			rootcomponents.KindCookieBanner,
		},
		"components/card": {
			rootcomponents.KindCard,
		},
		"components/carousel": {
			rootcomponents.KindCarousel,
			rootcomponents.KindCardCarousel,
		},
		"components/chatbubble": {
			rootcomponents.KindChatBubble,
			rootcomponents.KindTypingIndicator,
		},
		"components/codeblock": {
			rootcomponents.KindCodeBlock,
		},
		"components/dependencies": {
			rootcomponents.KindDependencies,
			rootcomponents.KindDependenciesMinimal,
		},
		"components/kbd": {
			rootcomponents.KindKbd,
		},
		"components/table": {
			rootcomponents.KindTable,
			rootcomponents.KindTableHeadContent,
			rootcomponents.KindTableRows,
			rootcomponents.KindTableRow,
			rootcomponents.KindTablePaginationNav,
			rootcomponents.KindTableImageCell,
		},
	}

	require.Len(t, expected, 11)
	displayCatalogEntries := make([]string, 0, len(expected))
	for _, page := range catalog.ComponentPages() {
		if page.Section == "Display" {
			displayCatalogEntries = append(displayCatalogEntries, page.Key)
		}
	}
	require.Len(t, displayCatalogEntries, 11)
	for _, key := range displayCatalogEntries {
		require.Containsf(t, expected, key, "unexpected Display catalog entry %q", key)
	}

	for key, wantKinds := range expected {
		entry, ok := Demos[key]
		require.Truef(t, ok, "missing Display registry entry %q", key)
		require.NotEmptyf(t, entry.API, "%s must register structured API metadata", key)

		seenIDs := make(map[string]struct{}, len(entry.API))
		gotKinds := make([]rootcomponents.Kind, 0, len(entry.API))
		for _, section := range entry.API {
			require.NotEmptyf(t, section.ID, "%s contains an empty API section ID", key)
			_, duplicate := seenIDs[section.ID]
			require.Falsef(t, duplicate, "%s contains duplicate API section ID %q", key, section.ID)
			seenIDs[section.ID] = struct{}{}
			if section.Kind != "" {
				gotKinds = append(gotKinds, section.Kind)
			}
		}
		require.ElementsMatchf(t, wantKinds, gotKinds, "%s API section Kinds", key)
	}
}

func TestAtomicInputAPIMetadataRegistered(t *testing.T) {
	expected := map[string][]rootcomponents.Kind{
		"components/button": {
			rootcomponents.KindButton,
		},
		"components/checkbox": {
			rootcomponents.KindCheckbox,
			rootcomponents.KindCheckboxGroup,
		},
		"components/fileinput": {
			rootcomponents.KindFileInput,
		},
		"components/palette": {
			rootcomponents.KindPalette,
		},
		"components/radio": {
			rootcomponents.KindRadio,
			rootcomponents.KindRadioBar,
			rootcomponents.KindRadioGroup,
		},
		"components/range": {
			rootcomponents.KindRange,
		},
		"components/rating": {
			rootcomponents.KindRating,
			rootcomponents.KindRatingDisplay,
		},
		"components/select": {
			rootcomponents.KindSelect,
		},
		"components/tags-list": {
			rootcomponents.KindTagsList,
		},
		"components/textarea": {
			rootcomponents.KindTextarea,
			rootcomponents.KindTextareaWithActions,
		},
		"components/text-input": {
			rootcomponents.KindTextInput,
		},
		"components/toggle": {
			rootcomponents.KindToggle,
		},
	}

	require.Len(t, expected, 12)
	for key, wantKinds := range expected {
		entry, ok := Demos[key]
		require.Truef(t, ok, "missing atomic Input registry entry %q", key)
		require.NotEmptyf(t, entry.API, "%s must register structured API metadata", key)

		seenIDs := make(map[string]struct{}, len(entry.API))
		gotKinds := make([]rootcomponents.Kind, 0, len(entry.API))
		for _, section := range entry.API {
			require.NotEmptyf(t, section.ID, "%s contains an empty API section ID", key)
			_, duplicate := seenIDs[section.ID]
			require.Falsef(t, duplicate, "%s contains duplicate API section ID %q", key, section.ID)
			seenIDs[section.ID] = struct{}{}
			if section.Kind != "" {
				gotKinds = append(gotKinds, section.Kind)
			}
		}
		require.ElementsMatchf(t, wantKinds, gotKinds, "%s API section Kinds", key)
	}
}

func TestComplexInputAPIMetadataRegistered(t *testing.T) {
	expected := map[string][]rootcomponents.Kind{
		"components/combobox": {
			rootcomponents.KindCombobox,
		},
		"components/form": {
			rootcomponents.KindForm,
			rootcomponents.KindFormSection,
			rootcomponents.KindFormCollapsibleSection,
			rootcomponents.KindFormFlipSection,
			rootcomponents.KindFormSubSection,
			rootcomponents.KindFormFieldGroup,
			rootcomponents.KindFormErrors,
		},
		"components/schema-form": {
			rootcomponents.KindSchemaFormFields,
		},
		"components/search": {
			rootcomponents.KindSearch,
			rootcomponents.KindSearchField,
			rootcomponents.KindSearchModal,
		},
		"components/structured-input": {
			rootcomponents.KindStructuredInput,
		},
	}

	require.Len(t, expected, 5)
	for key, wantKinds := range expected {
		entry, ok := Demos[key]
		require.Truef(t, ok, "missing complex Input registry entry %q", key)
		require.NotEmptyf(t, entry.API, "%s must register structured API metadata", key)

		var catalogKinds []rootcomponents.Kind
		for _, page := range catalog.ComponentPages() {
			if page.Key == key {
				catalogKinds = page.Kinds
				break
			}
		}
		require.Equalf(t, wantKinds, catalogKinds, "%s catalog Kinds", key)

		seenIDs := make(map[string]struct{}, len(entry.API))
		gotKinds := make([]rootcomponents.Kind, 0, len(entry.API))
		for _, section := range entry.API {
			require.NotEmptyf(t, section.ID, "%s contains an empty API section ID", key)
			_, duplicate := seenIDs[section.ID]
			require.Falsef(t, duplicate, "%s contains duplicate API section ID %q", key, section.ID)
			seenIDs[section.ID] = struct{}{}
			if section.Kind != "" {
				gotKinds = append(gotKinds, section.Kind)
			}
		}
		require.ElementsMatchf(t, wantKinds, gotKinds, "%s API section Kinds", key)
	}
}

func TestComplexInputAPIRegistryUsesPageSectionSlices(t *testing.T) {
	expected := map[string][]demo.APISection{
		"components/combobox":         comboboxAPISections,
		"components/form":             formAPISections,
		"components/schema-form":      schemaFormAPISections,
		"components/search":           searchAPISections,
		"components/structured-input": structuredInputAPISections,
	}

	require.Len(t, expected, 5)
	for key, pageSections := range expected {
		entry := Demos[key]
		require.NotEmpty(t, entry.API, key)
		require.NotEmpty(t, pageSections, key)
		require.Samef(
			t,
			&pageSections[0],
			&entry.API[0],
			"%s registry API must use the page's named metadata slice",
			key,
		)

		var rendered strings.Builder
		require.NoError(t, entry.Content().Render(context.Background(), &rendered))
		html := rendered.String()
		require.Equalf(t, 1, strings.Count(html, "data-api-reference"), "%s must render one structured API reference", key)
		for _, section := range pageSections {
			require.Equalf(
				t,
				1,
				strings.Count(html, `data-api-section="`+section.ID+`"`),
				"%s must render its %q API section exactly once",
				key,
				section.ID,
			)
		}
	}
}

func TestComplexInputStructAPIsDocumentEveryExportedFieldExactlyOnce(t *testing.T) {
	expectedTypes := []reflect.Type{
		reflect.TypeFor[combobox.Config](),
		reflect.TypeFor[combobox.Option](),
		reflect.TypeFor[combobox.Source](),
		reflect.TypeFor[combobox.State](),
		reflect.TypeFor[form.Config](),
		reflect.TypeFor[form.HTMXConfig](),
		reflect.TypeFor[form.FooterConfig](),
		reflect.TypeFor[form.CancelHTMXConfig](),
		reflect.TypeFor[form.SectionConfig](),
		reflect.TypeFor[form.CollapsibleSectionConfig](),
		reflect.TypeFor[form.FlipSectionConfig](),
		reflect.TypeFor[form.SubSectionConfig](),
		reflect.TypeFor[form.FieldGroupConfig](),
		reflect.TypeFor[form.FieldMeta](),
		reflect.TypeFor[form.ValidationConfig](),
		reflect.TypeFor[form.FormErrorsConfig](),
		reflect.TypeFor[form.FormErrorItem](),
		reflect.TypeFor[schemaform.FieldsConfig](),
		reflect.TypeFor[schemaform.Field](),
		reflect.TypeFor[search.Config](),
		reflect.TypeFor[search.Item](),
		reflect.TypeFor[structuredinput.Config](),
		reflect.TypeFor[structuredinput.Column](),
		reflect.TypeFor[structuredinput.Option](),
	}

	sectionsByType := make(map[reflect.Type][]demo.APISection, len(expectedTypes))
	for _, sections := range complexInputAPISectionSlices() {
		for _, section := range sections {
			require.NotEmpty(t, section.ID)
			require.NotEmpty(t, section.Title)
			require.NotEmptyf(t, section.Description, "%s section description", section.Title)
			for _, prop := range section.Props {
				require.NotEmptyf(t, prop.Default, "%s.%s default", section.Title, prop.Name)
				require.NotEmptyf(t, prop.Description, "%s.%s description", section.Title, prop.Name)
			}
			if section.StructType != nil {
				sectionsByType[section.StructType] = append(sectionsByType[section.StructType], section)
			}
		}
	}
	require.Len(t, sectionsByType, len(expectedTypes))

	for _, typ := range expectedTypes {
		sections := sectionsByType[typ]
		require.Lenf(t, sections, 1, "%s must have exactly one StructAPI section", typ)
		section := sections[0]

		wantFields := make([]string, 0, typ.NumField())
		for index := range typ.NumField() {
			field := typ.Field(index)
			if field.IsExported() {
				wantFields = append(wantFields, field.Name)
			}
		}
		gotFields := make([]string, 0, len(section.Props))
		for _, prop := range section.Props {
			gotFields = append(gotFields, prop.Name)
		}
		require.ElementsMatchf(t, wantFields, gotFields, "%s exported fields", typ)
		require.Lenf(t, gotFields, len(wantFields), "%s must document every field once", typ)
	}
}

func TestComplexInputMetadataCapturesSourceTruthAndPublicSignatures(t *testing.T) {
	require.True(t, apiProp(t, comboboxAPISections, "Config", "ID").Required)
	require.True(t, apiProp(t, comboboxAPISections, "Config", "Name").Required)
	require.Contains(t, apiProp(t, comboboxAPISections, "Config", "Required").Default, "no rendered effect")
	require.Contains(t, apiProp(t, comboboxAPISections, "Source", "LazyEndpoint").Description, "value is not rendered")
	require.Contains(t, apiProp(t, comboboxAPISections, "State", "Deps").Description, "not read by the renderer")

	require.Equal(t, "nil (prevention enabled)", apiProp(t, formAPISections, "Config", "PreventEnterSubmit").Default)
	require.Equal(t, `"post" when Action is non-empty`, apiProp(t, formAPISections, "Config", "Method").Default)
	require.Contains(t, apiProp(t, formAPISections, "FooterConfig", "CancelHTMX").Description, "does not suppress CancelHref")
	require.Contains(t, apiProp(t, formAPISections, "FieldGroupConfig", "Input").Description, "first non-nil")
	require.Contains(t, apiProp(t, formAPISections, "FieldGroupConfig", "FileInput").Description, "last built-in")

	require.Equal(t, `"values"`, apiProp(t, schemaFormAPISections, "FieldsConfig", "NamePrefix").Default)
	require.Contains(t, apiProp(t, schemaFormAPISections, "Field", "Value").Description, "wins over Default")
	require.Contains(t, apiProp(t, schemaFormAPISections, "Field", "ArrayDefault").Description, "KindArray")

	require.Equal(t, `"search"`, apiProp(t, searchAPISections, "Config", "ID").Default)
	require.Equal(t, "4 when <= 0", apiProp(t, searchAPISections, "Config", "MaxResults").Default)
	require.Equal(t, "120 when <= 0", apiProp(t, searchAPISections, "Config", "DescriptionMaxLength").Default)
	require.Contains(t, apiProp(t, searchAPISections, "Config", "ItemsURL").Description, "instead of Items")
	require.Contains(t, apiProp(t, searchAPISections, "Item", "Attrs").Description, "duplicate attributes")
	require.Contains(t, apiProp(t, searchAPISections, "Item", "Attrs").Description, "revalidated")
	require.Contains(t, apiProp(t, searchAPISections, "Item", "Href").Description, "revalidated client-side")
	require.Contains(t, apiProp(t, searchAPISections, "Item methods", "Item.SafeHref").Description, "revalidates both DOM dataset and client-item navigation")

	require.Equal(t, `"Add row"`, apiProp(t, structuredInputAPISections, "Config", "AddActionLabel").Default)
	require.Contains(t, apiProp(t, structuredInputAPISections, "Config", "Columns").Description, "repeatable row")
	require.Equal(t, "ColumnText when empty", apiProp(t, structuredInputAPISections, "Column", "Type").Default)
	require.Contains(t, apiProp(t, structuredInputAPISections, "Column", "Default").Description, "first option value")
	require.Contains(t, apiProp(t, structuredInputAPISections, "Option", "Label").Description, "falls back to Value")

	expectedSignatures := map[string]string{
		"Config.Validate":       "func (c Config) Validate() error",
		"Config.InitialState":   "func (c Config) InitialState() State",
		"OptionsProvider":       "type OptionsProvider func(ctx context.Context, search string, deps map[string]string) ([]Option, error)",
		"Handler":               "func Handler(cfg Config, provider OptionsProvider) http.Handler",
		"AllowModeManaged":      `const AllowModeManaged AllowMode = "managed"`,
		"AllowModeDisabled":     `const AllowModeDisabled AllowMode = "disabled"`,
		"FlattenAllowList":      "func FlattenAllowList(m *map[string]any) map[string]AllowMode",
		"Walk":                  "func Walk(schema map[string]any, defaults, values map[string]any, allowList map[string]AllowMode) []Field",
		"FallbackFromDefaults":  "func FallbackFromDefaults(defaults, values map[string]any, allowList map[string]AllowMode) []Field",
		"PruneDisabled":         "func PruneDisabled(values map[string]any, allowList map[string]AllowMode) map[string]any",
		"HasOnlySimpleScalars":  "func HasOnlySimpleScalars(fields []Field) bool",
		"Item.SearchText":       "func (item Item) SearchText() string",
		"Item.NormalizedMethod": "func (item Item) NormalizedMethod() string",
		"Item.SafeHref":         "func (item Item) SafeHref() string",
	}
	gotSignatures := make(map[string]string, len(expectedSignatures))
	for _, sections := range complexInputAPISectionSlices() {
		for _, section := range sections {
			for _, prop := range section.Props {
				if prop.Signature == "" {
					continue
				}
				_, duplicate := gotSignatures[prop.Name]
				require.Falsef(t, duplicate, "public signature %s documented more than once", prop.Name)
				gotSignatures[prop.Name] = prop.Signature
			}
		}
	}
	require.Equal(t, expectedSignatures, gotSignatures)
}

func TestComplexInputConstructorsAreDocumentedExactlyOnce(t *testing.T) {
	expected := map[string]rootcomponents.Kind{
		"combobox.Combobox(cfg Config, state State) Instance":                                   rootcomponents.KindCombobox,
		"form.Form(cfg Config) Instance":                                                        rootcomponents.KindForm,
		"form.Section(cfg SectionConfig) SectionInstance":                                       rootcomponents.KindFormSection,
		"form.CollapsibleSection(cfg CollapsibleSectionConfig) CollapsibleSectionInstance":      rootcomponents.KindFormCollapsibleSection,
		"form.FlipSection(cfg FlipSectionConfig, readView templ.Component) FlipSectionInstance": rootcomponents.KindFormFlipSection,
		"form.SubSection(cfg SubSectionConfig) SubSectionInstance":                              rootcomponents.KindFormSubSection,
		"form.FieldGroup(cfg FieldGroupConfig) FieldGroupInstance":                              rootcomponents.KindFormFieldGroup,
		"form.FormErrors(cfg FormErrorsConfig) FormErrorsInstance":                              rootcomponents.KindFormErrors,
		"schemaform.Fields(cfg FieldsConfig) Instance":                                          rootcomponents.KindSchemaFormFields,
		"search.Search(cfg Config) Instance":                                                    rootcomponents.KindSearch,
		"search.SearchField(cfg Config) FieldInstance":                                          rootcomponents.KindSearchField,
		"search.SearchModal(cfg Config) ModalInstance":                                          rootcomponents.KindSearchModal,
		"structuredinput.StructuredInput(cfg Config) Instance":                                  rootcomponents.KindStructuredInput,
	}

	got := make(map[string]rootcomponents.Kind, len(expected))
	for _, sections := range complexInputAPISectionSlices() {
		for _, section := range sections {
			if section.Constructor == "" {
				continue
			}
			_, duplicate := got[section.Constructor]
			require.Falsef(t, duplicate, "constructor %s documented more than once", section.Constructor)
			got[section.Constructor] = section.Kind
		}
	}
	require.Equal(t, expected, got)
}

func TestComplexInputMetadataMatchesRepresentativeRenderBranches(t *testing.T) {
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
	require.Equal(t, optionalCombobox.String(), requiredCombobox.String(), "Config.Required currently has no rendered effect")

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

func TestStructuredInputCatalogCopyDescribesActualDataModel(t *testing.T) {
	for _, page := range catalog.ComponentPages() {
		if page.Key != "components/structured-input" {
			continue
		}
		require.Contains(t, page.Description, "repeatable")
		require.Contains(t, page.Description, "typed columns")
		require.NotContains(t, page.Description, "prefixes")
		require.NotContains(t, page.Description, "suffixes")
		require.NotContains(t, page.Description, "segmented")
		return
	}
	t.Fatal("structured-input catalog entry not found")
}

func TestAtomicInputAPIRegistryUsesPageSectionSlices(t *testing.T) {
	expected := map[string][]demo.APISection{
		"components/button":     buttonAPISections,
		"components/checkbox":   checkboxAPISections,
		"components/fileinput":  fileInputAPISections,
		"components/palette":    paletteAPISections,
		"components/radio":      radioAPISections,
		"components/range":      rangeAPISections,
		"components/rating":     ratingAPISections,
		"components/select":     selectAPISections,
		"components/tags-list":  tagsListAPISections,
		"components/textarea":   textareaAPISections,
		"components/text-input": textInputAPISections,
		"components/toggle":     toggleAPISections,
	}

	require.Len(t, expected, 12)
	for key, pageSections := range expected {
		entry := Demos[key]
		require.NotEmpty(t, entry.API, key)
		require.NotEmpty(t, pageSections, key)
		require.Samef(
			t,
			&pageSections[0],
			&entry.API[0],
			"%s registry API must use the page's named metadata slice",
			key,
		)

		var rendered strings.Builder
		require.NoError(t, entry.Content().Render(context.Background(), &rendered))
		html := rendered.String()
		require.Equalf(t, 1, strings.Count(html, "data-api-reference"), "%s must render one structured API reference", key)
		for _, section := range pageSections {
			require.Equalf(
				t,
				1,
				strings.Count(html, `data-api-section="`+section.ID+`"`),
				"%s must render its %q API section exactly once",
				key,
				section.ID,
			)
		}
	}
}

func TestAtomicInputStructAPIsDocumentEveryExportedFieldExactlyOnce(t *testing.T) {
	expectedTypes := []reflect.Type{
		reflect.TypeFor[button.HTMXConfig](),
		reflect.TypeFor[button.AlpineConfig](),
		reflect.TypeFor[checkbox.Config](),
		reflect.TypeFor[checkbox.GroupConfig](),
		reflect.TypeFor[fileinput.Config](),
		reflect.TypeFor[palette.Config](),
		reflect.TypeFor[palette.AlpineConfig](),
		reflect.TypeFor[radio.Config](),
		reflect.TypeFor[radio.GroupConfig](),
		reflect.TypeFor[radio.HTMXConfig](),
		reflect.TypeFor[radio.AlpineConfig](),
		reflect.TypeFor[rangeinput.Config](),
		reflect.TypeFor[rangeinput.Tick](),
		reflect.TypeFor[rating.Config](),
		reflect.TypeFor[rating.DisplayConfig](),
		reflect.TypeFor[selectfield.Config](),
		reflect.TypeFor[selectfield.Option](),
		reflect.TypeFor[selectfield.AlpineConfig](),
		reflect.TypeFor[tagslist.Config](),
		reflect.TypeFor[textarea.Config](),
		reflect.TypeFor[textinput.Config](),
		reflect.TypeFor[toggle.Config](),
	}

	sectionsByType := make(map[reflect.Type][]demo.APISection, len(expectedTypes))
	for _, sections := range atomicInputAPISectionSlices() {
		for _, section := range sections {
			require.NotEmpty(t, section.ID)
			require.NotEmpty(t, section.Title)
			require.NotEmptyf(t, section.Description, "%s section description", section.Title)
			for _, prop := range section.Props {
				require.NotEmptyf(t, prop.Default, "%s.%s default", section.Title, prop.Name)
				require.NotEmptyf(t, prop.Description, "%s.%s description", section.Title, prop.Name)
			}
			if section.StructType != nil {
				sectionsByType[section.StructType] = append(sectionsByType[section.StructType], section)
			}
		}
	}
	require.Len(t, sectionsByType, len(expectedTypes))

	for _, typ := range expectedTypes {
		sections := sectionsByType[typ]
		require.Lenf(t, sections, 1, "%s must have exactly one StructAPI section", typ)
		section := sections[0]

		wantFields := make([]string, 0, typ.NumField())
		for index := range typ.NumField() {
			field := typ.Field(index)
			if field.IsExported() {
				wantFields = append(wantFields, field.Name)
			}
		}
		gotFields := make([]string, 0, len(section.Props))
		for _, prop := range section.Props {
			gotFields = append(gotFields, prop.Name)
		}
		require.ElementsMatchf(t, wantFields, gotFields, "%s exported fields", typ)
		require.Lenf(t, gotFields, len(wantFields), "%s must document every field once", typ)
	}
}

func TestAtomicInputMetadataCapturesRenderedDefaultsAndPublicSignatures(t *testing.T) {
	require.Equal(t, "3 when 0 or outside 1..10", apiProp(t, textareaAPISections, "Config", "Rows").Default)
	require.Equal(t, "nil", apiProp(t, textareaAPISections, "Config", "InputAttrs").Default)
	require.Contains(t, apiProp(t, textareaAPISections, "Config", "HelperText").Description, "neither renderer adds a counter")
	require.Equal(t, `"Please Select"`, apiProp(t, selectAPISections, "Config", "Placeholder").Default)
	require.Contains(t, apiProp(t, selectAPISections, "Config", "InputAttrs").Default, "not rendered")
	require.Equal(t, `"Add a tag..."`, apiProp(t, tagsListAPISections, "Config", "Placeholder").Default)
	require.Contains(t, apiProp(t, tagsListAPISections, "Config", "Values").Description, "duplicate")
	require.Equal(t, `"" (checkbox uses the browser default value)`, apiProp(t, toggleAPISections, "Config", "Value").Default)
	require.Equal(t, "nil", apiProp(t, toggleAPISections, "Config", "InputAttrs").Default)
	require.Contains(t, apiProp(t, fileInputAPISections, "Config", "Accept").Description, "does not validate")
	require.Contains(t, apiProp(t, paletteAPISections, "Config", "HideNeutral").Description, "white and black")
	require.Equal(t, `"change" when a verb is set; otherwise omitted`, apiProp(t, radioAPISections, "HTMXConfig", "Trigger").Default)
	require.Equal(t, "21 generated ticks from Min to Max", apiProp(t, rangeAPISections, "Config", "Ticks").Default)
	require.Equal(t, `"0 stars" aria-label; "Rating" when visibly shown`, apiProp(t, ratingAPISections, "DisplayConfig", "Label").Default)
	require.Contains(t, apiProp(t, textInputAPISections, "Config", "Type").Allowed, "TypeDateTimeLocal")

	expectedSignatures := map[string]string{
		"WithTone":            "func WithTone(tone Tone) Option",
		"WithSize":            "func WithSize(size Size) Option",
		"WithType":            "func WithType(buttonType string) Option",
		"Disabled":            "func Disabled() Option",
		"WithID":              "func WithID(id string) Option",
		"WithRootClass":       "func WithRootClass(class string) Option",
		"WithHTMX":            "func WithHTMX(htmx *HTMXConfig) Option",
		"WithAlpine":          "func WithAlpine(alpine *AlpineConfig) Option",
		"WithLoadingText":     "func WithLoadingText(text string) Option",
		"RadioBar":            "func RadioBar() BarInstance",
		"TextareaWithActions": "func TextareaWithActions(cfg Config) WithActionsInstance",
	}
	gotSignatures := make(map[string]string, len(expectedSignatures))
	for _, sections := range atomicInputAPISectionSlices() {
		for _, section := range sections {
			for _, prop := range section.Props {
				if prop.Signature != "" {
					_, duplicate := gotSignatures[prop.Name]
					require.Falsef(t, duplicate, "public signature %s documented more than once", prop.Name)
					gotSignatures[prop.Name] = prop.Signature
				}
			}
		}
	}
	require.Equal(t, expectedSignatures, gotSignatures)
}

func TestRadioMetadataDistinguishesRadioGroupApplicability(t *testing.T) {
	require.Equal(
		t,
		"Normal standalone Radio with non-empty HelperText renders aria-describedby on the input and the matching helper span ID. Segmented Radio takes precedence and renders neither the typed aria-describedby attribute nor a helper span. RadioGroup renders aria-describedby on the input but does not give its helper span a matching ID.",
		apiProp(t, radioAPISections, "Config", "HelperTextID").Description,
	)
	require.Equal(
		t,
		"Selects the bordered wrapper for a non-segmented standalone Radio. Segmented Radio ignores it. RadioGroup ignores the wrapper layout, but Container still changes the grouped item's inner input background through shared input classes.",
		apiProp(t, radioAPISections, "Config", "Container").Description,
	)
	require.Equal(
		t,
		"Selects the standalone segmented-pill renderer, intended for RadioBar; RadioGroup ignores it and renders its own standard item path.",
		apiProp(t, radioAPISections, "Config", "Segmented").Description,
	)
	require.Equal(
		t,
		"Appends CSS classes to a standalone Radio root; RadioGroup ignores it.",
		apiProp(t, radioAPISections, "Config", "RootClass").Description,
	)
}

func TestRadioInputAttrsMetadataDoesNotPromiseOverrides(t *testing.T) {
	description := apiProp(t, radioAPISections, "Config", "InputAttrs").Description
	require.Equal(
		t,
		"Additional non-conflicting attributes appended to the native input. Conflicts with modeled keys create duplicate attributes rather than overriding them; use typed fields for name, value, disabled, checked, HTMX, and Alpine behavior.",
		description,
	)
	require.NotContains(t, description, "able to override")
}

func TestDisplayAPIRegistryUsesPageSectionSlices(t *testing.T) {
	expected := map[string][]demo.APISection{
		"components/accordion":    accordionAPISections,
		"components/avatar":       avatarAPISections,
		"components/badge":        badgeAPISections,
		"components/banner":       bannerAPISections,
		"components/card":         cardAPISections,
		"components/carousel":     carouselAPISections,
		"components/chatbubble":   chatBubbleAPISections,
		"components/codeblock":    codeBlockAPISections,
		"components/dependencies": dependenciesAPISections,
		"components/kbd":          kbdAPISections,
		"components/table":        tableAPISections,
	}

	require.Len(t, expected, 11)
	for key, pageSections := range expected {
		entry := Demos[key]
		require.NotEmpty(t, entry.API, key)
		require.NotEmpty(t, pageSections, key)
		require.Samef(
			t,
			&pageSections[0],
			&entry.API[0],
			"%s registry API must use the page's named metadata slice",
			key,
		)

		var rendered strings.Builder
		require.NoError(t, entry.Content().Render(context.Background(), &rendered))
		html := rendered.String()
		require.Equalf(t, 1, strings.Count(html, "data-api-reference"), "%s must render one structured API reference", key)
		for _, section := range pageSections {
			require.Equalf(
				t,
				1,
				strings.Count(html, `data-api-section="`+section.ID+`"`),
				"%s must render its %q API section exactly once",
				key,
				section.ID,
			)
		}
	}
}

func TestDisplayStructAPIsDocumentEveryExportedFieldExactlyOnce(t *testing.T) {
	expectedTypes := []reflect.Type{
		reflect.TypeFor[accordion.AccordionConfig](),
		reflect.TypeFor[accordion.AccordionItem](),
		reflect.TypeFor[avatar.Config](),
		reflect.TypeFor[avatar.StackConfig](),
		reflect.TypeFor[badge.Config](),
		reflect.TypeFor[banner.Config](),
		reflect.TypeFor[banner.CTAConfig](),
		reflect.TypeFor[banner.CookieBannerConfig](),
		reflect.TypeFor[card.Config](),
		reflect.TypeFor[carousel.Config](),
		reflect.TypeFor[carousel.CardConfig](),
		reflect.TypeFor[carousel.Slide](),
		reflect.TypeFor[carousel.AutoplayConfig](),
		reflect.TypeFor[carousel.HTMXConfig](),
		reflect.TypeFor[chatbubble.Config](),
		reflect.TypeFor[codeblock.Config](),
		reflect.TypeFor[table.Config](),
		reflect.TypeFor[table.Column](),
		reflect.TypeFor[table.Cell](),
		reflect.TypeFor[table.Row](),
		reflect.TypeFor[table.RowHTMXConfig](),
		reflect.TypeFor[table.PaginationConfig](),
		reflect.TypeFor[table.InfiniteScrollConfig](),
		reflect.TypeFor[table.FilterConfig](),
		reflect.TypeFor[table.Filter](),
		reflect.TypeFor[table.FilterOption](),
		reflect.TypeFor[table.FilterOptionsHTMXConfig](),
		reflect.TypeFor[table.FilterHTMXConfig](),
		reflect.TypeFor[table.HTMXConfig](),
	}

	sectionsByType := make(map[reflect.Type][]demo.APISection, len(expectedTypes))
	for _, sections := range displayAPISectionSlices() {
		for _, section := range sections {
			require.NotEmpty(t, section.ID)
			require.NotEmpty(t, section.Title)
			require.NotEmptyf(t, section.Description, "%s section description", section.Title)
			for _, prop := range section.Props {
				require.NotEmptyf(t, prop.Default, "%s.%s default", section.Title, prop.Name)
				require.NotEmptyf(t, prop.Description, "%s.%s description", section.Title, prop.Name)
			}
			if section.StructType != nil {
				sectionsByType[section.StructType] = append(sectionsByType[section.StructType], section)
			}
		}
	}
	require.Len(t, sectionsByType, len(expectedTypes))

	for _, typ := range expectedTypes {
		sections := sectionsByType[typ]
		require.Lenf(t, sections, 1, "%s must have exactly one StructAPI section", typ)
		section := sections[0]

		wantFields := make([]string, 0, typ.NumField())
		for index := range typ.NumField() {
			field := typ.Field(index)
			if field.IsExported() {
				wantFields = append(wantFields, field.Name)
			}
		}
		gotFields := make([]string, 0, len(section.Props))
		for _, prop := range section.Props {
			gotFields = append(gotFields, prop.Name)
		}
		require.ElementsMatchf(t, wantFields, gotFields, "%s exported fields", typ)
		require.Lenf(t, gotFields, len(wantFields), "%s must document every field once", typ)
	}
}

func TestDisplayMetadataCapturesRenderedDefaultsAndPublicSignatures(t *testing.T) {
	require.Equal(t, `"table"`, apiProp(t, tableAPISections, "Config", "ID").Default)
	require.Equal(t, `"load"`, apiProp(t, tableAPISections, "Config", "LazyTrigger").Default)
	require.Equal(t, []string{"AppearanceDefault", "AppearanceStriped"}, apiProp(t, tableAPISections, "Config", "Appearance").Allowed)
	require.Contains(t, apiProp(t, tableAPISections, "Cell", "BadgeColor").Description, "badge.Badge")
	require.Equal(t, `"innerHTML"`, apiProp(t, tableAPISections, "RowHTMXConfig", "Swap").Default)
	require.Equal(t, "false when collapsible; otherwise visible", apiProp(t, tableAPISections, "FilterConfig", "InitiallyExpanded").Default)
	require.Equal(t, `"#<table-id>-tbody"`, apiProp(t, tableAPISections, "FilterHTMXConfig", "Target").Default)
	require.Equal(t, `"innerHTML"`, apiProp(t, tableAPISections, "FilterHTMXConfig", "Swap").Default)

	require.Equal(
		t,
		"func (cfg Config) NextSortDir(key string) SortDir",
		apiProp(t, tableAPISections, "Config helpers", "NextSortDir").Signature,
	)
	require.Equal(
		t,
		"func ImageCell(imageURL string, label string, detail string) ImageCellInstance",
		apiProp(t, tableAPISections, "ImageCell", "ImageCell").Signature,
	)

	require.Equal(t, "4000ms when enabled", apiProp(t, carouselAPISections, "AutoplayConfig", "Interval").Default)
	require.Equal(t, `"load"`, apiProp(t, carouselAPISections, "HTMXConfig", "Trigger").Default)
	require.Equal(t, `"innerHTML"`, apiProp(t, carouselAPISections, "HTMXConfig", "Swap").Default)

	require.Equal(t, `"Cookie Consent"`, apiProp(t, bannerAPISections, "CookieBannerConfig", "Title").Default)
	require.Equal(t, `"Accept"`, apiProp(t, bannerAPISections, "CookieBannerConfig", "AcceptLabel").Default)
	require.Equal(t, `"Decline"`, apiProp(t, bannerAPISections, "CookieBannerConfig", "RejectLabel").Default)

	require.Equal(
		t,
		"func NotificationBadge(count int) NotificationBadgeInstance",
		apiProp(t, badgeAPISections, "NotificationBadge", "NotificationBadge").Signature,
	)
	require.Equal(t, "count <= 0 renders nothing; counts > 99 render 99+", apiProp(t, badgeAPISections, "NotificationBadge", "NotificationBadge").Default)

	require.Equal(t, "SizeMD", apiProp(t, kbdAPISections, "Kbd options", "WithSize").Default)
	require.Equal(
		t,
		"func WithAttrs(attrs templ.Attributes) Option",
		apiProp(t, kbdAPISections, "Kbd options", "WithAttrs").Signature,
	)
}

func TestAccordionMetadataExplainsItemIDNamespace(t *testing.T) {
	configID := apiProp(t, accordionAPISections, "AccordionConfig", "ID")
	require.Contains(t, configID.Description, "root element")
	require.Contains(t, configID.Description, "does not namespace")

	itemID := apiProp(t, accordionAPISections, "AccordionItem", "ID")
	require.Contains(t, itemID.Description, "namespace source")
	require.Contains(t, itemID.Description, "multiple accordions")
}

func TestCarouselMetadataMarksStaticOnlyFields(t *testing.T) {
	for _, name := range []string{"Autoplay", "Touch", "AspectRatio", "Height"} {
		description := apiProp(t, carouselAPISections, "Config", name).Description
		require.Containsf(t, description, "static mode", "Config.%s", name)
		require.Containsf(t, description, "ignored when HTMX is non-nil", "Config.%s", name)
	}
}

func TestTableMetadataDocumentsConditionalOptionalFields(t *testing.T) {
	totalPages := apiProp(t, tableAPISections, "PaginationConfig", "TotalPages")
	require.False(t, totalPages.Required)
	require.Equal(t, "0 (no numbered navigation)", totalPages.Default)
	require.Contains(t, totalPages.Description, "traditional mode")
	require.Contains(t, totalPages.Description, "ignored in infinite-scroll mode")

	optionValue := apiProp(t, tableAPISections, "FilterOption", "Value")
	require.False(t, optionValue.Required)
	require.Equal(t, `""`, optionValue.Default)
	require.Contains(t, optionValue.Description, "empty string is valid")
}

func displayAPISectionSlices() [][]demo.APISection {
	return [][]demo.APISection{
		accordionAPISections,
		avatarAPISections,
		badgeAPISections,
		bannerAPISections,
		cardAPISections,
		carouselAPISections,
		chatBubbleAPISections,
		codeBlockAPISections,
		dependenciesAPISections,
		kbdAPISections,
		tableAPISections,
	}
}

func atomicInputAPISectionSlices() [][]demo.APISection {
	return [][]demo.APISection{
		buttonAPISections,
		checkboxAPISections,
		fileInputAPISections,
		paletteAPISections,
		radioAPISections,
		rangeAPISections,
		ratingAPISections,
		selectAPISections,
		tagsListAPISections,
		textareaAPISections,
		textInputAPISections,
		toggleAPISections,
	}
}

func complexInputAPISectionSlices() [][]demo.APISection {
	return [][]demo.APISection{
		comboboxAPISections,
		formAPISections,
		schemaFormAPISections,
		searchAPISections,
		structuredInputAPISections,
	}
}

func feedbackNavigationAPISectionSlices() [][]demo.APISection {
	return [][]demo.APISection{
		alertAPISections,
		toastAPISections,
		modalAPISections,
		drawerAPISections,
		spinnerAPISections,
		stepsAPISections,
		tooltipAPISections,
		breadcrumbsAPISections,
		dropdownAPISections,
		linkAPISections,
		navbarAPISections,
		paginationAPISections,
		sidebarAPISections,
		tabsAPISections,
	}
}

func isComponentNamedScalar(typ reflect.Type) bool {
	if typ.Name() == "" ||
		!strings.HasPrefix(typ.PkgPath(), "github.com/araihu/goshtoso/components/") {
		return false
	}
	switch typ.Kind() {
	case reflect.String,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	default:
		return false
	}
}

func apiProp(
	t *testing.T,
	sections []demo.APISection,
	sectionTitle string,
	propName string,
) demo.APIPropDoc {
	t.Helper()
	for _, section := range sections {
		if section.Title != sectionTitle {
			continue
		}
		for _, prop := range section.Props {
			if prop.Name == propName {
				return prop
			}
		}
		require.FailNowf(t, "missing API prop", "%s.%s", sectionTitle, propName)
	}
	require.FailNowf(t, "missing API section", "%s", sectionTitle)
	return demo.APIPropDoc{}
}
