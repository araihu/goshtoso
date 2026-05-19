package server

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/araihu/goshtoso/components/form"
	"github.com/araihu/goshtoso/components/form/validation"
	"github.com/araihu/goshtoso/components/textinput"
	"github.com/araihu/goshtoso/internal/pages/demo/components"
)

// takenSlugs simulates a database of existing slugs for the demo
var takenSlugs = map[string]bool{
	"admin": true,
	"test":  true,
	"demo":  true,
}

// buildDemoFormDef constructs the FormDef for the form validation demo.
// Exported so the demo page templ can call it too.
func buildDemoFormDef() *validation.FormDef {
	nameField := &form.FieldGroupConfig{
		Label:    "Name",
		Required: true,
		Hints:    []string{"At least 3 characters."},
		Input:    &textinput.Config{Name: "name", Placeholder: "My Project"},
	}
	slugField := &form.FieldGroupConfig{
		Label:    "Slug",
		Required: true,
		Hints:    []string{"URL-friendly identifier. Auto-generated from name."},
		Input:    &textinput.Config{Name: "slug", Placeholder: "my-project"},
	}
	emailField := &form.FieldGroupConfig{
		Label:    "Email",
		Required: true,
		Hints:    []string{"Contact email for notifications."},
		Input:    &textinput.Config{Name: "email", Type: textinput.TypeEmail, Placeholder: "you@example.com"},
	}

	def := &validation.FormDef{
		FormID:   "demo-validation",
		Endpoint: "/api/components/form-validation",
		Fields: map[string]*validation.FieldDef{
			"name":  {Name: "name", FieldGroup: nameField, OnChange: true},
			"slug":  {Name: "slug", FieldGroup: slugField, OnChange: true, DependsOn: []string{"name"}},
			"email": {Name: "email", FieldGroup: emailField, OnChange: true},
		},
	}
	def.Bind()
	return def
}

var slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func validateDemoField(ctx validation.ValidationContext, name string, fg *form.FieldGroupConfig) bool {
	switch name {
	case "name":
		val := ctx.FormValues["name"]
		if val == "" {
			fg.Errors = []string{"Name is required."}
			fg.Input.State = textinput.StateError
			return false
		}
		if len(val) < 3 {
			fg.Errors = []string{"Name must be at least 3 characters."}
			fg.Input.State = textinput.StateError
			return false
		}
		fg.Errors = nil
		fg.Input.State = textinput.StateSuccess
		return true

	case "slug":
		val := ctx.FormValues["slug"]
		// Auto-generate from name if this is a dependency trigger
		if ctx.Type == validation.ValidationDependency && val == "" {
			nameVal := ctx.FormValues["name"]
			generated := strings.ToLower(nameVal)
			generated = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(generated, "-")
			generated = strings.Trim(generated, "-")
			fg.Input.Value = generated
			val = generated
		}
		if val == "" {
			fg.Errors = []string{"Slug is required."}
			fg.Input.State = textinput.StateError
			return false
		}
		if !slugRegex.MatchString(val) {
			fg.Errors = []string{"Slug must be lowercase alphanumeric with hyphens."}
			fg.Input.State = textinput.StateError
			return false
		}
		if takenSlugs[val] {
			fg.Errors = []string{"This slug is already taken."}
			fg.Input.State = textinput.StateError
			return false
		}
		fg.Errors = nil
		fg.Input.State = textinput.StateSuccess
		return true

	case "email":
		val := ctx.FormValues["email"]
		if val == "" {
			fg.Errors = []string{"Email is required."}
			fg.Input.State = textinput.StateError
			return false
		}
		if !strings.Contains(val, "@") || !strings.Contains(val, ".") {
			fg.Errors = []string{"Please enter a valid email address."}
			fg.Input.State = textinput.StateError
			return false
		}
		fg.Errors = nil
		fg.Input.State = textinput.StateSuccess
		return true
	}
	return true
}

func (s *Server) handleFormValidation(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Render the demo page (same as /components/form-validation)
		components.FormValidationDemoPage().Render(r.Context(), w)
		return
	}

	def := buildDemoFormDef()
	result := validation.Handle(r, def, validateDemoField)

	if validation.IsFieldValidation(r) {
		validation.RenderFieldResponse(r.Context(), w, *result)
		return
	}

	// Full form submit
	if !result.Valid {
		// Re-render the form section with validation errors
		w.Header().Set("Content-Type", "text/html")
		components.FormValidationFormSection(
			def.Fields["name"].FieldGroup,
			def.Fields["slug"].FieldGroup,
			def.Fields["email"].FieldGroup,
		).Render(r.Context(), w)
		return
	}

	// Success
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<div class="p-6 text-center text-success font-medium">Form submitted successfully!</div>`))
}
