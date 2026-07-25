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

<!-- compile-current: dimensions -->
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

<!-- compile-current: button -->
```go
button.Button(
    button.WithTone(button.ToneDanger),
    button.WithSize(button.SizeSmall),
    button.WithType("submit"),
    button.WithID("remove"),
)
```

Link before:

```go
link.Link(link.Config{
    Href:   "/docs",
    Style:  link.StyleButton,
    Target: "_blank",
})
```

Link after:

<!-- compile-current: link -->
```go
link.Link("/docs",
    link.WithAppearance(link.AppearanceButton),
    link.WithTarget("_blank"),
)
```

Kbd before:

```go
kbd.Kbd(kbd.Config{Text: "K", Label: "Command K", Size: kbd.SizeSM})
```

Kbd after:

<!-- compile-current: kbd -->
```go
kbd.Kbd("K", kbd.WithLabel("Command K"), kbd.WithSize(kbd.SizeSM))
```

Tooltip before:

```go
tooltip.Tooltip(tooltip.Config{
    ID:          "save-help",
    Label:       "Saves changes",
    Position:    tooltip.Bottom,
    TriggerMode: tooltip.Click,
})
```

Tooltip after:

<!-- compile-current: tooltip -->
```go
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

<!-- compile-current: modal -->
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

Sixty-seven same-name public constructors changed their return type from
`templ.Component` to an exported concrete value that implements
`components.Component`. Seven constructors are new or renamed split
primitives: `banner.CookieBanner`, `carousel.CardCarousel`,
`modal.AlertDialog`, `rating.RatingDisplay`, `toast.MessageToast`,
`toast.OOBMessageToast`, and `toast.ToastContainer`. Together they form the
current 74-constructor inventory. Ordinary rendering remains source-compatible
because each concrete value still has
`Render(context.Context, io.Writer) error`. Code that stores same-name
constructors as function values, declares exact return types, or mocks
`templ.Component` factories must update its type signatures.

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

<!-- compile-current: kind -->
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

#### Complete removed public method inventory

The following table is the exact qualified method inventory removed since
`v0.0.11`. A method listed as renderer-owned has no public replacement: remove
the call and configure the documented field, attribute, slot, or composition
hook instead. This list intentionally names every method even when several
share the same migration action.

| Removed public methods | Migration or action |
| --- | --- |
| `accordion.AccordionConfig.ContainerClasses` | Use `AccordionConfig.RootClass`; the base class list is renderer-owned. |
| `accordion.AccordionItemData.CollapsedClasses`, `accordion.AccordionItemData.ContentClasses`, `accordion.AccordionItemData.ExpandedClasses`, `accordion.AccordionItemData.ItemButtonClasses`, `accordion.AccordionItemData.ItemContainerClasses` | Stop constructing `AccordionItemData`; pass `AccordionConfig` and `AccordionItem` values to `accordion.Accordion`. |
| `alert.Config.ContainerClasses`, `alert.Config.IconBadgeClasses`, `alert.Config.InnerClasses`, `alert.Config.LinkClasses`, `alert.Config.ListClasses`, `alert.Config.PrimaryActionClasses`, `alert.Config.TitleClasses` | Configure Alert content, action, Tone, and `RootClass`; all class assembly is renderer-owned. |
| `avatar.Config.BorderClasses`, `avatar.Config.HasImage`, `avatar.Config.HasInitials`, `avatar.Config.RadiusClasses`, `avatar.Config.ResolvedInitials`, `avatar.Config.ShapeClasses`, `avatar.Config.SizeClasses`, `avatar.Config.SpinnerSizeClasses`, `avatar.Config.StatusClasses`, `avatar.Config.StatusSizeClasses`, `avatar.Config.VariantClasses`, `avatar.Config.VariantFillClasses` | Set the documented image, initials, shape, size, Tone, and status fields; fallback selection and classes are renderer-owned. |
| `badge.Config.IndicatorClasses`, `badge.Config.IsSoft`, `badge.Config.SizeClasses`, `badge.Config.SizeTextClass`, `badge.Config.SoftInnerClasses`, `badge.Config.SoftVariantClasses`, `badge.Config.VariantClasses` | Set `Tone`, `Appearance`, and `Size`; renderer-owned predicates and classes have no public replacement. |
| `banner.Config.CTAClasses`, `banner.Config.ContainerClasses`, `banner.Config.CookieContainerClasses`, `banner.Config.LinkClasses`, `banner.Config.TextClasses` | Configure Banner fields and `RootClass`; use `CookieBanner` for consent behavior. Class assembly is renderer-owned. |
| `card.Config.ContainerClasses`, `card.Config.ContentClasses`, `card.Config.DescriptionClasses`, `card.Config.HasImage`, `card.Config.HasPrice`, `card.Config.HasRating`, `card.Config.ImageClasses`, `card.Config.ImageContainerClasses`, `card.Config.TagClasses`, `card.Config.TitleClasses` | Configure `card.Card`, its slots, and `RootClass`; compose price content and `rating.RatingDisplay` instead of calling render predicates. |
| `chatbubble.Config.BubbleClasses`, `chatbubble.Config.DataMine`, `chatbubble.Config.HasAvatar`, `chatbubble.Config.HasHeader`, `chatbubble.Config.IsMine`, `chatbubble.Config.RowClasses` | Set `Side`, sender/avatar fields, `Grouped`, and `RootClass`; row state and classes are renderer-owned. |
| `checkbox.Config.InputClasses`, `checkbox.Config.SvgClasses` | Use Checkbox configuration and `RootClass`; native-input and icon classes are renderer-owned. |
| `codeblock.Config.GetID`, `codeblock.Config.GetLabel` | Read owned input fields in application code if needed. Rendering still derives Label from `Language` and generates a stable content-based ID when `ID` is empty. |
| `combobox.Config.DepsSelector`, `combobox.Config.HXIncludeSelector`, `combobox.Config.IsClientMode`, `combobox.State.IsSelected` | Use `Config.Validate`, `Config.InitialState`, and `combobox.Handler`; selector, mode, and render-state decisions are internal. |
| `drawer.Config.EnterEnd`, `drawer.Config.EnterStart`, `drawer.Config.GetBodyID`, `drawer.Config.OverlayClasses`, `drawer.Config.PanelClasses`, `drawer.Config.StateVar`, `drawer.Config.TitleID` | Configure Drawer ID, position, classes, and Alpine/HTMX hooks through public fields; derived IDs, state names, transitions, and classes are renderer-owned. |
| `dropdown.Config.ButtonClasses`, `dropdown.Config.DangerClasses`, `dropdown.Config.DisabledClasses`, `dropdown.Config.GetTriggerMode`, `dropdown.Config.HasDividers`, `dropdown.Config.HasIcons`, `dropdown.Config.HasShortcuts`, `dropdown.Config.IsContextMenu`, `dropdown.Config.ItemClasses`, `dropdown.Config.MenuClasses`, `dropdown.Config.UseIconOnlyTrigger`, `dropdown.Item.IsButton` | Set `TriggerMode`, sections/items, `TriggerIcon`, and `TriggerIconOnly`. Empty `TriggerMode` still renders click mode; layout predicates and classes are internal. |
| `fileinput.Config.BrowseLabelClasses`, `fileinput.Config.ContainerClasses`, `fileinput.Config.DropZoneClasses`, `fileinput.Config.HelperTextClasses`, `fileinput.Config.IsUpload`, `fileinput.Config.LabelClasses`, `fileinput.Config.UploadButtonClasses`, `fileinput.Config.UploadContainerClasses`, `fileinput.Config.UploadControlClasses`, `fileinput.Config.UploadFileNameClasses` | Select `AppearanceDropZone` or `AppearanceUpload` and use documented class/attribute fields; mode checks and classes are renderer-owned. |
| `form.FormErrorsConfig.GetID`, `form.FormErrorsConfig.GetTitle` | Set `ID` or `Title` when application code needs explicit values. Rendering still defaults them to `form-errors` and `Validation failed`. |
| `kbd.Config.AccessibleLabel`, `kbd.Config.IconClasses`, `kbd.Config.RootClasses`, `kbd.Config.SizeClasses` | Replace `kbd.Config` with `kbd.Kbd(text, options...)`; use `WithLabel`, `WithIcon`, `WithRootClass`, and `WithSize`. |
| `modal.Config.AlertCTAClasses`, `modal.Config.DialogClasses`, `modal.Config.HeaderClasses`, `modal.Config.IconBadgeClasses`, `modal.Config.StateVar`, `modal.Config.TitleID`, `modal.Config.TriggerClasses` | Use `Modal` or `AlertDialog` plus their public configuration. Classes, Alpine state, and derived IDs are renderer-owned. |
| `navbar.Config.LeftActions`, `navbar.Config.NavClasses`, `navbar.Config.RightActions` | Supply documented Navbar items/actions and classes; action partitioning and base classes are renderer-owned. |
| `pagination.Config.EllipsisClasses`, `pagination.Config.ListClasses`, `pagination.Config.NavClasses`, `pagination.Config.PageClasses`, `pagination.Config.PrevNextClasses`, `pagination.Config.SwapStrategy` | Set `NavClass` and `HTMX.Swap`; rendering still defaults an empty swap to `innerHTML`. Other class helpers are internal. |
| `palette.Config.ContainerClasses` | Use Palette configuration and `RootClass`; base classes are renderer-owned. |
| `radio.Config.HasAlpine`, `radio.Config.HasHTMX`, `radio.Config.InputClasses`, `radio.Config.SegmentedLabelClasses` | Configure `Alpine`, `HTMX`, `Tone`, and public class/attribute fields; wiring predicates and classes are renderer-owned. |
| `radio.HTMXConfig.EffectiveTrigger`, `radio.HTMXConfig.HasHxVerb` | Set `HTMXConfig.Trigger` explicitly when application code needs the value. Rendering uses it, otherwise defaults to `change` when any HTMX verb is set, otherwise emits no trigger. |
| `range.Config.AlpineData`, `range.Config.ControlClasses`, `range.Config.IconClasses`, `range.Config.InputClasses`, `range.Config.LabelClasses`, `range.Config.MaxOrDefault`, `range.Config.MinOrDefault`, `range.Config.RootClasses`, `range.Config.StepOrDefault`, `range.Config.TickClasses`, `range.Config.TickLabels`, `range.Config.ValueClasses`, `range.Config.ValueOrDefault` | Use Range fields and class/attribute hooks directly. Empty `Value`, `Min`, `Max`, and `Step` still render as `0`, `0`, `100`, and `1`; Alpine data, ticks, and classes are internal. |
| `rating.Config.ActiveIconClasses`, `rating.Config.BindClass`, `rating.Config.ControlClasses`, `rating.Config.EmojiIcon`, `rating.Config.IconClasses`, `rating.Config.InactiveIconClasses`, `rating.Config.InputID`, `rating.Config.IsActive`, `rating.Config.ResolvedID`, `rating.Config.ResolvedLabel`, `rating.Config.ResolvedMax`, `rating.Config.ResolvedName`, `rating.Config.ResolvedValue`, `rating.Config.RootClasses`, `rating.Config.ValueLabel`, `rating.Config.XData` | Use `Rating` for input or `RatingDisplay` for output and set their documented fields. ID/name/label/value resolution, clamping, icons, Alpine data, and classes are renderer-owned. |
| `search.Config.DialogClasses`, `search.Config.GetDescriptionMaxLength`, `search.Config.GetEmptyText`, `search.Config.GetEscapeText`, `search.Config.GetID`, `search.Config.GetLabel`, `search.Config.GetMaxResults`, `search.Config.GetPlaceholder`, `search.Config.GetShortcutText`, `search.Config.RootClasses`, `search.Config.TriggerClasses` | Set the corresponding Config fields directly. Rendering still defaults ID `search`, Label/Placeholder `Search`, ShortcutText `⌘ K`, EscapeText `Esc`, EmptyText `No results found.`, MaxResults `4`, and DescriptionMaxLength `120`; classes are internal. |
| `select.Config.ContainerClasses`, `select.Config.GetPlaceholder`, `select.Config.IsEffectivelyDisabled`, `select.Config.LabelClasses`, `select.Config.SelectClasses`, `select.Config.SelectedValue`, `select.Config.TriggerClasses` | Set Select fields and `RootClass`; empty Placeholder still renders `Please Select`, and disabled/readonly/selection/class resolution is renderer-owned. |
| `sidebar.Config.ContainerClasses`, `sidebar.Config.NavClasses` | Configure Sidebar and `RootClass`; base layout classes are renderer-owned. |
| `sidebar.OverlayConfig.BackdropClasses`, `sidebar.OverlayConfig.PanelClasses`, `sidebar.OverlayConfig.PanelID`, `sidebar.OverlayConfig.RootClasses`, `sidebar.OverlayConfig.StateVar`, `sidebar.OverlayConfig.TriggerClasses`, `sidebar.OverlayConfig.TriggerLabelText` | Configure `sidebar.Overlay` through public fields. Derived panel/state IDs, fallback trigger text, and classes are renderer-owned. |
| `spinner.Config.FillClasses`, `spinner.Config.SizeClasses` | Set Spinner `Tone` and `Size`; class assembly is renderer-owned. |
| `structuredinput.Column.DefaultValue`, `structuredinput.Column.EntryAccessor`, `structuredinput.Column.NameBinding`, `structuredinput.Config.ContainerClasses`, `structuredinput.Config.EntriesJSON`, `structuredinput.Config.GetAddLabel`, `structuredinput.Config.NewRowJSON`, `structuredinput.Config.NormalizedColumns`, `structuredinput.Option.OptionLabel` | Supply Columns, Options, Entries, and `AddActionLabel`; rendering still falls back to `Add row`, option Value labels, and first select-option defaults. Normalization, bindings, JSON, and classes are internal. |
| `table.Config.CellClasses`, `table.Config.CheckboxClasses`, `table.Config.ColCount`, `table.Config.ContainerClasses`, `table.Config.FilterBarID`, `table.Config.GetID`, `table.Config.HTMXEndpointValue`, `table.Config.HTMXTargetValue`, `table.Config.HasActionableRows`, `table.Config.HasActions`, `table.Config.HasExpandableRows`, `table.Config.HasFilters`, `table.Config.HasLinkedRows`, `table.Config.HasSortableColumns`, `table.Config.HeaderCellClasses`, `table.Config.LazyLoadTrigger`, `table.Config.PaginationBaseURL`, `table.Config.PaginationID`, `table.Config.RowClasses`, `table.Config.SortableHeaderClasses`, `table.Config.TableClasses`, `table.Config.TbodyClasses`, `table.Config.TbodyID`, `table.Config.TheadClasses`, `table.Config.TheadID` | Keep using documented Table fields and the five supported URL/sort helpers. IDs, HTMX values, feature predicates, triggers, column counts, and classes are renderer-owned. |
| `table.FilterConfig.ResolvedHxSwap`, `table.FilterConfig.ResolvedHxTarget` | Set `FilterHTMXConfig` values explicitly when needed; default target/swap resolution is renderer-owned. |
| `table.PaginationConfig.GetContainerHeight`, `table.PaginationConfig.IsContained`, `table.PaginationConfig.IsInfiniteScroll`, `table.PaginationConfig.NextPage`, `table.PaginationConfig.PaginationPages` | Configure pagination and use public Table fragments for server responses; containment, infinite-scroll, page-window, and next-page calculations are internal. |
| `table.Row.ClickableRole`, `table.Row.HasHTMXAction`, `table.Row.IsActionable` | Set `Row.Link`, `LinkMode`, `OnClick`, or `HTMX`; actionability and role derivation are renderer-owned. |
| `tagslist.Config.AlpineData`, `tagslist.Config.ContainerClasses`, `tagslist.Config.GetAddLabel`, `tagslist.Config.GetPlaceholder` | Set `AddActionLabel`, `Placeholder`, Values, Name, and `RootClass`. Rendering still defaults to `Add` and `Add a tag...`; Alpine data and classes are internal. |
| `textarea.Config.ContainerClasses`, `textarea.Config.GetRows`, `textarea.Config.HelperTextClasses`, `textarea.Config.LabelClasses`, `textarea.Config.TextareaClasses` | Set `Rows`, State, and `RootClass`; zero Rows still renders `3`, while validation and class selection are renderer-owned. |
| `textinput.Config.ContainerClasses`, `textinput.Config.GetType`, `textinput.Config.HasMask`, `textinput.Config.HasMaxLength`, `textinput.Config.HasPattern`, `textinput.Config.HelperTextClasses`, `textinput.Config.InputClasses`, `textinput.Config.IsPassword`, `textinput.Config.IsSearch`, `textinput.Config.LabelClasses`, `textinput.Config.MaxLengthStr` | Set the documented Type, Mask, Pattern, MaxLength, State, and class fields; empty Type still renders `text`. Predicates, formatting, and classes are renderer-owned. |
| `toast.Config.BgClass`, `toast.Config.BorderClass`, `toast.Config.HasAction`, `toast.Config.IconBgClass`, `toast.Config.TitleClass` | Set Toast `Tone`, content, and action fields or use the MessageToast split; action predicates and semantic classes are renderer-owned. |
| `toggle.Config.LabelClasses`, `toggle.Config.ToggleClasses` | Set Toggle Tone, Appearance, Size, and class/attribute fields; class assembly is renderer-owned. |

Intentional non-rendering helpers such as validation, URL construction,
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

6. Find every removed public method name. The qualified table above resolves
   false positives when two packages used the same method name:

   ```bash
   removed_methods='AccessibleLabel|ActiveIconClasses|AlertCTAClasses|AlpineData|BackdropClasses|BgClass|BindClass|BorderClass|BorderClasses|BrowseLabelClasses|BubbleClasses|ButtonClasses|CTAClasses|CellClasses|CheckboxClasses|ClickableRole|ColCount|CollapsedClasses|ContainerClasses|ContentClasses'
   removed_methods+='|ControlClasses|CookieContainerClasses|DangerClasses|DataMine|DefaultValue|DepsSelector|DescriptionClasses|DialogClasses|DisabledClasses|DropZoneClasses|EffectiveTrigger|EllipsisClasses|EmojiIcon|EnterEnd|EnterStart|EntriesJSON|EntryAccessor|ExpandedClasses|FillClasses|FilterBarID'
   removed_methods+='|GetAddLabel|GetBodyID|GetContainerHeight|GetDescriptionMaxLength|GetEmptyText|GetEscapeText|GetID|GetLabel|GetMaxResults|GetPlaceholder|GetRows|GetShortcutText|GetTitle|GetTriggerMode|GetType|HTMXEndpointValue|HTMXTargetValue|HXIncludeSelector|HasAction|HasActionableRows'
   removed_methods+='|HasActions|HasAlpine|HasAvatar|HasDividers|HasExpandableRows|HasFilters|HasHTMX|HasHTMXAction|HasHeader|HasHxVerb|HasIcons|HasImage|HasInitials|HasLinkedRows|HasMask|HasMaxLength|HasPattern|HasPrice|HasRating|HasShortcuts'
   removed_methods+='|HasSortableColumns|HeaderCellClasses|HeaderClasses|HelperTextClasses|IconBadgeClasses|IconBgClass|IconClasses|ImageClasses|ImageContainerClasses|InactiveIconClasses|IndicatorClasses|InnerClasses|InputClasses|InputID|IsActionable|IsActive|IsButton|IsClientMode|IsContained|IsContextMenu'
   removed_methods+='|IsEffectivelyDisabled|IsInfiniteScroll|IsMine|IsPassword|IsSearch|IsSelected|IsSoft|IsUpload|ItemButtonClasses|ItemClasses|ItemContainerClasses|LabelClasses|LazyLoadTrigger|LeftActions|LinkClasses|ListClasses|MaxLengthStr|MaxOrDefault|MenuClasses|MinOrDefault'
   removed_methods+='|NameBinding|NavClasses|NewRowJSON|NextPage|NormalizedColumns|OptionLabel|OverlayClasses|PageClasses|PaginationBaseURL|PaginationID|PaginationPages|PanelClasses|PanelID|PrevNextClasses|PrimaryActionClasses|RadiusClasses|ResolvedHxSwap|ResolvedHxTarget|ResolvedID|ResolvedInitials'
   removed_methods+='|ResolvedLabel|ResolvedMax|ResolvedName|ResolvedValue|RightActions|RootClasses|RowClasses|SegmentedLabelClasses|SelectClasses|SelectedValue|ShapeClasses|SizeClasses|SizeTextClass|SoftInnerClasses|SoftVariantClasses|SortableHeaderClasses|SpinnerSizeClasses|StateVar|StatusClasses|StatusSizeClasses'
   removed_methods+='|StepOrDefault|SvgClasses|SwapStrategy|TableClasses|TagClasses|TbodyClasses|TbodyID|TextClasses|TextareaClasses|TheadClasses|TheadID|TickClasses|TickLabels|TitleClass|TitleClasses|TitleID|ToggleClasses|TriggerClasses|TriggerLabelText|UploadButtonClasses'
   removed_methods+='|UploadContainerClasses|UploadControlClasses|UploadFileNameClasses|UseIconOnlyTrigger|ValueClasses|ValueLabel|ValueOrDefault|VariantClasses|VariantFillClasses|XData'
   rg -n "\\.(${removed_methods})[[:space:]]*\\(" .
   ```

7. Find raw constant comparisons that may depend on the changed empty-string
   defaults:

   ```bash
   rg -n '"(default|solid|ellipsis)"|string\([^)]*(Appearance|Mode)' .
   ```

8. Update exact function types from `func(...) templ.Component` where the
   assigned constructor now has a concrete return type. Prefer accepting
   `components.Component` values unless the concrete type matters.
9. Run `go test ./...` in every module, regenerate the consumer application's
   templ output, and render-test affected pages in light/dark and supported
   themes.
10. Consult the generated
   `.agents/skills/using-goshtoso/references/components-reference.md` for the
   current public constructors, options, enums, structs, and fields.

The release's documentation contract enforces all 42 component pages, all 74
Kinds, all public configuration/data structs and fields, constructor
signatures, effective defaults, and generated-reference parity. Its
smoke-test contract directly loads all 42 pages, navigates all 42 destinations
through the actual HTMX path, and checks a representative light/dark
Goshtoso/Minimal theme matrix. Those guarantees complement, rather than
replace, consumer tests for application-specific behavior.
