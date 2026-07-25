package components_test

import (
	"testing"

	"github.com/araihu/goshtoso/components"
	"github.com/araihu/goshtoso/components/button"
	"github.com/araihu/goshtoso/components/checkbox"
	"github.com/araihu/goshtoso/components/combobox"
	"github.com/araihu/goshtoso/components/fileinput"
	"github.com/araihu/goshtoso/components/form"
	"github.com/araihu/goshtoso/components/palette"
	"github.com/araihu/goshtoso/components/radio"
	rangecomponent "github.com/araihu/goshtoso/components/range"
	"github.com/araihu/goshtoso/components/rating"
	"github.com/araihu/goshtoso/components/schemaform"
	"github.com/araihu/goshtoso/components/search"
	selectcomponent "github.com/araihu/goshtoso/components/select"
	"github.com/araihu/goshtoso/components/structuredinput"
	"github.com/araihu/goshtoso/components/tagslist"
	"github.com/araihu/goshtoso/components/textarea"
	"github.com/araihu/goshtoso/components/textinput"
	"github.com/araihu/goshtoso/components/toggle"
	"github.com/stretchr/testify/require"
)

func inputRenderables() map[components.Kind]components.Component {
	return map[components.Kind]components.Component{
		components.KindButton:                 button.Button(),
		components.KindCheckbox:               checkbox.Checkbox(checkbox.Config{}),
		components.KindCheckboxGroup:          checkbox.CheckboxGroup(checkbox.GroupConfig{}),
		components.KindCombobox:               combobox.Combobox(combobox.Config{}, combobox.State{}),
		components.KindFileInput:              fileinput.FileInput(fileinput.Config{}),
		components.KindForm:                   form.Form(form.Config{}),
		components.KindFormSection:            form.Section(form.SectionConfig{}),
		components.KindFormCollapsibleSection: form.CollapsibleSection(form.CollapsibleSectionConfig{}),
		components.KindFormFlipSection:        form.FlipSection(form.FlipSectionConfig{}, nil),
		components.KindFormSubSection:         form.SubSection(form.SubSectionConfig{}),
		components.KindFormFieldGroup:         form.FieldGroup(form.FieldGroupConfig{}),
		components.KindFormErrors:             form.FormErrors(form.FormErrorsConfig{}),
		components.KindPalette:                palette.Palette(palette.Config{}),
		components.KindRadio:                  radio.Radio(radio.Config{}),
		components.KindRadioBar:               radio.RadioBar(),
		components.KindRadioGroup:             radio.RadioGroup(radio.GroupConfig{}),
		components.KindRange:                  rangecomponent.Range(rangecomponent.Config{}),
		components.KindRating:                 rating.Rating(rating.Config{}),
		components.KindRatingDisplay:          rating.RatingDisplay(rating.DisplayConfig{}),
		components.KindSchemaFormFields:       schemaform.Fields(schemaform.FieldsConfig{}),
		components.KindSearch:                 search.Search(search.Config{}),
		components.KindSearchField:            search.SearchField(search.Config{}),
		components.KindSearchModal:            search.SearchModal(search.Config{}),
		components.KindSelect:                 selectcomponent.Select(selectcomponent.Config{}),
		components.KindStructuredInput:        structuredinput.StructuredInput(structuredinput.Config{}),
		components.KindTagsList:               tagslist.TagsList(tagslist.Config{}),
		components.KindTextarea:               textarea.Textarea(textarea.Config{}),
		components.KindTextareaWithActions:    textarea.TextareaWithActions(textarea.Config{}),
		components.KindTextInput:              textinput.TextInput(textinput.Config{}),
		components.KindToggle:                 toggle.Toggle(toggle.Config{}),
	}
}

func TestInputRenderablesExposeKinds(t *testing.T) {
	values := inputRenderables()
	require.Len(t, values, 30)
	for want, value := range values {
		require.Equal(t, want, value.Kind())
	}
}
