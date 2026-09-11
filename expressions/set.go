// Package expressions supplies application-owned UI text to Goshtoso renders.
// It does not select languages, load translations, or maintain global defaults.
package expressions

import "strconv"

// Set groups optional expressions by component. Empty strings and nil functions
// inherit the next layer. Treat sets and captured function state as immutable
// when sharing them between concurrent renders.
type Set struct {
	ActionGroup     ActionGroup
	Alert           Alert
	Avatar          Avatar
	Badge           Badge
	Banner          Banner
	Breadcrumbs     Breadcrumbs
	Carousel        Carousel
	ChatBubble      ChatBubble
	CodeBlock       CodeBlock
	Combobox        Combobox
	Drawer          Drawer
	Dropdown        Dropdown
	FileInput       FileInput
	Form            Form
	Modal           Modal
	Navbar          Navbar
	Pagination      Pagination
	Palette         Palette
	Popover         Popover
	Rating          Rating
	SchemaForm      SchemaForm
	SchemaTree      SchemaTree
	Search          Search
	Select          Select
	Sidebar         Sidebar
	Skeleton        Skeleton
	SplitButton     SplitButton
	Steps           Steps
	StructuredInput StructuredInput
	Table           Table
	Tabs            Tabs
	TagsList        TagsList
	Textarea        Textarea
	TextInput       TextInput
	Toast           Toast
	Toolbar         Toolbar
	Tooltip         Tooltip
	AppShell        AppShell
	EmptyState      EmptyState
}

// ActionGroup contains optional expressions for the ActionGroup component family.
type ActionGroup struct {
	// Label defaults to "Actions".
	Label string
	// OverflowLabel defaults to "More actions".
	OverflowLabel string
}

// Alert contains optional expressions for the Alert component family.
type Alert struct {
	// DismissLabel defaults to "dismiss alert".
	DismissLabel string
	// DismissActionLabel defaults to "Dismiss".
	DismissActionLabel string
}

// Avatar contains optional expressions for the Avatar component family.
type Avatar struct {
	// GroupLabel defaults to "Avatar group".
	GroupLabel string
}

// Badge contains optional expressions for the Badge component family.
type Badge struct {
	// NotificationLabel defaults to "notification".
	NotificationLabel string
}

// Banner contains optional expressions for the Banner component family.
type Banner struct {
	// DismissLabel defaults to "dismiss banner".
	DismissLabel string
	// ConsentLabel defaults to "Cookie consent".
	ConsentLabel string
	// Title defaults to "Cookie Consent".
	Title string
	// AcceptLabel defaults to "Accept".
	AcceptLabel string
	// RejectLabel defaults to "Decline".
	RejectLabel string
}

// Breadcrumbs contains optional expressions for the Breadcrumbs component family.
type Breadcrumbs struct {
	// Label defaults to "breadcrumb".
	Label string
}

// Carousel contains optional expressions for the Carousel component family.
type Carousel struct {
	// LoadingText defaults to "Loading carousel...".
	LoadingText string
	// PreviousLabel defaults to "previous slide".
	PreviousLabel string
	// NextLabel defaults to "next slide".
	NextLabel string
	// SlidesLabel defaults to "slides".
	SlidesLabel string
	// PauseLabel defaults to "pause carousel".
	PauseLabel string
	// PlayLabel defaults to "play carousel".
	PlayLabel string
	// SlideLabel formats a complete message.
	SlideLabel func(int) string
}

// ChatBubble contains optional expressions for the ChatBubble component family.
type ChatBubble struct {
	// TypingLabel defaults to "typing".
	TypingLabel string
	// SendingLabel defaults to "Sending".
	SendingLabel string
	// DeliveredLabel defaults to "Delivered".
	DeliveredLabel string
	// SeenLabel defaults to "Seen".
	SeenLabel string
	// BotLabel defaults to "BOT".
	BotLabel string
}

// CodeBlock contains optional expressions for the CodeBlock component family.
type CodeBlock struct {
	// CopyLabel defaults to "Copy".
	CopyLabel string
	// CopiedLabel defaults to "Copied!".
	CopiedLabel string
	// ErrorText defaults to "Unable to copy".
	ErrorText string
	// CopyAriaLabel formats a complete message.
	CopyAriaLabel func(string) string
}

// Combobox contains optional expressions for the Combobox component family.
type Combobox struct {
	// Placeholder defaults to "Select…".
	Placeholder string
	// ClearLabel defaults to "Clear all".
	ClearLabel string
	// SearchPlaceholder defaults to "Search…".
	SearchPlaceholder string
	// EmptyText defaults to "No matches found".
	EmptyText string
	// ErrorText defaults to "Failed to load.".
	ErrorText string
	// RetryLabel defaults to "Retry".
	RetryLabel string
	// SelectedLabel formats a complete message.
	SelectedLabel func(int) string
}

// Drawer contains optional expressions for the Drawer component family.
type Drawer struct {
	// CloseLabel defaults to "Close".
	CloseLabel string
}

// Dropdown contains optional expressions for the Dropdown component family.
type Dropdown struct {
	// ContextMenuLabel defaults to "context menu".
	ContextMenuLabel string
	// OpenMenuLabel defaults to "open menu".
	OpenMenuLabel string
}

// FileInput contains optional expressions for the FileInput component family.
type FileInput struct {
	// BrowseLabel defaults to "Browse".
	BrowseLabel string
	// DropText defaults to " or drag and drop here".
	DropText string
	// EmptyText defaults to "No file selected".
	EmptyText string
}

// Form contains optional expressions for the Form component family.
type Form struct {
	// EditLabel defaults to "Edit".
	EditLabel string
	// DoneLabel defaults to "Done".
	DoneLabel string
	// ErrorTitle defaults to "Validation failed".
	ErrorTitle string
}

// Modal contains optional expressions for the Modal component family.
type Modal struct {
	// CloseLabel defaults to "close modal".
	CloseLabel string
	// DialogCloseLabel defaults to "Close dialog".
	DialogCloseLabel string
}

// Navbar contains optional expressions for the Navbar component family.
type Navbar struct {
	// Label defaults to "main navigation".
	Label string
	// OpenMenuLabel defaults to "Open mobile menu".
	OpenMenuLabel string
	// CloseMenuLabel defaults to "Close mobile menu".
	CloseMenuLabel string
	// UserMenuLabel defaults to "user menu".
	UserMenuLabel string
	// SecondaryLabel defaults to "secondary navigation".
	SecondaryLabel string
}

// Pagination contains optional expressions for the Pagination component family.
type Pagination struct {
	// Label defaults to "pagination".
	Label string
	// PreviousLabel defaults to "previous page".
	PreviousLabel string
	// NextLabel defaults to "next page".
	NextLabel string
	// MoreLabel defaults to "more pages".
	MoreLabel string
	// PreviousText defaults to "Previous".
	PreviousText string
	// NextText defaults to "Next".
	NextText string
	// PageLabel formats a complete message.
	PageLabel func(int) string
}

// Palette contains optional expressions for the Palette component family.
type Palette struct {
	// Placeholder defaults to "Pick a color".
	Placeholder string
	// ResetLabel defaults to "Reset".
	ResetLabel string
	// CustomLabel defaults to "Custom color".
	CustomLabel string
	// WhiteLabel defaults to "white".
	WhiteLabel string
	// BlackLabel defaults to "black".
	BlackLabel string
}

// Popover contains optional expressions for the Popover component family.
type Popover struct {
	// TriggerLabel defaults to "Open".
	TriggerLabel string
}

// Rating contains optional expressions for the Rating component family.
type Rating struct {
	// Label defaults to "Rating".
	Label string
	// VeryDissatisfiedLabel defaults to "very dissatisfied".
	VeryDissatisfiedLabel string
	// DissatisfiedLabel defaults to "dissatisfied".
	DissatisfiedLabel string
	// NeutralLabel defaults to "neutral".
	NeutralLabel string
	// SatisfiedLabel defaults to "satisfied".
	SatisfiedLabel string
	// VerySatisfiedLabel defaults to "very satisfied".
	VerySatisfiedLabel string
	// StarLabel formats a complete message.
	StarLabel func(int) string
}

// SchemaForm contains optional expressions for the SchemaForm component family.
type SchemaForm struct {
	// RequiredLabel defaults to "required".
	RequiredLabel string
	// ManagedTitle defaults to "Managed by platform — cannot override".
	ManagedTitle string
	// ManagedLabel defaults to "managed".
	ManagedLabel string
	// ItemPlaceholder defaults to "item".
	ItemPlaceholder string
	// AddLabel defaults to "+ Add item".
	AddLabel string
	// DefaultLabel formats a complete message.
	DefaultLabel func(string) string
}

// SchemaTree contains optional expressions for the SchemaTree component family.
type SchemaTree struct {
	// RequiredLabel defaults to "required".
	RequiredLabel string
	// OptionalLabel defaults to "optional".
	OptionalLabel string
	// NullableLabel defaults to "nullable".
	NullableLabel string
	// DeprecatedLabel defaults to "deprecated".
	DeprecatedLabel string
}

// Search contains optional expressions for the Search component family.
type Search struct {
	// Label defaults to "Search".
	Label string
	// Placeholder defaults to "Search".
	Placeholder string
	// ShortcutText defaults to "⌘ K".
	ShortcutText string
	// EscapeText defaults to "Esc".
	EscapeText string
	// EmptyText defaults to "No results found.".
	EmptyText string
	// ResultsLabel formats a complete message.
	ResultsLabel func(string) string
}

// Select contains optional expressions for the Select component family.
type Select struct {
	// Placeholder defaults to "Please Select".
	Placeholder string
	// ListLabel formats a complete message.
	ListLabel func(string) string
}

// Sidebar contains optional expressions for the Sidebar component family.
type Sidebar struct {
	// Label defaults to "sidebar navigation".
	Label string
	// SearchLabel defaults to "Search".
	SearchLabel string
	// SkipLabel defaults to "skip to the main content".
	SkipLabel string
	// ActiveLabel defaults to "active".
	ActiveLabel string
	// TriggerLabel defaults to "Open sidebar".
	TriggerLabel string
}

// Skeleton contains optional expressions for the Skeleton component family.
type Skeleton struct {
	// Label defaults to "Loading content".
	Label string
}

// SplitButton contains optional expressions for the SplitButton component family.
type SplitButton struct {
	// MenuLabel defaults to "More actions".
	MenuLabel string
}

// Steps contains optional expressions for the Steps component family.
type Steps struct {
	// Label defaults to "progress".
	Label string
	// CompletedLabel defaults to "completed".
	CompletedLabel string
}

// StructuredInput contains optional expressions for the StructuredInput component family.
type StructuredInput struct {
	// AddLabel defaults to "Add row".
	AddLabel string
	// RemoveLabel defaults to "Remove row".
	RemoveLabel string
}

// Table contains optional expressions for the Table component family.
type Table struct {
	// LoadingText defaults to "Loading...".
	LoadingText string
	// ActionsLabel defaults to "Actions".
	ActionsLabel string
	// LoadMoreLabel defaults to "Load more".
	LoadMoreLabel string
	// FiltersLabel defaults to "Filters".
	FiltersLabel string
	// OpenRowLabel formats a complete message.
	OpenRowLabel func(string) string
}

// Tabs contains optional expressions for the Tabs component family.
type Tabs struct {
	// OptionsLabel defaults to "tab options".
	OptionsLabel string
	// LoadingText defaults to "Loading...".
	LoadingText string
}

// TagsList contains optional expressions for the TagsList component family.
type TagsList struct {
	// AddLabel defaults to "Add".
	AddLabel string
	// Placeholder defaults to "Add a tag...".
	Placeholder string
	// RemoveLabel defaults to "Remove tag".
	RemoveLabel string
}

// Textarea contains optional expressions for the Textarea component family.
type Textarea struct {
	// SendLabel defaults to "send".
	SendLabel string
	// EmojiLabel defaults to "Emojis".
	EmojiLabel string
	// AttachLabel defaults to "Attach a file".
	AttachLabel string
	// VoiceLabel defaults to "Send voice".
	VoiceLabel string
	// SendText defaults to "Send".
	SendText string
}

// TextInput contains optional expressions for the TextInput component family.
type TextInput struct {
	// ShowPasswordLabel defaults to "Show password".
	ShowPasswordLabel string
	// SearchLabel defaults to "search".
	SearchLabel string
}

// Toast contains optional expressions for the Toast component family.
type Toast struct {
	// DismissLabel defaults to "Dismiss".
	DismissLabel string
	// DismissAriaLabel defaults to "dismiss notification".
	DismissAriaLabel string
}

// Toolbar contains optional expressions for the Toolbar component family.
type Toolbar struct {
	// Label defaults to "Page tools".
	Label string
}

// Tooltip contains optional expressions for the Tooltip component family.
type Tooltip struct {
	// TriggerLabel defaults to "Hover Me".
	TriggerLabel string
}

// AppShell contains optional expressions for the AppShell component family.
type AppShell struct {
	// SkipLinkLabel defaults to "Skip to main content".
	SkipLinkLabel string
}

// EmptyState contains optional expressions for the EmptyState component family.
type EmptyState struct {
	// Title defaults to "Nothing here yet".
	Title string
	// Description defaults to "Items will appear here when they are available.".
	Description string
}

// English returns a fresh set of the built-in English defaults.
func English() Set {
	return Set{
		ActionGroup: ActionGroup{
			Label:         "Actions",
			OverflowLabel: "More actions",
		},
		Alert: Alert{
			DismissLabel:       "dismiss alert",
			DismissActionLabel: "Dismiss",
		},
		Avatar: Avatar{
			GroupLabel: "Avatar group",
		},
		Badge: Badge{
			NotificationLabel: "notification",
		},
		Banner: Banner{
			DismissLabel: "dismiss banner",
			ConsentLabel: "Cookie consent",
			Title:        "Cookie Consent",
			AcceptLabel:  "Accept",
			RejectLabel:  "Decline",
		},
		Breadcrumbs: Breadcrumbs{
			Label: "breadcrumb",
		},
		Carousel: Carousel{
			LoadingText:   "Loading carousel...",
			PreviousLabel: "previous slide",
			NextLabel:     "next slide",
			SlidesLabel:   "slides",
			PauseLabel:    "pause carousel",
			PlayLabel:     "play carousel",
			SlideLabel:    func(n int) string { return "slide " + strconv.Itoa(n) },
		},
		ChatBubble: ChatBubble{
			TypingLabel:    "typing",
			SendingLabel:   "Sending",
			DeliveredLabel: "Delivered",
			SeenLabel:      "Seen",
			BotLabel:       "BOT",
		},
		CodeBlock: CodeBlock{
			CopyLabel:     "Copy",
			CopiedLabel:   "Copied!",
			ErrorText:     "Unable to copy",
			CopyAriaLabel: func(label string) string { return "Copy " + label + " code" },
		},
		Combobox: Combobox{
			Placeholder:       "Select…",
			ClearLabel:        "Clear all",
			SearchPlaceholder: "Search…",
			EmptyText:         "No matches found",
			ErrorText:         "Failed to load.",
			RetryLabel:        "Retry",
			SelectedLabel:     func(n int) string { return strconv.Itoa(n) + " selected" },
		},
		Drawer: Drawer{
			CloseLabel: "Close",
		},
		Dropdown: Dropdown{
			ContextMenuLabel: "context menu",
			OpenMenuLabel:    "open menu",
		},
		FileInput: FileInput{
			BrowseLabel: "Browse",
			DropText:    " or drag and drop here",
			EmptyText:   "No file selected",
		},
		Form: Form{
			EditLabel:  "Edit",
			DoneLabel:  "Done",
			ErrorTitle: "Validation failed",
		},
		Modal: Modal{
			CloseLabel:       "close modal",
			DialogCloseLabel: "Close dialog",
		},
		Navbar: Navbar{
			Label:          "main navigation",
			OpenMenuLabel:  "Open mobile menu",
			CloseMenuLabel: "Close mobile menu",
			UserMenuLabel:  "user menu",
			SecondaryLabel: "secondary navigation",
		},
		Pagination: Pagination{
			Label:         "pagination",
			PreviousLabel: "previous page",
			NextLabel:     "next page",
			MoreLabel:     "more pages",
			PreviousText:  "Previous",
			NextText:      "Next",
			PageLabel:     func(n int) string { return "page " + strconv.Itoa(n) },
		},
		Palette: Palette{
			Placeholder: "Pick a color",
			ResetLabel:  "Reset",
			CustomLabel: "Custom color",
			WhiteLabel:  "white",
			BlackLabel:  "black",
		},
		Popover: Popover{
			TriggerLabel: "Open",
		},
		Rating: Rating{
			Label:                 "Rating",
			VeryDissatisfiedLabel: "very dissatisfied",
			DissatisfiedLabel:     "dissatisfied",
			NeutralLabel:          "neutral",
			SatisfiedLabel:        "satisfied",
			VerySatisfiedLabel:    "very satisfied",
			StarLabel: func(n int) string {
				if n == 1 {
					return "one star"
				}
				return strconv.Itoa(n) + " stars"
			},
		},
		SchemaForm: SchemaForm{
			RequiredLabel:   "required",
			ManagedTitle:    "Managed by platform — cannot override",
			ManagedLabel:    "managed",
			ItemPlaceholder: "item",
			AddLabel:        "+ Add item",
			DefaultLabel:    func(value string) string { return "default: " + value },
		},
		SchemaTree: SchemaTree{
			RequiredLabel:   "required",
			OptionalLabel:   "optional",
			NullableLabel:   "nullable",
			DeprecatedLabel: "deprecated",
		},
		Search: Search{
			Label:        "Search",
			Placeholder:  "Search",
			ShortcutText: "⌘ K",
			EscapeText:   "Esc",
			EmptyText:    "No results found.",
			ResultsLabel: func(label string) string { return label + " results" },
		},
		Select: Select{
			Placeholder: "Please Select",
			ListLabel:   func(label string) string { return label + " list" },
		},
		Sidebar: Sidebar{
			Label:        "sidebar navigation",
			SearchLabel:  "Search",
			SkipLabel:    "skip to the main content",
			ActiveLabel:  "active",
			TriggerLabel: "Open sidebar",
		},
		Skeleton: Skeleton{
			Label: "Loading content",
		},
		SplitButton: SplitButton{
			MenuLabel: "More actions",
		},
		Steps: Steps{
			Label:          "progress",
			CompletedLabel: "completed",
		},
		StructuredInput: StructuredInput{
			AddLabel:    "Add row",
			RemoveLabel: "Remove row",
		},
		Table: Table{
			LoadingText:   "Loading...",
			ActionsLabel:  "Actions",
			LoadMoreLabel: "Load more",
			FiltersLabel:  "Filters",
			OpenRowLabel: func(id string) string {
				if id == "" {
					return "Open row"
				}
				return "Open row " + id
			},
		},
		Tabs: Tabs{
			OptionsLabel: "tab options",
			LoadingText:  "Loading...",
		},
		TagsList: TagsList{
			AddLabel:    "Add",
			Placeholder: "Add a tag...",
			RemoveLabel: "Remove tag",
		},
		Textarea: Textarea{
			SendLabel:   "send",
			EmojiLabel:  "Emojis",
			AttachLabel: "Attach a file",
			VoiceLabel:  "Send voice",
			SendText:    "Send",
		},
		TextInput: TextInput{
			ShowPasswordLabel: "Show password",
			SearchLabel:       "search",
		},
		Toast: Toast{
			DismissLabel:     "Dismiss",
			DismissAriaLabel: "dismiss notification",
		},
		Toolbar: Toolbar{
			Label: "Page tools",
		},
		Tooltip: Tooltip{
			TriggerLabel: "Hover Me",
		},
		AppShell: AppShell{
			SkipLinkLabel: "Skip to main content",
		},
		EmptyState: EmptyState{
			Title:       "Nothing here yet",
			Description: "Items will appear here when they are available.",
		},
	}
}

// Merge overlays nonempty expressions onto base, without modifying either set.
func Merge(base, overrides Set) Set {
	base.ActionGroup = mergeActionGroup(base.ActionGroup, overrides.ActionGroup)
	base.Alert = mergeAlert(base.Alert, overrides.Alert)
	base.Avatar = mergeAvatar(base.Avatar, overrides.Avatar)
	base.Badge = mergeBadge(base.Badge, overrides.Badge)
	base.Banner = mergeBanner(base.Banner, overrides.Banner)
	base.Breadcrumbs = mergeBreadcrumbs(base.Breadcrumbs, overrides.Breadcrumbs)
	base.Carousel = mergeCarousel(base.Carousel, overrides.Carousel)
	base.ChatBubble = mergeChatBubble(base.ChatBubble, overrides.ChatBubble)
	base.CodeBlock = mergeCodeBlock(base.CodeBlock, overrides.CodeBlock)
	base.Combobox = mergeCombobox(base.Combobox, overrides.Combobox)
	base.Drawer = mergeDrawer(base.Drawer, overrides.Drawer)
	base.Dropdown = mergeDropdown(base.Dropdown, overrides.Dropdown)
	base.FileInput = mergeFileInput(base.FileInput, overrides.FileInput)
	base.Form = mergeForm(base.Form, overrides.Form)
	base.Modal = mergeModal(base.Modal, overrides.Modal)
	base.Navbar = mergeNavbar(base.Navbar, overrides.Navbar)
	base.Pagination = mergePagination(base.Pagination, overrides.Pagination)
	base.Palette = mergePalette(base.Palette, overrides.Palette)
	base.Popover = mergePopover(base.Popover, overrides.Popover)
	base.Rating = mergeRating(base.Rating, overrides.Rating)
	base.SchemaForm = mergeSchemaForm(base.SchemaForm, overrides.SchemaForm)
	base.SchemaTree = mergeSchemaTree(base.SchemaTree, overrides.SchemaTree)
	base.Search = mergeSearch(base.Search, overrides.Search)
	base.Select = mergeSelect(base.Select, overrides.Select)
	base.Sidebar = mergeSidebar(base.Sidebar, overrides.Sidebar)
	base.Skeleton = mergeSkeleton(base.Skeleton, overrides.Skeleton)
	base.SplitButton = mergeSplitButton(base.SplitButton, overrides.SplitButton)
	base.Steps = mergeSteps(base.Steps, overrides.Steps)
	base.StructuredInput = mergeStructuredInput(base.StructuredInput, overrides.StructuredInput)
	base.Table = mergeTable(base.Table, overrides.Table)
	base.Tabs = mergeTabs(base.Tabs, overrides.Tabs)
	base.TagsList = mergeTagsList(base.TagsList, overrides.TagsList)
	base.Textarea = mergeTextarea(base.Textarea, overrides.Textarea)
	base.TextInput = mergeTextInput(base.TextInput, overrides.TextInput)
	base.Toast = mergeToast(base.Toast, overrides.Toast)
	base.Toolbar = mergeToolbar(base.Toolbar, overrides.Toolbar)
	base.Tooltip = mergeTooltip(base.Tooltip, overrides.Tooltip)
	base.AppShell = mergeAppShell(base.AppShell, overrides.AppShell)
	base.EmptyState = mergeEmptyState(base.EmptyState, overrides.EmptyState)
	return base
}

func mergeActionGroup(base, overrides ActionGroup) ActionGroup {
	if overrides.Label != "" {
		base.Label = overrides.Label
	}
	if overrides.OverflowLabel != "" {
		base.OverflowLabel = overrides.OverflowLabel
	}
	return base
}

func mergeAlert(base, overrides Alert) Alert {
	if overrides.DismissLabel != "" {
		base.DismissLabel = overrides.DismissLabel
	}
	if overrides.DismissActionLabel != "" {
		base.DismissActionLabel = overrides.DismissActionLabel
	}
	return base
}

func mergeAvatar(base, overrides Avatar) Avatar {
	if overrides.GroupLabel != "" {
		base.GroupLabel = overrides.GroupLabel
	}
	return base
}

func mergeBadge(base, overrides Badge) Badge {
	if overrides.NotificationLabel != "" {
		base.NotificationLabel = overrides.NotificationLabel
	}
	return base
}

func mergeBanner(base, overrides Banner) Banner {
	if overrides.DismissLabel != "" {
		base.DismissLabel = overrides.DismissLabel
	}
	if overrides.ConsentLabel != "" {
		base.ConsentLabel = overrides.ConsentLabel
	}
	if overrides.Title != "" {
		base.Title = overrides.Title
	}
	if overrides.AcceptLabel != "" {
		base.AcceptLabel = overrides.AcceptLabel
	}
	if overrides.RejectLabel != "" {
		base.RejectLabel = overrides.RejectLabel
	}
	return base
}

func mergeBreadcrumbs(base, overrides Breadcrumbs) Breadcrumbs {
	if overrides.Label != "" {
		base.Label = overrides.Label
	}
	return base
}

func mergeCarousel(base, overrides Carousel) Carousel {
	if overrides.LoadingText != "" {
		base.LoadingText = overrides.LoadingText
	}
	if overrides.PreviousLabel != "" {
		base.PreviousLabel = overrides.PreviousLabel
	}
	if overrides.NextLabel != "" {
		base.NextLabel = overrides.NextLabel
	}
	if overrides.SlidesLabel != "" {
		base.SlidesLabel = overrides.SlidesLabel
	}
	if overrides.PauseLabel != "" {
		base.PauseLabel = overrides.PauseLabel
	}
	if overrides.PlayLabel != "" {
		base.PlayLabel = overrides.PlayLabel
	}
	if overrides.SlideLabel != nil {
		base.SlideLabel = overrides.SlideLabel
	}
	return base
}

func mergeChatBubble(base, overrides ChatBubble) ChatBubble {
	if overrides.TypingLabel != "" {
		base.TypingLabel = overrides.TypingLabel
	}
	if overrides.SendingLabel != "" {
		base.SendingLabel = overrides.SendingLabel
	}
	if overrides.DeliveredLabel != "" {
		base.DeliveredLabel = overrides.DeliveredLabel
	}
	if overrides.SeenLabel != "" {
		base.SeenLabel = overrides.SeenLabel
	}
	if overrides.BotLabel != "" {
		base.BotLabel = overrides.BotLabel
	}
	return base
}

func mergeCodeBlock(base, overrides CodeBlock) CodeBlock {
	if overrides.CopyLabel != "" {
		base.CopyLabel = overrides.CopyLabel
	}
	if overrides.CopiedLabel != "" {
		base.CopiedLabel = overrides.CopiedLabel
	}
	if overrides.ErrorText != "" {
		base.ErrorText = overrides.ErrorText
	}
	if overrides.CopyAriaLabel != nil {
		base.CopyAriaLabel = overrides.CopyAriaLabel
	}
	return base
}

func mergeCombobox(base, overrides Combobox) Combobox {
	if overrides.Placeholder != "" {
		base.Placeholder = overrides.Placeholder
	}
	if overrides.ClearLabel != "" {
		base.ClearLabel = overrides.ClearLabel
	}
	if overrides.SearchPlaceholder != "" {
		base.SearchPlaceholder = overrides.SearchPlaceholder
	}
	if overrides.EmptyText != "" {
		base.EmptyText = overrides.EmptyText
	}
	if overrides.ErrorText != "" {
		base.ErrorText = overrides.ErrorText
	}
	if overrides.RetryLabel != "" {
		base.RetryLabel = overrides.RetryLabel
	}
	if overrides.SelectedLabel != nil {
		base.SelectedLabel = overrides.SelectedLabel
	}
	return base
}

func mergeDrawer(base, overrides Drawer) Drawer {
	if overrides.CloseLabel != "" {
		base.CloseLabel = overrides.CloseLabel
	}
	return base
}

func mergeDropdown(base, overrides Dropdown) Dropdown {
	if overrides.ContextMenuLabel != "" {
		base.ContextMenuLabel = overrides.ContextMenuLabel
	}
	if overrides.OpenMenuLabel != "" {
		base.OpenMenuLabel = overrides.OpenMenuLabel
	}
	return base
}

func mergeFileInput(base, overrides FileInput) FileInput {
	if overrides.BrowseLabel != "" {
		base.BrowseLabel = overrides.BrowseLabel
	}
	if overrides.DropText != "" {
		base.DropText = overrides.DropText
	}
	if overrides.EmptyText != "" {
		base.EmptyText = overrides.EmptyText
	}
	return base
}

func mergeForm(base, overrides Form) Form {
	if overrides.EditLabel != "" {
		base.EditLabel = overrides.EditLabel
	}
	if overrides.DoneLabel != "" {
		base.DoneLabel = overrides.DoneLabel
	}
	if overrides.ErrorTitle != "" {
		base.ErrorTitle = overrides.ErrorTitle
	}
	return base
}

func mergeModal(base, overrides Modal) Modal {
	if overrides.CloseLabel != "" {
		base.CloseLabel = overrides.CloseLabel
	}
	if overrides.DialogCloseLabel != "" {
		base.DialogCloseLabel = overrides.DialogCloseLabel
	}
	return base
}

func mergeNavbar(base, overrides Navbar) Navbar {
	if overrides.Label != "" {
		base.Label = overrides.Label
	}
	if overrides.OpenMenuLabel != "" {
		base.OpenMenuLabel = overrides.OpenMenuLabel
	}
	if overrides.CloseMenuLabel != "" {
		base.CloseMenuLabel = overrides.CloseMenuLabel
	}
	if overrides.UserMenuLabel != "" {
		base.UserMenuLabel = overrides.UserMenuLabel
	}
	if overrides.SecondaryLabel != "" {
		base.SecondaryLabel = overrides.SecondaryLabel
	}
	return base
}

func mergePagination(base, overrides Pagination) Pagination {
	if overrides.Label != "" {
		base.Label = overrides.Label
	}
	if overrides.PreviousLabel != "" {
		base.PreviousLabel = overrides.PreviousLabel
	}
	if overrides.NextLabel != "" {
		base.NextLabel = overrides.NextLabel
	}
	if overrides.MoreLabel != "" {
		base.MoreLabel = overrides.MoreLabel
	}
	if overrides.PreviousText != "" {
		base.PreviousText = overrides.PreviousText
	}
	if overrides.NextText != "" {
		base.NextText = overrides.NextText
	}
	if overrides.PageLabel != nil {
		base.PageLabel = overrides.PageLabel
	}
	return base
}

func mergePalette(base, overrides Palette) Palette {
	if overrides.Placeholder != "" {
		base.Placeholder = overrides.Placeholder
	}
	if overrides.ResetLabel != "" {
		base.ResetLabel = overrides.ResetLabel
	}
	if overrides.CustomLabel != "" {
		base.CustomLabel = overrides.CustomLabel
	}
	if overrides.WhiteLabel != "" {
		base.WhiteLabel = overrides.WhiteLabel
	}
	if overrides.BlackLabel != "" {
		base.BlackLabel = overrides.BlackLabel
	}
	return base
}

func mergePopover(base, overrides Popover) Popover {
	if overrides.TriggerLabel != "" {
		base.TriggerLabel = overrides.TriggerLabel
	}
	return base
}

func mergeRating(base, overrides Rating) Rating {
	if overrides.Label != "" {
		base.Label = overrides.Label
	}
	if overrides.VeryDissatisfiedLabel != "" {
		base.VeryDissatisfiedLabel = overrides.VeryDissatisfiedLabel
	}
	if overrides.DissatisfiedLabel != "" {
		base.DissatisfiedLabel = overrides.DissatisfiedLabel
	}
	if overrides.NeutralLabel != "" {
		base.NeutralLabel = overrides.NeutralLabel
	}
	if overrides.SatisfiedLabel != "" {
		base.SatisfiedLabel = overrides.SatisfiedLabel
	}
	if overrides.VerySatisfiedLabel != "" {
		base.VerySatisfiedLabel = overrides.VerySatisfiedLabel
	}
	if overrides.StarLabel != nil {
		base.StarLabel = overrides.StarLabel
	}
	return base
}

func mergeSchemaForm(base, overrides SchemaForm) SchemaForm {
	if overrides.RequiredLabel != "" {
		base.RequiredLabel = overrides.RequiredLabel
	}
	if overrides.ManagedTitle != "" {
		base.ManagedTitle = overrides.ManagedTitle
	}
	if overrides.ManagedLabel != "" {
		base.ManagedLabel = overrides.ManagedLabel
	}
	if overrides.ItemPlaceholder != "" {
		base.ItemPlaceholder = overrides.ItemPlaceholder
	}
	if overrides.AddLabel != "" {
		base.AddLabel = overrides.AddLabel
	}
	if overrides.DefaultLabel != nil {
		base.DefaultLabel = overrides.DefaultLabel
	}
	return base
}

func mergeSchemaTree(base, overrides SchemaTree) SchemaTree {
	if overrides.RequiredLabel != "" {
		base.RequiredLabel = overrides.RequiredLabel
	}
	if overrides.OptionalLabel != "" {
		base.OptionalLabel = overrides.OptionalLabel
	}
	if overrides.NullableLabel != "" {
		base.NullableLabel = overrides.NullableLabel
	}
	if overrides.DeprecatedLabel != "" {
		base.DeprecatedLabel = overrides.DeprecatedLabel
	}
	return base
}

func mergeSearch(base, overrides Search) Search {
	if overrides.Label != "" {
		base.Label = overrides.Label
	}
	if overrides.Placeholder != "" {
		base.Placeholder = overrides.Placeholder
	}
	if overrides.ShortcutText != "" {
		base.ShortcutText = overrides.ShortcutText
	}
	if overrides.EscapeText != "" {
		base.EscapeText = overrides.EscapeText
	}
	if overrides.EmptyText != "" {
		base.EmptyText = overrides.EmptyText
	}
	if overrides.ResultsLabel != nil {
		base.ResultsLabel = overrides.ResultsLabel
	}
	return base
}

func mergeSelect(base, overrides Select) Select {
	if overrides.Placeholder != "" {
		base.Placeholder = overrides.Placeholder
	}
	if overrides.ListLabel != nil {
		base.ListLabel = overrides.ListLabel
	}
	return base
}

func mergeSidebar(base, overrides Sidebar) Sidebar {
	if overrides.Label != "" {
		base.Label = overrides.Label
	}
	if overrides.SearchLabel != "" {
		base.SearchLabel = overrides.SearchLabel
	}
	if overrides.SkipLabel != "" {
		base.SkipLabel = overrides.SkipLabel
	}
	if overrides.ActiveLabel != "" {
		base.ActiveLabel = overrides.ActiveLabel
	}
	if overrides.TriggerLabel != "" {
		base.TriggerLabel = overrides.TriggerLabel
	}
	return base
}

func mergeSkeleton(base, overrides Skeleton) Skeleton {
	if overrides.Label != "" {
		base.Label = overrides.Label
	}
	return base
}

func mergeSplitButton(base, overrides SplitButton) SplitButton {
	if overrides.MenuLabel != "" {
		base.MenuLabel = overrides.MenuLabel
	}
	return base
}

func mergeSteps(base, overrides Steps) Steps {
	if overrides.Label != "" {
		base.Label = overrides.Label
	}
	if overrides.CompletedLabel != "" {
		base.CompletedLabel = overrides.CompletedLabel
	}
	return base
}

func mergeStructuredInput(base, overrides StructuredInput) StructuredInput {
	if overrides.AddLabel != "" {
		base.AddLabel = overrides.AddLabel
	}
	if overrides.RemoveLabel != "" {
		base.RemoveLabel = overrides.RemoveLabel
	}
	return base
}

func mergeTable(base, overrides Table) Table {
	if overrides.LoadingText != "" {
		base.LoadingText = overrides.LoadingText
	}
	if overrides.ActionsLabel != "" {
		base.ActionsLabel = overrides.ActionsLabel
	}
	if overrides.LoadMoreLabel != "" {
		base.LoadMoreLabel = overrides.LoadMoreLabel
	}
	if overrides.FiltersLabel != "" {
		base.FiltersLabel = overrides.FiltersLabel
	}
	if overrides.OpenRowLabel != nil {
		base.OpenRowLabel = overrides.OpenRowLabel
	}
	return base
}

func mergeTabs(base, overrides Tabs) Tabs {
	if overrides.OptionsLabel != "" {
		base.OptionsLabel = overrides.OptionsLabel
	}
	if overrides.LoadingText != "" {
		base.LoadingText = overrides.LoadingText
	}
	return base
}

func mergeTagsList(base, overrides TagsList) TagsList {
	if overrides.AddLabel != "" {
		base.AddLabel = overrides.AddLabel
	}
	if overrides.Placeholder != "" {
		base.Placeholder = overrides.Placeholder
	}
	if overrides.RemoveLabel != "" {
		base.RemoveLabel = overrides.RemoveLabel
	}
	return base
}

func mergeTextarea(base, overrides Textarea) Textarea {
	if overrides.SendLabel != "" {
		base.SendLabel = overrides.SendLabel
	}
	if overrides.EmojiLabel != "" {
		base.EmojiLabel = overrides.EmojiLabel
	}
	if overrides.AttachLabel != "" {
		base.AttachLabel = overrides.AttachLabel
	}
	if overrides.VoiceLabel != "" {
		base.VoiceLabel = overrides.VoiceLabel
	}
	if overrides.SendText != "" {
		base.SendText = overrides.SendText
	}
	return base
}

func mergeTextInput(base, overrides TextInput) TextInput {
	if overrides.ShowPasswordLabel != "" {
		base.ShowPasswordLabel = overrides.ShowPasswordLabel
	}
	if overrides.SearchLabel != "" {
		base.SearchLabel = overrides.SearchLabel
	}
	return base
}

func mergeToast(base, overrides Toast) Toast {
	if overrides.DismissLabel != "" {
		base.DismissLabel = overrides.DismissLabel
	}
	if overrides.DismissAriaLabel != "" {
		base.DismissAriaLabel = overrides.DismissAriaLabel
	}
	return base
}

func mergeToolbar(base, overrides Toolbar) Toolbar {
	if overrides.Label != "" {
		base.Label = overrides.Label
	}
	return base
}

func mergeTooltip(base, overrides Tooltip) Tooltip {
	if overrides.TriggerLabel != "" {
		base.TriggerLabel = overrides.TriggerLabel
	}
	return base
}

func mergeAppShell(base, overrides AppShell) AppShell {
	if overrides.SkipLinkLabel != "" {
		base.SkipLinkLabel = overrides.SkipLinkLabel
	}
	return base
}

func mergeEmptyState(base, overrides EmptyState) EmptyState {
	if overrides.Title != "" {
		base.Title = overrides.Title
	}
	if overrides.Description != "" {
		base.Description = overrides.Description
	}
	return base
}
