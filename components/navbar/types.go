package navbar

import (
	"fmt"
	"slices"
	"strings"

	"github.com/a-h/templ"
)

// NavLink represents a navigation link in the navbar
type NavLink struct {
	// Label is the display text
	Label string
	// Href is the link URL
	Href string
	// Active marks this link as the current page
	Active bool
	// LinkAttrs are extra HTML attributes on the <a> tag
	LinkAttrs templ.Attributes
}

// UserProfile holds user information for the avatar dropdown
type UserProfile struct {
	// Name is the user's display name
	Name string
	// Email is the user's email address
	Email string
	// Avatar is an optional component rendered as the avatar trigger button content.
	// When nil, a default user icon is rendered.
	Avatar templ.Component
}

// UserMenuItem represents a single item in the user dropdown menu
type UserMenuItem struct {
	// Label is the display text
	Label string
	// Href is the link URL
	Href string
	// Icon is an optional icon rendered before the label
	Icon templ.Component
	// LinkAttrs are extra HTML attributes on the <a> tag
	LinkAttrs templ.Attributes
	// Danger renders the item in danger color (e.g., sign out)
	Danger bool
}

// ActionPosition controls where an action item is rendered in the navbar
type ActionPosition string

const (
	// ActionLeft renders the action after the brand (default)
	ActionLeft ActionPosition = "left"
	// ActionRight renders the action before the user avatar
	ActionRight ActionPosition = "right"
)

// ActionItem is a custom component rendered in the navbar at a configurable position.
type ActionItem struct {
	// Content is the component to render
	Content templ.Component
	// Position controls placement: "left" (after brand) or "right" (before avatar).
	// Default: "left"
	Position ActionPosition
}

// SecondaryCurrent identifies the current secondary navigation item.
type SecondaryCurrent string

const (
	// SecondaryCurrentNone leaves the link inactive.
	SecondaryCurrentNone SecondaryCurrent = ""
	// SecondaryCurrentPage marks the destination as the current page.
	SecondaryCurrentPage SecondaryCurrent = "page"
	// SecondaryCurrentLocation marks the destination as the current location.
	SecondaryCurrentLocation SecondaryCurrent = "location"
)

// SecondaryConfig holds the optional secondary navbar row configuration.
type SecondaryConfig struct {
	// Links are consumer-owned primitive links rendered in order as the secondary navigation region when Content is nil.
	// Links cannot be combined with Content. The default is no links.
	Links []SecondaryLink
	// Actions are consumer-owned components rendered in the secondary action region.
	// Nil actions are invalid, and the field cannot be combined with Content. The default is no actions.
	Actions []templ.Component
	// Content is a consumer-owned escape hatch rendered exactly once as the secondary row's only content.
	// It cannot be combined with Links or Actions; nil selects primitive link/action rendering.
	Content templ.Component
	// AriaLabel names the secondary navigation landmark. It defaults to "secondary navigation"
	// and must differ from the primary "main navigation" landmark.
	AriaLabel string
	// Scrollable enables horizontal scrolling for the primitive link container. The default is false.
	Scrollable bool
	// RootClass adds consumer-owned classes to the neutral secondary-row root after package defaults.
	// The default is no additional class.
	RootClass string
	// RootAttrs supplies consumer-owned allowlisted attributes for the neutral secondary-row root.
	// Structural attributes are reserved, and class is merged with package classes. The default is empty.
	RootAttrs templ.Attributes
}

// SecondaryLink is a primitive secondary navigation link rendered by Navbar.
type SecondaryLink struct {
	// Label is the required visible text rendered inside the anchor.
	Label string
	// Href is the required destination URL rendered on the anchor.
	Href string
	// Current identifies the link's current-location state. The default is SecondaryCurrentNone;
	// page or location renders the corresponding aria-current value, with at most one current link.
	Current SecondaryCurrent
	// LinkAttrs supplies consumer-owned allowlisted attributes for the primitive anchor.
	// Class is merged with package classes, while structural and action/mutation attributes are rejected.
	LinkAttrs templ.Attributes
}

// ValidationError reports a stable validation failure path and reason.
type ValidationError struct {
	Path   string
	Reason string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("navbar: invalid %s: %s", e.Path, e.Reason)
}

// Config holds configuration for the navbar component
type Config struct {
	// Brand is the logo/brand component (left side)
	Brand templ.Component
	// BrandHref is the link for the brand (default: "/")
	BrandHref string
	// Links are the desktop navigation links
	Links []NavLink
	// Actions are custom components (e.g., dark mode toggle, theme selector)
	// rendered at configurable positions. Default position is left (after brand).
	Actions []ActionItem
	// User holds user profile data for the avatar dropdown (nil = no avatar)
	User *UserProfile
	// UserMenu contains dropdown items under the avatar
	UserMenu []UserMenuItem
	// Secondary configures an optional second navbar row.
	Secondary *SecondaryConfig
	// NavClass allows additional CSS classes on the outer <nav>.
	NavClass string
	// NavAttrs are extra HTML attributes on the <nav> element
	NavAttrs templ.Attributes
}

// Validate rejects invalid secondary navbar configuration before rendering.
func (cfg Config) Validate() error {
	if cfg.Secondary == nil {
		return nil
	}
	if err := cfg.Secondary.validateCore(); err != nil {
		return err
	}
	if cfg.Secondary.hasContent() {
		if err := validatePrimaryNavAttrs(cfg.NavAttrs); err != nil {
			return err
		}
	}
	return cfg.Secondary.validateAttrMaps()
}

// Validate rejects invalid secondary-row configuration before rendering.
func (cfg SecondaryConfig) Validate() error {
	if err := cfg.validateCore(); err != nil {
		return err
	}
	return cfg.validateAttrMaps()
}

// LeftActions returns action items positioned on the left
func (cfg Config) leftActions() []ActionItem {
	var items []ActionItem
	for _, a := range cfg.Actions {
		if a.Position == "" || a.Position == ActionLeft {
			items = append(items, a)
		}
	}
	return items
}

// RightActions returns action items positioned on the right
func (cfg Config) rightActions() []ActionItem {
	var items []ActionItem
	for _, a := range cfg.Actions {
		if a.Position == ActionRight {
			items = append(items, a)
		}
	}
	return items
}

// NavClasses returns the CSS classes for the outer nav element
func (cfg Config) navClasses() string {
	base := "flex items-center justify-between border-b border-outline px-6 py-4 dark:border-outline-dark"
	if cfg.NavClass != "" {
		return base + " " + cfg.NavClass
	}
	return base
}

func (cfg Config) hasSecondaryContent() bool {
	return cfg.Secondary != nil && cfg.Secondary.hasContent()
}

func (cfg SecondaryConfig) hasContent() bool {
	return len(cfg.Links) > 0 || len(cfg.Actions) > 0 || cfg.Content != nil
}

func (cfg SecondaryConfig) validateCore() error {
	if cfg.Content != nil && (len(cfg.Links) > 0 || len(cfg.Actions) > 0) {
		return ValidationError{
			Path:   "Secondary.Content",
			Reason: "cannot be combined with Links or Actions",
		}
	}
	for i, action := range cfg.Actions {
		if action == nil {
			return ValidationError{
				Path:   fmt.Sprintf("Secondary.Actions[%d]", i),
				Reason: "must not be nil",
			}
		}
	}
	currentCount := 0
	for i, link := range cfg.Links {
		if strings.TrimSpace(link.Label) == "" {
			return ValidationError{
				Path:   fmt.Sprintf("Secondary.Links[%d].Label", i),
				Reason: "must not be blank after trimming whitespace",
			}
		}
		if strings.TrimSpace(link.Href) == "" {
			return ValidationError{
				Path:   fmt.Sprintf("Secondary.Links[%d].Href", i),
				Reason: "must not be blank after trimming whitespace",
			}
		}
		switch link.Current {
		case SecondaryCurrentNone, SecondaryCurrentPage, SecondaryCurrentLocation:
		default:
			return ValidationError{
				Path:   fmt.Sprintf("Secondary.Links[%d].Current", i),
				Reason: "must be empty, page, or location",
			}
		}
		if link.Current != SecondaryCurrentNone {
			currentCount++
		}
	}
	if currentCount > 1 {
		return ValidationError{
			Path:   "Secondary.Links",
			Reason: "at most one link may have Current",
		}
	}
	if normalized := normalizeLandmarkLabel(cfg.AriaLabel); normalized != "" && strings.EqualFold(normalized, "main navigation") {
		return ValidationError{
			Path:   "Secondary.AriaLabel",
			Reason: "must differ from main navigation",
		}
	}
	return nil
}

func (cfg SecondaryConfig) validateAttrMaps() error {
	if err := validateSecondaryRootAttrs(cfg.RootAttrs); err != nil {
		return err
	}
	for i, link := range cfg.Links {
		if err := validateSecondaryLinkAttrs(i, link.LinkAttrs); err != nil {
			return err
		}
	}
	return nil
}

func validatePrimaryNavAttrs(attrs templ.Attributes) error {
	if len(attrs) == 0 {
		return nil
	}
	keys := canonicalKeys(attrs)
	for _, key := range keys {
		switch key {
		case "aria-hidden":
			return ValidationError{
				Path:   `NavAttrs["aria-hidden"]`,
				Reason: "reserved when secondary content is present; primary landmark must remain exposed",
			}
		case "aria-label":
			return ValidationError{
				Path:   `NavAttrs["aria-label"]`,
				Reason: "reserved when secondary content is present; primary label is main navigation",
			}
		case "aria-labelledby":
			return ValidationError{
				Path:   `NavAttrs["aria-labelledby"]`,
				Reason: "reserved when secondary content is present; primary label is main navigation",
			}
		case "aria-roledescription":
			return ValidationError{
				Path:   `NavAttrs["aria-roledescription"]`,
				Reason: "reserved when secondary content is present; primary landmark role is component-owned",
			}
		case "role":
			return ValidationError{
				Path:   `NavAttrs["role"]`,
				Reason: "reserved when secondary content is present; primary element is navigation",
			}
		}
	}
	return nil
}

func validateSecondaryRootAttrs(attrs templ.Attributes) error {
	return validateAttrMap("Secondary.RootAttrs", attrs, rootAttrPolicy)
}

func validateSecondaryLinkAttrs(index int, attrs templ.Attributes) error {
	return validateAttrMap(fmt.Sprintf("Secondary.Links[%d].LinkAttrs", index), attrs, linkAttrPolicy)
}

type attrPolicy struct {
	classReason       string
	duplicateReason   string
	reservedReason    string
	unsupportedReason string
	actionReason      string
	isAllowed         func(string) bool
	isReserved        func(string) bool
	isAction          func(string) bool
}

var rootAttrPolicy = attrPolicy{
	classReason:       "class value must be a string",
	duplicateReason:   "duplicate case-insensitive attribute keys",
	reservedReason:    "reserved attribute",
	unsupportedReason: "unsupported secondary-root attribute",
	isAllowed:         isAllowedSecondaryRootAttr,
	isReserved:        isReservedSecondaryRootAttr,
}

var linkAttrPolicy = attrPolicy{
	classReason:       "class value must be a string",
	duplicateReason:   "duplicate case-insensitive attribute keys",
	reservedReason:    "reserved attribute",
	unsupportedReason: "unsupported primitive-link attribute",
	actionReason:      "action or mutation attribute must be placed in Actions",
	isAllowed:         isAllowedSecondaryLinkAttr,
	isReserved:        isReservedSecondaryLinkAttr,
	isAction:          isActionSecondaryLinkAttr,
}

func validateAttrMap(path string, attrs templ.Attributes, policy attrPolicy) error {
	if len(attrs) == 0 {
		return nil
	}
	duplicates := duplicateCanonicalKeys(attrs)
	if len(duplicates) > 0 {
		return ValidationError{
			Path:   fmt.Sprintf(`%s["%s"]`, path, duplicates[0]),
			Reason: policy.duplicateReason,
		}
	}
	values := canonicalAttrValues(attrs)
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	for _, key := range keys {
		value := values[key]
		if key == "class" {
			if _, ok := value.(string); !ok {
				return ValidationError{
					Path:   fmt.Sprintf(`%s["class"]`, path),
					Reason: policy.classReason,
				}
			}
		}
		if policy.isReserved != nil && policy.isReserved(key) {
			return ValidationError{
				Path:   fmt.Sprintf(`%s["%s"]`, path, key),
				Reason: policy.reservedReason,
			}
		}
		if policy.isAction != nil && policy.isAction(key) {
			return ValidationError{
				Path:   fmt.Sprintf(`%s["%s"]`, path, key),
				Reason: policy.actionReason,
			}
		}
		if policy.isAllowed == nil || policy.isAllowed(key) {
			continue
		}
		return ValidationError{
			Path:   fmt.Sprintf(`%s["%s"]`, path, key),
			Reason: policy.unsupportedReason,
		}
	}
	return nil
}

func normalizeLandmarkLabel(label string) string {
	return strings.Join(strings.Fields(label), " ")
}

func canonicalKeys(attrs templ.Attributes) []string {
	keys := make([]string, 0, len(attrs))
	for key := range attrs {
		keys = append(keys, canonicalAttrKey(key))
	}
	slices.Sort(keys)
	return slices.Compact(keys)
}

func duplicateCanonicalKeys(attrs templ.Attributes) []string {
	counts := map[string]int{}
	for key := range attrs {
		counts[canonicalAttrKey(key)]++
	}
	var duplicates []string
	for key, count := range counts {
		if count > 1 {
			duplicates = append(duplicates, key)
		}
	}
	slices.Sort(duplicates)
	return duplicates
}

func canonicalAttrValues(attrs templ.Attributes) map[string]any {
	canonical := make(map[string]any, len(attrs))
	for key, value := range attrs {
		canonical[canonicalAttrKey(key)] = value
	}
	return canonical
}

func canonicalAttrKey(key string) string {
	var lower strings.Builder
	lower.Grow(len(key))
	for i := 0; i < len(key); i++ {
		char := key[i]
		if char >= 'A' && char <= 'Z' {
			char += 'a' - 'A'
		}
		lower.WriteByte(char)
	}
	return lower.String()
}

func isAllowedSecondaryRootAttr(key string) bool {
	switch key {
	case "id", "class",
		"aria-describedby", "aria-description", "aria-details", "aria-errormessage",
		"aria-keyshortcuts", "aria-live", "aria-atomic", "aria-busy", "aria-relevant":
		return true
	}
	return strings.HasPrefix(key, "data-") || strings.HasPrefix(key, "hx-") || strings.HasPrefix(key, "x-")
}

func isReservedSecondaryRootAttr(key string) bool {
	switch key {
	case "role", "aria-label", "aria-labelledby", "aria-hidden", "aria-current", "aria-selected", "aria-controls", "tabindex":
		return true
	}
	return key == "data-navbar-secondary"
}

func isAllowedSecondaryLinkAttr(key string) bool {
	switch key {
	case "id", "class", "target", "rel", "download", "type", "hreflang", "referrerpolicy",
		"aria-label", "aria-describedby", "aria-labelledby",
		"hx-get", "hx-target", "hx-swap", "hx-push-url", "hx-select", "hx-indicator", "hx-confirm":
		return true
	}
	return strings.HasPrefix(key, "data-")
}

func isReservedSecondaryLinkAttr(key string) bool {
	switch key {
	case "href", "role", "tabindex", "aria-current", "aria-selected", "aria-controls":
		return true
	}
	return false
}

func isActionSecondaryLinkAttr(key string) bool {
	switch key {
	case "ping", "action", "method", "formaction", "formenctype", "formmethod", "formnovalidate", "formtarget":
		return true
	}
	if strings.HasPrefix(key, "x-") || strings.HasPrefix(key, "on") {
		return true
	}
	if strings.HasPrefix(key, "hx-") && !isAllowedSecondaryLinkAttr(key) {
		return true
	}
	return false
}

func sanitizeSecondaryRootAttrs(attrs templ.Attributes) templ.Attributes {
	return sanitizeAttrMap(attrs, "class")
}

func sanitizeSecondaryLinkAttrs(attrs templ.Attributes) templ.Attributes {
	return sanitizeAttrMap(attrs, "class")
}

func sanitizeAttrMap(attrs templ.Attributes, omit ...string) templ.Attributes {
	if len(attrs) == 0 {
		return nil
	}
	skip := make(map[string]struct{}, len(omit))
	for _, key := range omit {
		skip[key] = struct{}{}
	}
	canonical := canonicalAttrValues(attrs)
	out := make(templ.Attributes, len(canonical))
	for key, value := range canonical {
		if _, ok := skip[key]; ok {
			continue
		}
		out[key] = value
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func classValue(attrs templ.Attributes) string {
	if len(attrs) == 0 {
		return ""
	}
	values := canonicalAttrValues(attrs)
	if class, ok := values["class"].(string); ok {
		return class
	}
	return ""
}

func mergeClasses(parts ...string) string {
	tokens := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			tokens = append(tokens, part)
		}
	}
	return strings.Join(tokens, " ")
}

func (cfg SecondaryConfig) rootClasses() string {
	return mergeClasses(
		"flex min-w-0 flex-wrap items-end gap-2 border-b border-outline bg-surface px-6 dark:border-outline-dark dark:bg-surface-dark",
		cfg.RootClass,
		classValue(cfg.RootAttrs),
	)
}

func (cfg SecondaryConfig) rootAttrs() templ.Attributes {
	return sanitizeSecondaryRootAttrs(cfg.RootAttrs)
}

func (cfg SecondaryConfig) normalizedAriaLabel() string {
	if label := normalizeLandmarkLabel(cfg.AriaLabel); label != "" {
		return label
	}
	return "secondary navigation"
}

func (cfg SecondaryConfig) linksNavClasses() string {
	if cfg.Scrollable {
		return "min-w-0 flex-1"
	}
	return "min-w-0 flex-1"
}

func (cfg SecondaryConfig) linksContainerClasses() string {
	if cfg.Scrollable {
		return "flex min-w-0 flex-1 items-end gap-1 overflow-x-auto overflow-y-visible px-2 py-2 whitespace-nowrap"
	}
	return "flex min-w-0 flex-1 flex-wrap items-end gap-1"
}

func (cfg SecondaryConfig) actionsClasses() string {
	return "flex shrink-0 items-center gap-2"
}

func secondaryLinkClasses(current SecondaryCurrent, extra string) string {
	base := "inline-flex min-h-11 min-w-11 shrink-0 items-center border-b-2 bg-surface px-3 py-2 text-sm whitespace-nowrap transition-colors motion-reduce:transition-none focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary dark:bg-surface-dark dark:focus-visible:outline-primary-dark"
	if current != SecondaryCurrentNone {
		return mergeClasses(
			base,
			"border-primary font-semibold text-on-surface-strong hover:border-primary hover:text-on-surface-strong active:border-primary active:text-on-surface-strong dark:border-primary-dark dark:text-on-surface-dark-strong dark:hover:border-primary-dark dark:hover:text-on-surface-dark-strong dark:active:border-primary-dark dark:active:text-on-surface-dark-strong",
			extra,
		)
	}
	return mergeClasses(
		base,
		"border-transparent text-on-surface hover:border-outline-strong hover:border-b-outline-strong hover:!border-b-outline-strong hover:text-on-surface-strong active:border-primary active:text-on-surface-strong dark:text-on-surface-dark dark:hover:border-outline dark:hover:border-b-outline-dark-strong dark:hover:!border-b-outline-dark-strong dark:hover:text-on-surface-dark-strong dark:active:border-primary-dark dark:active:text-on-surface-dark-strong",
		extra,
	)
}

// LinkClasses returns the CSS classes for a nav link
func linkClasses(active bool) string {
	if active {
		return "font-bold text-primary underline-offset-2 hover:text-primary focus:outline-hidden focus:underline dark:text-primary-dark dark:hover:text-primary-dark"
	}
	return "font-medium text-on-surface underline-offset-2 hover:text-primary focus:outline-hidden focus:underline dark:text-on-surface-dark dark:hover:text-primary-dark"
}

// MenuItemClasses returns the CSS classes for a user menu item
func menuItemClasses(danger bool) string {
	if danger {
		return "block bg-surface-alt px-4 py-2 text-sm text-danger hover:bg-danger/5 focus-visible:bg-danger/10 focus-visible:outline-hidden dark:bg-surface-dark-alt dark:hover:bg-danger/10"
	}
	return "block bg-surface-alt px-4 py-2 text-sm text-on-surface hover:bg-surface-dark-alt/5 hover:text-on-surface-strong focus-visible:bg-surface-dark-alt/10 focus-visible:text-on-surface-strong focus-visible:outline-hidden dark:bg-surface-dark-alt dark:text-on-surface-dark dark:hover:bg-surface-alt/5 dark:hover:text-on-surface-dark-strong dark:focus-visible:bg-surface-alt/10 dark:focus-visible:text-on-surface-dark-strong"
}
