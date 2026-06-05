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

func TestSelectData_EscapesControlCharactersInAllAlpineStrings(t *testing.T) {
	cfg := Config{
		Placeholder: "pick\nregion\r\t\u2028\u2029",
		Options: []Option{{
			Value:    "value'\n\r\t\\\u2028\u2029</script>",
			Label:    "label'\n\r\t\\\u2028\u2029</script>",
			Selected: true,
		}},
	}

	data := selectData(cfg)

	assert.Contains(t, data, `placeholder: 'pick\nregion\r\t\u2028\u2029'`)
	assert.Contains(t, data, `{value:'value\'\n\r\t\\\u2028\u2029</script>',label:'label\'\n\r\t\\\u2028\u2029</script>'}`)
	assert.Contains(t, data, `selectedValues: ['value\'\n\r\t\\\u2028\u2029</script>']`)
	assert.NotContains(t, data, "value'\n")
	assert.NotContains(t, data, "\r")
	assert.NotContains(t, data, "\u2028")
	assert.NotContains(t, data, "\u2029")
}
