package validation

import (
	"slices"
	"strings"

	"github.com/araihu/goshtoso/components/form"
)

// Bind populates Meta and Validation on each FieldGroupConfig based on the FormDef.
// Must be called after constructing the FormDef and before rendering the form.
func (fd *FormDef) Bind() {
	for name, fdef := range fd.Fields {
		fg := fdef.FieldGroup

		// Set metadata for reconstruction
		fg.Meta = &form.FieldMeta{
			FormID:    fd.FormID,
			FieldName: name,
			DependsOn: strings.Join(fdef.DependsOn, ","),
		}

		// Auto-set ID if empty (required for OOB swaps)
		if fg.ID == "" {
			fg.ID = "goshtoso-field-" + name
		}

		// Set validation config for OnChange fields
		if fdef.OnChange && fg.Validation == nil {
			endpoint := fd.Endpoint
			fg.Validation = &form.ValidationConfig{
				Endpoint: endpoint,
				Trigger:  "change",
			}
		}
	}
}

// Dependents returns field names that depend on the given field.
// Non-transitive: only direct dependents are returned.
func (fd *FormDef) Dependents(fieldName string) []string {
	var deps []string
	for name, fdef := range fd.Fields {
		if slices.Contains(fdef.DependsOn, fieldName) {
			deps = append(deps, name)
		}
	}
	return deps
}
