package selectfield

import (
	"html"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelectData_EscapesOptionAndPlaceholderStrings(t *testing.T) {
	cfg := Config{
		Placeholder: `Pick Bob's \ region & team`,
		Options: []Option{
			{Value: `us'east\1`, Label: `Bob's \ Team & Co.`},
			{Value: `eu-west`, Label: `Europe`},
		},
	}

	data := selectData(cfg)

	assert.Contains(t, data, `placeholder: 'Pick Bob\'s \\ region & team'`)
	assert.Contains(t, data, `{value:'us\'east\\1',label:'Bob\'s \\ Team & Co.'}`)
	assert.Contains(t, data, "selectedValues: []")
	assert.NotContains(t, data, "selectedValues: null")
}

func TestSelectedValueJS_EmptySelectionIsArray(t *testing.T) {
	assert.Equal(t, "[]", selectedValueJS(nil))
	assert.Equal(t, "[]", selectedValueJS([]Option{{Value: "a", Label: "A"}}))
}

func TestSelect_RenderedOptionIDExpressionEscapesConfigID(t *testing.T) {
	rendered := renderSelect(t, Config{
		ID:      `choice'\x`,
		Options: []Option{{Value: "a", Label: "A"}},
	}, nil)
	browserHTML := html.UnescapeString(rendered)

	assert.Contains(t, browserHTML, `x-bind:id="'choice\'\\x-option-' + index"`)
}

func TestSelectOptionsEscapeNewlinesInAlpineStrings(t *testing.T) {
	rendered := renderSelect(t, Config{
		ID: "choice",
		Options: []Option{{
			Value: "safe",
			Label: "first line\nalert(1)",
		}},
	}, nil)
	browserHTML := html.UnescapeString(rendered)

	assert.Contains(t, browserHTML, `label:'first line\nalert(1)'`)
	assert.NotContains(t, browserHTML, "label:'first line\nalert(1)'")
}
