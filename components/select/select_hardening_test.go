package selectfield

import (
	"encoding/json"
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
	assert.Contains(t, browserHTML, `x-data="goshtosoSelect($el)"`)
	assert.Contains(t, browserHTML, `data-select-config=`)
	assert.NotContains(t, browserHTML, "<script")
}

func TestSelect_RenderedReducedMotionTransitionContract(t *testing.T) {
	for _, cfg := range []Config{
		{
			ID:      "choice",
			Options: []Option{{Value: "a", Label: "A"}},
		},
		{
			ID:        "custom-choice",
			Shell:     true,
			ValueExpr: "choice",
		},
	} {
		rendered := html.UnescapeString(renderSelect(t, cfg, nil))

		assert.Contains(t, rendered, `x-transition:enter="transition ease-out duration-150 motion-reduce:transition-none"`)
		assert.Contains(t, rendered, `x-transition:enter-start="opacity-0 scale-95 motion-reduce:opacity-100 motion-reduce:scale-100"`)
		assert.Contains(t, rendered, `x-transition:enter-end="opacity-100 scale-100"`)
		assert.Contains(t, rendered, `x-transition:leave="transition ease-in duration-100 motion-reduce:transition-none"`)
		assert.Contains(t, rendered, `x-transition:leave-start="opacity-100 scale-100"`)
		assert.Contains(t, rendered, `x-transition:leave-end="opacity-0 scale-95 motion-reduce:opacity-100 motion-reduce:scale-100"`)
		assert.Contains(t, rendered, `class="size-5 shrink-0 transition-transform duration-150 motion-reduce:transition-none"`)
	}
}

func TestSelectFactoryDataJSONPreservesOptionAndPlaceholderStrings(t *testing.T) {
	cfg := Config{
		Placeholder: `Pick Bob's \ region & team`,
		Options: []Option{
			{Value: `us'east\1`, Label: `Bob's \ Team & Co.`},
			{Value: `eu-west`, Label: `Europe`},
		},
	}

	data := cfg.factoryDataJSON()

	var decoded factoryData
	assert.NoError(t, json.Unmarshal([]byte(data), &decoded))
	assert.Equal(t, cfg.Placeholder, decoded.Placeholder)
	assert.Equal(t, []factoryOption{{Value: `us'east\1`, Label: `Bob's \ Team & Co.`}, {Value: "eu-west", Label: "Europe"}}, decoded.Options)
	assert.Empty(t, decoded.SelectedValues)
}

func TestSelectFactoryDataJSON_EmptySelectionIsArray(t *testing.T) {
	assert.Contains(t, Config{}.factoryDataJSON(), `"selectedValues":[]`)
	assert.Contains(t, Config{Options: []Option{{Value: "a", Label: "A"}}}.factoryDataJSON(), `"selectedValues":[]`)
}

func TestSelect_RenderedOptionIDExpressionEscapesConfigID(t *testing.T) {
	rendered := renderSelect(t, Config{
		ID:      `choice'\x`,
		Options: []Option{{Value: "a", Label: "A"}},
	}, nil)
	browserHTML := html.UnescapeString(rendered)

	assert.Contains(t, browserHTML, `x-bind:id="'choice\'\\x-option-' + index"`)
}

func TestSelectOptionsWithNewlinesStayInEncodedFactoryData(t *testing.T) {
	cfg := Config{
		ID: "choice",
		Options: []Option{{
			Value: "safe",
			Label: "first line\nalert(1)",
		}},
	}
	rendered := renderSelect(t, cfg, nil)
	assert.Contains(t, rendered, `data-select-config="`+cfg.factoryData()+`"`)
	assert.NotContains(t, rendered, "first line\nalert(1)")
}

func TestSelectFactoryDataJSONEscapesControlCharacters(t *testing.T) {
	cfg := Config{
		Placeholder: "pick\nregion\r\t\u2028\u2029",
		Options: []Option{{
			Value:    "value'\n\r\t\\\u2028\u2029</script>",
			Label:    "label'\n\r\t\\\u2028\u2029</script>",
			Selected: true,
		}},
	}

	data := cfg.factoryDataJSON()

	var decoded factoryData
	assert.NoError(t, json.Unmarshal([]byte(data), &decoded))
	assert.Equal(t, cfg.Placeholder, decoded.Placeholder)
	assert.Equal(t, []factoryOption{{Value: cfg.Options[0].Value, Label: cfg.Options[0].Label, Selected: true}}, decoded.Options)
	assert.Equal(t, []string{cfg.Options[0].Value}, decoded.SelectedValues)
}
