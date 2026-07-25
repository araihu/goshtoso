# Migrating the component API from v0.0.11

This guide covers the breaking component API changes in the current
**Unreleased** development line, using `v0.0.11` as the exact comparison base.
The base tag resolves to commit
`10b4dcbf3da3c1dd534d8d2baa949d043b9d0f1f`.

Goshtoso is still alpha. This migration deliberately removes accidental public
surface instead of carrying compatibility aliases. Upgrade all Goshtoso
packages together, then compile and render-test the application before
deploying it.

## What changed

The component model now separates three ideas:

- A `components.Component` is a renderable primitive with a stable `Kind()`.
  `components.AllKinds()` returns the complete Kind registry in stable order.
- A configuration dimension names one independent choice. `Tone` is semantic
  intent, `Appearance` is visual treatment, and `Mode` is a behavior or
  presentation strategy.
- A theme is the application-wide token system. It is neither a component Kind
  nor a configuration dimension.

There is no universal `Variant` or `Variant()` API. The old word mixed semantic
color, visual treatment, behavior, layout, and different component contracts.
Dimensions stay package-local, so `badge.ToneDanger` and `alert.ToneDanger`
remain distinct Go types even though they express the same word.

The release also introduces split primitives when semantics, content,
accessibility, DOM ownership, or interaction differ; selective functional
options for four atomic components; concrete return types for every public
renderable; and a curated set of supported helpers.

## Source-breaking changes

### Searchable old symbol to new symbol or action

No row below implies an alias. Every old symbol in this section was removed.

#### Dimension types, fields, and constants

| Old symbol | New symbol or action |
| --- | --- |
| `accordion.Variant`, `accordion.AccordionConfig.Variant` | `accordion.Appearance`, `accordion.AccordionConfig.Appearance` |
| `accordion.Default` | `accordion.AppearanceDefault` |
| `accordion.NoBackground` | `accordion.AppearancePlain` |
| `accordion.Split` | `accordion.AppearanceSplit` |
| `accordion.SingleOpen` | Remove it; `AllowMultiple == false` is the single-open behavior. |
| `alert.Variant`, `alert.Config.Variant` | `alert.Tone`, `alert.Config.Tone` |
| `alert.Info`, `alert.Success`, `alert.Warning`, `alert.Danger` | `alert.ToneInfo`, `alert.ToneSuccess`, `alert.ToneWarning`, `alert.ToneDanger` |
| `avatar.Variant`, `avatar.Config.Variant` | `avatar.Tone`, `avatar.Config.Tone` |
| `avatar.Default`, `avatar.Inverse`, `avatar.Primary`, `avatar.Secondary`, `avatar.Info`, `avatar.Success`, `avatar.Warning`, `avatar.Danger` | Prefix each constant with `Tone`: `avatar.ToneDefault` through `avatar.ToneDanger`. |
| `badge.Variant`, `badge.Config.Variant` | `badge.Tone`, `badge.Config.Tone` |
| `badge.Default`, `badge.Inverse`, `badge.Primary`, `badge.Secondary`, `badge.Info`, `badge.Success`, `badge.Warning`, `badge.Danger` | Prefix each constant with `Tone`: `badge.ToneDefault` through `badge.ToneDanger`. |
| `badge.Style`, `badge.Config.Style` | `badge.Appearance`, `badge.Config.Appearance` |
| `badge.StyleSolid`, `badge.StyleSoft` | `badge.AppearanceSolid`, `badge.AppearanceSoft` |
| `banner.Variant`, `banner.Config.Variant` | `banner.Tone`, `banner.Config.Tone` |
| `banner.Default`, `banner.Primary`, `banner.Info`, `banner.Success`, `banner.Warning`, `banner.Danger` | Prefix each constant with `Tone`: `banner.ToneDefault` through `banner.ToneDanger`. |
| `button.Variant`, `button.Config.Variant` | `button.Tone` and `button.WithTone(...)`; `button.Config` itself was removed. |
| `button.Primary`, `button.Secondary`, `button.Alternate`, `button.Inverse`, `button.Info`, `button.Danger`, `button.Warning`, `button.Success` | Prefix each constant with `Tone`: `button.TonePrimary` through `button.ToneSuccess`. |
| `card.Variant`, `card.Config.Variant` | `card.Appearance`, `card.Config.Appearance` |
| `card.Default`, `card.Primary` | `card.AppearanceDefault`, `card.AppearancePrimary` |
| `carousel.Variant`, `carousel.Config.Variant` | Remove the dimension; use slide content inference or a split primitive. |
| `carousel.Default` | Omit the old field; plain slides remain the default. |
| `carousel.WithText`, `carousel.WithCTA` | Populate `Slide.Title` / `Slide.Description`, or a complete `Slide.CTALabel` plus `Slide.CTAHref`; overlay rendering is inferred. |
| `carousel.OnCard` | Use `carousel.CardCarousel(carousel.CardConfig{...})`. |
| `chatbubble.Config.AvatarVariant` | `chatbubble.Config.AvatarTone` with the typed `avatar.Tone` value. |
| `checkbox.Variant`, `checkbox.Config.Variant` | `checkbox.Tone`, `checkbox.Config.Tone` |
| `checkbox.Primary`, `checkbox.Secondary`, `checkbox.Info`, `checkbox.Success`, `checkbox.Warning`, `checkbox.Danger` | Prefix each constant with `Tone`. |
| `fileinput.Variant`, `fileinput.Config.Variant` | `fileinput.Appearance`, `fileinput.Config.Appearance` |
| `fileinput.VariantDropZone`, `fileinput.VariantUpload` | `fileinput.AppearanceDropZone`, `fileinput.AppearanceUpload` |
| `link.Style`, `link.Config.Style` | `link.Appearance` and `link.WithAppearance(...)`; `link.Config` itself was removed. |
| `link.StyleText`, `link.StyleButton` | `link.AppearanceText`, `link.AppearanceButton` |
| `modal.Variant`, `modal.Config.Variant` | `modal.Tone`, but only on `modal.AlertDialogConfig.Tone`. |
| `modal.Default`, `modal.Success`, `modal.Info`, `modal.Warning`, `modal.Danger` | Prefix each constant with `Tone`: `modal.ToneDefault` through `modal.ToneDanger`. |
| `pagination.Variant`, `pagination.Config.Variant` | `pagination.Mode`, `pagination.Config.Mode` |
| `pagination.WithEllipsis`, `pagination.Simple` | `pagination.ModeEllipsis`, `pagination.ModeSimple` |
| `radio.Variant`, `radio.Config.Variant` | `radio.Tone`, `radio.Config.Tone` |
| `radio.Primary`, `radio.Secondary`, `radio.Info`, `radio.Success`, `radio.Warning`, `radio.Danger` | Prefix each constant with `Tone`. |
| `rating.Style`, `rating.Config.Style` | `rating.Appearance`, `rating.Config.Appearance` |
| `rating.StyleStars`, `rating.StyleEmoji` | `rating.AppearanceStars`, `rating.AppearanceEmoji` |
| `spinner.Variant`, `spinner.Config.Variant` | `spinner.Tone`, `spinner.Config.Tone` |
| `spinner.Default`, `spinner.Primary`, `spinner.Secondary`, `spinner.Info`, `spinner.Success`, `spinner.Warning`, `spinner.Danger` | Prefix each constant with `Tone`. |
| `table.Variant`, `table.Config.Variant` | `table.Appearance`, `table.Config.Appearance` |
| `table.Default`, `table.Striped` | `table.AppearanceDefault`, `table.AppearanceStriped` |
| `table.WithCheckbox` | Remove it; use `table.Config.ShowCheckbox`. |
| `table.FilterVariant`, `table.FilterConfig.Variant` | `table.FilterAppearance`, `table.FilterConfig.Appearance` |
| `table.FilterVariantBar`, `table.FilterVariantInline` | `table.FilterAppearanceBar`, `table.FilterAppearanceInline` |
| `toast.Variant`, `toast.Config.Variant` | `toast.Tone`, `toast.Config.Tone` |
| `toast.Info`, `toast.Success`, `toast.Warning`, `toast.Danger` | Prefix each constant with `Tone`. |
| `toast.Message` | Use the split `toast.MessageToast(toast.MessageConfig{...})` primitive. |
| `toggle.Variant`, `toggle.Config.Variant` | `toggle.Tone`, `toggle.Config.Tone` |
| `toggle.Primary`, `toggle.Secondary`, `toggle.Info`, `toggle.Success`, `toggle.Warning`, `toggle.Danger` | Prefix each constant with `Tone`. |
| `toggle.Style`, `toggle.Config.Style` | `toggle.Appearance`, `toggle.Config.Appearance` |
| `toggle.StyleDefault`, `toggle.StyleContainer` | `toggle.AppearanceDefault`, `toggle.AppearanceContainer` |
| `tooltip.Trigger`, `tooltip.Config.TriggerMode` | `tooltip.Activation` and `tooltip.WithActivation(...)`; `tooltip.Config` itself was removed. |
| `tooltip.Hover`, `tooltip.Click` | `tooltip.ActivationHover`, `tooltip.ActivationClick` |
| `tooltip.Top`, `tooltip.Bottom`, `tooltip.Left`, `tooltip.Right` | `tooltip.PositionTop`, `tooltip.PositionBottom`, `tooltip.PositionLeft`, `tooltip.PositionRight` |

#### Functional-option constructors

| Old symbol or field | New symbol or action |
| --- | --- |
| `button.Config`, `button.Button(cfg)` | `button.Option`, `button.Button(options ...button.Option)` |
| `button.Config.Variant` | `button.WithTone(...)` |
| `button.Config.Size` | `button.WithSize(...)` |
| `button.Config.Type` | `button.WithType(...)` |
| `button.Config.Disabled` | `button.Disabled()` when true; omit it when false. |
| `button.Config.ID` | `button.WithID(...)` |
| `button.Config.RootClass` | `button.WithRootClass(...)` |
| `button.Config.HTMX` | `button.WithHTMX(...)` |
| `button.Config.Alpine` | `button.WithAlpine(...)` |
| `button.Config.LoadingText` | `button.WithLoadingText(...)` |
| `link.Config`, `link.Link(cfg)` | `link.Option`, `link.Link(href, options ...link.Option)` |
| `link.Config.Href` | Required first `href` argument. |
| `link.Config.Target`, `link.Config.Rel`, `link.Config.Role`, `link.Config.ID` | `link.WithTarget`, `WithRel`, `WithRole`, `WithID` |
| `link.Config.Style`, `link.Config.Size` | `link.WithAppearance`, `link.WithSize` |
| `link.Config.Icon`, `link.Config.IconPosition` | `link.WithIcon`, `link.WithIconPosition` |
| `link.Config.Class`, `link.Config.Attrs` | `link.WithRootClass`, `link.WithAttrs` |
| `kbd.Config`, `kbd.Kbd(cfg)` | `kbd.Option`, `kbd.Kbd(text, options ...kbd.Option)` |
| `kbd.Config.Text` | Required first `text` argument. |
| `kbd.Config.Label`, `kbd.Config.Size`, `kbd.Config.Icon` | `kbd.WithLabel`, `WithSize`, `WithIcon` |
| `kbd.Config.Class`, `kbd.Config.Attrs` | `kbd.WithRootClass`, `kbd.WithAttrs` |
| `tooltip.Config`, `tooltip.Tooltip(cfg)` | `tooltip.Option`, `tooltip.Tooltip(id, label, options ...tooltip.Option)` |
| `tooltip.Config.ID`, `tooltip.Config.Label` | Required first and second constructor arguments. |
| `tooltip.Config.Description`, `tooltip.Config.Position`, `tooltip.Config.TriggerMode` | `tooltip.WithDescription`, `WithPosition`, `WithActivation` |
| `tooltip.Config.TriggerLabel`, `tooltip.Config.Trigger` | `tooltip.WithTriggerLabel`, `WithTrigger` |

#### Split primitives and removed fields

| Old symbol or shape | New symbol or action |
| --- | --- |
| `banner.Config.CookieBanner`, `banner.Config.CookieConfig` | Call `banner.CookieBanner(banner.CookieBannerConfig{...})`; ordinary `banner.Banner` no longer owns consent behavior. |
| `carousel.OnCard` | Call `carousel.CardCarousel(carousel.CardConfig{...})`. |
| `modal.Config.AlertMode`, `modal.Config.Variant` | Call `modal.AlertDialog(modal.AlertDialogConfig{Tone: ...})`. |
| `rating.Config.ReadOnly` | Call `rating.RatingDisplay(rating.DisplayConfig{...})`; keep `rating.Rating` for interactive form input. |
| `rating.Config.Class`, `rating.Config.Attrs` | `RootClass`, `RootAttrs` on the selected `Config` or `DisplayConfig`. |
| `toast.Config.Sender`, `toast.Message` | Call `toast.MessageToast(toast.MessageConfig{Sender: ...})`. |
| `toast.Container` | `toast.ToastContainer` |
| Message OOB rendering through `toast.OOBToast(toast.Config{Variant: toast.Message, ...})` | `toast.OOBMessageToast(toast.MessageConfig{...})` |
| `card.Config.Price`, `card.Config.Rating` | Remove the inert fields; compose price content in `Footer` and use `rating.RatingDisplay`. |
| `combobox.Option.Meta`, `combobox.Option.Img`, `combobox.Option.Initials`, `combobox.Option.Badge`, `combobox.Option.BadgeColor` | Remove the inert presentation fields. `combobox.Option` now supports `Value`, `Label`, and `Disabled`. |

### Before and after: configuration dimensions

Before:

```go
badge.Badge(badge.Config{
    Label:   "Requires attention",
    Variant: badge.Danger,
    Style:   badge.StyleSoft,
})

pagination.Pagination(pagination.Config{
    Variant:    pagination.Simple,
    CurrentPage: 2,
    TotalPages:  5,
})
```

After:

```go
badge.Badge(badge.Config{
    Label:      "Requires attention",
    Tone:       badge.ToneDanger,
    Appearance: badge.AppearanceSoft,
})

pagination.Pagination(pagination.Config{
    Mode:        pagination.ModeSimple,
    CurrentPage: 2,
    TotalPages:  5,
})
```

`Tone`, `Appearance`, and `Mode` can change independently and tell the reader
what kind of choice is being made.

### Before and after: functional options

Button before:

```go
button.Button(button.Config{
    Variant: button.Danger,
    Size:    button.SizeSmall,
    Type:    "submit",
    ID:      "remove",
})
```

Button after:

```go
button.Button(
    button.WithTone(button.ToneDanger),
    button.WithSize(button.SizeSmall),
    button.WithType("submit"),
    button.WithID("remove"),
)
```

Link before and after:

```go
// Before
link.Link(link.Config{
    Href:   "/docs",
    Style:  link.StyleButton,
    Target: "_blank",
})

// After
link.Link("/docs",
    link.WithAppearance(link.AppearanceButton),
    link.WithTarget("_blank"),
)
```

Kbd before and after:

```go
// Before
kbd.Kbd(kbd.Config{Text: "K", Label: "Command K", Size: kbd.SizeSM})

// After
kbd.Kbd("K", kbd.WithLabel("Command K"), kbd.WithSize(kbd.SizeSM))
```

Tooltip before and after:

```go
// Before
tooltip.Tooltip(tooltip.Config{
    ID:          "save-help",
    Label:       "Saves changes",
    Position:    tooltip.Bottom,
    TriggerMode: tooltip.Click,
})

// After
tooltip.Tooltip("save-help", "Saves changes",
    tooltip.WithPosition(tooltip.PositionBottom),
    tooltip.WithActivation(tooltip.ActivationClick),
)
```

### How to choose split primitives

Use one primitive with a dimension when semantic role, content model,
interaction, lifecycle, DOM responsibility, and server contract remain the
same. Use split primitives when one of those contracts changes materially.

| Choose | When |
| --- | --- |
| `banner.Banner` | A dismissible or persistent page banner with optional CTA. |
| `banner.CookieBanner` | A fixed consent dialog with accept and reject actions. |
| `carousel.Carousel` | Image slides with optional inferred overlay/CTA, autoplay, or lazy HTMX content. |
| `carousel.CardCarousel` | Article/card slides with their own wrapper and body contract. |
| `modal.Modal` | A general dialog with primary and secondary actions. |
| `modal.AlertDialog` | An `alertdialog` with semantic `Tone`, one action, and dismissal. |
| `rating.Rating` | An interactive radio-group form control. |
| `rating.RatingDisplay` | Read-only output with `role="img"` and no form lifecycle. |
| `toast.Toast` / `toast.OOBToast` | Semantic notification with `Tone`, title, and message. |
| `toast.MessageToast` / `toast.OOBMessageToast` | Sender-oriented message content and a text dismiss control. |

Modal before:

```go
modal.Modal(modal.Config{
    ID:           "remove-profile",
    Title:        "Remove profile?",
    TriggerLabel: "Remove",
    PrimaryLabel: "Remove",
    AlertMode:    true,
    Variant:      modal.Danger,
})
```

Modal after:

```go
modal.AlertDialog(modal.AlertDialogConfig{
    ID:           "remove-profile",
    Title:        "Remove profile?",
    TriggerLabel: "Remove",
    ActionLabel:  "Remove",
    Tone:         modal.ToneDanger,
})
```

For the other splits, move cookie settings to `CookieBannerConfig`, card
carousel settings to `CardConfig`, read-only rating settings to `DisplayConfig`,
and sender/message settings to `MessageConfig`.

### Concrete return types and stable Kind identity

Every public renderable used to return `templ.Component`. Every one now returns
an exported concrete value that implements `components.Component`. Ordinary
rendering remains source-compatible because the concrete value still has
`Render(context.Context, io.Writer) error`. Code that stores constructors as
function values, declares exact return types, or mocks `templ.Component`
factories must update its type signatures.

The complete constructor-to-result inventory is:

| Package | Constructors and concrete return types |
| --- | --- |
| `accordion` | `Accordion` → `accordion.Instance` |
| `alert` | `Alert` → `alert.Instance` |
| `avatar` | `Avatar` → `avatar.Instance`; `AvatarStack` → `avatar.StackInstance` |
| `badge` | `Badge` → `badge.Instance`; `NotificationBadge` → `badge.NotificationBadgeInstance`; `NotificationDot` → `badge.NotificationDotInstance`; `AnimatingDot` → `badge.AnimatingDotInstance` |
| `banner` | `Banner` → `banner.Instance`; `CookieBanner` → `banner.CookieBannerInstance` |
| `breadcrumbs` | `Breadcrumbs` → `breadcrumbs.Instance` |
| `button` | `Button` → `button.Instance` |
| `card` | `Card` → `card.Instance` |
| `carousel` | `Carousel` → `carousel.Instance`; `CardCarousel` → `carousel.CardCarouselInstance` |
| `chatbubble` | `ChatBubble` → `chatbubble.Instance`; `TypingIndicator` → `chatbubble.TypingIndicatorInstance` |
| `checkbox` | `Checkbox` → `checkbox.Instance`; `CheckboxGroup` → `checkbox.GroupInstance` |
| `codeblock` | `CodeBlock` → `codeblock.Instance` |
| `combobox` | `Combobox` → `combobox.Instance` |
| `drawer` | `Drawer` → `drawer.Instance` |
| `dropdown` | `Dropdown` → `dropdown.Instance` |
| `fileinput` | `FileInput` → `fileinput.Instance` |
| `form` | `Form` → `form.Instance`; `Section` → `form.SectionInstance`; `CollapsibleSection` → `form.CollapsibleSectionInstance`; `FlipSection` → `form.FlipSectionInstance`; `SubSection` → `form.SubSectionInstance`; `FieldGroup` → `form.FieldGroupInstance`; `FormErrors` → `form.FormErrorsInstance` |
| `head` | `Dependencies` → `head.Instance`; `DependenciesMinimal` → `head.MinimalInstance` |
| `kbd` | `Kbd` → `kbd.Instance` |
| `link` | `Link` → `link.Instance` |
| `modal` | `Modal` → `modal.Instance`; `AlertDialog` → `modal.AlertDialogInstance` |
| `navbar` | `Navbar` → `navbar.Instance` |
| `pagination` | `Pagination` → `pagination.Instance` |
| `palette` | `Palette` → `palette.Instance` |
| `radio` | `Radio` → `radio.Instance`; `RadioBar` → `radio.BarInstance`; `RadioGroup` → `radio.GroupInstance` |
| `range` | `Range` → `range.Instance` |
| `rating` | `Rating` → `rating.Instance`; `RatingDisplay` → `rating.DisplayInstance` |
| `schemaform` | `Fields` → `schemaform.Instance` |
| `search` | `Search` → `search.Instance`; `SearchField` → `search.FieldInstance`; `SearchModal` → `search.ModalInstance` |
| `select` | `Select` → `selectfield.Instance` (the import path is `components/select`) |
| `sidebar` | `Sidebar` → `sidebar.Instance`; `Overlay` → `sidebar.OverlayInstance` |
| `spinner` | `Spinner` → `spinner.Instance` |
| `steps` | `Steps` → `steps.Instance` |
| `structuredinput` | `StructuredInput` → `structuredinput.Instance` |
| `table` | `Table` → `table.Instance`; `TableHeadContent` → `table.TableHeadContentInstance`; `TableRows` → `table.TableRowsInstance`; `TableRow` → `table.TableRowInstance`; `TablePaginationNav` → `table.TablePaginationNavInstance`; `ImageCell` → `table.ImageCellInstance` |
| `tabs` | `Tabs` → `tabs.Instance` |
| `tagslist` | `TagsList` → `tagslist.Instance` |
| `textarea` | `Textarea` → `textarea.Instance`; `TextareaWithActions` → `textarea.WithActionsInstance` |
| `textinput` | `TextInput` → `textinput.Instance` |
| `toast` | `Toast` → `toast.Instance`; `OOBToast` → `toast.OOBInstance`; `MessageToast` → `toast.MessageInstance`; `OOBMessageToast` → `toast.OOBMessageInstance`; `ToastContainer` → `toast.ContainerInstance` |
| `toggle` | `Toggle` → `toggle.Instance` |
| `tooltip` | `Tooltip` → `tooltip.Instance` |

Store mixed primitives through the shared interface and switch on stable Kind:

```go
pageComponents := []components.Component{
    badge.Badge(badge.Config{Label: "New"}),
    button.Button(),
}

for _, component := range pageComponents {
    switch component.Kind() {
    case components.KindBadge:
        // Badge-specific orchestration.
    case components.KindButton:
        // Button-specific orchestration.
    }
}
```

`Kind` values are stable kebab-case identities for registries, diagnostics,
documentation tooling, and switches. They are not CSS classes or HTML element
names. `AllKinds()` returns a copy, so callers may safely sort or filter it.

### Curated removed and private surface

The library now exports renderables and intentional behavior helpers, not its
CSS assembly or template implementation. There are no compatibility aliases
for the removed symbols.

#### Accordion, Card, Combobox, and Table

| Removed symbol | Migration |
| --- | --- |
| `accordion.AccordionItemData` and its fields `accordion.AccordionItemData.Item`, `accordion.AccordionItemData.Index`, `accordion.AccordionItemData.AllowMultiple`, `accordion.AccordionItemData.Variant`, `accordion.AccordionItemData.ContainerID` and methods `ItemContainerClasses`, `ItemButtonClasses`, `ExpandedClasses`, `CollapsedClasses`, `ContentClasses` | This was generated-template render data. Construct `AccordionConfig` and `AccordionItem`; do not construct per-item internals. |
| `accordion.AccordionConfig.ContainerClasses` | Add consumer classes through `RootClass`; rendered class assembly is private. |
| `card.Config.Price`, `card.Config.Rating`, `card.Config.HasPrice`, `card.Config.HasRating` | These fields/predicates did not drive rendering. Compose price markup in `Footer` and use `rating.RatingDisplay`. |
| `card.StarRating` | `rating.RatingDisplay` |
| `card.Config.ContainerClasses`, `ContentClasses`, `DescriptionClasses`, `HasImage`, `ImageClasses`, `ImageContainerClasses`, `TagClasses`, `TitleClasses` | Use `card.Card` plus `RootClass`; class and render predicates are private. |
| `combobox.Body`, `combobox.OptionsList`, `combobox.ClientScript`, `combobox.ProviderError`, `combobox.BodyOOB`, `combobox.TriggerLabelOOB` | Use `combobox.Combobox` for the public primitive and `combobox.Handler` for the supported server workflow. The fragments are private implementation details. |
| `combobox.Config.DepsSelector`, `combobox.Config.HXIncludeSelector`, `combobox.Config.IsClientMode`, `combobox.State.IsSelected` | Use `Config.Validate`, `Config.InitialState`, and the public handler contract; render-only state helpers are private. |
| `table.ActionButton`, `table.StatusBadge` | Compose `button.Button` and `badge.Badge`. |
| `table.TableHead`, `table.TableBody`, `table.TablePagination` | Use `table.Table` for initial rendering. Server responses may use the supported `TableHeadContent`, `TableRows`, `TableRow`, and `TablePaginationNav` primitives. |
| `table.ActionButtonClasses`, `table.StatusBadgeClasses`, `table.ColumnCellClasses`, `table.ColumnHeaderClasses`, `table.BadgeCellClasses` | Compose components or use public config fields; class assembly is private. |
| `table.Config.CellClasses`, `CheckboxClasses`, `ColCount`, `ContainerClasses`, `FilterBarID`, `GetID`, `HTMXEndpointValue`, `HTMXTargetValue`, `HasActionableRows`, `HasActions`, `HasExpandableRows`, `HasFilters`, `HasLinkedRows`, `HasSortableColumns`, `HeaderCellClasses`, `LazyLoadTrigger`, `PaginationBaseURL`, `PaginationID`, `RowClasses`, `SortableHeaderClasses`, `TableClasses`, `TbodyClasses`, `TbodyID`, `TheadClasses`, `TheadID` | Removed/private render and resolved-value helpers. Keep using the documented `Config` fields. |
| `table.FilterConfig.ResolvedHxSwap`, `table.FilterConfig.ResolvedHxTarget` | Configure `FilterHTMXConfig`; resolved values are private. |
| `table.PaginationConfig.GetContainerHeight`, `IsContained`, `IsInfiniteScroll`, `NextPage`, `PaginationPages` | Configure pagination and render through the public table primitives. |
| `table.Row.ClickableRole`, `table.Row.HasHTMXAction`, `table.Row.IsActionable` | Configure `Row.Link`, `LinkMode`, `OnClick`, or `HTMX`; render predicates are private. |

Five table URL/sort helpers intentionally remain public:
`Config.IsSortedBy`, `Config.NextPageURL`, `Config.NextSortDir`,
`Config.PageURL`, and `Config.SortURL`.

#### Other exact top-level removals

| Removed symbol | Migration |
| --- | --- |
| `avatar.UserIcon` | Supply `Config.Icon` when a custom icon is needed, or use the component's built-in fallback. |
| `navbar.LinkClasses`, `navbar.MenuItemClasses` | Use `Navbar` configuration/attributes; class assembly is private. |
| `radio.BadgeClasses` | Use `Config.BadgeColor` for the supported radio label treatment or compose `badge.Badge`. |
| `search.JSString` | Remove direct calls; JavaScript escaping is internal to Search. |
| `select.ToOptions` | Map the consumer slice to `[]selectfield.Option` in application code. |
| `tabs.ActiveClasses`, `tabs.InactiveClasses`, `tabs.BadgeActiveClasses`, `tabs.BadgeInactiveClasses` | Use `Tabs` configuration; state class assembly is private. |
| `palette.DefaultHues`, `palette.DefaultShades` | Omit `Config.Hues`/`Shades` for built-in defaults or pass owned slices. Mutable package defaults are private. |
| `rating.EmojiOption`, its fields `rating.EmojiOption.Value`, `rating.EmojiOption.Label`, `rating.EmojiOption.Icon`, and `rating.DefaultEmojiOptions` | Emoji render data is private. Choose `rating.AppearanceEmoji`; custom emoji sets are not a supported API. |

Across the remaining component packages, exported CSS-class assemblers,
default/ID resolvers, and render-only predicates were similarly privatized.
If an application called a method whose result was fed back into Goshtoso
markup, remove that call and use the documented config field or composition
hook. Intentional non-rendering helpers such as validation, URL construction,
initial-state creation, initials derivation, and schema transforms remain
public and appear in the generated component reference.

## Behavior and effective defaults

Source breaks and runtime changes are separate. Renaming a semantic color to
`Tone` preserves its underlying word. The following behavior changes require
review even after code compiles:

- `Carousel` no longer obeys a single presentation flag. It renders an overlay
  when a slide has a title, description, or a complete CTA. A CTA link renders
  only when both `CTALabel` and `CTAHref` are present. Card rendering moved to
  `CardCarousel`.
- `Modal` keeps ordinary `role="dialog"` behavior. `AlertDialog` renders
  `role="alertdialog"`, owns semantic `Tone`, has one action, and has a
  close/dismiss control.
- `Rating` is always an interactive radio group. `RatingDisplay` is
  non-interactive, uses `role="img"`, has no form inputs, and uses an explicit
  `Label` before its value-derived accessible-label fallback.
- `Toast` client events distinguish `kind: "toast"` plus `tone` from
  `kind: "message-toast"` plus sender/message data. Message toasts no longer
  render a hard-coded Reply button; the action exists only when `ActionLabel`
  is supplied. `DismissLabel` defaults to `Dismiss`.
- Custom Tooltip triggers now place `aria-describedby` on the actual focusable
  descendant, preserve consumer-owned state across initialization, and avoid
  promoting disabled controls to interactive wrappers.
- Search revalidates both typed and attribute-provided navigation targets at
  the final client-side navigation sink.

The constructor and zero-value effective defaults that commonly affect an
upgrade are:

| API | Effective default |
| --- | --- |
| `button.Button()` | `TonePrimary`, `SizeMedium`, native `type="button"` |
| `link.Link(href)` | `AppearanceText`, `SizeMedium`, `IconTrailing`; `href` is now a required argument |
| `kbd.Kbd(text)` | `SizeMD`; `text` is now a required argument |
| `tooltip.Tooltip(id, label)` | `PositionTop`, `ActivationHover`, trigger label `Hover Me`; ID and accessible label are now explicit arguments |
| `accordion.Appearance` | zero value / `AppearanceDefault` renders the standard treatment; `AllowMultiple == false` is single-open |
| `badge.Appearance` | zero value / `AppearanceSolid` renders solid |
| `card.Appearance` | zero value / `AppearanceDefault` renders the standard card |
| `fileinput.Appearance` | zero value / `AppearanceDropZone` |
| `pagination.Mode` | zero value / `ModeEllipsis` |
| `table.Appearance`, `table.FilterAppearance` | zero values / standard table and filter bar |
| `toggle.Appearance` | zero value / `AppearanceDefault` |
| `rating.Config.Max`, `rating.DisplayConfig.Max` | `5`; out-of-range values are clamped |
| `banner.CookieBannerConfig` | title `Cookie Consent`, default cookie icon, actions `Accept` and `Decline` |
| `toast.Config.DisplayDuration`, `toast.MessageConfig.DisplayDuration`, `toast.ContainerConfig.DisplayDuration` | `8000ms`; a negative per-toast duration keeps a server-rendered toast visible |

Several zero/default constants changed their string representation while
keeping the same effective rendering: `accordion.AppearanceDefault`,
`badge.AppearanceSolid`, `card.AppearanceDefault`,
`pagination.ModeEllipsis`, `table.AppearanceDefault`, and
`toggle.AppearanceDefault` are `""`. Do not persist or compare their raw
strings; use the typed constants.

The documentation contract now records effective rendered behavior rather than
pretending every Go zero value is literal output. In particular, verify these
current contracts in downstream code:

- `select.Config.InputAttrs` is exported but the custom Select renderer does
  not currently place it on a DOM element.
- `combobox.Config.Required` is not rendered as `required` or
  `aria-required`; `Source.LazyEndpoint` selects server mode, while
  `OptionsEndpoint` is the URL used for search/retry requests; `State.Deps` is
  handler state and is not read by the renderer.
- `schemaform.Field.Name` is populated by transforms, while rendering derives
  native names and IDs from `Field.Path`.
- `form.FooterConfig.CancelHTMX` adds HTMX attributes but does not remove a
  simultaneous `CancelHref`. `FieldGroupConfig` has no built-in Select field;
  its first non-nil supported built-in field wins.
- Checkbox checked/disabled and helper-text branches are conditional. Do not
  assume every combination emits every `Name`, `Value`, `checked`, and
  `disabled` attribute.
- Alert action `DismissLabel` changes copy; it does not attach an application
  dismiss handler. Toast OOB helpers target `toast-container-oob`, independent
  of a custom `ContainerConfig.ID`.
- Dropdown `Section` has items but no rendered heading. Sidebar item fields
  apply differently to ordinary, nested, and disclosure branches. Lazy Tabs
  currently expose both declarative `hx-get` and an Alpine `htmx.ajax` path.

These are behavior/effective-default notes, not new compatibility promises.
Test the branches an application depends on.

## Mechanical upgrade checklist

1. Upgrade all Goshtoso imports together and format the result.
2. Find old dimension names and struct fields:

   ```bash
   rg -n '\.Variant\b|\bVariant[A-Z]|\b[A-Z][A-Za-z]+Variant\b|\.Style\b' .
   ```

3. Find the four removed config constructors:

   ```bash
   rg -n '\b(button|link|kbd|tooltip)\.Config\b' .
   ```

4. Find split-primitive and removed-field call sites:

   ```bash
   rg -n 'AlertMode|CookieBanner|OnCard|ReadOnly|toast\.Message\b|\.Sender\b|HasPrice|HasRating' .
   ```

5. Find accidental or private implementation entry points:

   ```bash
   rg -n 'AccordionItemData|card\.StarRating' .
   rg -n 'table\.(ActionButton|StatusBadge|TableHead|TableBody|TablePagination)\b' .
   rg -n 'combobox\.(Body|OptionsList|ClientScript|ProviderError|BodyOOB|TriggerLabelOOB)\b' .
   ```

6. Find raw constant comparisons that may depend on the changed empty-string
   defaults:

   ```bash
   rg -n '"(default|solid|ellipsis)"|string\([^)]*(Appearance|Mode)' .
   ```

7. Update exact function types from `func(...) templ.Component` where the
   assigned constructor now has a concrete return type. Prefer accepting
   `components.Component` values unless the concrete type matters.
8. Run `go test ./...` in every module, regenerate the consumer application's
   templ output, and render-test affected pages in light/dark and supported
   themes.
9. Consult the generated
   `.agents/skills/using-goshtoso/references/components-reference.md` for the
   current public constructors, options, enums, structs, and fields.

The release's documentation contract enforces all 42 component pages, all 74
Kinds, all public configuration/data structs and fields, constructor
signatures, effective defaults, and generated-reference parity. Its
smoke-test contract directly loads all 42 pages, navigates all 42 destinations
through the actual HTMX path, and checks a representative light/dark
Goshtoso/Minimal theme matrix. Those guarantees complement, rather than
replace, consumer tests for application-specific behavior.
