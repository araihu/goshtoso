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
	// AriaLabel defaults to "Actions".
	AriaLabel string
	// OverflowAriaLabel defaults to "More actions".
	OverflowAriaLabel string
}

// Alert contains optional expressions for the Alert component family.
type Alert struct {
	// DismissAriaLabel defaults to "dismiss alert".
	DismissAriaLabel string
	// DismissLabel defaults to "Dismiss".
	DismissLabel string
}

// Avatar contains optional expressions for the Avatar component family.
type Avatar struct {
	// GroupAriaLabel defaults to "Avatar group".
	GroupAriaLabel string
}

// Badge contains optional expressions for the Badge component family.
type Badge struct {
	// NotificationAriaLabel defaults to "notification".
	NotificationAriaLabel string
}

// Banner contains optional expressions for the Banner component family.
type Banner struct {
	// DismissAriaLabel defaults to "dismiss banner".
	DismissAriaLabel string
	// ConsentAriaLabel defaults to "Cookie consent".
	ConsentAriaLabel string
	// Title defaults to "Cookie Consent".
	Title string
	// AcceptLabel defaults to "Accept".
	AcceptLabel string
	// RejectLabel defaults to "Decline".
	RejectLabel string
}

// Breadcrumbs contains optional expressions for the Breadcrumbs component family.
type Breadcrumbs struct {
	// AriaLabel defaults to "breadcrumb".
	AriaLabel string
}

// Carousel contains optional expressions for the Carousel component family.
type Carousel struct {
	// LoadingText defaults to "Loading carousel...".
	LoadingText string
	// PreviousAriaLabel defaults to "previous slide".
	PreviousAriaLabel string
	// NextAriaLabel defaults to "next slide".
	NextAriaLabel string
	// SlidesAriaLabel defaults to "slides".
	SlidesAriaLabel string
	// PauseAriaLabel defaults to "pause carousel".
	PauseAriaLabel string
	// PlayAriaLabel defaults to "play carousel".
	PlayAriaLabel string
	// SlideLabel returns the carousel counter text for slide. Rendering prepares values
	// from zero through the number of slides; visible slide numbers start at one.
	SlideLabel func(slide int) string
}

// ChatBubble contains optional expressions for the ChatBubble component family.
type ChatBubble struct {
	// TypingAriaLabel defaults to "typing".
	TypingAriaLabel string
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
	// CopyAriaLabel returns the copy button’s accessible name for the code block header
	// label, or its ID when no label is available.
	CopyAriaLabel func(label string) string
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
	// SelectedLabel returns the complete selection summary for count selected values. It
	// must handle zero and positive counts; client components prepare all possible counts
	// during rendering.
	SelectedLabel func(count int) string
}

// Drawer contains optional expressions for the Drawer component family.
type Drawer struct {
	// CloseAriaLabel defaults to "Close".
	CloseAriaLabel string
}

// Dropdown contains optional expressions for the Dropdown component family.
type Dropdown struct {
	// ContextMenuAriaLabel defaults to "context menu".
	ContextMenuAriaLabel string
	// OpenMenuAriaLabel defaults to "open menu".
	OpenMenuAriaLabel string
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
	// CloseAriaLabel defaults to "close modal".
	CloseAriaLabel string
	// DialogCloseAriaLabel defaults to "Close dialog".
	DialogCloseAriaLabel string
}

// Navbar contains optional expressions for the Navbar component family.
type Navbar struct {
	// AriaLabel defaults to "main navigation".
	AriaLabel string
	// OpenMenuAriaLabel defaults to "Open mobile menu".
	OpenMenuAriaLabel string
	// CloseMenuAriaLabel defaults to "Close mobile menu".
	CloseMenuAriaLabel string
	// UserMenuAriaLabel defaults to "user menu".
	UserMenuAriaLabel string
	// SecondaryAriaLabel defaults to "secondary navigation".
	SecondaryAriaLabel string
}

// Pagination contains optional expressions for the Pagination component family.
type Pagination struct {
	// AriaLabel defaults to "pagination".
	AriaLabel string
	// PreviousAriaLabel defaults to "previous page".
	PreviousAriaLabel string
	// NextAriaLabel defaults to "next page".
	NextAriaLabel string
	// MoreAriaLabel defaults to "more pages".
	MoreAriaLabel string
	// PreviousLabel defaults to "Previous".
	PreviousLabel string
	// NextLabel defaults to "Next".
	NextLabel string
	// PageAriaLabel returns the accessible name for a page link. Page numbers start at one.
	PageAriaLabel func(page int) string
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
	// StarLabel returns the complete label for a star rating value. Values include zero
	// for an unrated display and run through the configured maximum.
	StarLabel func(value int) string
}

// SchemaForm contains optional expressions for the SchemaForm component family.
type SchemaForm struct {
	// RequiredAriaLabel defaults to "required".
	RequiredAriaLabel string
	// ManagedTitle defaults to "Managed by platform — cannot override".
	ManagedTitle string
	// ManagedLabel defaults to "managed".
	ManagedLabel string
	// ItemPlaceholder defaults to "item".
	ItemPlaceholder string
	// AddLabel defaults to "+ Add item".
	AddLabel string
	// DefaultLabel returns the default-value annotation for a schema field. The argument is
	// the field’s default value as supplied by the application.
	DefaultLabel func(value string) string
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
	// ResultsAriaLabel returns the results region’s accessible name from the search
	// control’s resolved label.
	ResultsAriaLabel func(label string) string
}

// Select contains optional expressions for the Select component family.
type Select struct {
	// Placeholder defaults to "Please Select".
	Placeholder string
	// ListAriaLabel returns the option list’s accessible name from the select control’s
	// configured label, which may be empty.
	ListAriaLabel func(label string) string
}

// Sidebar contains optional expressions for the Sidebar component family.
type Sidebar struct {
	// AriaLabel defaults to "sidebar navigation".
	AriaLabel string
	// SearchAriaLabel defaults to "Search".
	SearchAriaLabel string
	// SkipLabel defaults to "skip to the main content".
	SkipLabel string
	// ActiveLabel defaults to "active".
	ActiveLabel string
	// TriggerLabel defaults to "Open sidebar".
	TriggerLabel string
}

// Skeleton contains optional expressions for the Skeleton component family.
type Skeleton struct {
	// AriaLabel defaults to "Loading content".
	AriaLabel string
}

// SplitButton contains optional expressions for the SplitButton component family.
type SplitButton struct {
	// MenuAriaLabel defaults to "More actions".
	MenuAriaLabel string
}

// Steps contains optional expressions for the Steps component family.
type Steps struct {
	// AriaLabel defaults to "progress".
	AriaLabel string
	// CompletedLabel defaults to "completed".
	CompletedLabel string
}

// StructuredInput contains optional expressions for the StructuredInput component family.
type StructuredInput struct {
	// AddLabel defaults to "Add row".
	AddLabel string
	// RemoveAriaLabel defaults to "Remove row".
	RemoveAriaLabel string
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
	// OpenRowLabel returns the screen-reader text for opening a table row. rowID is empty
	// when the row has no ID.
	OpenRowLabel func(rowID string) string
}

// Tabs contains optional expressions for the Tabs component family.
type Tabs struct {
	// OptionsAriaLabel defaults to "tab options".
	OptionsAriaLabel string
	// LoadingText defaults to "Loading...".
	LoadingText string
}

// TagsList contains optional expressions for the TagsList component family.
type TagsList struct {
	// AddLabel defaults to "Add".
	AddLabel string
	// Placeholder defaults to "Add a tag...".
	Placeholder string
	// RemoveAriaLabel defaults to "Remove tag".
	RemoveAriaLabel string
}

// Textarea contains optional expressions for the Textarea component family.
type Textarea struct {
	// SendAriaLabel defaults to "send".
	SendAriaLabel string
	// EmojiAriaLabel defaults to "Emojis".
	EmojiAriaLabel string
	// AttachAriaLabel defaults to "Attach a file".
	AttachAriaLabel string
	// VoiceAriaLabel defaults to "Send voice".
	VoiceAriaLabel string
	// SendLabel defaults to "Send".
	SendLabel string
}

// TextInput contains optional expressions for the TextInput component family.
type TextInput struct {
	// ShowPasswordAriaLabel defaults to "Show password".
	ShowPasswordAriaLabel string
	// SearchAriaLabel defaults to "search".
	SearchAriaLabel string
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
	// AriaLabel defaults to "Page tools".
	AriaLabel string
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
			AriaLabel:         "Actions",
			OverflowAriaLabel: "More actions",
		},
		Alert: Alert{
			DismissAriaLabel: "dismiss alert",
			DismissLabel:     "Dismiss",
		},
		Avatar: Avatar{
			GroupAriaLabel: "Avatar group",
		},
		Badge: Badge{
			NotificationAriaLabel: "notification",
		},
		Banner: Banner{
			DismissAriaLabel: "dismiss banner",
			ConsentAriaLabel: "Cookie consent",
			Title:            "Cookie Consent",
			AcceptLabel:      "Accept",
			RejectLabel:      "Decline",
		},
		Breadcrumbs: Breadcrumbs{
			AriaLabel: "breadcrumb",
		},
		Carousel: Carousel{
			LoadingText:       "Loading carousel...",
			PreviousAriaLabel: "previous slide",
			NextAriaLabel:     "next slide",
			SlidesAriaLabel:   "slides",
			PauseAriaLabel:    "pause carousel",
			PlayAriaLabel:     "play carousel",
			SlideLabel:        func(n int) string { return "slide " + strconv.Itoa(n) },
		},
		ChatBubble: ChatBubble{
			TypingAriaLabel: "typing",
			SendingLabel:    "Sending",
			DeliveredLabel:  "Delivered",
			SeenLabel:       "Seen",
			BotLabel:        "BOT",
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
			CloseAriaLabel: "Close",
		},
		Dropdown: Dropdown{
			ContextMenuAriaLabel: "context menu",
			OpenMenuAriaLabel:    "open menu",
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
			CloseAriaLabel:       "close modal",
			DialogCloseAriaLabel: "Close dialog",
		},
		Navbar: Navbar{
			AriaLabel:          "main navigation",
			OpenMenuAriaLabel:  "Open mobile menu",
			CloseMenuAriaLabel: "Close mobile menu",
			UserMenuAriaLabel:  "user menu",
			SecondaryAriaLabel: "secondary navigation",
		},
		Pagination: Pagination{
			AriaLabel:         "pagination",
			PreviousAriaLabel: "previous page",
			NextAriaLabel:     "next page",
			MoreAriaLabel:     "more pages",
			PreviousLabel:     "Previous",
			NextLabel:         "Next",
			PageAriaLabel:     func(n int) string { return "page " + strconv.Itoa(n) },
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
			RequiredAriaLabel: "required",
			ManagedTitle:      "Managed by platform — cannot override",
			ManagedLabel:      "managed",
			ItemPlaceholder:   "item",
			AddLabel:          "+ Add item",
			DefaultLabel:      func(value string) string { return "default: " + value },
		},
		SchemaTree: SchemaTree{
			RequiredLabel:   "required",
			OptionalLabel:   "optional",
			NullableLabel:   "nullable",
			DeprecatedLabel: "deprecated",
		},
		Search: Search{
			Label:            "Search",
			Placeholder:      "Search",
			ShortcutText:     "⌘ K",
			EscapeText:       "Esc",
			EmptyText:        "No results found.",
			ResultsAriaLabel: func(label string) string { return label + " results" },
		},
		Select: Select{
			Placeholder:   "Please Select",
			ListAriaLabel: func(label string) string { return label + " list" },
		},
		Sidebar: Sidebar{
			AriaLabel:       "sidebar navigation",
			SearchAriaLabel: "Search",
			SkipLabel:       "skip to the main content",
			ActiveLabel:     "active",
			TriggerLabel:    "Open sidebar",
		},
		Skeleton: Skeleton{
			AriaLabel: "Loading content",
		},
		SplitButton: SplitButton{
			MenuAriaLabel: "More actions",
		},
		Steps: Steps{
			AriaLabel:      "progress",
			CompletedLabel: "completed",
		},
		StructuredInput: StructuredInput{
			AddLabel:        "Add row",
			RemoveAriaLabel: "Remove row",
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
			OptionsAriaLabel: "tab options",
			LoadingText:      "Loading...",
		},
		TagsList: TagsList{
			AddLabel:        "Add",
			Placeholder:     "Add a tag...",
			RemoveAriaLabel: "Remove tag",
		},
		Textarea: Textarea{
			SendAriaLabel:   "send",
			EmojiAriaLabel:  "Emojis",
			AttachAriaLabel: "Attach a file",
			VoiceAriaLabel:  "Send voice",
			SendLabel:       "Send",
		},
		TextInput: TextInput{
			ShowPasswordAriaLabel: "Show password",
			SearchAriaLabel:       "search",
		},
		Toast: Toast{
			DismissLabel:     "Dismiss",
			DismissAriaLabel: "dismiss notification",
		},
		Toolbar: Toolbar{
			AriaLabel: "Page tools",
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
	if overrides.AriaLabel != "" {
		base.AriaLabel = overrides.AriaLabel
	}
	if overrides.OverflowAriaLabel != "" {
		base.OverflowAriaLabel = overrides.OverflowAriaLabel
	}
	return base
}

func mergeAlert(base, overrides Alert) Alert {
	if overrides.DismissAriaLabel != "" {
		base.DismissAriaLabel = overrides.DismissAriaLabel
	}
	if overrides.DismissLabel != "" {
		base.DismissLabel = overrides.DismissLabel
	}
	return base
}

func mergeAvatar(base, overrides Avatar) Avatar {
	if overrides.GroupAriaLabel != "" {
		base.GroupAriaLabel = overrides.GroupAriaLabel
	}
	return base
}

func mergeBadge(base, overrides Badge) Badge {
	if overrides.NotificationAriaLabel != "" {
		base.NotificationAriaLabel = overrides.NotificationAriaLabel
	}
	return base
}

func mergeBanner(base, overrides Banner) Banner {
	if overrides.DismissAriaLabel != "" {
		base.DismissAriaLabel = overrides.DismissAriaLabel
	}
	if overrides.ConsentAriaLabel != "" {
		base.ConsentAriaLabel = overrides.ConsentAriaLabel
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
	if overrides.AriaLabel != "" {
		base.AriaLabel = overrides.AriaLabel
	}
	return base
}

func mergeCarousel(base, overrides Carousel) Carousel {
	if overrides.LoadingText != "" {
		base.LoadingText = overrides.LoadingText
	}
	if overrides.PreviousAriaLabel != "" {
		base.PreviousAriaLabel = overrides.PreviousAriaLabel
	}
	if overrides.NextAriaLabel != "" {
		base.NextAriaLabel = overrides.NextAriaLabel
	}
	if overrides.SlidesAriaLabel != "" {
		base.SlidesAriaLabel = overrides.SlidesAriaLabel
	}
	if overrides.PauseAriaLabel != "" {
		base.PauseAriaLabel = overrides.PauseAriaLabel
	}
	if overrides.PlayAriaLabel != "" {
		base.PlayAriaLabel = overrides.PlayAriaLabel
	}
	if overrides.SlideLabel != nil {
		base.SlideLabel = overrides.SlideLabel
	}
	return base
}

func mergeChatBubble(base, overrides ChatBubble) ChatBubble {
	if overrides.TypingAriaLabel != "" {
		base.TypingAriaLabel = overrides.TypingAriaLabel
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
	if overrides.CloseAriaLabel != "" {
		base.CloseAriaLabel = overrides.CloseAriaLabel
	}
	return base
}

func mergeDropdown(base, overrides Dropdown) Dropdown {
	if overrides.ContextMenuAriaLabel != "" {
		base.ContextMenuAriaLabel = overrides.ContextMenuAriaLabel
	}
	if overrides.OpenMenuAriaLabel != "" {
		base.OpenMenuAriaLabel = overrides.OpenMenuAriaLabel
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
	if overrides.CloseAriaLabel != "" {
		base.CloseAriaLabel = overrides.CloseAriaLabel
	}
	if overrides.DialogCloseAriaLabel != "" {
		base.DialogCloseAriaLabel = overrides.DialogCloseAriaLabel
	}
	return base
}

func mergeNavbar(base, overrides Navbar) Navbar {
	if overrides.AriaLabel != "" {
		base.AriaLabel = overrides.AriaLabel
	}
	if overrides.OpenMenuAriaLabel != "" {
		base.OpenMenuAriaLabel = overrides.OpenMenuAriaLabel
	}
	if overrides.CloseMenuAriaLabel != "" {
		base.CloseMenuAriaLabel = overrides.CloseMenuAriaLabel
	}
	if overrides.UserMenuAriaLabel != "" {
		base.UserMenuAriaLabel = overrides.UserMenuAriaLabel
	}
	if overrides.SecondaryAriaLabel != "" {
		base.SecondaryAriaLabel = overrides.SecondaryAriaLabel
	}
	return base
}

func mergePagination(base, overrides Pagination) Pagination {
	if overrides.AriaLabel != "" {
		base.AriaLabel = overrides.AriaLabel
	}
	if overrides.PreviousAriaLabel != "" {
		base.PreviousAriaLabel = overrides.PreviousAriaLabel
	}
	if overrides.NextAriaLabel != "" {
		base.NextAriaLabel = overrides.NextAriaLabel
	}
	if overrides.MoreAriaLabel != "" {
		base.MoreAriaLabel = overrides.MoreAriaLabel
	}
	if overrides.PreviousLabel != "" {
		base.PreviousLabel = overrides.PreviousLabel
	}
	if overrides.NextLabel != "" {
		base.NextLabel = overrides.NextLabel
	}
	if overrides.PageAriaLabel != nil {
		base.PageAriaLabel = overrides.PageAriaLabel
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
	if overrides.RequiredAriaLabel != "" {
		base.RequiredAriaLabel = overrides.RequiredAriaLabel
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
	if overrides.ResultsAriaLabel != nil {
		base.ResultsAriaLabel = overrides.ResultsAriaLabel
	}
	return base
}

func mergeSelect(base, overrides Select) Select {
	if overrides.Placeholder != "" {
		base.Placeholder = overrides.Placeholder
	}
	if overrides.ListAriaLabel != nil {
		base.ListAriaLabel = overrides.ListAriaLabel
	}
	return base
}

func mergeSidebar(base, overrides Sidebar) Sidebar {
	if overrides.AriaLabel != "" {
		base.AriaLabel = overrides.AriaLabel
	}
	if overrides.SearchAriaLabel != "" {
		base.SearchAriaLabel = overrides.SearchAriaLabel
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
	if overrides.AriaLabel != "" {
		base.AriaLabel = overrides.AriaLabel
	}
	return base
}

func mergeSplitButton(base, overrides SplitButton) SplitButton {
	if overrides.MenuAriaLabel != "" {
		base.MenuAriaLabel = overrides.MenuAriaLabel
	}
	return base
}

func mergeSteps(base, overrides Steps) Steps {
	if overrides.AriaLabel != "" {
		base.AriaLabel = overrides.AriaLabel
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
	if overrides.RemoveAriaLabel != "" {
		base.RemoveAriaLabel = overrides.RemoveAriaLabel
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
	if overrides.OptionsAriaLabel != "" {
		base.OptionsAriaLabel = overrides.OptionsAriaLabel
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
	if overrides.RemoveAriaLabel != "" {
		base.RemoveAriaLabel = overrides.RemoveAriaLabel
	}
	return base
}

func mergeTextarea(base, overrides Textarea) Textarea {
	if overrides.SendAriaLabel != "" {
		base.SendAriaLabel = overrides.SendAriaLabel
	}
	if overrides.EmojiAriaLabel != "" {
		base.EmojiAriaLabel = overrides.EmojiAriaLabel
	}
	if overrides.AttachAriaLabel != "" {
		base.AttachAriaLabel = overrides.AttachAriaLabel
	}
	if overrides.VoiceAriaLabel != "" {
		base.VoiceAriaLabel = overrides.VoiceAriaLabel
	}
	if overrides.SendLabel != "" {
		base.SendLabel = overrides.SendLabel
	}
	return base
}

func mergeTextInput(base, overrides TextInput) TextInput {
	if overrides.ShowPasswordAriaLabel != "" {
		base.ShowPasswordAriaLabel = overrides.ShowPasswordAriaLabel
	}
	if overrides.SearchAriaLabel != "" {
		base.SearchAriaLabel = overrides.SearchAriaLabel
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
	if overrides.AriaLabel != "" {
		base.AriaLabel = overrides.AriaLabel
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
