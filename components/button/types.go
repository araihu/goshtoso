package button

// Variant represents button style variants
type Variant string

const (
	Primary   Variant = "primary"
	Secondary Variant = "secondary"
	Alternate Variant = "alternate"
	Inverse   Variant = "inverse"
	Info      Variant = "info"
	Danger    Variant = "danger"
	Warning   Variant = "warning"
	Success   Variant = "success"
)

// Size represents button sizes
type Size string

const (
	SizeSmall  Size = "sm"
	SizeMedium Size = "md"
	SizeLarge  Size = "lg"
	SizeXLarge Size = "xl"
)

// HTMXConfig holds HTMX attributes for server-side interactions
type HTMXConfig struct {
	// Get is the URL for an HTMX GET request.
	Get string
	// Post is the URL for an HTMX POST request.
	Post string
	// Put is the URL for an HTMX PUT request.
	Put string
	// Delete is the URL for an HTMX DELETE request.
	Delete string
	// Patch is the URL for an HTMX PATCH request.
	Patch string
	// Target is the CSS selector for the element to swap the response into.
	Target string
	// Swap is the htmx swap strategy (e.g. "innerHTML", "outerHTML").
	Swap string
	// Trigger is the htmx trigger that initiates the request.
	Trigger string
	// Indicator is the CSS selector for the loading indicator element.
	Indicator string
	// PushURL pushes the request URL to the browser history.
	PushURL bool
	// Confirm is a confirmation message shown before the request is sent.
	Confirm string
	// Vals is additional values to submit with the request as JSON.
	Vals string
}

// AlpineConfig holds Alpine.js directives for client-side interactions
type AlpineConfig struct {
	// OnClick is the Alpine x-on:click expression.
	OnClick string
	// BindDisabled is the x-bind:disabled expression.
	BindDisabled string
	// Show is the x-show expression controlling visibility.
	Show string
	// Transition enables x-transition on the element.
	Transition bool
	// Data is the x-data expression for component state.
	Data string
}

// Config holds all configuration for a Button component
type Config struct {
	// Variant is the visual style variant of the button.
	Variant Variant
	// Size is the button size, defaults to SizeMedium.
	Size Size
	// Type is the HTML button type attribute (e.g. "button", "submit").
	Type string
	// Disabled disables the button, preventing interaction.
	Disabled bool
	// ID is the HTML id attribute for the button element.
	ID string
	// Class is additional CSS classes appended to the button.
	RootClass string
	// HTMX is an optional HTMX config for server-side interaction.
	HTMX *HTMXConfig
	// Alpine is an optional Alpine.js config for client-side behavior.
	Alpine *AlpineConfig
	// LoadingText is the text shown during an HTMX request.
	LoadingText string
}

// variantClasses returns the Tailwind utility classes for a variant
func variantClasses(variant Variant) string {
	switch variant {
	case Primary:
		// Black background, white text
		return "bg-primary text-on-primary border-primary dark:bg-primary-dark dark:text-on-primary-dark dark:border-primary-dark"
	case Secondary:
		// Dark gray background, white text
		return "bg-secondary text-on-secondary border-secondary dark:bg-secondary-dark dark:text-on-secondary-dark dark:border-secondary-dark"
	case Alternate:
		// Light gray background, dark text
		return "bg-surface-alt text-on-surface-strong border-surface-alt dark:bg-surface-dark-alt dark:text-on-surface-dark-strong dark:border-surface-dark-alt"
	case Inverse:
		// Black background (in light mode), white in dark mode
		return "bg-surface-dark text-on-surface-dark border-surface-dark dark:bg-surface dark:text-on-surface dark:border-surface"
	case Info:
		// Sky blue background, white text
		return "bg-info text-on-info border-info dark:bg-info-dark dark:text-on-info-dark dark:border-info-dark"
	case Danger:
		// Red background, white text
		return "bg-danger text-on-danger border-danger dark:bg-danger-dark dark:text-on-danger-dark dark:border-danger-dark"
	case Warning:
		// Yellow/amber background, dark text
		return "bg-warning text-on-warning border-warning dark:bg-warning-dark dark:text-on-warning-dark dark:border-warning-dark"
	case Success:
		// Green background, dark text
		return "bg-success text-on-success border-success dark:bg-success-dark dark:text-on-success-dark dark:border-success-dark"
	default:
		return "bg-primary text-on-primary border-primary"
	}
}

// sizeClasses returns the size-specific classes
func sizeClasses(size Size) string {
	switch size {
	case SizeSmall:
		return "px-4 py-2 text-xs"
	case SizeLarge:
		return "px-4 py-2 text-base"
	case SizeXLarge:
		return "px-4 py-2 text-lg"
	default: // Medium
		return "px-4 py-2 text-sm"
	}
}

// buttonClasses returns all CSS classes for the button
func buttonClasses(cfg Config) string {
	// Base classes
	base := "whitespace-nowrap rounded-2xl font-medium tracking-wide transition hover:opacity-75 text-center focus-visible:outline-2 focus-visible:outline-offset-2 active:opacity-100 active:outline-offset-0 disabled:opacity-75 disabled:cursor-not-allowed border"

	// Variant classes
	variant := variantClasses(cfg.Variant)

	// Size classes
	size := sizeClasses(cfg.Size)

	// Outline color based on variant
	outline := "focus-visible:outline-" + string(cfg.Variant)

	return base + " " + variant + " " + size + " " + outline + " " + cfg.RootClass
}
