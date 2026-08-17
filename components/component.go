package components

import (
	"slices"

	"github.com/a-h/templ"
)

// Kind identifies a Goshtoso component type.
type Kind string

// Component is a renderable Goshtoso component with a stable identity.
type Component interface {
	templ.Component
	Kind() Kind
}

const (
	KindAccordion              Kind = "accordion"
	KindAlert                  Kind = "alert"
	KindAvatar                 Kind = "avatar"
	KindAvatarStack            Kind = "avatar-stack"
	KindBadge                  Kind = "badge"
	KindNotificationBadge      Kind = "notification-badge"
	KindNotificationDot        Kind = "notification-dot"
	KindAnimatingDot           Kind = "animating-dot"
	KindBanner                 Kind = "banner"
	KindCookieBanner           Kind = "cookie-banner"
	KindBreadcrumbs            Kind = "breadcrumbs"
	KindButton                 Kind = "button"
	KindCard                   Kind = "card"
	KindCarousel               Kind = "carousel"
	KindCardCarousel           Kind = "card-carousel"
	KindChatBubble             Kind = "chat-bubble"
	KindTypingIndicator        Kind = "typing-indicator"
	KindCheckbox               Kind = "checkbox"
	KindCheckboxGroup          Kind = "checkbox-group"
	KindCodeBlock              Kind = "code-block"
	KindInlineCode             Kind = "inline-code"
	KindCombobox               Kind = "combobox"
	KindDrawer                 Kind = "drawer"
	KindDropdown               Kind = "dropdown"
	KindPopover                Kind = "popover"
	KindSplitButton            Kind = "split-button"
	KindActionGroup            Kind = "action-group"
	KindFileInput              Kind = "file-input"
	KindForm                   Kind = "form"
	KindFormSection            Kind = "form-section"
	KindFormCollapsibleSection Kind = "form-collapsible-section"
	KindFormFlipSection        Kind = "form-flip-section"
	KindFormSubSection         Kind = "form-sub-section"
	KindFormFieldGroup         Kind = "form-field-group"
	KindFormErrors             Kind = "form-errors"
	KindDependencies           Kind = "dependencies"
	KindDependenciesMinimal    Kind = "dependencies-minimal"
	KindKbd                    Kind = "kbd"
	KindLink                   Kind = "link"
	KindModal                  Kind = "modal"
	KindAlertDialog            Kind = "alert-dialog"
	KindNavbar                 Kind = "navbar"
	KindPagination             Kind = "pagination"
	KindPalette                Kind = "palette"
	KindRadio                  Kind = "radio"
	KindRadioBar               Kind = "radio-bar"
	KindRadioGroup             Kind = "radio-group"
	KindRange                  Kind = "range"
	KindRating                 Kind = "rating"
	KindRatingDisplay          Kind = "rating-display"
	KindSchemaFormFields       Kind = "schema-form-fields"
	KindSearch                 Kind = "search"
	KindSearchField            Kind = "search-field"
	KindSearchModal            Kind = "search-modal"
	KindSelect                 Kind = "select"
	KindSidebar                Kind = "sidebar"
	KindSidebarOverlay         Kind = "sidebar-overlay"
	KindSpinner                Kind = "spinner"
	KindSteps                  Kind = "steps"
	KindStructuredInput        Kind = "structured-input"
	KindTable                  Kind = "table"
	KindTableHeadContent       Kind = "table-head-content"
	KindTableRows              Kind = "table-rows"
	KindTableRow               Kind = "table-row"
	KindTablePaginationNav     Kind = "table-pagination-nav"
	KindTableImageCell         Kind = "table-image-cell"
	KindTabs                   Kind = "tabs"
	KindTagsList               Kind = "tags-list"
	KindTextarea               Kind = "textarea"
	KindTextareaWithActions    Kind = "textarea-with-actions"
	KindTextInput              Kind = "text-input"
	KindToastContainer         Kind = "toast-container"
	KindToast                  Kind = "toast"
	KindMessageToast           Kind = "message-toast"
	KindOOBToast               Kind = "oob-toast"
	KindOOBMessageToast        Kind = "oob-message-toast"
	KindToggle                 Kind = "toggle"
	KindTooltip                Kind = "tooltip"
	KindAppShell               Kind = "app-shell"
	KindPageHeader             Kind = "page-header"
	KindToolbar                Kind = "toolbar"
	KindPanel                  Kind = "panel"
	KindEmptyState             Kind = "empty-state"
	KindSkeleton               Kind = "skeleton"
	KindIcon                   Kind = "icon"
)

var allKinds = []Kind{
	KindAccordion,
	KindAlert,
	KindAvatar,
	KindAvatarStack,
	KindBadge,
	KindNotificationBadge,
	KindNotificationDot,
	KindAnimatingDot,
	KindBanner,
	KindCookieBanner,
	KindBreadcrumbs,
	KindButton,
	KindCard,
	KindCarousel,
	KindCardCarousel,
	KindChatBubble,
	KindTypingIndicator,
	KindCheckbox,
	KindCheckboxGroup,
	KindCodeBlock,
	KindInlineCode,
	KindCombobox,
	KindDrawer,
	KindDropdown,
	KindPopover,
	KindSplitButton,
	KindActionGroup,
	KindFileInput,
	KindForm,
	KindFormSection,
	KindFormCollapsibleSection,
	KindFormFlipSection,
	KindFormSubSection,
	KindFormFieldGroup,
	KindFormErrors,
	KindDependencies,
	KindDependenciesMinimal,
	KindKbd,
	KindLink,
	KindModal,
	KindAlertDialog,
	KindNavbar,
	KindPagination,
	KindPalette,
	KindRadio,
	KindRadioBar,
	KindRadioGroup,
	KindRange,
	KindRating,
	KindRatingDisplay,
	KindSchemaFormFields,
	KindSearch,
	KindSearchField,
	KindSearchModal,
	KindSelect,
	KindSidebar,
	KindSidebarOverlay,
	KindSpinner,
	KindSteps,
	KindStructuredInput,
	KindTable,
	KindTableHeadContent,
	KindTableRows,
	KindTableRow,
	KindTablePaginationNav,
	KindTableImageCell,
	KindTabs,
	KindTagsList,
	KindTextarea,
	KindTextareaWithActions,
	KindTextInput,
	KindToastContainer,
	KindToast,
	KindMessageToast,
	KindOOBToast,
	KindOOBMessageToast,
	KindToggle,
	KindTooltip,
	KindAppShell,
	KindPageHeader,
	KindToolbar,
	KindPanel,
	KindEmptyState,
	KindSkeleton,
	KindIcon,
}

// AllKinds returns every supported component kind in stable order.
func AllKinds() []Kind {
	return slices.Clone(allKinds)
}
