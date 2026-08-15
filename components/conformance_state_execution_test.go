package components_test

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/components/banner"
	"github.com/araihu/goshtoso/components/card"
	"github.com/araihu/goshtoso/components/chatbubble"
	"github.com/araihu/goshtoso/components/checkbox"
	"github.com/araihu/goshtoso/components/codeblock"
	"github.com/araihu/goshtoso/components/drawer"
	"github.com/araihu/goshtoso/components/dropdown"
	"github.com/araihu/goshtoso/components/head"
	"github.com/araihu/goshtoso/components/icon"
	"github.com/araihu/goshtoso/components/link"
	"github.com/araihu/goshtoso/components/panel"
	"github.com/araihu/goshtoso/components/search"
	"github.com/araihu/goshtoso/components/skeleton"
	"github.com/araihu/goshtoso/components/table"
	"github.com/araihu/goshtoso/components/textinput"
	"github.com/araihu/goshtoso/components/tooltip"
	"github.com/stretchr/testify/require"
)

// TestConformanceLedgerPreviouslyImplicitStates executes every source-derived
// public state whose named constant was previously only reached through an
// implicit default. Explicit defaults must render byte-identically to the zero
// value; distinct states must expose the expected rendered semantic hook.
func TestConformanceLedgerPreviouslyImplicitStates(t *testing.T) {
	t.Run("banner PositionRelative", func(t *testing.T) {
		explicit := renderState(t, banner.Banner(banner.Config{Description: "notice", Position: banner.PositionRelative}))
		implicit := renderState(t, banner.Banner(banner.Config{Description: "notice"}))
		require.Equal(t, implicit, explicit)
		require.NotContains(t, explicit, "fixed top-0")
	})
	t.Run("card InteractionDefault", func(t *testing.T) {
		explicit := renderState(t, card.Card(card.Config{Title: "Card", Interaction: card.InteractionDefault}))
		implicit := renderState(t, card.Card(card.Config{Title: "Card"}))
		require.Equal(t, implicit, explicit)
		require.NotContains(t, explicit, "hover:translate-y-1.5")
	})

	for _, state := range []struct {
		name string
		got  chatbubble.Status
		want string
	}{
		{name: "StatusDelivered", got: chatbubble.StatusDelivered, want: "Delivered"},
		{name: "StatusSending", got: chatbubble.StatusSending, want: "Sending"},
	} {
		t.Run("chatbubble "+state.name, func(t *testing.T) {
			html := renderState(t, chatbubble.ChatBubble(chatbubble.Config{Message: "hello", Status: state.got}))
			require.Contains(t, html, state.want)
		})
	}
	t.Run("chatbubble StatusNone", func(t *testing.T) {
		explicit := renderState(t, chatbubble.ChatBubble(chatbubble.Config{Message: "hello", Status: chatbubble.StatusNone}))
		implicit := renderState(t, chatbubble.ChatBubble(chatbubble.Config{Message: "hello"}))
		require.Equal(t, implicit, explicit)
		require.NotContains(t, explicit, "Sending")
		require.NotContains(t, explicit, "Delivered")
	})

	t.Run("checkbox AnimationNone", func(t *testing.T) {
		explicit := renderState(t, checkbox.Checkbox(checkbox.Config{ID: "check", Label: "Check", Animation: checkbox.AnimationNone}))
		implicit := renderState(t, checkbox.Checkbox(checkbox.Config{ID: "check", Label: "Check"}))
		require.Equal(t, implicit, explicit)
		require.NotContains(t, explicit, "origin-bottom")
	})
	t.Run("checkbox IconCheck", func(t *testing.T) {
		explicit := renderState(t, checkbox.Checkbox(checkbox.Config{ID: "check", Label: "Check", Icon: checkbox.IconCheck}))
		implicit := renderState(t, checkbox.Checkbox(checkbox.Config{ID: "check", Label: "Check"}))
		require.Equal(t, implicit, explicit)
		require.Contains(t, explicit, "M4.5 12.75l6 6 9-13.5")
	})
	t.Run("codeblock DensityDefault", func(t *testing.T) {
		explicit := renderState(t, codeblock.CodeBlock(codeblock.Config{Language: "go", Code: "package p", Density: codeblock.DensityDefault}))
		implicit := renderState(t, codeblock.CodeBlock(codeblock.Config{Language: "go", Code: "package p"}))
		require.Equal(t, implicit, explicit)
		require.NotContains(t, explicit, "codeblock-compact")
	})

	for _, state := range []struct {
		name string
		got  drawer.Height
		want string
	}{
		{name: "HeightSM", got: drawer.HeightSM, want: "max-h-[320px]"},
		{name: "HeightMD", got: drawer.HeightMD, want: "max-h-[420px]"},
		{name: "HeightXL", got: drawer.HeightXL, want: "max-h-[720px]"},
		{name: "HeightFull", got: drawer.HeightFull, want: "max-h-full md:max-h-[90vh]"},
	} {
		t.Run("drawer "+state.name, func(t *testing.T) {
			html := renderState(t, drawer.Drawer(drawer.Config{ID: "state-drawer", Title: "State", Side: drawer.SideTop, Height: state.got}))
			require.Contains(t, html, state.want)
		})
	}
	t.Run("drawer WidthMD", func(t *testing.T) {
		explicit := renderState(t, drawer.Drawer(drawer.Config{ID: "state-drawer", Title: "State", Width: drawer.WidthMD}))
		implicit := renderState(t, drawer.Drawer(drawer.Config{ID: "state-drawer", Title: "State"}))
		require.Equal(t, implicit, explicit)
		require.Contains(t, explicit, "max-w-[420px]")
	})

	t.Run("dropdown AlignStart", func(t *testing.T) {
		explicit := renderState(t, dropdown.Dropdown(dropdown.Config{ID: "menu", Label: "Menu", MenuAlign: dropdown.AlignStart}))
		implicit := renderState(t, dropdown.Dropdown(dropdown.Config{ID: "menu", Label: "Menu"}))
		require.Equal(t, implicit, explicit)
		require.Contains(t, explicit, "left-0")
	})
	t.Run("dropdown TriggerClick", func(t *testing.T) {
		explicit := renderState(t, dropdown.Dropdown(dropdown.Config{ID: "menu", Label: "Menu", TriggerMode: dropdown.TriggerClick}))
		implicit := renderState(t, dropdown.Dropdown(dropdown.Config{ID: "menu", Label: "Menu"}))
		require.Equal(t, implicit, explicit)
		require.Contains(t, explicit, `x-on:click="isOpen = ! isOpen"`)
	})

	t.Run("head OpenGraphTypeWebsite", func(t *testing.T) {
		config := metadataStateConfig()
		config.OpenGraphType = head.OpenGraphTypeWebsite
		explicit := renderState(t, head.Metadata(config))
		config.OpenGraphType = ""
		implicit := renderState(t, head.Metadata(config))
		require.Equal(t, implicit, explicit)
		require.Contains(t, explicit, `property="og:type" content="website"`)
	})
	t.Run("icon ModeExternal", func(t *testing.T) {
		config := icon.Config{SpriteURL: "/assets/icons.svg", Symbol: icon.Symbol("check"), Label: "Check", Mode: icon.ModeExternal}
		explicit := renderState(t, icon.Icon(config))
		config.Mode = ""
		implicit := renderState(t, icon.Icon(config))
		require.Equal(t, implicit, explicit)
		require.Contains(t, explicit, `/assets/icons.svg#check`)
	})
	t.Run("link AppearanceText", func(t *testing.T) {
		explicit := renderState(t, link.Link("/docs", link.WithAppearance(link.AppearanceText)))
		implicit := renderState(t, link.Link("/docs"))
		require.Equal(t, implicit, explicit)
		require.Contains(t, explicit, "hover:underline")
	})
	t.Run("panel DensityDefault", func(t *testing.T) {
		explicit := renderState(t, panel.Panel(panel.Config{Density: panel.DensityDefault, Body: templ.Raw("body")}))
		implicit := renderState(t, panel.Panel(panel.Config{Body: templ.Raw("body")}))
		require.Equal(t, implicit, explicit)
		require.Contains(t, explicit, "px-5 py-4")
	})
	t.Run("search MatchModeSubstring", func(t *testing.T) {
		explicit := renderState(t, search.Search(search.Config{ID: "state-search", MatchMode: search.MatchModeSubstring}))
		implicit := renderState(t, search.Search(search.Config{ID: "state-search"}))
		require.Equal(t, implicit, explicit)
		require.Contains(t, explicit, `data-search-match-mode="substring"`)
	})
	t.Run("skeleton ShapeText", func(t *testing.T) {
		explicit := renderState(t, skeleton.Skeleton(skeleton.Config{Shape: skeleton.ShapeText}))
		implicit := renderState(t, skeleton.Skeleton(skeleton.Config{}))
		require.Equal(t, implicit, explicit)
		require.Equal(t, 3, strings.Count(explicit, `aria-hidden="true"`))
	})

	t.Run("table LinkSPA", func(t *testing.T) {
		config := table.Config{Columns: []table.Column{{Key: "name", Label: "Name"}}}
		row := table.Row{ID: "row", Cells: map[string]table.Cell{"name": {Text: "Row"}}, Link: "/rows/1", LinkMode: table.LinkSPA}
		explicit := renderState(t, table.TableRow(config, row))
		row.LinkMode = ""
		implicit := renderState(t, table.TableRow(config, row))
		require.Equal(t, implicit, explicit)
		require.Contains(t, explicit, `hx-target="#main-content-area"`)
	})
	t.Run("table PaginationTraditional", func(t *testing.T) {
		config := table.Config{ID: "state-table", Columns: []table.Column{{Key: "name", Label: "Name"}}, Pagination: &table.PaginationConfig{Mode: table.PaginationTraditional, CurrentPage: 1, TotalPages: 2}}
		explicit := renderState(t, table.Table(config))
		config.Pagination.Mode = ""
		implicit := renderState(t, table.Table(config))
		require.Equal(t, implicit, explicit)
		require.Contains(t, explicit, "state-table-pagination")
	})

	for _, state := range []struct {
		name string
		got  textinput.InputType
		want string
	}{
		{name: "TypeDate", got: textinput.TypeDate, want: "date"},
		{name: "TypeDateTimeLocal", got: textinput.TypeDateTimeLocal, want: "datetime-local"},
		{name: "TypeNumber", got: textinput.TypeNumber, want: "number"},
		{name: "TypeURL", got: textinput.TypeURL, want: "url"},
	} {
		t.Run("textinput "+state.name, func(t *testing.T) {
			html := renderState(t, textinput.TextInput(textinput.Config{ID: "state-input", Label: "State", Type: state.got}))
			require.Contains(t, html, `type="`+state.want+`"`)
		})
	}

	t.Run("tooltip ActivationHover", func(t *testing.T) {
		explicit := renderState(t, tooltip.Tooltip("state-tip", "Tip", tooltip.WithActivation(tooltip.ActivationHover)))
		implicit := renderState(t, tooltip.Tooltip("state-tip", "Tip"))
		require.Equal(t, implicit, explicit)
		require.Contains(t, explicit, "peer-hover:opacity-100")
	})
	t.Run("tooltip PositionTop", func(t *testing.T) {
		explicit := renderState(t, tooltip.Tooltip("state-tip", "Tip", tooltip.WithPosition(tooltip.PositionTop)))
		implicit := renderState(t, tooltip.Tooltip("state-tip", "Tip"))
		require.Equal(t, implicit, explicit)
		require.Contains(t, explicit, "bottom-full")
	})
}

func renderState(t *testing.T, component interface {
	Render(context.Context, io.Writer) error
}) string {
	t.Helper()
	var output bytes.Buffer
	require.NoError(t, component.Render(context.Background(), &output))
	return output.String()
}

func metadataStateConfig() head.MetadataConfig {
	return head.MetadataConfig{
		Title:        "State proof",
		Description:  "Source-derived state proof",
		CanonicalURL: "https://example.com/state",
		Image: head.SocialImage{
			URL:      "https://example.com/state.png",
			MIMEType: "image/png",
			Width:    1200,
			Height:   630,
			Alt:      "State proof",
		},
	}
}
