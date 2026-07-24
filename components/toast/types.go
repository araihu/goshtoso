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

// Tone represents a toast notification's semantic color treatment.
type Tone string

const (
	ToneInfo    Tone = "info"
	ToneSuccess Tone = "success"
	ToneWarning Tone = "warning"
	ToneDanger  Tone = "danger"
)

// Sender represents the avatar and name shown in notification-style toasts.
type Sender struct {
	// Name is the sender display name.
	Name string
	// Avatar is the sender avatar image URL.
	Avatar string
}

// HTMXConfig holds HTMX attributes for a toast action button.
type HTMXConfig struct {
	// Get is the HTMX GET request URL the action button fires.
	Get string
	// Post is the HTMX POST request URL the action button fires.
	Post string
	// Target configures the action button's HTMX swap target.
	Target string
	// Swap configures the action button's HTMX swap strategy.
	Swap string
}

// Config holds configuration for a single toast notification.
// Used for server-side rendered toasts (including HTMX OOB swaps).
type Config struct {
	// Tone determines the color scheme and icon.
	Tone Tone
	// Title is the notification heading.
	Title string
	// Message is the notification body text
	Message string
	// DisplayDuration in milliseconds (default 8000); negative keeps a server-rendered toast visible until dismissed manually.
	DisplayDuration int
	// ActionLabel, when set, renders an inline action button in the toast (e.g.
	// "Undo"). Clicking it fires the configured HTMX request and dismisses the
	// toast. The close (dismiss) button is always present regardless.
	ActionLabel string
	// ActionHTMX configures the action button's HTMX request.
	ActionHTMX *HTMXConfig
}

// MessageConfig holds configuration for a sender-oriented message toast.
type MessageConfig struct {
	// Sender is the person or system that sent the message.
	Sender Sender
	// Message is the notification body text.
	Message string
	// DisplayDuration in milliseconds (default 8000); negative keeps a server-rendered toast visible until dismissed manually.
	DisplayDuration int
	// ActionLabel, when set, renders an inline action button.
	ActionLabel string
	// ActionHTMX configures the action button's HTMX request.
	ActionHTMX *HTMXConfig
	// DismissLabel labels the text dismiss control. Defaults to "Dismiss".
	DismissLabel string
}

// HasAction reports whether the toast should render an inline action button.
func (cfg Config) HasAction() bool { return cfg.ActionLabel != "" }

func (cfg MessageConfig) hasAction() bool { return cfg.ActionLabel != "" }

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

func (cfg MessageConfig) effectiveDuration() int {
	if cfg.DisplayDuration > 0 {
		return cfg.DisplayDuration
	}
	return 8000
}

func (cfg MessageConfig) effectiveDismissLabel() string {
	if cfg.DismissLabel != "" {
		return cfg.DismissLabel
	}
	return "Dismiss"
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

// BorderClass returns the border color class for the tone.
func (cfg Config) BorderClass() string {
	switch cfg.Tone {
	case ToneInfo:
		return "border-info"
	case ToneSuccess:
		return "border-success"
	case ToneWarning:
		return "border-warning"
	case ToneDanger:
		return "border-danger"
	default:
		return "border-info"
	}
}

// BgClass returns the inner background class for the tone.
func (cfg Config) BgClass() string {
	switch cfg.Tone {
	case ToneInfo:
		return "bg-info/10"
	case ToneSuccess:
		return "bg-success/10"
	case ToneWarning:
		return "bg-warning/10"
	case ToneDanger:
		return "bg-danger/10"
	default:
		return "bg-info/10"
	}
}

// IconBgClass returns the icon badge background class
func (cfg Config) IconBgClass() string {
	switch cfg.Tone {
	case ToneInfo:
		return "bg-info/15 text-info"
	case ToneSuccess:
		return "bg-success/15 text-success"
	case ToneWarning:
		return "bg-warning/15 text-warning"
	case ToneDanger:
		return "bg-danger/15 text-danger"
	default:
		return "bg-info/15 text-info"
	}
}

// TitleClass returns the title text color class
func (cfg Config) TitleClass() string {
	switch cfg.Tone {
	case ToneInfo:
		return "text-info"
	case ToneSuccess:
		return "text-success"
	case ToneWarning:
		return "text-warning"
	case ToneDanger:
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
            var kind = data.kind === 'message-toast' ? 'message-toast' : 'toast';
            var notification = { id: id, kind: kind, tone: data.tone || 'info', sender: data.sender || null, title: data.title || null, message: data.message || null };

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
func singleToastAlpineData(duration int, persistent bool) string {
	if persistent {
		return `{ isVisible: true }`
	}
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
