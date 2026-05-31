package chatbubble

import "github.com/a-h/templ"

// Side determines bubble alignment and color.
type Side string

const (
	Received Side = "received" // left-aligned, neutral bubble
	Sent     Side = "sent"     // right-aligned, primary bubble
	Auto     Side = "auto"     // alignment decided at runtime via data-mine (live rooms); renders received by default
)

// Status is the optional delivery status shown under a sent/own bubble.
type Status string

const (
	StatusNone      Status = ""
	StatusSending   Status = "Sending"
	StatusDelivered Status = "Delivered"
	StatusSeen      Status = "Seen"
)

// Config holds configuration for the chat bubble component.
type Config struct {
	// Side determines alignment + color (received/sent/auto).
	Side Side
	// Message is the bubble text.
	Message string
	// SenderName is shown above the bubble (hidden when Grouped).
	SenderName string
	// Timestamp is shown next to the sender name, e.g. "11:32 AM".
	Timestamp string
	// Status is shown under sent/own bubbles, e.g. "Delivered".
	Status Status
	// AvatarSrc is an optional avatar image URL.
	AvatarSrc string
	// AvatarInitials is the initials fallback when no image is set.
	AvatarInitials string
	// AvatarVariant is an avatar.Variant token (e.g. "info"); used when no Src.
	AvatarVariant string
	// ShowAvatar renders the avatar column.
	ShowAvatar bool
	// Grouped marks a consecutive message: tighten spacing, hide avatar + name.
	Grouped bool
	// Sender, for Side=Auto, is emitted as data-sender for client mine-detection.
	Sender string
	// IsBot renders a small BOT badge by the name.
	IsBot bool
	// Class allows additional CSS classes on the row.
	Class string
	// Attrs allows arbitrary HTML attributes on the row element (e.g. hx-*).
	Attrs templ.Attributes
}

// IsMine reports whether the bubble is statically "mine" (sent).
// Auto is resolved client-side, so it is not statically mine.
func (cfg Config) IsMine() bool {
	return cfg.Side == Sent
}

// DataMine returns the static data-mine value for the row.
func (cfg Config) DataMine() string {
	if cfg.IsMine() {
		return "true"
	}
	return "false"
}

// RowClasses returns the classes for the flex row that carries data-mine.
// A `group` so the bubble can react via group-data-[mine=true]: variants.
// flex-row-reverse on mine flips both the avatar column and bubble alignment.
func (cfg Config) RowClasses() string {
	base := "group flex w-full items-end gap-2 justify-start" +
		" data-[mine=true]:flex-row-reverse data-[mine=true]:justify-end"
	if cfg.Grouped {
		base += " mt-0.5"
	} else {
		base += " mt-4"
	}
	if cfg.Class != "" {
		base += " " + cfg.Class
	}
	return base
}

// BubbleClasses returns the classes for the message bubble.
// Received look is the default; the group-data-[mine=true]: variants supply
// the sent look. Tails flip with data-mine: received tail bottom-left
// (rounded-es-sm), sent tail bottom-right (rounded-ee-sm).
func (cfg Config) BubbleClasses() string {
	return "max-w-[75%] px-4 py-2 text-sm break-words" +
		" bg-surface-alt text-on-surface dark:bg-surface-dark-alt dark:text-on-surface-dark" +
		" rounded-2xl rounded-es-sm" +
		" group-data-[mine=true]:bg-primary group-data-[mine=true]:text-on-primary" +
		" group-data-[mine=true]:dark:bg-primary-dark group-data-[mine=true]:dark:text-on-primary-dark" +
		" group-data-[mine=true]:rounded-es-2xl group-data-[mine=true]:rounded-ee-sm"
}

// HasAvatar reports whether the avatar column should render.
func (cfg Config) HasAvatar() bool {
	return cfg.ShowAvatar && !cfg.Grouped
}

// HasHeader reports whether the sender header line should render.
func (cfg Config) HasHeader() bool {
	return cfg.SenderName != "" && !cfg.Grouped
}
