package form

import (
	"regexp"
	"strings"
	"testing"

	"github.com/araihu/goshtoso/components/checkbox"
	"github.com/araihu/goshtoso/components/combobox"
	"github.com/araihu/goshtoso/components/fileinput"
	selectfield "github.com/araihu/goshtoso/components/select"
	"github.com/araihu/goshtoso/components/structuredinput"
	"github.com/araihu/goshtoso/components/tagslist"
	"github.com/araihu/goshtoso/components/textarea"
	"github.com/araihu/goshtoso/components/textinput"
	"github.com/araihu/goshtoso/components/toggle"
)

var renderedIDPattern = regexp.MustCompile(`\sid="([^"]+)"`)

func TestFieldGroupRendersUniqueWrapperAndControlIDs(t *testing.T) {
	tests := []struct {
		name          string
		cfg           FieldGroupConfig
		labelTarget   string
		groupLabelled bool
	}{
		{
			name:        "text input",
			cfg:         FieldGroupConfig{ID: "email", Label: "Email", Input: &textinput.Config{Name: "email"}},
			labelTarget: "email-input",
		},
		{
			name:        "textarea",
			cfg:         FieldGroupConfig{ID: "note", Label: "Note", Textarea: &textarea.Config{Name: "note"}},
			labelTarget: "note-input",
		},
		{
			name: "combobox",
			cfg: FieldGroupConfig{ID: "branch", Label: "Branch", Combobox: &combobox.Config{
				Name:   "branch",
				Source: combobox.Source{Static: []combobox.Option{{Value: "central", Label: "Central"}}},
			}},
			labelTarget: "branch-control-trigger",
		},
		{
			name: "select",
			cfg: FieldGroupConfig{ID: "country", Label: "Country", Select: &selectfield.Config{
				Name:    "country",
				Options: []selectfield.Option{{Value: "br", Label: "Brazil"}},
			}},
			labelTarget: "country-input-trigger",
		},
		{
			name:        "toggle",
			cfg:         FieldGroupConfig{ID: "notify", Label: "Notify patron", Toggle: &toggle.Config{Name: "notify"}},
			labelTarget: "notify-input",
		},
		{
			name:        "checkbox",
			cfg:         FieldGroupConfig{ID: "confirm", Label: "Confirm handoff", Checkbox: &checkbox.Config{Name: "confirm"}},
			labelTarget: "confirm-input",
		},
		{
			name:        "tags list",
			cfg:         FieldGroupConfig{ID: "labels", Label: "Labels", TagsList: &tagslist.Config{Name: "labels"}},
			labelTarget: "labels-control-input",
		},
		{
			name: "structured input",
			cfg: FieldGroupConfig{ID: "samples", Label: "Samples", StructuredInput: &structuredinput.Config{
				Name:    "samples",
				Columns: []structuredinput.Column{{Key: "seal", Label: "Seal"}},
			}},
			groupLabelled: true,
		},
		{
			name:        "file input",
			cfg:         FieldGroupConfig{ID: "manifest", Label: "Manifest", FileInput: &fileinput.Config{Name: "manifest"}},
			labelTarget: "manifest-input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html := render(t, FieldGroup(tt.cfg))
			assertUniqueRenderedIDs(t, html)
			mustContain(t, html, `id="`+tt.cfg.ID+`"`)
			if tt.groupLabelled {
				mustContain(t, html,
					`role="group"`,
					`aria-labelledby="`+tt.cfg.controlID()+`-label"`,
					`id="`+tt.cfg.controlID()+`-label"`,
				)
				mustNotContain(t, html, `<label`)
				return
			}
			mustContain(t, html, `for="`+tt.labelTarget+`"`, `id="`+tt.labelTarget+`"`)
		})
	}
}

func TestFieldGroupAssociatesErrorsAndHintsWithBuiltInControls(t *testing.T) {
	tests := []struct {
		name       string
		cfg        FieldGroupConfig
		controlID  string
		controlTag string
	}{
		{
			name:       "text input",
			cfg:        FieldGroupConfig{ID: "email", Label: "Email", Errors: []string{"Enter an email"}, Hints: []string{"Work address"}, Input: &textinput.Config{Name: "email"}},
			controlID:  "email-input",
			controlTag: "input",
		},
		{
			name:       "textarea",
			cfg:        FieldGroupConfig{ID: "note", Label: "Note", Errors: []string{"Add a note"}, Hints: []string{"Explain the exception"}, Textarea: &textarea.Config{Name: "note"}},
			controlID:  "note-input",
			controlTag: "textarea",
		},
		{
			name: "combobox",
			cfg: FieldGroupConfig{ID: "branch", Label: "Branch", Errors: []string{"Choose a branch"}, Hints: []string{"Pickup location"}, Combobox: &combobox.Config{
				Name:   "branch",
				Source: combobox.Source{Static: []combobox.Option{{Value: "central", Label: "Central"}}},
			}},
			controlID:  "branch-control-trigger",
			controlTag: "button",
		},
		{
			name: "select",
			cfg: FieldGroupConfig{ID: "country", Label: "Country", Errors: []string{"Choose a country"}, Hints: []string{"Shipping destination"}, Select: &selectfield.Config{
				Name:    "country",
				Options: []selectfield.Option{{Value: "br", Label: "Brazil"}},
			}},
			controlID:  "country-input-trigger",
			controlTag: "button",
		},
		{
			name: "file input preserves component helper",
			cfg: FieldGroupConfig{ID: "manifest", Label: "Manifest", Errors: []string{"Upload the manifest"}, Hints: []string{"Signed files only"}, FileInput: &fileinput.Config{
				Name:       "manifest",
				HelperText: "PDF up to 5 MB",
			}},
			controlID:  "manifest-input",
			controlTag: "input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html := render(t, FieldGroup(tt.cfg))
			control := openingTagWithID(t, html, tt.controlTag, tt.controlID)
			mustContain(t, control, `aria-invalid="true"`)
			describedBy := `aria-describedby="` + tt.cfg.controlID() + `-errors ` + tt.cfg.controlID() + `-hints`
			if tt.cfg.FileInput != nil {
				describedBy += ` ` + tt.cfg.controlID() + `-helper`
			}
			mustContain(t, control, describedBy+`"`)
			mustContain(t, html,
				`id="`+tt.cfg.controlID()+`-errors"`,
				`id="`+tt.cfg.controlID()+`-hints"`,
			)
		})
	}
}

func TestFormErrorsLinksSummaryToFocusableFieldTargets(t *testing.T) {
	branch := FieldGroupConfig{ID: "branch", Combobox: &combobox.Config{Name: "branch"}}
	html := render(t, FormErrors(FormErrorsConfig{
		ID: "validation-summary",
		Items: []FormErrorItem{
			{Path: "Email", Message: "Enter an email", TargetID: "email"},
			{Message: "Choose a branch", TargetID: branch.FocusTargetID()},
		},
	}))

	mustContain(t, html,
		`id="validation-summary"`,
		`tabindex="-1"`,
		`x-init="$el.focus()"`,
		`href="#email"`,
		`>Email</a>`,
		`href="#branch-control-trigger"`,
		`>Choose a branch</a>`,
	)
}

func TestFieldGroupRequiredStateReachesBuiltInControls(t *testing.T) {
	tests := []struct {
		name       string
		cfg        FieldGroupConfig
		controlTag string
	}{
		{name: "text input", cfg: FieldGroupConfig{ID: "name", Required: true, Input: &textinput.Config{Name: "name"}}, controlTag: "input"},
		{name: "combobox", cfg: FieldGroupConfig{ID: "branch", Required: true, Combobox: &combobox.Config{Name: "branch"}}, controlTag: "button"},
		{name: "select", cfg: FieldGroupConfig{ID: "country", Required: true, Select: &selectfield.Config{Name: "country"}}, controlTag: "button"},
		{name: "toggle", cfg: FieldGroupConfig{ID: "notify", Required: true, Toggle: &toggle.Config{Name: "notify"}}, controlTag: "input"},
		{name: "checkbox", cfg: FieldGroupConfig{ID: "confirm", Required: true, Checkbox: &checkbox.Config{Name: "confirm"}}, controlTag: "input"},
		{name: "tags list", cfg: FieldGroupConfig{ID: "labels", Required: true, TagsList: &tagslist.Config{Name: "labels"}}, controlTag: "input"},
		{name: "structured input", cfg: FieldGroupConfig{ID: "samples", Required: true, StructuredInput: &structuredinput.Config{Name: "samples"}}, controlTag: "div"},
		{name: "file input", cfg: FieldGroupConfig{ID: "manifest", Required: true, FileInput: &fileinput.Config{Name: "manifest"}}, controlTag: "input"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html := render(t, FieldGroup(tt.cfg))
			control := openingTagWithID(t, html, tt.controlTag, tt.cfg.FocusTargetID())
			mustContain(t, control, `aria-required="true"`)
		})
	}
}

func TestFieldGroupPreservesLegacyRootIDWithoutDuplicatingControlID(t *testing.T) {
	html := render(t, FieldGroup(FieldGroupConfig{
		ID:    "email",
		Label: "Email",
		Input: &textinput.Config{Name: "email"},
	}))

	assertUniqueRenderedIDs(t, html)
	mustContain(t, html,
		`id="email"`,
		`id="email-input"`,
		`for="email-input"`,
	)
}

func TestFieldGroupPreservesLegacyRootWhenNestedCompositeReusesItsID(t *testing.T) {
	cfg := FieldGroupConfig{
		ID:    "provider",
		Label: "Provider",
		Combobox: &combobox.Config{
			ID:     "provider",
			Name:   "provider",
			Source: combobox.Source{Static: []combobox.Option{{Value: "local", Label: "Local"}}},
		},
	}
	html := render(t, FieldGroup(cfg))

	assertUniqueRenderedIDs(t, html)
	mustContain(t, html,
		`id="provider"`,
		`id="provider-control"`,
		`id="provider-control-trigger"`,
		`for="provider-control-trigger"`,
	)
	if got := cfg.FocusTargetID(); got != "provider-control-trigger" {
		t.Fatalf("FocusTargetID() = %q, want provider-control-trigger", got)
	}
}

func TestFieldGroupWithoutControlIDDoesNotEmitEmptyDescriptionIDs(t *testing.T) {
	html := render(t, FieldGroup(FieldGroupConfig{
		Label:  "Custom field",
		Errors: []string{"Check this value"},
		Hints:  []string{"Use the domain format"},
	}))

	mustNotContain(t, html, `id=""`, `aria-describedby=""`)
}

func TestBoundValidationFieldPreservesSwapTargetWithoutDuplicatingControlID(t *testing.T) {
	html := render(t, FieldGroup(FieldGroupConfig{
		ID:    "goshtoso-field-name",
		Label: "Name",
		Meta:  &FieldMeta{FormID: "project", FieldName: "name"},
		Input: &textinput.Config{Name: "name"},
	}))

	assertUniqueRenderedIDs(t, html)
	mustContain(t, html,
		`id="goshtoso-field-name"`,
		`id="goshtoso-field-name-input"`,
		`for="goshtoso-field-name-input"`,
	)
}

func assertUniqueRenderedIDs(t *testing.T, html string) {
	t.Helper()
	seen := map[string]struct{}{}
	for _, match := range renderedIDPattern.FindAllStringSubmatch(html, -1) {
		id := match[1]
		if _, duplicate := seen[id]; duplicate {
			t.Fatalf("rendered duplicate id %q in:\n%s", id, html)
		}
		seen[id] = struct{}{}
	}
}

func openingTagWithID(t *testing.T, html, tag, id string) string {
	t.Helper()
	start := strings.Index(html, "<"+tag)
	for start >= 0 {
		end := strings.Index(html[start:], ">")
		if end < 0 {
			break
		}
		opening := html[start : start+end+1]
		if strings.Contains(opening, `id="`+id+`"`) {
			return opening
		}
		next := strings.Index(html[start+end+1:], "<"+tag)
		if next < 0 {
			break
		}
		start += end + 1 + next
	}
	t.Fatalf("rendered HTML has no <%s> with id %q in:\n%s", tag, id, html)
	return ""
}
