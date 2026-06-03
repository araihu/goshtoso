package form

import (
	"fmt"

	"github.com/araihu/goshtoso/components/checkbox"
	"github.com/araihu/goshtoso/components/combobox"
	"github.com/araihu/goshtoso/components/fileinput"
	"github.com/araihu/goshtoso/components/structuredinput"
	"github.com/araihu/goshtoso/components/tagslist"
	"github.com/araihu/goshtoso/components/textarea"
	"github.com/araihu/goshtoso/components/textinput"
	"github.com/araihu/goshtoso/components/toggle"
)

// Config holds the form configuration
type Config struct {
	// ID is the HTML form ID (required for external submit via form="..." attribute)
	ID string
	// Action is the POST action URL for native form submission
	Action string
	// Method is the HTTP method ("post" default, "get", "dialog")
	Method string
	// Class allows additional CSS classes on the form element
	Class string
	// HTMX enables HTMX-based submission (alternative to native Action)
	HTMX *HTMXConfig
	// PreventEnterSubmit prevents Enter key from submitting the form.
	// Default true — set to false to allow Enter submission.
	PreventEnterSubmit *bool
	// Footer renders Cancel + Submit buttons at the bottom.
	// Nil = no footer (useful for modal forms where the modal provides buttons).
	Footer *FooterConfig
}

// shouldPreventEnter returns true if enter-key submission should be blocked
func (c Config) shouldPreventEnter() bool {
	if c.PreventEnterSubmit == nil {
		return true // default: prevent
	}
	return *c.PreventEnterSubmit
}

// getMethod returns the form method, defaulting to "post"
func (c Config) getMethod() string {
	if c.Method == "" {
		return "post"
	}
	return c.Method
}

// HTMXConfig configures HTMX-based form submission
type HTMXConfig struct {
	Post     string // hx-post
	Get      string // hx-get
	Put      string // hx-put
	Delete   string // hx-delete
	Target   string // hx-target
	Swap     string // hx-swap
	Encoding string // hx-encoding (e.g. "multipart/form-data")
}

// FooterConfig configures the form footer with action buttons
type FooterConfig struct {
	// SubmitText is the submit button label (e.g. "Create", "Save")
	SubmitText string
	// CancelText is the cancel button label (e.g. "Cancel")
	CancelText string
	// CancelHref is the cancel link URL (plain navigation)
	CancelHref string
	// CancelHTMX enables HTMX-powered cancel (SPA navigation). Overrides CancelHref when set.
	CancelHTMX *CancelHTMXConfig
	// SubmitDisabled is an Alpine.js expression for x-bind:disabled on submit
	SubmitDisabled string
	// Sticky makes the footer stick to the bottom of the viewport while scrolling (default: false)
	Sticky bool
}

// CancelHTMXConfig configures HTMX attributes on the cancel link for SPA navigation.
type CancelHTMXConfig struct {
	Get     string // hx-get
	Target  string // hx-target
	Swap    string // hx-swap
	PushURL bool   // hx-push-url
}

// footerClasses returns the CSS classes for the footer container
func (c FooterConfig) footerClasses() string {
	base := "flex justify-end gap-3 mt-6 pt-4 border-t border-outline dark:border-outline-dark"
	if c.Sticky {
		base += " sticky bottom-0 bg-surface dark:bg-surface-dark pb-2"
	}
	return base
}

// SectionConfig holds configuration for a regular form section
type SectionConfig struct {
	// ID is the section element ID (used for HTMX targeting)
	ID string
	// Title is the section heading
	Title string
	// Class allows additional CSS classes
	Class string
	// OOB enables hx-swap-oob="true" for HTMX out-of-band updates
	OOB bool
	// Columns controls the grid layout: "1" for single column, "2" (default) for responsive 2-column
	Columns string
}

// gridClasses returns the CSS grid classes based on column config
func (c SectionConfig) gridClasses() string {
	if c.Columns == "1" {
		return "grid grid-cols-1 gap-x-6 gap-y-5"
	}
	return "grid grid-cols-1 md:grid-cols-2 items-start gap-x-6 gap-y-6"
}

// CollapsibleSectionConfig extends SectionConfig for accordion-style sections
type CollapsibleSectionConfig struct {
	SectionConfig
	// Collapsed sets the initial state (true = starts collapsed)
	Collapsed bool
	// Summary is the text shown when collapsed (e.g. "Using defaults")
	Summary string
}

// alpineData returns the Alpine x-data string
func (c CollapsibleSectionConfig) alpineData() string {
	if c.Collapsed {
		return "{ isExpanded: false }"
	}
	return "{ isExpanded: true }"
}

// FlipSectionConfig configures a flip-card section (read-only front / editable back)
type FlipSectionConfig struct {
	SectionConfig
	// Flipped starts in edit mode if true (default: false = read-only mode)
	Flipped bool
	// EditLabel is the button text to enter edit mode (default: "Edit")
	EditLabel string
	// DoneLabel is the button text to exit edit mode (default: "Done")
	DoneLabel string
}

// getEditLabel returns the edit button label with default
func (c FlipSectionConfig) getEditLabel() string {
	if c.EditLabel == "" {
		return "Edit"
	}
	return c.EditLabel
}

// getDoneLabel returns the done button label with default
func (c FlipSectionConfig) getDoneLabel() string {
	if c.DoneLabel == "" {
		return "Done"
	}
	return c.DoneLabel
}

// alpineData returns the Alpine x-data string
func (c FlipSectionConfig) alpineData() string {
	return fmt.Sprintf("{ isEditing: %t }", c.Flipped)
}

// SubSectionConfig configures a nested subsection within a section
type SubSectionConfig struct {
	// ID is the subsection element ID
	ID string
	// Title is the subsection heading
	Title string
	// Class allows additional CSS classes
	Class string
	// Columns controls the grid layout: "1" for single column, "2" (default)
	Columns string
}

// gridClasses returns the CSS grid classes
func (c SubSectionConfig) gridClasses() string {
	if c.Columns == "1" {
		return "grid grid-cols-1 gap-x-6 gap-y-5"
	}
	return "grid grid-cols-1 md:grid-cols-2 items-start gap-x-6 gap-y-6"
}

// FieldGroupConfig configures a field wrapper with label, errors, and hints.
// Set one of the built-in field types (Input, Select, Combobox, etc.) to render
// a Goshtoso component automatically. If none are set, uses { children... }.
type FieldGroupConfig struct {
	// ID is the field ID (used for label's "for" attribute)
	ID string
	// Label is the field label text
	Label string
	// Required shows a red asterisk next to the label
	Required bool
	// Errors are validation error messages displayed below the field
	Errors []string
	// Hints are helper text messages displayed below errors
	Hints []string
	// Class allows additional CSS classes on the wrapper
	Class string
	// Validation enables HTMX-based field validation
	Validation *ValidationConfig

	// Built-in Goshtoso field types (mutually exclusive — first non-nil wins).
	// If none are set, FieldGroup renders { children... } instead.
	Input           *textinput.Config
	Combobox        *combobox.Config
	Textarea        *textarea.Config
	Toggle          *toggle.Config
	Checkbox        *checkbox.Config
	TagsList        *tagslist.Config
	StructuredInput *structuredinput.Config
	FileInput       *fileinput.Config

	// OOB enables hx-swap-oob="true" on the wrapper div for out-of-band HTMX updates.
	OOB bool
	// Meta holds validation metadata rendered as data-* attributes.
	// Set by validation.FormDef.Bind(); do not set directly.
	Meta *FieldMeta
}

// FieldMeta holds metadata embedded as data-* attributes on the FieldGroup wrapper div.
// Used by the validation utility to identify fields and their dependencies.
type FieldMeta struct {
	// FormID identifies which FormDef to use for reconstruction.
	FormID string
	// FieldName is the canonical field name (matches the <input name="...">).
	FieldName string
	// DependsOn is a comma-separated list of field names this field depends on.
	DependsOn string
}

// ValidationConfig configures HTMX field validation
type ValidationConfig struct {
	// Endpoint is the hx-post URL for validation
	Endpoint string
	// Target is an optional hx-target (for section-level re-render)
	Target string
	// Trigger is the hx-trigger event (default: "change")
	Trigger string
}

// getTrigger returns the validation trigger with default
func (c ValidationConfig) getTrigger() string {
	if c.Trigger == "" {
		return "change"
	}
	return c.Trigger
}
