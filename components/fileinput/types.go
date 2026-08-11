package fileinput

import (
	"maps"
	"strings"

	"github.com/a-h/templ"
)

// Appearance controls the visual treatment of the file input.
type Appearance string

const (
	// AppearanceDropZone renders the drag-and-drop drop zone. This is the default.
	AppearanceDropZone Appearance = ""
	// AppearanceUpload renders a compact text-input-style upload control.
	AppearanceUpload Appearance = "upload"
)

// Config holds configuration for the file input component
type Config struct {
	// Appearance controls the visual treatment (default: drop zone).
	Appearance Appearance
	// ID is the HTML id for the file input element
	ID string
	// Name is the form field name
	Name string
	// Label text displayed above the drop zone (e.g. "Cover Picture")
	Label string
	// Accept restricts file types (e.g. "image/*", ".pdf,.doc")
	Accept string
	// HelperText shown below the drop zone (e.g. "PNG, JPG, WebP - Max 5MB")
	HelperText string
	// Required marks the input as required
	Required bool
	// Disabled disables the input
	Disabled bool
	// RootClass allows additional CSS classes on the outer container.
	RootClass string
	// InputAttrs are extra attributes applied to the <input> element (e.g. hx-post, x-on:change).
	InputAttrs templ.Attributes
}

func (cfg Config) inputAttributes() templ.Attributes {
	attrs := make(templ.Attributes, len(cfg.InputAttrs)+1)
	maps.Copy(attrs, cfg.InputAttrs)
	if cfg.HelperText != "" && cfg.ID != "" {
		existing, _ := attrs["aria-describedby"].(string)
		attrs["aria-describedby"] = mergeAttributeTokens(existing, cfg.ID+"-helper")
	}
	return attrs
}

func mergeAttributeTokens(values ...string) string {
	seen := map[string]bool{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		for token := range strings.FieldsSeq(value) {
			if !seen[token] {
				seen[token] = true
				result = append(result, token)
			}
		}
	}
	return strings.Join(result, " ")
}

// ContainerClasses returns CSS classes for the outermost wrapper div
func (cfg Config) containerClasses() string {
	base := "flex w-full max-w-xl flex-col gap-1 text-center"
	if cfg.RootClass != "" {
		return base + " " + cfg.RootClass
	}
	return base
}

// UploadContainerClasses returns classes for the compact upload wrapper.
func (cfg Config) uploadContainerClasses() string {
	base := "flex w-full max-w-xl flex-col gap-1 text-left text-on-surface dark:text-on-surface-dark"
	if cfg.RootClass != "" {
		return base + " " + cfg.RootClass
	}
	return base
}

// LabelClasses returns CSS classes for the label text above the drop zone
func (cfg Config) labelClasses() string {
	return "w-fit pl-0.5 text-sm text-on-surface dark:text-on-surface-dark"
}

// DropZoneClasses returns the static (non-dynamic) CSS classes for the drop zone
func (cfg Config) dropZoneClasses() string {
	base := "flex w-full flex-col items-center justify-center gap-2 rounded-radius border border-dashed p-8 text-on-surface dark:text-on-surface-dark"
	if cfg.Disabled {
		return base + " [&>*]:opacity-50 cursor-not-allowed"
	}
	return base
}

// BrowseLabelClasses returns CSS classes for the "Browse" label link
func (cfg Config) browseLabelClasses() string {
	return "font-medium text-primary group-focus-within:underline dark:text-primary-dark cursor-pointer"
}

// IsUpload returns true when the compact upload appearance should render.
func (cfg Config) isUpload() bool {
	return cfg.Appearance == AppearanceUpload
}

// UploadControlClasses returns classes for the compact upload control.
func (cfg Config) uploadControlClasses() string {
	base := "flex w-full items-stretch overflow-hidden rounded-radius border border-control-outline bg-surface-alt text-sm text-on-surface focus-within:outline-2 focus-within:outline-offset-2 focus-within:outline-primary dark:border-control-outline-dark dark:bg-surface-dark-alt/50 dark:text-on-surface-dark dark:focus-within:outline-primary-dark"
	if cfg.Disabled {
		return base + " cursor-not-allowed"
	}
	return base + " cursor-pointer"
}

// UploadFileNameClasses returns classes for the compact file-name display.
func (cfg Config) uploadFileNameClasses() string {
	base := "flex min-w-0 flex-1 items-center px-2 py-2 text-on-surface-muted dark:text-on-surface-dark-muted"
	if cfg.Disabled {
		return base + " opacity-75"
	}
	return base
}

// UploadButtonClasses returns classes for the compact Browse affordance.
func (cfg Config) uploadButtonClasses() string {
	base := "flex shrink-0 items-center border-l border-outline bg-surface px-3 py-2 font-medium text-primary dark:border-outline-dark dark:bg-surface-dark dark:text-primary-dark"
	if cfg.Disabled {
		return base + " cursor-not-allowed opacity-75"
	}
	return base
}

// HelperTextClasses returns classes for helper text below the field.
func (cfg Config) helperTextClasses() string {
	return "pl-0.5 text-xs text-on-surface-muted dark:text-on-surface-dark-muted"
}
