package splitbutton

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/components"
	"github.com/araihu/goshtoso/components/dropdown"
	"github.com/stretchr/testify/require"
)

func testIcon(label string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := io.WriteString(w, `<svg data-icon="`+label+`" aria-hidden="true"></svg>`)
		return err
	})
}

func renderSplitButton(t *testing.T, cfg Config) string {
	t.Helper()
	var output bytes.Buffer
	require.NoError(t, SplitButton(cfg).Render(context.Background(), &output))
	return output.String()
}

func TestSplitButtonComposesPrimaryActionAndMenu(t *testing.T) {
	html := renderSplitButton(t, Config{
		ID: "page-actions",
		Primary: Action{
			Label: "Publish",
			Href:  "/publish",
			Icon:  testIcon("publish"),
		},
		MenuLabel: "More actions",
		Sections: []dropdown.Section{{Items: []dropdown.Item{{
			Label:   "Open docs",
			Caption: "Open in another tab",
			Href:    "/docs",
			Target:  "_blank",
			Rel:     "noopener noreferrer",
			Icon:    testIcon("docs"),
		}}}},
	})

	for _, want := range []string{
		`href="/publish"`,
		`data-icon="publish"`,
		`aria-label="More actions"`,
		`target="_blank"`,
		`rel="noopener noreferrer"`,
		`data-icon="docs"`,
		`data-popover-root`,
		`rounded-radius`,
		`data-split-button-divider`,
		`w-px shrink-0 self-stretch bg-outline dark:bg-outline-dark`,
		`rounded-r-none`,
		`border-l-0`,
		`rounded-l-none`,
	} {
		require.Contains(t, html, want)
	}
	require.NotContains(t, html, "overflow-clip")
}

func TestSplitButtonIdentity(t *testing.T) {
	require.Equal(t, components.KindSplitButton, SplitButton(Config{}).Kind())
}
