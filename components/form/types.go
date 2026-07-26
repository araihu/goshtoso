package form

import (
	"fmt"
	"strings"

	"github.com/a-h/templ"
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

// Config holds the form configuration
type Config struct {
	// ID is the HTML form ID (required for external submit via form="..." attribute)
	ID string
	// Action is the POST action URL for native form submission
	Action string
	// Method is the HTTP method ("post" default, "get", "dialog")
	Method string
	// RootClass allows additional CSS classes on the form element.
	RootClass string
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
	// SubmitLabel is the submit button label (e.g. "Create", "Save")
	SubmitLabel string
	// CancelLabel is the cancel button label (e.g. "Cancel")
	CancelLabel string
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
	// RootClass allows additional CSS classes on the wrapper.
	RootClass string
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
	// RootClass allows additional CSS classes on the root.
	RootClass string
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
	// ID preserves the historical field-wrapper ID used by CSS, HTMX, and OOB
	// targets. Built-in controls derive a collision-free ID when their own ID is
	// empty. Use FocusTargetID after construction or validation binding when an
	// error summary needs the real focusable control ID.
	ID string
	// RootID overrides the wrapper ID used for HTMX targeting. When empty,
	// FieldGroup preserves ID as the historical wrapper target or derives a
	// wrapper from an explicitly identified built-in control.
	RootID string
	// Label is the field label text
	Label string
	// Required marks the field visually and exposes accessible required state.
	// Native required validation is also enabled where the built-in control
	// supports it; composite collection controls still require server validation.
	Required bool
	// Errors are validation error messages displayed below the field
	Errors []string
	// Hints are helper text messages displayed below errors
	Hints []string
	// RootClass allows additional CSS classes on the wrapper.
	RootClass string
	// Validation enables HTMX-based field validation
	Validation *ValidationConfig

	// Built-in Goshtoso field types (mutually exclusive — first non-nil wins).
	// If none are set, FieldGroup renders { children... } instead.
	Input           *textinput.Config
	Combobox        *combobox.Config
	Select          *selectfield.Config
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

func (c FieldGroupConfig) controlID() string {
	if id := c.explicitControlID(); id != "" {
		if c.RootID == "" && c.ID != "" && id == c.ID {
			return c.derivedControlID()
		}
		return id
	}
	if c.ID != "" && c.hasBuiltIn() {
		return c.derivedControlID()
	}
	return c.ID
}

func (c FieldGroupConfig) derivedControlID() string {
	if c.Combobox != nil || c.TagsList != nil || c.StructuredInput != nil {
		return c.ID + "-control"
	}
	return c.ID + "-input"
}

func (c FieldGroupConfig) explicitControlID() string {
	switch {
	case c.Input != nil && c.Input.ID != "":
		return c.Input.ID
	case c.Combobox != nil && c.Combobox.ID != "":
		return c.Combobox.ID
	case c.Select != nil && c.Select.ID != "":
		return c.Select.ID
	case c.Textarea != nil && c.Textarea.ID != "":
		return c.Textarea.ID
	case c.Toggle != nil && c.Toggle.ID != "":
		return c.Toggle.ID
	case c.Checkbox != nil && c.Checkbox.ID != "":
		return c.Checkbox.ID
	case c.TagsList != nil && c.TagsList.ID != "":
		return c.TagsList.ID
	case c.StructuredInput != nil && c.StructuredInput.ID != "":
		return c.StructuredInput.ID
	case c.FileInput != nil && c.FileInput.ID != "":
		return c.FileInput.ID
	default:
		return ""
	}
}

func (c FieldGroupConfig) hasBuiltIn() bool {
	return c.Input != nil || c.Combobox != nil || c.Select != nil || c.Textarea != nil ||
		c.Toggle != nil || c.Checkbox != nil || c.TagsList != nil ||
		c.StructuredInput != nil || c.FileInput != nil
}

func (c FieldGroupConfig) rootID() string {
	if c.RootID != "" {
		return c.RootID
	}
	// ID has historically named the swappable field wrapper. Keep that contract
	// for CSS, HTMX, OOB, and validation consumers.
	if c.ID != "" {
		return c.ID
	}
	if id := c.controlID(); id != "" {
		return id + "-field"
	}
	return ""
}

func (c FieldGroupConfig) labelTargetID() string {
	id := c.controlID()
	if id == "" || c.StructuredInput != nil {
		return ""
	}
	if c.Combobox != nil || c.Select != nil {
		return id + "-trigger"
	}
	if c.TagsList != nil {
		return id + "-input"
	}
	return id
}

// FocusTargetID returns the actual focus target for labels and linked form
// error summaries. It accounts for composite controls whose public ID names a
// component root rather than its trigger or text entry. An empty result means
// the caller must provide a target for custom children.
func (c FieldGroupConfig) FocusTargetID() string {
	if c.StructuredInput != nil {
		return c.controlID()
	}
	return c.labelTargetID()
}

func (c FieldGroupConfig) labelID() string {
	if id := c.controlID(); id != "" {
		return id + "-label"
	}
	return ""
}

func (c FieldGroupConfig) errorsID() string {
	if len(c.Errors) == 0 || c.controlID() == "" {
		return ""
	}
	return c.controlID() + "-errors"
}

func (c FieldGroupConfig) hintsID() string {
	if len(c.Hints) == 0 || c.controlID() == "" {
		return ""
	}
	return c.controlID() + "-hints"
}

func (c FieldGroupConfig) describedBy() string {
	return strings.Join(nonEmpty(c.errorsID(), c.hintsID()), " ")
}

func (c FieldGroupConfig) fieldAttrs(existing templ.Attributes) templ.Attributes {
	attrs := cloneAttributes(existing)
	if describedBy := c.describedBy(); describedBy != "" {
		attrs["aria-describedby"] = joinTokens(attributeString(attrs["aria-describedby"]), describedBy)
	}
	if len(c.Errors) > 0 {
		attrs["aria-invalid"] = "true"
	}
	if c.Required {
		attrs["aria-required"] = "true"
	}
	return attrs
}

func (c FieldGroupConfig) nativeFieldAttrs(existing templ.Attributes) templ.Attributes {
	attrs := c.fieldAttrs(existing)
	if c.Required {
		attrs["required"] = true
	}
	return attrs
}

func (c FieldGroupConfig) inputConfig() textinput.Config {
	cfg := *c.Input
	cfg.ID = c.controlID()
	if c.Label != "" {
		cfg.Label = ""
	}
	cfg.Required = cfg.Required || c.Required
	if len(c.Errors) > 0 {
		cfg.State = textinput.StateError
	}
	cfg.InputAttrs = c.fieldAttrs(cfg.InputAttrs)
	return cfg
}

func (c FieldGroupConfig) textareaConfig() textarea.Config {
	cfg := *c.Textarea
	cfg.ID = c.controlID()
	if c.Label != "" {
		cfg.Label = ""
	}
	cfg.Required = cfg.Required || c.Required
	if len(c.Errors) > 0 {
		cfg.State = textarea.StateError
	}
	cfg.InputAttrs = c.fieldAttrs(cfg.InputAttrs)
	return cfg
}

func (c FieldGroupConfig) comboboxConfig() combobox.Config {
	cfg := *c.Combobox
	cfg.ID = c.controlID()
	if c.Label != "" {
		cfg.Label = ""
	}
	cfg.Required = cfg.Required || c.Required
	cfg.TriggerAttrs = c.fieldAttrs(cfg.TriggerAttrs)
	return cfg
}

func (c FieldGroupConfig) selectConfig() selectfield.Config {
	cfg := *c.Select
	cfg.ID = c.controlID()
	if c.Label != "" {
		cfg.Label = ""
	}
	cfg.Required = cfg.Required || c.Required
	if len(c.Errors) > 0 {
		cfg.State = selectfield.StateError
	}
	cfg.TriggerAttrs = c.fieldAttrs(cfg.TriggerAttrs)
	return cfg
}

func (c FieldGroupConfig) toggleConfig() toggle.Config {
	cfg := *c.Toggle
	cfg.ID = c.controlID()
	if c.Label != "" {
		cfg.Label = ""
	}
	cfg.InputAttrs = c.nativeFieldAttrs(cfg.InputAttrs)
	return cfg
}

func (c FieldGroupConfig) checkboxConfig() checkbox.Config {
	cfg := *c.Checkbox
	cfg.ID = c.controlID()
	if c.Label != "" {
		cfg.Label = ""
	}
	cfg.InputAttrs = c.nativeFieldAttrs(cfg.InputAttrs)
	return cfg
}

func (c FieldGroupConfig) tagsListConfig() tagslist.Config {
	cfg := *c.TagsList
	cfg.ID = c.controlID()
	cfg.InputAttrs = c.fieldAttrs(cfg.InputAttrs)
	return cfg
}

func (c FieldGroupConfig) structuredInputConfig() structuredinput.Config {
	cfg := *c.StructuredInput
	cfg.ID = c.controlID()
	cfg.RootAttrs = c.fieldAttrs(cfg.RootAttrs)
	if _, ok := cfg.RootAttrs["tabindex"]; !ok {
		cfg.RootAttrs["tabindex"] = "-1"
	}
	return cfg
}

func (c FieldGroupConfig) fileInputConfig() fileinput.Config {
	cfg := *c.FileInput
	cfg.ID = c.controlID()
	if c.Label != "" {
		cfg.Label = ""
	}
	cfg.Required = cfg.Required || c.Required
	cfg.InputAttrs = c.fieldAttrs(cfg.InputAttrs)
	return cfg
}

func cloneAttributes(existing templ.Attributes) templ.Attributes {
	attrs := make(templ.Attributes, len(existing)+2)
	for key, value := range existing {
		attrs[key] = value
	}
	return attrs
}

func attributeString(value any) string {
	valueString, _ := value.(string)
	return valueString
}

func joinTokens(values ...string) string {
	seen := map[string]bool{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		for _, token := range strings.Fields(value) {
			if !seen[token] {
				seen[token] = true
				result = append(result, token)
			}
		}
	}
	return strings.Join(result, " ")
}

func nonEmpty(values ...string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value != "" {
			result = append(result, value)
		}
	}
	return result
}

// getTrigger returns the validation trigger with default
func (c ValidationConfig) getTrigger() string {
	if c.Trigger == "" {
		return "change"
	}
	return c.Trigger
}
