package validation

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/araihu/goshtoso/components/checkbox"
	"github.com/araihu/goshtoso/components/form"
	"github.com/araihu/goshtoso/components/textarea"
	"github.com/araihu/goshtoso/components/textinput"
	"github.com/araihu/goshtoso/components/toggle"
)

// ---------------------------------------------------------------------------
// Bind tests
// ---------------------------------------------------------------------------

func TestBind_SetsMetaOnAllFields(t *testing.T) {
	fd := &FormDef{
		FormID:   "myform",
		Endpoint: "/validate",
		Fields: map[string]*FieldDef{
			"name": {
				Name:       "name",
				FieldGroup: &form.FieldGroupConfig{},
				OnChange:   true,
			},
			"email": {
				Name:       "email",
				FieldGroup: &form.FieldGroupConfig{},
				DependsOn:  []string{"name"},
				OnChange:   true,
			},
		},
	}

	fd.Bind()

	for name, fdef := range fd.Fields {
		require.NotNil(t, fdef.FieldGroup.Meta, "Meta should be set for field %s", name)
		assert.Equal(t, "myform", fdef.FieldGroup.Meta.FormID)
		assert.Equal(t, name, fdef.FieldGroup.Meta.FieldName)
	}
	assert.Equal(t, "name", fd.Fields["email"].FieldGroup.Meta.DependsOn)
}

func TestBind_SetsValidationEndpoint(t *testing.T) {
	fd := &FormDef{
		FormID:   "f1",
		Endpoint: "/api/validate",
		Fields: map[string]*FieldDef{
			"field1": {
				Name:       "field1",
				FieldGroup: &form.FieldGroupConfig{},
				OnChange:   true,
			},
		},
	}

	fd.Bind()

	require.NotNil(t, fd.Fields["field1"].FieldGroup.Validation)
	assert.Equal(t, "/api/validate", fd.Fields["field1"].FieldGroup.Validation.Endpoint)
}

func TestBind_AutoSetsFieldGroupID(t *testing.T) {
	fd := &FormDef{
		FormID: "f1",
		Fields: map[string]*FieldDef{
			"username": {
				Name:       "username",
				FieldGroup: &form.FieldGroupConfig{},
			},
		},
	}

	fd.Bind()

	assert.Equal(t, "goshtoso-field-username", fd.Fields["username"].FieldGroup.ID)
}

func TestBind_PreservesExistingID(t *testing.T) {
	fd := &FormDef{
		FormID: "f1",
		Fields: map[string]*FieldDef{
			"username": {
				Name:       "username",
				FieldGroup: &form.FieldGroupConfig{ID: "custom-id"},
			},
		},
	}

	fd.Bind()

	assert.Equal(t, "custom-id", fd.Fields["username"].FieldGroup.ID)
}

func TestBind_PreservesExistingEndpoint(t *testing.T) {
	customValidation := &form.ValidationConfig{
		Endpoint: "/custom/endpoint",
		Trigger:  "input",
	}
	fd := &FormDef{
		FormID:   "f1",
		Endpoint: "/default/endpoint",
		Fields: map[string]*FieldDef{
			"field1": {
				Name:       "field1",
				FieldGroup: &form.FieldGroupConfig{Validation: customValidation},
				OnChange:   true,
			},
		},
	}

	fd.Bind()

	// Bind only sets Validation when it's nil, so the custom one is preserved
	assert.Equal(t, "/custom/endpoint", fd.Fields["field1"].FieldGroup.Validation.Endpoint)
	assert.Equal(t, "input", fd.Fields["field1"].FieldGroup.Validation.Trigger)
}

func TestBind_SkipsValidationForNonOnChange(t *testing.T) {
	fd := &FormDef{
		FormID:   "f1",
		Endpoint: "/validate",
		Fields: map[string]*FieldDef{
			"static": {
				Name:       "static",
				FieldGroup: &form.FieldGroupConfig{},
				OnChange:   false,
			},
		},
	}

	fd.Bind()

	assert.Nil(t, fd.Fields["static"].FieldGroup.Validation)
}

// ---------------------------------------------------------------------------
// Dependents tests
// ---------------------------------------------------------------------------

func TestDependents_FindsDirect(t *testing.T) {
	fd := &FormDef{
		Fields: map[string]*FieldDef{
			"A": {Name: "A", FieldGroup: &form.FieldGroupConfig{}},
			"B": {Name: "B", FieldGroup: &form.FieldGroupConfig{}, DependsOn: []string{"A"}},
		},
	}

	deps := fd.Dependents("A")
	assert.Equal(t, []string{"B"}, deps)
}

func TestDependents_NonTransitive(t *testing.T) {
	fd := &FormDef{
		Fields: map[string]*FieldDef{
			"A": {Name: "A", FieldGroup: &form.FieldGroupConfig{}},
			"B": {Name: "B", FieldGroup: &form.FieldGroupConfig{}, DependsOn: []string{"A"}},
			"C": {Name: "C", FieldGroup: &form.FieldGroupConfig{}, DependsOn: []string{"B"}},
		},
	}

	deps := fd.Dependents("A")
	assert.Equal(t, []string{"B"}, deps)

	// C is NOT a dependent of A, only of B
	depsOfB := fd.Dependents("B")
	assert.Equal(t, []string{"C"}, depsOfB)
}

func TestDependents_Multiple(t *testing.T) {
	fd := &FormDef{
		Fields: map[string]*FieldDef{
			"A": {Name: "A", FieldGroup: &form.FieldGroupConfig{}},
			"B": {Name: "B", FieldGroup: &form.FieldGroupConfig{}, DependsOn: []string{"A"}},
			"C": {Name: "C", FieldGroup: &form.FieldGroupConfig{}, DependsOn: []string{"A"}},
		},
	}

	deps := fd.Dependents("A")
	sort.Strings(deps)
	assert.Equal(t, []string{"B", "C"}, deps)
}

func TestDependents_NoDeps(t *testing.T) {
	fd := &FormDef{
		Fields: map[string]*FieldDef{
			"A": {Name: "A", FieldGroup: &form.FieldGroupConfig{}},
			"B": {Name: "B", FieldGroup: &form.FieldGroupConfig{}},
		},
	}

	deps := fd.Dependents("A")
	assert.Empty(t, deps)
}

func TestDependents_UnknownField(t *testing.T) {
	fd := &FormDef{
		Fields: map[string]*FieldDef{
			"A": {Name: "A", FieldGroup: &form.FieldGroupConfig{}},
		},
	}

	deps := fd.Dependents("nonexistent")
	assert.Empty(t, deps)
}

// ---------------------------------------------------------------------------
// PopulateValues tests
// ---------------------------------------------------------------------------

func TestPopulateValues_TextInput(t *testing.T) {
	fd := &FormDef{
		Fields: map[string]*FieldDef{
			"name": {
				Name:       "name",
				FieldGroup: &form.FieldGroupConfig{Input: &textinput.Config{}},
			},
		},
	}

	fd.PopulateValues(map[string]string{"name": "Alice"})

	assert.Equal(t, "Alice", fd.Fields["name"].FieldGroup.Input.Value)
}

func TestPopulateValues_Textarea(t *testing.T) {
	fd := &FormDef{
		Fields: map[string]*FieldDef{
			"bio": {
				Name:       "bio",
				FieldGroup: &form.FieldGroupConfig{Textarea: &textarea.Config{}},
			},
		},
	}

	fd.PopulateValues(map[string]string{"bio": "Hello world"})

	assert.Equal(t, "Hello world", fd.Fields["bio"].FieldGroup.Textarea.Value)
}

func TestPopulateValues_Toggle(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"on value", "on", true},
		{"empty value", "", false},
		{"other value", "off", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fd := &FormDef{
				Fields: map[string]*FieldDef{
					"active": {
						Name:       "active",
						FieldGroup: &form.FieldGroupConfig{Toggle: &toggle.Config{}},
					},
				},
			}

			fd.PopulateValues(map[string]string{"active": tt.value})

			assert.Equal(t, tt.expected, fd.Fields["active"].FieldGroup.Toggle.Checked)
		})
	}
}

func TestPopulateValues_Checkbox(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"on value", "on", true},
		{"empty value", "", false},
		{"other value", "no", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fd := &FormDef{
				Fields: map[string]*FieldDef{
					"agree": {
						Name:       "agree",
						FieldGroup: &form.FieldGroupConfig{Checkbox: &checkbox.Config{}},
					},
				},
			}

			fd.PopulateValues(map[string]string{"agree": tt.value})

			assert.Equal(t, tt.expected, fd.Fields["agree"].FieldGroup.Checkbox.Checked)
		})
	}
}

func TestPopulateValues_MissingField(t *testing.T) {
	fd := &FormDef{
		Fields: map[string]*FieldDef{
			"name": {
				Name:       "name",
				FieldGroup: &form.FieldGroupConfig{Input: &textinput.Config{}},
			},
		},
	}

	// No "name" in values map — should default to empty string, no panic
	fd.PopulateValues(map[string]string{"other": "value"})

	assert.Equal(t, "", fd.Fields["name"].FieldGroup.Input.Value)
}

func TestPopulateValues_NilFieldTypes(t *testing.T) {
	fd := &FormDef{
		Fields: map[string]*FieldDef{
			"custom": {
				Name:       "custom",
				FieldGroup: &form.FieldGroupConfig{}, // no built-in type
			},
		},
	}

	// Should not panic
	assert.NotPanics(t, func() {
		fd.PopulateValues(map[string]string{"custom": "value"})
	})
}

// ---------------------------------------------------------------------------
// Handle tests
// ---------------------------------------------------------------------------

func newPostRequest(values url.Values, headers map[string]string) *http.Request {
	body := values.Encode()
	r := httptest.NewRequest("POST", "/validate", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for k, v := range headers {
		r.Header.Set(k, v)
	}
	return r
}

func TestHandle_Submit_AllFieldsValidated(t *testing.T) {
	fd := &FormDef{
		FormID: "f1",
		Fields: map[string]*FieldDef{
			"a": {Name: "a", FieldGroup: &form.FieldGroupConfig{Input: &textinput.Config{}}},
			"b": {Name: "b", FieldGroup: &form.FieldGroupConfig{Input: &textinput.Config{}}},
			"c": {Name: "c", FieldGroup: &form.FieldGroupConfig{Input: &textinput.Config{}}},
		},
	}

	validated := make(map[string]bool)
	hook := func(ctx ValidationContext, name string, fg *form.FieldGroupConfig) bool {
		validated[name] = true
		assert.Equal(t, ValidationSubmit, ctx.Type)
		return true
	}

	values := url.Values{"a": {"1"}, "b": {"2"}, "c": {"3"}}
	r := newPostRequest(values, nil)

	Handle(r, fd, hook)

	assert.True(t, validated["a"])
	assert.True(t, validated["b"])
	assert.True(t, validated["c"])
}

func TestHandle_Submit_ValidResult(t *testing.T) {
	fd := &FormDef{
		FormID: "f1",
		Fields: map[string]*FieldDef{
			"name": {Name: "name", FieldGroup: &form.FieldGroupConfig{Input: &textinput.Config{}}},
		},
	}

	hook := func(_ ValidationContext, _ string, _ *form.FieldGroupConfig) bool {
		return true
	}

	values := url.Values{"name": {"Alice"}}
	r := newPostRequest(values, nil)

	result := Handle(r, fd, hook)
	assert.True(t, result.Valid)
}

func TestHandle_Submit_InvalidResult(t *testing.T) {
	fd := &FormDef{
		FormID: "f1",
		Fields: map[string]*FieldDef{
			"name":  {Name: "name", FieldGroup: &form.FieldGroupConfig{Input: &textinput.Config{}}},
			"email": {Name: "email", FieldGroup: &form.FieldGroupConfig{Input: &textinput.Config{}}},
		},
	}

	hook := func(_ ValidationContext, name string, fg *form.FieldGroupConfig) bool {
		if name == "email" {
			fg.Errors = []string{"required"}
			return false
		}
		return true
	}

	values := url.Values{"name": {"Alice"}, "email": {""}}
	r := newPostRequest(values, nil)

	result := Handle(r, fd, hook)
	assert.False(t, result.Valid)
}

func TestHandle_Submit_DeterministicOrder(t *testing.T) {
	fd := &FormDef{
		FormID: "f1",
		Fields: map[string]*FieldDef{
			"z_field": {Name: "z_field", FieldGroup: &form.FieldGroupConfig{Input: &textinput.Config{}}},
			"a_field": {Name: "a_field", FieldGroup: &form.FieldGroupConfig{Input: &textinput.Config{}}},
			"m_field": {Name: "m_field", FieldGroup: &form.FieldGroupConfig{Input: &textinput.Config{}}},
		},
	}

	var order []string
	hook := func(_ ValidationContext, name string, _ *form.FieldGroupConfig) bool {
		order = append(order, name)
		return true
	}

	r := newPostRequest(url.Values{}, nil)
	Handle(r, fd, hook)

	assert.Equal(t, []string{"a_field", "m_field", "z_field"}, order)
}

func TestHandle_FieldChange_OnlyTriggerAndDeps(t *testing.T) {
	fd := &FormDef{
		FormID: "f1",
		Fields: map[string]*FieldDef{
			"country":   {Name: "country", FieldGroup: &form.FieldGroupConfig{Input: &textinput.Config{}}, OnChange: true},
			"region":    {Name: "region", FieldGroup: &form.FieldGroupConfig{Input: &textinput.Config{}}, DependsOn: []string{"country"}, OnChange: true},
			"unrelated": {Name: "unrelated", FieldGroup: &form.FieldGroupConfig{Input: &textinput.Config{}}, OnChange: true},
		},
	}

	validated := make(map[string]bool)
	hook := func(_ ValidationContext, name string, _ *form.FieldGroupConfig) bool {
		validated[name] = true
		return true
	}

	values := url.Values{
		"country":               {"US"},
		"region":                {""},
		"unrelated":             {"x"},
		"X-Goshtoso-Validation": {"field"},
	}
	r := newPostRequest(values, map[string]string{"HX-Trigger-Name": "country"})

	Handle(r, fd, hook)

	assert.True(t, validated["country"], "trigger field should be validated")
	assert.True(t, validated["region"], "dependent field should be validated")
	assert.False(t, validated["unrelated"], "unrelated field should NOT be validated")
}

func TestHandle_FieldChange_SetsCorrectType(t *testing.T) {
	fd := &FormDef{
		FormID: "f1",
		Fields: map[string]*FieldDef{
			"country": {Name: "country", FieldGroup: &form.FieldGroupConfig{Input: &textinput.Config{}}, OnChange: true},
			"region":  {Name: "region", FieldGroup: &form.FieldGroupConfig{Input: &textinput.Config{}}, DependsOn: []string{"country"}, OnChange: true},
		},
	}

	typesByField := make(map[string]ValidationType)
	hook := func(ctx ValidationContext, name string, _ *form.FieldGroupConfig) bool {
		typesByField[name] = ctx.Type
		return true
	}

	values := url.Values{
		"country":               {"US"},
		"region":                {""},
		"X-Goshtoso-Validation": {"field"},
	}
	r := newPostRequest(values, map[string]string{"HX-Trigger-Name": "country"})

	Handle(r, fd, hook)

	assert.Equal(t, ValidationFieldChange, typesByField["country"])
	assert.Equal(t, ValidationDependency, typesByField["region"])
}

func TestHandle_FieldChange_PrimaryAndDependents(t *testing.T) {
	fd := &FormDef{
		FormID: "f1",
		Fields: map[string]*FieldDef{
			"country": {Name: "country", FieldGroup: &form.FieldGroupConfig{Input: &textinput.Config{}}, OnChange: true},
			"region":  {Name: "region", FieldGroup: &form.FieldGroupConfig{Input: &textinput.Config{}}, DependsOn: []string{"country"}, OnChange: true},
			"city":    {Name: "city", FieldGroup: &form.FieldGroupConfig{Input: &textinput.Config{}}, DependsOn: []string{"country"}, OnChange: true},
		},
	}

	hook := func(_ ValidationContext, _ string, _ *form.FieldGroupConfig) bool { return true }

	values := url.Values{
		"country":               {"US"},
		"region":                {""},
		"city":                  {""},
		"X-Goshtoso-Validation": {"field"},
	}
	r := newPostRequest(values, map[string]string{"HX-Trigger-Name": "country"})

	result := Handle(r, fd, hook)

	require.NotNil(t, result.Primary)
	assert.Equal(t, "country", result.Primary.Name)

	depNames := make([]string, len(result.Dependents))
	for i, d := range result.Dependents {
		depNames[i] = d.Name
	}
	sort.Strings(depNames)
	assert.Equal(t, []string{"city", "region"}, depNames)
}

func TestHandle_HookMutatesConfig(t *testing.T) {
	fg := &form.FieldGroupConfig{Input: &textinput.Config{}}
	fd := &FormDef{
		FormID: "f1",
		Fields: map[string]*FieldDef{
			"email": {Name: "email", FieldGroup: fg},
		},
	}

	hook := func(_ ValidationContext, name string, fg *form.FieldGroupConfig) bool {
		fg.Errors = []string{"invalid email format"}
		fg.Input.State = textinput.StateError
		return false
	}

	values := url.Values{"email": {"bad"}}
	r := newPostRequest(values, nil)

	result := Handle(r, fd, hook)

	assert.False(t, result.Valid)
	assert.Equal(t, []string{"invalid email format"}, fg.Errors)
	assert.Equal(t, textinput.StateError, fg.Input.State)
}

func TestHandle_UnknownTriggerField(t *testing.T) {
	fd := &FormDef{
		FormID: "f1",
		Fields: map[string]*FieldDef{
			"name": {Name: "name", FieldGroup: &form.FieldGroupConfig{Input: &textinput.Config{}}},
		},
	}

	hook := func(_ ValidationContext, _ string, _ *form.FieldGroupConfig) bool {
		t.Fatal("hook should not be called for unknown field")
		return false
	}

	values := url.Values{
		"X-Goshtoso-Validation": {"field"},
	}
	r := newPostRequest(values, map[string]string{"HX-Trigger-Name": "nonexistent"})

	result := Handle(r, fd, hook)

	// Unknown trigger field means nothing to validate — result is valid
	assert.True(t, result.Valid)
	assert.Nil(t, result.Primary)
}

// ---------------------------------------------------------------------------
// IsFieldValidation tests
// ---------------------------------------------------------------------------

func TestIsFieldValidation_True(t *testing.T) {
	values := url.Values{"X-Goshtoso-Validation": {"field"}}
	r := newPostRequest(values, nil)

	assert.True(t, IsFieldValidation(r))
}

func TestIsFieldValidation_False(t *testing.T) {
	r := newPostRequest(url.Values{}, nil)

	assert.False(t, IsFieldValidation(r))
}

func TestIsFieldValidation_WrongValue(t *testing.T) {
	values := url.Values{"X-Goshtoso-Validation": {"submit"}}
	r := newPostRequest(values, nil)

	assert.False(t, IsFieldValidation(r))
}

// ---------------------------------------------------------------------------
// RenderFieldResponse tests
// ---------------------------------------------------------------------------

func TestRenderFieldResponse_PrimaryOnly(t *testing.T) {
	fg := &form.FieldGroupConfig{
		ID:    "goshtoso-field-name",
		Label: "Name",
		Input: &textinput.Config{Name: "name", Value: "Alice"},
	}

	result := Result{
		Valid:   true,
		Primary: &FieldDef{Name: "name", FieldGroup: fg},
	}

	w := httptest.NewRecorder()
	err := RenderFieldResponse(t.Context(), w, result)
	require.NoError(t, err)

	body := w.Body.String()
	assert.Contains(t, body, "goshtoso-field-name")
	assert.Equal(t, "text/html; charset=utf-8", w.Header().Get("Content-Type"))
}

func TestRenderFieldResponse_WithOOB(t *testing.T) {
	primary := &form.FieldGroupConfig{
		ID:    "goshtoso-field-country",
		Label: "Country",
		Input: &textinput.Config{Name: "country", Value: "US"},
	}
	dep := &form.FieldGroupConfig{
		ID:    "goshtoso-field-region",
		Label: "Region",
		Input: &textinput.Config{Name: "region"},
	}

	result := Result{
		Valid:   true,
		Primary: &FieldDef{Name: "country", FieldGroup: primary},
		Dependents: []*FieldDef{
			{Name: "region", FieldGroup: dep},
		},
	}

	w := httptest.NewRecorder()
	err := RenderFieldResponse(t.Context(), w, result)
	require.NoError(t, err)

	body := w.Body.String()
	assert.Contains(t, body, "goshtoso-field-country")
	assert.Contains(t, body, "goshtoso-field-region")
	assert.Contains(t, body, "hx-swap-oob")

	// OOB flag should be reset after rendering
	assert.False(t, dep.OOB, "OOB should be reset to false after render")
}

func TestRenderFieldResponse_NilPrimary(t *testing.T) {
	result := Result{
		Valid:   true,
		Primary: nil,
	}

	w := httptest.NewRecorder()
	err := RenderFieldResponse(t.Context(), w, result)
	require.NoError(t, err)

	body := w.Body.String()
	assert.Empty(t, body)
}
