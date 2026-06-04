package fileinput

import "github.com/a-h/templ"

// Variant controls the visual style of the file input.
type Variant string

const (
	// VariantDropZone renders the drag-and-drop drop zone. This is the default.
	VariantDropZone Variant = ""
	// VariantUpload renders a compact text-input-style upload control.
	VariantUpload Variant = "upload"
)

// Config holds configuration for the file input component
type Config struct {
	// Variant controls the visual style (default: drop zone)
	Variant Variant
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

// ContainerClasses returns CSS classes for the outermost wrapper div
func (cfg Config) ContainerClasses() string {
	base := "flex w-full max-w-xl flex-col gap-1 text-center"
	if cfg.RootClass != "" {
		return base + " " + cfg.RootClass
	}
	return base
}

// UploadContainerClasses returns classes for the compact upload wrapper.
func (cfg Config) UploadContainerClasses() string {
	base := "flex w-full max-w-xl flex-col gap-1 text-left text-on-surface dark:text-on-surface-dark"
	if cfg.RootClass != "" {
		return base + " " + cfg.RootClass
	}
	return base
}

// LabelClasses returns CSS classes for the label text above the drop zone
func (cfg Config) LabelClasses() string {
	return "w-fit pl-0.5 text-sm text-on-surface dark:text-on-surface-dark"
}

// DropZoneClasses returns the static (non-dynamic) CSS classes for the drop zone
func (cfg Config) DropZoneClasses() string {
	base := "flex w-full flex-col items-center justify-center gap-2 rounded-radius border border-dashed p-8 text-on-surface dark:text-on-surface-dark"
	if cfg.Disabled {
		return base + " opacity-50 cursor-not-allowed"
	}
	return base
}

// BrowseLabelClasses returns CSS classes for the "Browse" label link
func (cfg Config) BrowseLabelClasses() string {
	return "font-medium text-primary group-focus-within:underline dark:text-primary-dark cursor-pointer"
}

// IsUpload returns true when the compact upload variant should render.
func (cfg Config) IsUpload() bool {
	return cfg.Variant == VariantUpload
}

// UploadControlClasses returns classes for the compact upload control.
func (cfg Config) UploadControlClasses() string {
	base := "flex w-full items-stretch overflow-hidden rounded-radius border border-outline bg-surface-alt text-sm text-on-surface focus-within:outline-2 focus-within:outline-offset-2 focus-within:outline-primary dark:border-outline-dark dark:bg-surface-dark-alt/50 dark:text-on-surface-dark dark:focus-within:outline-primary-dark"
	if cfg.Disabled {
		return base + " cursor-not-allowed opacity-75"
	}
	return base + " cursor-pointer"
}

// UploadFileNameClasses returns classes for the compact file-name display.
func (cfg Config) UploadFileNameClasses() string {
	return "flex min-w-0 flex-1 items-center px-2 py-2 text-on-surface-muted dark:text-on-surface-dark-muted"
}

// UploadButtonClasses returns classes for the compact Browse affordance.
func (cfg Config) UploadButtonClasses() string {
	base := "flex shrink-0 items-center border-l border-outline bg-surface px-3 py-2 font-medium text-primary dark:border-outline-dark dark:bg-surface-dark dark:text-primary-dark"
	if cfg.Disabled {
		return base + " cursor-not-allowed"
	}
	return base
}

// HelperTextClasses returns classes for helper text below the field.
func (cfg Config) HelperTextClasses() string {
	return "pl-0.5 text-xs text-on-surface/60 dark:text-on-surface-dark/60"
}
