package demo

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/components"
	"github.com/stretchr/testify/require"
)

type sampleConfig struct {
	Label string
	Count int
}

func TestStructAPIRecordsReflectedType(t *testing.T) {
	section := StructAPI[sampleConfig](
		components.KindButton,
		"Config",
		"sample.New",
		"Configures the sample.",
		[]APIPropDoc{
			{Name: "Label", Default: `""`, Description: "Visible label.", Required: true},
			{Name: "Count", Default: "0", Allowed: []string{"0", "1"}, Description: "Visible count."},
		},
	)

	require.Equal(t, "config", section.ID)
	require.Equal(t, reflect.TypeFor[sampleConfig](), section.StructType)
	require.Equal(t, components.KindButton, section.Kind)
	require.Equal(t, "sample.New", section.Constructor)
}

func TestStructuredAPIHelpersEnforceSignatureMode(t *testing.T) {
	require.Panics(t, func() {
		StructAPI[sampleConfig](
			components.KindButton,
			"Config",
			"sample.New",
			"",
			[]APIPropDoc{{Name: "Label", Signature: "string"}},
		)
	})

	for _, build := range []func() APISection{
		func() APISection {
			return OptionsAPI(
				components.KindButton,
				"Options",
				"button.Button",
				"",
				[]APIPropDoc{{Name: "WithTone"}},
			)
		},
		func() APISection {
			return FunctionsAPI(
				components.KindBadge,
				"Functions",
				"",
				"",
				[]APIPropDoc{{Name: "NotificationDot"}},
			)
		},
	} {
		require.Panics(t, func() {
			build()
		})
	}

	options := OptionsAPI(
		components.KindButton,
		"Options",
		"button.Button",
		"Functional options.",
		[]APIPropDoc{{Name: "WithTone", Signature: "func(Tone) Option"}},
	)
	functions := FunctionsAPI(
		components.KindBadge,
		"Functions",
		"",
		"Additional constructors.",
		[]APIPropDoc{{Name: "NotificationDot", Signature: "func() components.Component"}},
	)
	require.Nil(t, options.StructType)
	require.Nil(t, functions.StructType)
}

func TestStructuredAPIReferenceRendersExactTypesAndStableHooks(t *testing.T) {
	sections := []APISection{
		StructAPI[sampleConfig](
			components.KindButton,
			"Config",
			"sample.New",
			"Configures the sample.",
			[]APIPropDoc{
				{Name: "Label", Default: `""`, Description: "Visible label.", Required: true},
				{Name: "Count", Default: "0", Allowed: []string{"0", "1"}, Description: "Visible count."},
			},
		),
		FunctionsAPI(
			components.KindBadge,
			"Functions",
			"",
			"Additional constructors.",
			[]APIPropDoc{{Name: "NotificationDot", Signature: "func() components.Component", Default: "n/a", Description: "Renders a dot."}},
		),
	}

	var out strings.Builder
	require.NoError(t, StructuredAPIReference(sections).Render(context.Background(), &out))
	html := out.String()

	require.Contains(t, html, "data-api-reference")
	require.Contains(t, html, `data-api-section="config"`)
	require.Contains(t, html, `data-api-section="functions"`)
	require.Contains(t, html, ">string<")
	require.Contains(t, html, ">int<")
	require.Contains(t, html, "func() components.Component")
	require.Contains(t, html, "Required")
	require.Contains(t, html, "Allowed")
	require.Contains(t, html, "0, 1")
	require.Contains(t, html, "sample.New")
	require.Contains(t, html, string(components.KindButton))
}

func TestStructuredAPIReferenceMakesCollisionChainAnchorsGloballyUnique(t *testing.T) {
	sections := []APISection{
		{Title: "Functions"},
		{Title: "Functions 2"},
		{Title: "Functions"},
		{ID: "x", Title: "X"},
		{ID: "x-2", Title: "X 2"},
		{ID: "x-3", Title: "X 3"},
		{ID: "x", Title: "X again"},
		{ID: "repeat", Title: "Repeat"},
		{ID: "repeat", Title: "Repeat again"},
		{ID: "repeat", Title: "Repeat once more"},
	}
	originalIDs := []string{"", "", "", "x", "x-2", "x-3", "x", "repeat", "repeat", "repeat"}
	expectedIDs := []string{
		"functions",
		"functions-2",
		"functions-3",
		"x",
		"x-2",
		"x-3",
		"x-4",
		"repeat",
		"repeat-2",
		"repeat-3",
	}

	var out strings.Builder
	require.NoError(t, StructuredAPIReference(sections).Render(context.Background(), &out))
	html := out.String()

	remaining := html
	for _, id := range expectedIDs {
		sectionHook := `data-api-section="` + id + `"`
		headingID := `id="api-` + id + `"`
		require.Equal(t, 1, strings.Count(html, sectionHook), "section hook %q must be unique", id)
		require.Equal(t, 1, strings.Count(html, headingID), "heading ID %q must be unique", id)

		index := strings.Index(remaining, sectionHook)
		require.NotEqualf(t, -1, index, "section %q must preserve input order", id)
		remaining = remaining[index+len(sectionHook):]
	}

	for i := range sections {
		require.Equal(t, originalIDs[i], sections[i].ID, "renderer must not mutate caller section %d", i)
	}

	var second strings.Builder
	require.NoError(t, StructuredAPIReference(sections).Render(context.Background(), &second))
	require.Equal(t, html, second.String(), "anchor normalization must be deterministic")
}

func TestDemoRenderersExposeStableDocumentationHooks(t *testing.T) {
	component := ComponentDemo(
		ComponentDemoProps{
			Title:       "Sample",
			Description: "Sample description.",
		},
		templ.Raw("preview"),
		"sample.New()",
	)

	var componentOut strings.Builder
	require.NoError(t, component.Render(context.Background(), &componentOut))
	componentHTML := componentOut.String()
	require.Contains(t, componentHTML, "data-component-description")
	require.Contains(t, componentHTML, "data-demo-section")
	require.Contains(t, componentHTML, "data-demo-preview")
	require.Contains(t, componentHTML, "data-demo-code")
	require.NotContains(t, componentHTML, "data-api-reference")

	var sectionOut strings.Builder
	require.NoError(t, DemoSection(
		DemoSectionProps{Title: "Secondary", Description: "Secondary description."},
		templ.Raw("preview"),
		"sample.Secondary()",
	).Render(context.Background(), &sectionOut))
	sectionHTML := sectionOut.String()
	require.Contains(t, sectionHTML, "data-demo-section")
	require.Contains(t, sectionHTML, "data-demo-preview")
	require.Contains(t, sectionHTML, "data-demo-code")
}
