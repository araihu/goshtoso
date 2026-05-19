package validation

import (
	"github.com/araihu/goshtoso/components/form"
)

// ValidationType indicates why validation is running.
type ValidationType int

const (
	ValidationSubmit      ValidationType = iota // full form submission
	ValidationFieldChange                       // a single field changed
	ValidationDependency                        // field validated because a dependency changed
)

// ValidationContext is passed to the hook for each field being validated.
type ValidationContext struct {
	Type         ValidationType
	TriggerField string            // field that started this (empty for submit)
	FormValues   map[string]string // all parsed form values
}

// ValidateFunc is the user-provided validation hook.
// Called once per field being validated. Mutate fg to set Errors, State, etc.
// Return true if the field is valid.
type ValidateFunc func(ctx ValidationContext, name string, fg *form.FieldGroupConfig) bool

// FormDef describes a form for validation purposes.
type FormDef struct {
	FormID   string
	Endpoint string               // default validation endpoint
	Fields   map[string]*FieldDef
}

// FieldDef describes a single field for validation.
type FieldDef struct {
	Name       string
	FieldGroup *form.FieldGroupConfig // pointer — hook mutates this
	DependsOn  []string               // field names this depends on
	OnChange   bool                   // enable field-level validation on change
}

// Result holds the outcome of validation.
type Result struct {
	Valid      bool
	Primary    *FieldDef
	Dependents []*FieldDef
	AllFields  map[string]*FieldDef
}
