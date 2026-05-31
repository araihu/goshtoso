package toast

import (
	"fmt"
	"strings"
	"sync/atomic"
)

var idCounter atomic.Int64

// uniqueID generates a unique ID for server-rendered toasts
func uniqueID() int64 {
	return idCounter.Add(1)
}

// Variant represents toast notification style variants
type Variant string

const (
	Info    Variant = "info"
	Success Variant = "success"
	Warning Variant = "warning"
	Danger  Variant = "danger"
	Message Variant = "message"
)

// Sender represents the avatar and name shown in notification-style toasts.
type Sender struct {
	// Name is the sender display name.
	Name string
	// Avatar is the sender avatar image URL.
	Avatar string
}

// Config holds configuration for a single toast notification.
// Used for server-side rendered toasts (including HTMX OOB swaps).
type Config struct {
	// Variant determines the color scheme and icon
	Variant Variant
	// Title is the notification heading (not used for Message variant)
	Title string
	// Message is the notification body text
	Message string
	// Sender is used only for the Message variant
	Sender *Sender
	// DisplayDuration in milliseconds (default 8000)
	DisplayDuration int
	// ActionText, when set, renders an inline action button in the toast (e.g.
	// "Undo"). Clicking it fires the configured HTMX request and dismisses the
	// toast. The close (dismiss) button is always present regardless.
	ActionText string
	// ActionHxGet / ActionHxPost is the HTMX request URL the action button fires.
	ActionHxGet  string
	ActionHxPost string
	// ActionHxTarget / ActionHxSwap configure the action button's HTMX swap.
	ActionHxTarget string
	ActionHxSwap   string
}

// HasAction reports whether the toast should render an inline action button.
func (cfg Config) HasAction() bool { return cfg.ActionText != "" }

// ContainerConfig holds configuration for the toast container.
// The container is the fixed-position wrapper that holds stacking notifications.
type ContainerConfig struct {
	// ID is the element ID for HTMX OOB targeting (default "toast-container")
	ID string
	// DisplayDuration in milliseconds (default 8000)
	DisplayDuration int
}

// effectiveDuration returns the display duration, defaulting to 8000ms
func (cfg Config) effectiveDuration() int {
	if cfg.DisplayDuration > 0 {
		return cfg.DisplayDuration
	}
	return 8000
}

// effectiveID returns the container ID, defaulting to "toast-container"
func (cfg ContainerConfig) effectiveID() string {
	if cfg.ID != "" {
		return cfg.ID
	}
	return "toast-container"
}

// effectiveDuration returns the display duration, defaulting to 8000ms
func (cfg ContainerConfig) effectiveDuration() int {
	if cfg.DisplayDuration > 0 {
		return cfg.DisplayDuration
	}
	return 8000
}

// BorderClass returns the border color class for the variant
func (cfg Config) BorderClass() string {
	switch cfg.Variant {
	case Info:
		return "border-info"
	case Success:
		return "border-success"
	case Warning:
		return "border-warning"
	case Danger:
		return "border-danger"
	case Message:
		return "border-outline dark:border-outline-dark"
	default:
		return "border-info"
	}
}

// BgClass returns the inner background class for the variant
func (cfg Config) BgClass() string {
	switch cfg.Variant {
	case Info:
		return "bg-info/10"
	case Success:
		return "bg-success/10"
	case Warning:
		return "bg-warning/10"
	case Danger:
		return "bg-danger/10"
	case Message:
		return "bg-surface-alt dark:bg-surface-dark-alt"
	default:
		return "bg-info/10"
	}
}

// IconBgClass returns the icon badge background class
func (cfg Config) IconBgClass() string {
	switch cfg.Variant {
	case Info:
		return "bg-info/15 text-info"
	case Success:
		return "bg-success/15 text-success"
	case Warning:
		return "bg-warning/15 text-warning"
	case Danger:
		return "bg-danger/15 text-danger"
	default:
		return "bg-info/15 text-info"
	}
}

// TitleClass returns the title text color class
func (cfg Config) TitleClass() string {
	switch cfg.Variant {
	case Info:
		return "text-info"
	case Success:
		return "text-success"
	case Warning:
		return "text-warning"
	case Danger:
		return "text-danger"
	default:
		return "text-info"
	}
}

// containerAlpineData returns the Alpine.js x-data for the toast container
func containerAlpineData(cfg ContainerConfig) string {
	return fmt.Sprintf(`{
        notifications: [],
        displayDuration: %d,

        addNotification(data) {
            var id = Date.now();
            var notification = { id: id, variant: data.variant || 'info', sender: data.sender || null, title: data.title || null, message: data.message || null };

            if (this.notifications.length >= 20) {
                this.notifications.splice(0, this.notifications.length - 19);
            }

            this.notifications.push(notification);
        },
        removeNotification(id) {
            setTimeout(() => {
                this.notifications = this.notifications.filter(
                    (notification) => notification.id !== id
                );
            }, 400);
        }
    }`, cfg.effectiveDuration())
}

// singleToastAlpineData returns the Alpine.js x-data for an individual toast item
func singleToastAlpineData(duration int) string {
	return fmt.Sprintf(`{
        isVisible: false,
        timeout: null,
        init() {
            this.$nextTick(() => { this.isVisible = true });
            this.timeout = setTimeout(() => { this.isVisible = false; this.$dispatch('toast-dismiss', { id: this.$el.dataset.toastId }); }, %d);
        }
    }`, duration)
}

// jsEscapeSingle escapes single quotes and backslashes for safe JS string embedding
func jsEscapeSingle(s string) string {
	var result strings.Builder
	for _, c := range s {
		switch c {
		case '\'':
			result.WriteString(`\'`)
		case '\\':
			result.WriteString(`\\`)
		default:
			result.WriteString(string(c))
		}
	}
	return result.String()
}
