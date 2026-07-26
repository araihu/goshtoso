package selectfield

import (
	"html"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelect_RenderedARIAContract(t *testing.T) {
	rendered := renderSelect(t, Config{
		ID:    "choice",
		Label: "Choice",
		Options: []Option{
			{Value: "a", Label: "A"},
			{Value: "b", Label: "B", Selected: true},
		},
	}, nil)
	browserHTML := html.UnescapeString(rendered)

	assert.Contains(t, browserHTML, `id="choice-trigger"`)
	assert.Contains(t, browserHTML, `aria-controls="choice-listbox"`)
	assert.Contains(t, browserHTML, `<ul id="choice-listbox"`)
	assert.Equal(t, 1, strings.Count(browserHTML, `role="listbox"`), "a listbox must not contain another listbox")
	assert.Contains(t, browserHTML, `role="option"`)
	assert.Contains(t, browserHTML, `x-bind:aria-selected="selectedOption && selectedOption.value === item.value"`)
}

func TestSelect_ShellRenderedARIAContract(t *testing.T) {
	rendered := renderSelect(t, Config{
		ID:        "custom-choice",
		Label:     "Custom choice",
		Shell:     true,
		ValueExpr: "choice",
	}, nil)
	browserHTML := html.UnescapeString(rendered)

	assert.Contains(t, browserHTML, `aria-controls="custom-choice-listbox"`)
	assert.Contains(t, browserHTML, `id="custom-choice-listbox"`)
	assert.Equal(t, 1, strings.Count(browserHTML, `role="listbox"`))
}

func TestSelect_RenderedKeyboardAndExternalSyncContract(t *testing.T) {
	rendered := renderSelect(t, Config{
		ID:      "choice",
		Options: []Option{{Value: "a", Label: "A"}},
	}, nil)
	browserHTML := html.UnescapeString(rendered)

	assert.Contains(t, browserHTML, `x-ref="trigger"`)
	assert.Contains(t, browserHTML, `x-on:keydown.down.prevent="openFromTrigger(1)"`)
	assert.Contains(t, browserHTML, `x-on:keydown.up.prevent="openFromTrigger(-1)"`)
	assert.Contains(t, browserHTML, `x-on:keydown.down.prevent="moveActiveOption($el, 1)"`)
	assert.Contains(t, browserHTML, `x-on:keydown.up.prevent="moveActiveOption($el, -1)"`)
	assert.Contains(t, browserHTML, `x-on:input="if ($event.target === $refs.hiddenInput) syncFromInput($event.target.value)"`)
	assert.Contains(t, browserHTML, `x-on:change="if ($event.target === $refs.hiddenInput) syncFromInput($event.target.value)"`)
	assert.Contains(t, browserHTML, "openFromTrigger(direction)")
	assert.Contains(t, browserHTML, "moveActiveOption(current, direction)")
	assert.Contains(t, browserHTML, "syncFromInput(value)")
}

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
