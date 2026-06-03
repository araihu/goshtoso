package validation

import (
	"net/http"
	"sort"
)

// Handle is the main entry point for form validation.
// It parses the request, populates field values, runs the validation hook,
// and returns a Result.
func Handle(r *http.Request, def *FormDef, validate ValidateFunc) *Result {
	_ = r.ParseForm()

	// Parse all form values
	formValues := make(map[string]string, len(r.Form))
	for key, vals := range r.Form {
		if len(vals) > 0 {
			formValues[key] = vals[0]
		}
	}

	// Populate field configs with submitted values
	def.PopulateValues(formValues)

	// Determine validation type
	valType := ValidationSubmit
	triggerField := ""
	if IsFieldValidation(r) {
		valType = ValidationFieldChange
		triggerField = r.Header.Get("HX-Trigger-Name")
	}

	// Determine which fields to validate
	var fieldsToValidate []string
	switch valType {
	case ValidationSubmit:
		for name := range def.Fields {
			fieldsToValidate = append(fieldsToValidate, name)
		}
		sort.Strings(fieldsToValidate) // deterministic order
	case ValidationFieldChange:
		fieldsToValidate = append(fieldsToValidate, triggerField)
		fieldsToValidate = append(fieldsToValidate, def.Dependents(triggerField)...)
	}

	// Call the hook for each field
	result := &Result{Valid: true, AllFields: def.Fields}
	for _, name := range fieldsToValidate {
		fdef, ok := def.Fields[name]
		if !ok {
			continue
		}

		ctx := ValidationContext{
			Type:         valType,
			TriggerField: triggerField,
			FormValues:   formValues,
		}
		// If this field is being validated due to dependency, set the type
		if valType == ValidationFieldChange && name != triggerField {
			ctx.Type = ValidationDependency
		}

		if !validate(ctx, name, fdef.FieldGroup) {
			result.Valid = false
		}

		if name == triggerField {
			result.Primary = fdef
		} else if valType == ValidationFieldChange {
			result.Dependents = append(result.Dependents, fdef)
		}
	}

	return result
}

// IsFieldValidation returns true if the request is a field-level validation
// (as opposed to a full form submit).
func IsFieldValidation(r *http.Request) bool {
	_ = r.ParseForm()
	return r.FormValue("X-Goshtoso-Validation") == "field"
}

// PopulateValues fills field configs with values from the form submission.
func (fd *FormDef) PopulateValues(values map[string]string) {
	for name, fdef := range fd.Fields {
		val := values[name]
		fg := fdef.FieldGroup

		if fg.Input != nil {
			fg.Input.Value = val
		}
		if fg.Textarea != nil {
			fg.Textarea.Value = val
		}
		if fg.Toggle != nil {
			fg.Toggle.Checked = val == "on"
		}
		if fg.Checkbox != nil {
			fg.Checkbox.Checked = val == "on"
		}
		// Combobox, TagsList, StructuredInput, etc. can be added as needed
	}
}
