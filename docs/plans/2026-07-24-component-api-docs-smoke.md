# Component API, Documentation, and Smoke Tests Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use
> superpowers:subagent-driven-development (recommended) or
> superpowers:executing-plans to implement this plan task-by-task. Steps use
> checkbox (`- [ ]`) syntax for tracking.

**Goal:** Deliver a curated Goshtoso component API with runtime primitive
identity, explicit configuration dimensions, complete consumer documentation,
and registry-driven smoke proof for every component page.

**Architecture:** The root `components` package owns stable `Kind` values and
the shared renderable interface. Component packages retain typed configuration
and concrete renderable instances, while the site uses an independent catalog
plus reflection-backed API metadata to keep navigation, documentation, and
tests synchronized. Broad direct-load and HTMX-navigation smoke tests complement
focused component behavior tests.

**Tech Stack:** Go 1.26.5, templ, Tailwind CSS v4, HTMX, Alpine.js,
Playwright Go, testify.

## Global Constraints

- Work only in `/tmp/gs-component-docs-smoke-audit` on branch
  `codex/component-docs-smoke-audit`.
- Never hand-edit `*_templ.go`, `assets/styles.css`, generated theme CSS, or
  generated runtime constants.
- Run `templ generate` after every `.templ` entry-point change.
- Run `just css` only when markup introduces or removes Tailwind utilities.
- Run `go run ./cmd/skillgen` after public component types or entry points
  change.
- Breaking alpha APIs are allowed; do not add compatibility aliases for names
  this plan removes.
- Public dimension types remain package-specific. Do not introduce shared root
  `Tone`, `Appearance`, `Mode`, or `State` enums.
- Do not add a universal `Variant` field or `Variant()` method.
- Use `Dimension`, not `Axis`, in consumer documentation.
- A public default means effective rendered behavior, not the Go zero value.
- Config structs remain for data-heavy components. Functional options are
  limited to Button, Link, Kbd, and Tooltip in this pass.
- Keep root dependencies slim and never import `site/` from the library.
- Each task ends in a focused commit and leaves its affected tests green.

---

## Locked Public Renderable Inventory

The final library has exactly 74 intentional public renderable entry points:

| Package | Public renderables |
| --- | --- |
| accordion | `Accordion` |
| alert | `Alert` |
| avatar | `Avatar`, `AvatarStack` |
| badge | `Badge`, `NotificationBadge`, `NotificationDot`, `AnimatingDot` |
| banner | `Banner`, `CookieBanner` |
| breadcrumbs | `Breadcrumbs` |
| button | `Button` |
| card | `Card` |
| carousel | `Carousel`, `CardCarousel` |
| chatbubble | `ChatBubble`, `TypingIndicator` |
| checkbox | `Checkbox`, `CheckboxGroup` |
| codeblock | `CodeBlock` |
| combobox | `Combobox` |
| drawer | `Drawer` |
| dropdown | `Dropdown` |
| fileinput | `FileInput` |
| form | `Form`, `Section`, `CollapsibleSection`, `FlipSection`, `SubSection`, `FieldGroup`, `FormErrors` |
| head | `Dependencies`, `DependenciesMinimal` |
| kbd | `Kbd` |
| link | `Link` |
| modal | `Modal`, `AlertDialog` |
| navbar | `Navbar` |
| pagination | `Pagination` |
| palette | `Palette` |
| radio | `Radio`, `RadioBar`, `RadioGroup` |
| range | `Range` |
| rating | `Rating`, `RatingDisplay` |
| schemaform | `Fields` |
| search | `Search`, `SearchField`, `SearchModal` |
| select | `Select` |
| sidebar | `Sidebar`, `Overlay` |
| spinner | `Spinner` |
| steps | `Steps` |
| structuredinput | `StructuredInput` |
| table | `Table`, `TableHeadContent`, `TableRows`, `TableRow`, `TablePaginationNav`, `ImageCell` |
| tabs | `Tabs` |
| tagslist | `TagsList` |
| textarea | `Textarea`, `TextareaWithActions` |
| textinput | `TextInput` |
| toast | `ToastContainer`, `Toast`, `MessageToast`, `OOBToast`, `OOBMessageToast` |
| toggle | `Toggle` |
| tooltip | `Tooltip` |

Current renderables that become private:

- combobox: `Body`, `OptionsList`, `ClientScript`, `ProviderError`, `BodyOOB`,
  `TriggerLabelOOB`;
- table: `TableHead`, `TableBody`, `TablePagination`.

Current renderables that are removed:

- `card.StarRating`; consumers use `rating.RatingDisplay`;
- `table.ActionButton`; consumers use `button.Button`;
- `table.StatusBadge`; consumers use `badge.Badge`.

---

### Task 1: Add the root component identity contract

**Files:**

- Create: `components/component.go`
- Create: `components/component_test.go`

**Interfaces:**

- Produces: `type Kind string`
- Produces: `type Component interface { templ.Component; Kind() Kind }`
- Produces: `func AllKinds() []Kind`

- [ ] **Step 1: Write failing Kind registry tests**

```go
package components

func TestAllKindsAreStableAndUnique(t *testing.T) {
    kinds := AllKinds()
    require.Len(t, kinds, 74)
    seen := map[Kind]struct{}{}
    for _, kind := range kinds {
        require.NotEmpty(t, kind)
        _, duplicate := seen[kind]
        require.Falsef(t, duplicate, "duplicate Kind %q", kind)
        seen[kind] = struct{}{}
    }
}

func TestAllKindsReturnsCopy(t *testing.T) {
    kinds := AllKinds()
    kinds[0] = "mutated"
    require.NotEqual(t, Kind("mutated"), AllKinds()[0])
}
```

- [ ] **Step 2: Run the tests and verify the missing API failure**

Run: `go test ./components -run 'TestAllKinds' -count=1`

Expected: compile failure because `Kind` and `AllKinds` do not exist.

- [ ] **Step 3: Implement the interface and all 74 constants**

Use type-prefixed constant names and kebab-case values. The constant set is:

```go
const (
    KindAccordion             Kind = "accordion"
    KindAlert                 Kind = "alert"
    KindAvatar                Kind = "avatar"
    KindAvatarStack           Kind = "avatar-stack"
    KindBadge                 Kind = "badge"
    KindNotificationBadge     Kind = "notification-badge"
    KindNotificationDot       Kind = "notification-dot"
    KindAnimatingDot          Kind = "animating-dot"
    KindBanner                Kind = "banner"
    KindCookieBanner          Kind = "cookie-banner"
    KindBreadcrumbs           Kind = "breadcrumbs"
    KindButton                Kind = "button"
    KindCard                  Kind = "card"
    KindCarousel              Kind = "carousel"
    KindCardCarousel          Kind = "card-carousel"
    KindChatBubble            Kind = "chat-bubble"
    KindTypingIndicator       Kind = "typing-indicator"
    KindCheckbox              Kind = "checkbox"
    KindCheckboxGroup         Kind = "checkbox-group"
    KindCodeBlock             Kind = "code-block"
    KindCombobox              Kind = "combobox"
    KindDrawer                Kind = "drawer"
    KindDropdown              Kind = "dropdown"
    KindFileInput             Kind = "file-input"
    KindForm                  Kind = "form"
    KindFormSection           Kind = "form-section"
    KindFormCollapsibleSection Kind = "form-collapsible-section"
    KindFormFlipSection       Kind = "form-flip-section"
    KindFormSubSection        Kind = "form-sub-section"
    KindFormFieldGroup        Kind = "form-field-group"
    KindFormErrors            Kind = "form-errors"
    KindDependencies          Kind = "dependencies"
    KindDependenciesMinimal   Kind = "dependencies-minimal"
    KindKbd                   Kind = "kbd"
    KindLink                  Kind = "link"
    KindModal                 Kind = "modal"
    KindAlertDialog           Kind = "alert-dialog"
    KindNavbar                Kind = "navbar"
    KindPagination            Kind = "pagination"
    KindPalette               Kind = "palette"
    KindRadio                 Kind = "radio"
    KindRadioBar              Kind = "radio-bar"
    KindRadioGroup            Kind = "radio-group"
    KindRange                 Kind = "range"
    KindRating                Kind = "rating"
    KindRatingDisplay         Kind = "rating-display"
    KindSchemaFormFields      Kind = "schema-form-fields"
    KindSearch                Kind = "search"
    KindSearchField           Kind = "search-field"
    KindSearchModal           Kind = "search-modal"
    KindSelect                Kind = "select"
    KindSidebar               Kind = "sidebar"
    KindSidebarOverlay        Kind = "sidebar-overlay"
    KindSpinner               Kind = "spinner"
    KindSteps                 Kind = "steps"
    KindStructuredInput       Kind = "structured-input"
    KindTable                 Kind = "table"
    KindTableHeadContent      Kind = "table-head-content"
    KindTableRows             Kind = "table-rows"
    KindTableRow              Kind = "table-row"
    KindTablePaginationNav    Kind = "table-pagination-nav"
    KindTableImageCell        Kind = "table-image-cell"
    KindTabs                  Kind = "tabs"
    KindTagsList              Kind = "tags-list"
    KindTextarea              Kind = "textarea"
    KindTextareaWithActions   Kind = "textarea-with-actions"
    KindTextInput             Kind = "text-input"
    KindToastContainer        Kind = "toast-container"
    KindToast                 Kind = "toast"
    KindMessageToast          Kind = "message-toast"
    KindOOBToast              Kind = "oob-toast"
    KindOOBMessageToast       Kind = "oob-message-toast"
    KindToggle                Kind = "toggle"
    KindTooltip               Kind = "tooltip"
)
```

Declare `allKinds` in the same order and return `slices.Clone(allKinds)`.

- [ ] **Step 4: Run identity tests**

Run: `go test ./components -run 'TestAllKinds' -count=1`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add components/component.go components/component_test.go
git commit -m "feat: define component kind identity"
```

---

### Task 2: Rename semantic-color variants to Tone

**Files:**

- Modify: `components/alert/types.go`, `components/alert/alert.templ`,
  `components/alert/*_test.go`
- Modify: `components/avatar/types.go`, `components/avatar/avatar.templ`,
  `components/avatar/*_test.go`
- Modify: `components/badge/types.go`, `components/badge/badge.templ`,
  `components/badge/*_test.go`
- Modify: `components/banner/types.go`, `components/banner/banner.templ`,
  `components/banner/*_test.go`
- Modify: `components/checkbox/types.go`, `components/checkbox/checkbox.templ`,
  `components/checkbox/*_test.go`
- Modify: `components/radio/types.go`, `components/radio/radio.templ`,
  `components/radio/*_test.go`
- Modify: `components/spinner/types.go`, `components/spinner/spinner.templ`,
  `components/spinner/*_test.go`
- Modify: `components/toggle/types.go`, `components/toggle/toggle.templ`,
  `components/toggle/*_test.go`
- Modify: `components/chatbubble/types.go`,
  `components/chatbubble/chatbubble.templ`,
  `components/chatbubble/*_test.go`
- Modify: `components/search/types.go`
- Modify: affected-package callers from:
  `rg -l '\bVariant:|AvatarVariant|AnimatingDot\(' components site --glob '*.go' --glob '*.templ' --glob '!**/*_templ.go'`

**Interfaces:**

- Produces: package-local `Tone` types and `Tone...` constants.
- Removes: the corresponding `Variant` types, fields, and unprefixed
  constants.
- Changes: `chatbubble.Config.AvatarVariant string` to
  `AvatarTone avatar.Tone`.

- [ ] **Step 1: Change focused tests to the new names before source**

Example required shape:

```go
func TestDangerToneRendersDangerTreatment(t *testing.T) {
    html := render(t, Config{Tone: ToneDanger, Title: "Problem"})
    require.Contains(t, html, "border-danger")
}
```

Update every existing variant table so its key type is `Tone` and values are
type-prefixed constants.

- [ ] **Step 2: Verify the new tests fail to compile**

Run:

```bash
go test ./components/alert ./components/avatar ./components/badge \
  ./components/banner ./components/checkbox ./components/radio \
  ./components/spinner ./components/toggle ./components/chatbubble -count=1
(cd site && go test $(go list ./... | grep -v /tests/e2e) -count=1)
```

Expected: compile failures for undefined `Tone`, `ToneDanger`, and `Config.Tone`.

- [ ] **Step 3: Implement the exact mappings**

| Package | Constants |
| --- | --- |
| alert | `ToneInfo`, `ToneSuccess`, `ToneWarning`, `ToneDanger` |
| avatar | `ToneDefault`, `ToneInverse`, `TonePrimary`, `ToneSecondary`, `ToneInfo`, `ToneSuccess`, `ToneWarning`, `ToneDanger` |
| badge | same set as avatar |
| banner | `ToneDefault`, `TonePrimary`, `ToneInfo`, `ToneSuccess`, `ToneWarning`, `ToneDanger` |
| checkbox | `TonePrimary`, `ToneSecondary`, `ToneInfo`, `ToneSuccess`, `ToneWarning`, `ToneDanger` |
| radio | same set as checkbox |
| spinner | `ToneDefault`, `TonePrimary`, `ToneSecondary`, `ToneInfo`, `ToneSuccess`, `ToneWarning`, `ToneDanger` |
| toggle | same set as checkbox |

The string values remain the current wire/CSS values. Rename render-only helper
names such as `VariantClasses` to `toneClasses` while updating templates.
Change badge `AnimatingDot` to accept `Tone`. Change chat-bubble avatar
configuration to a typed `avatar.Tone`; no string cast remains. Rename
search's `methodBadgeVariant` helper to `methodBadgeTone` and return
`badge.Tone`.

- [ ] **Step 4: Regenerate and run focused tests**

```bash
templ generate
go run ./cmd/skillgen
go test ./components/alert ./components/avatar ./components/badge \
  ./components/banner ./components/checkbox ./components/radio \
  ./components/spinner ./components/toggle ./components/chatbubble -count=1
```

Expected: PASS.

- [ ] **Step 5: Prove old public vocabulary is absent in this group**

Run:

```bash
rg -n '\bVariant\b|AvatarVariant|VariantClasses' \
  components/{alert,avatar,badge,banner,checkbox,radio,spinner,toggle,chatbubble} \
  --glob '*.go' --glob '*.templ' --glob '!**/*_templ.go'
```

Expected: no matches.

- [ ] **Step 6: Commit**

```bash
git add components/alert components/avatar components/badge components/banner \
  components/checkbox components/radio components/spinner components/toggle \
  components/chatbubble components/search site .agents .claude
git commit -m "refactor: name semantic component tones"
```

---

### Task 3: Rename appearance and mode dimensions

**Files:**

- Modify: `components/accordion/types.go`,
  `components/accordion/accordion.templ`, tests
- Modify: `components/card/types.go`, `components/card/card.templ`, tests
- Modify: `components/fileinput/types.go`,
  `components/fileinput/fileinput.templ`, tests
- Modify: `components/pagination/types.go`,
  `components/pagination/pagination.templ`, tests
- Modify: `components/table/types.go`, `components/table/table.templ`, tests
- Modify: `components/badge/types.go`, `components/badge/badge.templ`, tests
- Modify: `components/toggle/types.go`, `components/toggle/toggle.templ`, tests
- Modify: affected-package callers from:
  `rg -l '\b(Variant|Style|FilterVariant):|SingleOpen|WithCheckbox|HasPrice|HasRating' components site --glob '*.go' --glob '*.templ' --glob '!**/*_templ.go'`

**Interfaces:**

- Accordion: `AppearanceDefault`, `AppearancePlain`, `AppearanceSplit`.
- Card: `AppearanceDefault`, `AppearancePrimary`.
- FileInput: `AppearanceDropZone`, `AppearanceUpload`.
- Pagination: `ModeEllipsis`, `ModeSimple`.
- Table: `AppearanceDefault`, `AppearanceStriped`.
- Table filters: `FilterAppearanceBar`, `FilterAppearanceInline`.
- Badge: `AppearanceSolid`, `AppearanceSoft`.
- Toggle: `AppearanceDefault`, `AppearanceContainer`.

- [ ] **Step 1: Rewrite focused tests around the named dimensions**

```go
func TestStripedAppearance(t *testing.T) {
    html := renderT(t, Table(Config{Appearance: AppearanceStriped}))
    require.Contains(t, html, "odd:bg-surface-alt")
}

func TestCheckboxSelectionIsBehaviorNotAppearance(t *testing.T) {
    html := renderT(t, Table(Config{ShowCheckbox: true}))
    require.Contains(t, html, `type="checkbox"`)
}
```

Add assertions that accordion single-open behavior is still the zero-value
`AllowMultiple == false`, not an appearance constant.

- [ ] **Step 2: Verify compile failures**

Run:

```bash
go test ./components/accordion ./components/card ./components/fileinput \
  ./components/pagination ./components/table ./components/badge \
  ./components/toggle -count=1
(cd site && go test $(go list ./... | grep -v /tests/e2e) -count=1)
```

Expected: compile failures for the new fields/types.

- [ ] **Step 3: Apply the exact public changes**

- Remove accordion `SingleOpen`.
- Remove table `WithCheckbox`; `ShowCheckbox` is the only checkbox-selection
  switch.
- Remove card `Price`, `Rating`, `HasPrice`, and `HasRating`.
- Rename badge `Style` to `Appearance`.
- Rename toggle `Style` to `Appearance`.
- Rename table `FilterVariant` to `FilterAppearance` and
  `FilterConfig.Variant` to `FilterConfig.Appearance`.
- Keep existing string values where they are meaningful; use the empty string
  for the existing zero/default behavior.

- [ ] **Step 4: Regenerate and run tests**

```bash
templ generate
go run ./cmd/skillgen
go test ./components/accordion ./components/card ./components/fileinput \
  ./components/pagination ./components/table ./components/badge \
  ./components/toggle -count=1
```

Expected: PASS.

- [ ] **Step 5: Verify removed names are gone**

```bash
rg -n '\bSingleOpen\b|\bWithCheckbox\b|\bFilterVariant\b|\.Variant\b|\.Style\b|HasPrice|HasRating' \
  components/{accordion,card,fileinput,pagination,table,badge,toggle} \
  --glob '*.go' --glob '*.templ' --glob '!**/*_templ.go'
```

Expected: no obsolete API matches.

- [ ] **Step 6: Commit**

```bash
git add components/accordion components/card components/fileinput \
  components/pagination components/table components/badge components/toggle \
  site .agents .claude
git commit -m "refactor: name component appearance and mode dimensions"
```

---

### Task 4: Split materially different structural and interaction primitives

**Files:**

- Modify: `components/carousel/types.go`,
  `components/carousel/carousel.templ`, carousel tests
- Modify: `components/modal/types.go`, `components/modal/modal.templ`, modal
  tests
- Modify: `components/toast/types.go`, `components/toast/toast.templ`, toast
  tests
- Modify: `components/banner/types.go`, `components/banner/banner.templ`,
  banner tests
- Modify: `components/rating/types.go`, `components/rating/rating.templ`,
  rating tests
- Modify: affected-package callers reported by:
  `rg -l 'carousel\.(OnCard|WithText|WithCTA)|AlertMode|modal\.(Default|Success|Info|Warning|Danger)|toast\.(Container|Info|Success|Warning|Danger|Message)|CookieBanner|CookieConfig|ReadOnly|\bVariant:' components site --glob '*.go' --glob '*.templ' --glob '!**/*_templ.go'`

**Interfaces:**

- Produces: `carousel.CardConfig` and `CardCarousel(CardConfig)`.
- Produces: `modal.AlertDialogConfig` and
  `AlertDialog(AlertDialogConfig)`.
- Produces: `toast.Tone`, `toast.Config`, `toast.MessageConfig`,
  `Toast`, `MessageToast`, `OOBToast`, and `OOBMessageToast`.
- Produces: `banner.CookieBannerConfig` and
  `CookieBanner(CookieBannerConfig)`.
- Produces: `rating.DisplayConfig` and
  `RatingDisplay(DisplayConfig)`.
- Removes: carousel `Variant`, modal `AlertMode`/`Variant`, toast
  `Variant`/`Message`, banner `Config.CookieBanner`/`CookieConfig`, and rating
  `Config.ReadOnly`.

- [ ] **Step 1: Write failing structural-contract tests**

```go
func TestAlertDialogUsesAlertDialogRole(t *testing.T) {
    html := render(t, AlertDialog(AlertDialogConfig{
        ID: "delete", Title: "Delete?", ActionLabel: "Delete",
        Tone: ToneDanger,
    }))
    require.Contains(t, html, `role="alertdialog"`)
    require.NotContains(t, html, "SecondaryLabel")
}

func TestCarouselInfersOverlayFromSlideContent(t *testing.T) {
    html := render(t, Carousel(Config{Slides: []Slide{{
        ImgSrc: "/x.webp", Title: "Release", CTALabel: "Read",
        CTAHref: "/release",
    }}}))
    require.Contains(t, html, "Release")
    require.Contains(t, html, "/release")
}

func TestMessageToastHasItsOwnConfig(t *testing.T) {
    html := renderToastComponent(t, MessageToast(MessageConfig{
        Sender: Sender{Name: "Ada"}, Message: "Review ready",
        ActionLabel: "Open",
    }))
    require.Contains(t, html, "Ada")
    require.Contains(t, html, "Open")
}

func TestCookieBannerOwnsDialogContract(t *testing.T) {
    html := render(t, CookieBanner(CookieBannerConfig{
        Description: "We use local storage.",
    }))
    require.Contains(t, html, `role="dialog"`)
    require.Contains(t, html, "Cookie Consent")
}

func TestRatingDisplayHasNoFormInputs(t *testing.T) {
    html := render(t, RatingDisplay(DisplayConfig{
        Value: 4,
        Label: "Four out of five",
    }))
    require.Contains(t, html, `role="img"`)
    require.NotContains(t, html, `type="radio"`)
}
```

- [ ] **Step 2: Run and verify compile failures**

Run:

```bash
go test ./components/carousel ./components/modal ./components/toast \
  ./components/banner ./components/rating -count=1
```

Expected: undefined new config types and constructors.

- [ ] **Step 3: Implement Carousel and CardCarousel**

`Carousel` keeps `Config` without a `Variant`. A slide with title,
description, or CTA content renders the overlay; a CTA renders only when both
label and href are non-empty. `CardCarousel` accepts:

```go
type CardConfig struct {
    ID        string
    Slides    []Slide
    Touch     bool
    Height    string
    RootClass string
}
```

It owns the article/card wrapper and body. It does not silently accept autoplay
or lazy-loader fields that the current card branch ignores.

- [ ] **Step 4: Implement Modal and AlertDialog**

`modal.Config` loses `AlertMode` and `Variant`. Add:

```go
type AlertDialogConfig struct {
    ID           string
    Title        string
    Body         string
    TriggerLabel string
    ActionLabel  string
    Action       *ButtonAction
    Tone         Tone
    PanelClass   string
}
```

`Tone` has `ToneDefault`, `ToneSuccess`, `ToneInfo`, `ToneWarning`, and
`ToneDanger`, preserving the current string values.

`AlertDialog` uses `role="alertdialog"` and has a single action plus dismiss
control. `Modal` retains primary and secondary actions and uses `role="dialog"`.

- [ ] **Step 5: Implement Toast and MessageToast**

```go
type Config struct {
    Tone            Tone
    Title           string
    Message         string
    DisplayDuration int
    ActionLabel     string
    ActionHTMX      *HTMXConfig
}

type MessageConfig struct {
    Sender          Sender
    Message         string
    DisplayDuration int
    ActionLabel     string
    ActionHTMX      *HTMXConfig
    DismissLabel    string
}
```

Toast `Tone` has `ToneInfo`, `ToneSuccess`, `ToneWarning`, and `ToneDanger`.
Rename `Container` to `ToastContainer`. Client notification events use
`kind: "toast"` plus `tone`, or `kind: "message-toast"` plus sender/message.
Remove the hard-coded no-op Reply action; render an action only when
`ActionLabel` is set.

- [ ] **Step 6: Implement CookieBanner and RatingDisplay**

`banner.Config` keeps normal announcement content, tone, position, CTA, and
dismiss behavior. Move consent-only fields into:

```go
type CookieBannerConfig struct {
    Title        string
    Description  string
    Icon         templ.Component
    AcceptLabel  string
    RejectLabel  string
    AcceptAction string
    RejectAction string
    RootClass    string
}
```

`CookieBanner` uses `role="dialog"` and preserves effective defaults
`"Cookie Consent"`, `"Accept"`, and `"Decline"`.

Rename rating `Style` to `Appearance` with `AppearanceStars` and
`AppearanceEmoji`. Rename `Class` to `RootClass` and `Attrs` to `RootAttrs`.
Interactive `Config` loses `ReadOnly`. Make `EmojiOption` and
`DefaultEmojiOptions` private because no public configuration accepts custom
emoji options. Add:

```go
type DisplayConfig struct {
    ID         string
    Value      int
    Max        int
    Label      string
    ShowLabel  bool
    Appearance Appearance
    Size       Size
    RootClass  string
    RootAttrs  templ.Attributes
}
```

`Rating` retains the radio-group form contract; `RatingDisplay` owns the
non-interactive `role="img"` rendering.

- [ ] **Step 7: Update all consumers and regenerate**

Run:

```bash
templ generate
go run ./cmd/skillgen
go test ./components/carousel ./components/modal ./components/toast \
  ./components/banner ./components/rating -count=1
(cd site && go test $(go list ./... | grep -v /tests/e2e) -count=1)
```

Expected: PASS.

- [ ] **Step 8: Commit**

```bash
git add components/carousel components/modal components/toast \
  components/banner components/rating \
  site/internal/pages/demo site/internal/server .agents .claude
git commit -m "refactor: split structural component primitives"
```

---

### Task 5: Adopt functional options for four atomic primitives

**Files:**

- Modify: `components/button/types.go`, `components/button/button.templ`,
  button tests and all `button.Button` callers
- Modify: `components/link/types.go`, `components/link/link.templ`, link tests
  and callers
- Modify: `components/kbd/types.go`, `components/kbd/kbd.templ`, kbd tests and
  callers
- Modify: `components/tooltip/types.go`,
  `components/tooltip/tooltip.templ`, tooltip tests and callers
- Modify: `internal/skillgen/main.go`
- Modify: `internal/skillgen/main_test.go`

**Interfaces:**

- `button.Button(options ...Option)`
- `link.Link(href string, options ...Option)`
- `kbd.Kbd(text string, options ...Option)`
- `tooltip.Tooltip(id string, label string, options ...Option)`

- [ ] **Step 1: Write option/default tests before changing constructors**

```go
func TestButtonOptionsApplyOverDefaults(t *testing.T) {
    html := renderComponent(t, Button(
        WithTone(ToneDanger),
        WithSize(SizeSmall),
        WithID("delete"),
    ), "Delete")
    require.Contains(t, html, `id="delete"`)
    require.Contains(t, html, "bg-danger")
}

func TestLinkRequiresHrefAndDefaultsToTextAppearance(t *testing.T) {
    html := renderWithChildren(t, Link("/docs"), "Docs")
    require.Contains(t, html, `href="/docs"`)
    require.Contains(t, html, "underline-offset-2")
}

func TestKbdRequiredText(t *testing.T) {
    html := render(t, Kbd("⌘K", WithSize(SizeSM)))
    require.Contains(t, html, "⌘K")
}

func TestTooltipRequiresIDAndLabel(t *testing.T) {
    html := render(t, Tooltip(
        "copy-url-tooltip",
        "Copies the URL",
        WithPosition(PositionBottom),
    ))
    require.Contains(t, html, "Copies the URL")
    require.Contains(t, html, `id="copy-url-tooltip"`)
}
```

- [ ] **Step 2: Verify constructor-signature failures**

```bash
go test ./components/button ./components/link ./components/kbd \
  ./components/tooltip -count=1
```

Expected: compile failures for `With...` functions.

- [ ] **Step 3: Implement exact option surfaces**

Use a private `config`, a sealed exported `Option` interface, and package-local
`optionFunc`:

```go
type Option interface {
    apply(*config)
}

type optionFunc func(*config)

func (fn optionFunc) apply(config *config) {
    fn(config)
}
```

Public functions:

```text
button: WithTone, WithSize, WithType, Disabled, WithID, WithRootClass,
        WithHTMX, WithAlpine, WithLoadingText
link:   WithTarget, WithRel, WithRole, WithID, WithAppearance, WithSize,
        WithIcon, WithIconPosition, WithRootClass, WithAttrs
kbd:    WithLabel, WithSize, WithIcon, WithRootClass, WithAttrs
tooltip: WithDescription, WithPosition, WithActivation, WithTriggerLabel,
         WithTrigger
```

For Button, replace `Variant` with `Tone` and type-prefixed `TonePrimary`,
`ToneSecondary`, `ToneAlternate`, `ToneInverse`, `ToneInfo`, `ToneDanger`,
`ToneWarning`, and `ToneSuccess`. Defaults are `TonePrimary`, `SizeMedium`, and
HTML type `"button"`; submit behavior is explicit through
`WithType("submit")`.

For Link, rename `Style` to `Appearance` with `AppearanceText` and
`AppearanceButton`; defaults are text appearance, medium size, and trailing
icon placement. For Kbd, default to `SizeMD`.

For Tooltip, make both ID and label required constructor arguments. Rename the
activation type from `Trigger` to `Activation` with `ActivationHover` and
`ActivationClick`; prefix placement constants as `PositionTop`,
`PositionBottom`, `PositionLeft`, and `PositionRight`. `Trigger` remains the
custom trigger component concept only through `WithTrigger`. Defaults are top
position, hover activation, and trigger label `"Hover Me"`.

- [ ] **Step 4: Teach skillgen to publish functional options**

Add a failing generator fixture containing `type Option interface` and
`func WithTone(tone Tone) Option`. Extend `pkgAPI` with option signatures,
collect exported top-level functions whose single result is the local
`Option`, and render them under an `**Options:**` line. Assert the generated
reference contains `WithTone(tone Tone)`.

Run:

```bash
go test ./internal/skillgen -count=1
```

Expected: PASS.

- [ ] **Step 5: Update all callers and generated usage examples**

Use the exact caller list from:

```bash
rg -l '@(button\.Button|link\.Link|kbd\.Kbd|tooltip\.Tooltip)\(' \
  components site --glob '*.templ'
rg -l '(button\.Button|link\.Link|kbd\.Kbd|tooltip\.Tooltip)\(' \
  components site --glob '*.go' --glob '!**/*_templ.go'
```

Every changed call must show required arguments directly and options only for
non-default behavior.

- [ ] **Step 6: Regenerate and test**

```bash
templ generate
go run ./cmd/skillgen
go test ./components/button ./components/link ./components/kbd \
  ./components/tooltip -count=1
(cd site && go test $(go list ./... | grep -v /tests/e2e) -count=1)
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add components/button components/link components/kbd components/tooltip \
  internal/skillgen site .agents .claude
git commit -m "refactor: add focused functional option APIs"
```

---

### Task 6: Give Display primitives concrete component instances

**Files:**

- Create: `components/{accordion,avatar,badge,banner,card,carousel,chatbubble,codeblock,head,kbd,table}/component.go`
- Modify: public entry points in the corresponding `.templ` files
- Create: `components/display_identity_test.go`
- Modify: `internal/skillgen/main.go`
- Modify: `internal/skillgen/main_test.go`
- Regenerate: `.agents/skills/using-goshtoso/references/components-reference.md`
- Regenerate: `.claude/skills/using-goshtoso/components-reference.md`

**Interfaces:**

- Consumes: `components.Component` and Display `Kind` constants.
- Produces: concrete `Instance` types plus specific secondary instance types.

- [ ] **Step 1: Teach skillgen about concrete component return types**

Add a failing `internal/skillgen` fixture with a local `Instance` type exposing
both `Kind()` and `Render(context.Context, io.Writer) error`, plus
`func Alert(config Config) Instance`. Assert the generated reference contains
`Alert(config Config)`.

Change the AST pass to collect local receiver type names that expose exported
`Kind` and `Render` methods. `isEntry` accepts either the current
`templ.Component` result or a single named local result in that receiver set.
This preserves generator coverage during the constructor migration without
adding a type-loading dependency or relying on an `Instance` suffix.

Run:

```bash
go test ./internal/skillgen -count=1
```

Expected: PASS for both legacy `templ.Component` and concrete component
constructors.

- [ ] **Step 2: Write a failing external-package identity test**

The test uses `package components_test` and constructs every Display entry from
the locked inventory:

```go
func displayRenderables() map[components.Kind]components.Component {
    return map[components.Kind]components.Component{
        components.KindAccordion: accordion.Accordion(accordion.AccordionConfig{}),
        components.KindAvatar: avatar.Avatar(avatar.Config{}),
        components.KindAvatarStack: avatar.AvatarStack(avatar.StackConfig{}),
    }
}

func TestDisplayRenderablesExposeKinds(t *testing.T) {
    values := displayRenderables()
    require.Len(t, values, 24)
    for want, value := range values {
        require.Equal(t, want, value.Kind())
    }
}
```

Fill `displayRenderables` with the 24 locked entries under accordion, avatar,
badge, banner, card, carousel, chatbubble, codeblock, head, kbd, and table. The
length assertion makes an omitted constructor fail visibly.

- [ ] **Step 3: Verify the test fails**

Run: `go test ./components -run TestDisplayRenderablesExposeKinds -count=1`

Expected: constructors return `templ.Component`, which lacks `Kind`.

- [ ] **Step 4: Add concrete wrappers**

For each primary entry point, rename the generated templ function to a private
`...Template` and delegate through a concrete value:

```go
type Instance struct{ cfg Config }

func Card(cfg Config) Instance { return Instance{cfg: cfg} }
func (Instance) Kind() components.Kind { return components.KindCard }
func (i Instance) Render(ctx context.Context, w io.Writer) error {
    return cardTemplate(i.cfg).Render(ctx, w)
}

var _ components.Component = Instance{}
```

Use distinct exported instance types for secondary primitives, such as
`StackInstance`, `NotificationBadgeInstance`, `CardCarouselInstance`, and
`TableRowsInstance`. Preserve templ child propagation by delegating with the
received render context.

- [ ] **Step 5: Regenerate and run Display tests**

```bash
templ generate
go run ./cmd/skillgen
go test ./components -run TestDisplayRenderablesExposeKinds -count=1
go test ./components/{accordion,avatar,badge,banner,card,carousel,chatbubble,codeblock,head,kbd,table} -count=1
(cd site && go test $(go list ./... | grep -v /tests/e2e) -count=1)
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add components internal/skillgen .agents .claude
git commit -m "feat: identify display component primitives"
```

---

### Task 7: Give Input primitives concrete component instances

**Files:**

- Create: `components/{button,checkbox,combobox,fileinput,form,radio,range,rating,palette,search,select,schemaform,structuredinput,tagslist,textinput,textarea,toggle}/component.go`
- Modify: public entry points in corresponding `.templ` files
- Create: `components/input_identity_test.go`

**Interfaces:**

- Produces all locked Input renderables as `components.Component`.

- [ ] **Step 1: Write the complete failing Input identity table**

Create `inputRenderables() map[components.Kind]components.Component` containing
the 30 locked entries under the 17 Input packages, including all seven form
primitives, three radio primitives, three search primitives, and
`TextareaWithActions`. Assert `require.Len(t, inputRenderables(), 30)` before
checking each map key against the returned Kind.

```go
require.Equal(t, components.KindFormFieldGroup,
    form.FieldGroup(form.FieldGroupConfig{}).Kind())
require.Equal(t, components.KindSearchModal,
    search.SearchModal(search.Config{}).Kind())
```

- [ ] **Step 2: Verify failure**

Run: `go test ./components -run TestInputRenderablesExposeKinds -count=1`

Expected: missing `Kind` methods.

- [ ] **Step 3: Add concrete wrappers and privatize combobox fragments**

Apply the Task 6 delegation pattern. Rename the six combobox fragments listed
in the locked inventory to lowercase and update `handler.go` plus same-package
tests. Only `Combobox` receives a Kind.

- [ ] **Step 4: Regenerate and run Input tests**

```bash
templ generate
go run ./cmd/skillgen
go test ./components -run TestInputRenderablesExposeKinds -count=1
go test ./components/{button,checkbox,combobox,fileinput,form,radio,range,rating,palette,search,select,schemaform,structuredinput,tagslist,textinput,textarea,toggle} -count=1
(cd site && go test $(go list ./... | grep -v /tests/e2e) -count=1)
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add components .agents .claude
git commit -m "feat: identify input component primitives"
```

---

### Task 8: Give Feedback and Navigation primitives concrete instances

**Files:**

- Create: `components/{alert,modal,toast,drawer,spinner,steps,tooltip,breadcrumbs,dropdown,link,navbar,pagination,sidebar,tabs}/component.go`
- Modify: public entry points in corresponding `.templ` files
- Create: `components/feedback_navigation_identity_test.go`

**Interfaces:**

- Produces all locked Feedback and Navigation renderables as
  `components.Component`.

- [ ] **Step 1: Write failing identity tables for all 14 packages**

Create `feedbackRenderables()` with the 12 locked Feedback entries and
`navigationRenderables()` with the eight locked Navigation entries. Include
`AlertDialog`, all five toast primitives, `sidebar.Overlay`, and the remaining
primary primitives. Assert those exact map lengths before checking each map key
against the returned Kind.

- [ ] **Step 2: Verify failure**

Run:

```bash
go test ./components \
  -run 'Test(Feedback|Navigation)RenderablesExposeKinds' -count=1
```

Expected: missing `Kind` methods.

- [ ] **Step 3: Add wrappers and compile-time assertions**

Apply the same concrete delegation pattern. Do not return a single shared
opaque wrapper.

- [ ] **Step 4: Regenerate and test**

```bash
templ generate
go run ./cmd/skillgen
go test ./components \
  -run 'Test(Feedback|Navigation)RenderablesExposeKinds' -count=1
go test ./components/{alert,modal,toast,drawer,spinner,steps,tooltip,breadcrumbs,dropdown,link,navbar,pagination,sidebar,tabs} -count=1
(cd site && go test $(go list ./... | grep -v /tests/e2e) -count=1)
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add components .agents .claude
git commit -m "feat: identify feedback and navigation primitives"
```

---

### Task 9: Curate accidental public helpers and render fragments

**Files:**

- Modify: `components/*/types.go`
- Modify: same-package unit tests that call renamed helpers
- Modify: `components/table/table.templ`
- Modify: `components/card/card.templ`
- Create: `components/public_renderable_inventory_test.go`

**Interfaces:**

- Keeps only deliberate non-render helper APIs.
- Proves the 74 constructors cover `AllKinds` exactly once.

- [ ] **Step 1: Add the failing all-renderable inventory test**

Merge the four identity maps created in Tasks 6–8 and compare the resulting
Kind set with `components.AllKinds()`:

```go
func TestPublicRenderableInventoryMatchesAllKinds(t *testing.T) {
    inventories := []map[components.Kind]components.Component{
        displayRenderables(),
        inputRenderables(),
        feedbackRenderables(),
        navigationRenderables(),
    }
    got := make([]components.Kind, 0, 74)
    for _, inventory := range inventories {
        for want, value := range inventory {
            require.Equal(t, want, value.Kind())
            got = append(got, value.Kind())
        }
    }
    require.ElementsMatch(t, components.AllKinds(), got)
    require.Len(t, got, 74)
    require.Len(t, got, len(components.AllKinds()))
}
```

- [ ] **Step 2: Remove and replace duplicate renderables**

- Replace `card.StarRating` examples with `rating.RatingDisplay`.
- Replace `table.ActionButton` with `button.Button`.
- Replace `table.StatusBadge` with `badge.Badge`.
- Keep `table.ImageCell`.
- Rename `TableHead`, `TableBody`, and `TablePagination` private; retain the
  four documented server fragments.
- Remove combobox `Option.Meta`, `Img`, `Initials`, `Badge`, and `BadgeColor`;
  the renderer consumes none of them. Keep `Value`, `Label`, and `Disabled`.
- Make palette `DefaultHues` and `DefaultShades` private immutable defaults;
  callers already override them through `Config.Hues` and `Config.Shades`.

- [ ] **Step 3: Privatize render-only helpers**

After this task, exported non-render helpers are limited to:

```text
avatar.GetInitials
codeblock.Render
combobox.Config.Validate
combobox.Config.InitialState
combobox.Handler
pagination.Config.HasPrevious
pagination.Config.HasNext
pagination.Config.PreviousPage
pagination.Config.NextPage
pagination.Config.PageURL
pagination.Config.Pages
schemaform.FlattenAllowList
schemaform.Walk
schemaform.FallbackFromDefaults
schemaform.PruneDisabled
schemaform.HasOnlySimpleScalars
search.Item.SearchText
search.Item.NormalizedMethod
search.Item.SafeHref
table.Config.IsSortedBy
table.Config.NextSortDir
table.Config.SortURL
table.Config.PageURL
table.Config.NextPageURL
form.FormDef.Bind
form.FormDef.Dependents
form.Handle
form.IsFieldValidation
form.FormDef.PopulateValues
form.RenderFieldResponse
```

`Render`, `Kind`, and `http.Handler.ServeHTTP` methods are additionally
exported by their required interfaces. Rename all CSS class assemblers,
resolved-ID/default helpers, and render predicates outside this list to
lowercase.

- [ ] **Step 4: Regenerate and run the complete library suite**

```bash
templ generate
go run ./cmd/skillgen
go test ./... -count=1
(cd site && go test $(go list ./... | grep -v /tests/e2e) -count=1)
```

Expected: PASS.

- [ ] **Step 5: Audit the resulting exported helper surface**

Run:

```bash
for d in components/*; do
  rg -n '^func (\([^)]*\) )?[A-Z][A-Za-z0-9_]*\(' "$d" \
    --glob '*.go' --glob '!**/*_test.go' --glob '!**/*_templ.go' || true
done
```

Expected: only locked constructors, interface methods, and the allowlist above.

- [ ] **Step 6: Commit**

```bash
git add components site .agents .claude
git commit -m "refactor: curate the supported component API"
```

---

### Task 10: Publish the consumer component model

**Files:**

- Create: `docs/COMPONENT_MODEL.md`
- Modify: `docs/COMPONENT_API_NAMING.md`
- Modify: `docs/USAGE.md`
- Modify: `README.md`
- Modify: `site/internal/pages/demo/components/getting_started.templ`
- Create: `site/internal/pages/demo/components/component_model.templ`
- Modify: `site/internal/pages/demo/components/registry.go`
- Modify: `site/internal/pages/demo/layout.templ`
- Modify: `.agents/skills/using-goshtoso/SKILL.md`
- Modify: `internal/skillgen/main.go`
- Modify: `internal/skillgen/main_test.go`
- Regenerate: `.agents/skills/using-goshtoso/references/components-reference.md`
- Regenerate: `.claude/skills/using-goshtoso/components-reference.md`

**Interfaces:**

- Produces: `/docs/component-model`.
- Documents: Theme, Primitive, Kind, configuration dimension, and the
  one-primitive-or-two rule.

- [ ] **Step 1: Add a failing route/render test**

Create `site/internal/pages/demo/components/component_model_test.go`:

```go
func TestComponentModelDocumentsConsumerVocabulary(t *testing.T) {
    html := render(t, componentModelContent())
    for _, phrase := range []string{
        "Theme", "Primitive", "Kind", "Configuration dimension",
        "There is no universal Variant", "One primitive or two",
    } {
        require.Contains(t, html, phrase)
    }
    require.Contains(t, html, "component.Kind()")
}
```

- [ ] **Step 2: Verify failure**

Run:

```bash
cd site
go test ./internal/pages/demo/components -run TestComponentModel -count=1
```

Expected: undefined page renderer.

- [ ] **Step 3: Write source and site guides**

Use the approved design vocabulary verbatim. Include:

```go
for _, component := range pageComponents {
    switch component.Kind() {
    case components.KindAlertDialog:
        // Alert-dialog orchestration.
    case components.KindTable:
        // Table orchestration.
    }
}
```

Show a separate theme/kind/dimensions example and state that `Dimension` is
documentation vocabulary, not a public Go type.

- [ ] **Step 4: Link every consumer channel**

- Add Component Model to the site top-level docs navigation and search.
- Link it from Getting Started, README, `docs/USAGE.md`, and the maintainer
  naming guide.
- Link the hand-written `.agents/skills/using-goshtoso/SKILL.md` to the guide.
- Extend `internal/skillgen.render` so both generated references begin with the
  concise Theme/Primitive/Kind/Dimension rationale and a
  `docs/COMPONENT_MODEL.md` link. Extend `internal/skillgen/main_test.go` to
  assert that generated introduction.

- [ ] **Step 5: Regenerate and test**

```bash
templ generate
go run ./cmd/skillgen
cd site && go test ./internal/pages/demo/components -count=1
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add README.md docs site .agents .claude internal/skillgen
git commit -m "docs: explain the Goshtoso component model"
```

---

### Task 11: Introduce the site catalog and structured API metadata

**Files:**

- Create: `site/internal/pages/catalog/catalog.go`
- Create: `site/internal/pages/catalog/catalog_test.go`
- Create: `site/internal/pages/demo/api_docs.go`
- Modify: `site/internal/pages/demo/component_demo.templ`
- Modify: `site/internal/pages/demo/layout.templ`
- Modify: `site/internal/pages/demo/components/registry.go`

**Interfaces:**

- Produces: `catalog.Entry`, `catalog.ComponentPages()`, `catalog.Lookup()`.
- Produces: `demo.APISection`, `demo.APIPropDoc`, `demo.StructAPI[T]`,
  `demo.OptionsAPI`, `demo.FunctionsAPI`.
- Changes: `components.DemoEntry` gains `API []demo.APISection`; component
  entries use keyed literals and point at the same section variables rendered
  by their pages.

- [ ] **Step 1: Write failing catalog and API-structure tests**

```go
func TestComponentCatalogHasEveryPageOnce(t *testing.T) {
    pages := ComponentPages()
    require.Len(t, pages, 42)
    seen := map[string]bool{}
    for _, page := range pages {
        require.False(t, seen[page.Path])
        require.NotEmpty(t, page.Description)
        require.NotEmpty(t, page.Section)
        seen[page.Path] = true
    }
}
```

```go
func TestStructAPIRecordsReflectedType(t *testing.T) {
    section := StructAPI[sampleConfig](
        components.KindButton, "Config", "sample.New", "", nil,
    )
    require.Equal(t, reflect.TypeFor[sampleConfig](), section.StructType)
}
```

- [ ] **Step 2: Verify failures**

```bash
cd site
go test ./internal/pages/catalog ./internal/pages/demo -count=1
```

Expected: packages/types do not exist.

- [ ] **Step 3: Implement the catalog**

```go
type Entry struct {
    Key         string
    Path        string
    Title       string
    Active      string
    Description string
    Section     string
    Order       int
    Kinds       []components.Kind
}
```

Populate exactly 42 component entries in current sidebar order. Map all 74
Kinds exactly once. Return cloned slices. Derive component sidebar sections,
previous/next navigation, search entries, metadata descriptions, and component
count from this catalog. Remove the component copies from
`getSidebarSections`, `orderedComponents`, `searchItemDescription`, and
`demoDescription`.

- [ ] **Step 4: Implement structured API metadata**

```go
type APIPropDoc struct {
    Name        string
    Signature   string
    Default     string
    Allowed     []string
    Description string
    Required    bool
}

type APISection struct {
    ID          string
    Title       string
    Description string
    Constructor string
    Kind        components.Kind
    StructType  reflect.Type
    Props       []APIPropDoc
}
```

`StructAPI[T]` records `reflect.TypeFor[T]()` and requires `Signature == ""`
for struct fields. `OptionsAPI` and `FunctionsAPI` use explicit signatures and
no reflected struct; the former labels option functions while the latter
covers constructors and deliberate non-render helpers. The renderer computes
exact struct field types from reflection and renders `Allowed` values for typed
dimensions.

- [ ] **Step 5: Add stable documentation hooks**

Add:

```text
data-component-description
data-demo-section
data-demo-preview
data-demo-code
data-api-reference
data-api-section="<section-id>"
```

Keep the current flat `PropDoc` and `APIReference` untouched while pages are
migrated. Add `StructuredAPIReference([]APISection)`, which renders multiple
named sections with a Required column or badge. Add the stable preview/code
hooks to `ComponentDemo`, and add `data-api-reference` to both the flat and
structured API renderers, so legacy and migrated pages satisfy the same
smoke-test selectors.

- [ ] **Step 6: Regenerate and test**

```bash
templ generate
cd site
go test ./internal/pages/catalog ./internal/pages/demo \
  ./internal/pages/demo/components -count=1
```

Expected: PASS. Existing pages continue using the flat renderer; Tasks 12–15
move them explicitly to `StructuredAPIReference`.

- [ ] **Step 7: Commit**

```bash
git add site/internal/pages/catalog site/internal/pages/demo
git commit -m "feat: structure the component documentation catalog"
```

---

### Task 12: Migrate Display API references and correct their claims

**Files:**

- Modify:
  `site/internal/pages/demo/components/{accordion,avatar,badge,banner,card,carousel,chatbubble,codeblock,dependencies,kbd,table}.templ`

**Interfaces:**

- Produces named `[]demo.APISection` values used by both the page and registry.

- [ ] **Step 1: Add failing registry expectations for Display pages**

In `site/internal/pages/demo/components/api_contract_test.go`, assert each
Display entry has non-empty `API` metadata and the expected Kind list.

- [ ] **Step 2: Verify failure**

Run:

```bash
cd site
go test ./internal/pages/demo/components \
  -run TestDisplayAPIMetadataRegistered -count=1
```

Expected: Display entries have no structured API.

- [ ] **Step 3: Migrate every Display public type**

Document:

```text
accordion: AccordionConfig, AccordionItem
avatar: Config, StackConfig
badge: Config and the three notification-dot/count constructors
banner: Config, CTAConfig, CookieBannerConfig, Banner, CookieBanner
card: Config
carousel: Config, CardConfig, Slide, AutoplayConfig, HTMXConfig
chatbubble: Config and TypingIndicator
codeblock: Config
dependencies: Dependencies and DependenciesMinimal
kbd: constructor plus every functional option
table: Config, Column, Cell, Row, RowHTMXConfig, PaginationConfig,
       InfiniteScrollConfig, FilterConfig, Filter, FilterOption,
       FilterOptionsHTMXConfig, FilterHTMXConfig, HTMXConfig,
       TableHeadContent, TableRows, TableRow, TablePaginationNav, ImageCell
```

Remove the table checkbox-variant claim. Replace table action/status examples
with Button and Badge. Document all effective defaults, not raw zero values.

- [ ] **Step 4: Regenerate and test**

```bash
templ generate
cd site && go test ./internal/pages/demo/components -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add site/internal/pages/demo/components
git commit -m "docs: complete display component API references"
```

---

### Task 13: Migrate atomic Input API references

**Files:**

- Modify:
  `site/internal/pages/demo/components/{button,checkbox,fileinput,palette,radio,range,rating,select,tagslist,textarea,textinput,toggle}.templ`

- [ ] **Step 1: Add failing API-registration expectations**

Assert all 12 pages have structured API sections and their catalog Kinds.

- [ ] **Step 2: Verify failure**

Run:

```bash
cd site
go test ./internal/pages/demo/components \
  -run TestAtomicInputAPIMetadataRegistered -count=1
```

- [ ] **Step 3: Document exact public surfaces**

```text
button: constructor plus every functional option, HTMXConfig, AlpineConfig
checkbox: Config, GroupConfig
fileinput: Config
palette: Config, AlpineConfig
radio: Config, GroupConfig, HTMXConfig, AlpineConfig, RadioBar
range: Config, Tick
rating: Config, DisplayConfig, Rating, RatingDisplay
select: Config, Option, AlpineConfig
tagslist: Config
textarea: Config, TextareaWithActions
textinput: Config
toggle: Config
```

Correct the known defaults:

- textarea zero `Rows` renders `3`;
- select zero `Placeholder` renders `"Please Select"`;
- tags-list zero `Placeholder` renders `"Add a tag..."`;
- include `textarea.Config.InputAttrs`;
- include `toggle.Config.Value` and `toggle.Config.InputAttrs`;
- remove claims that textarea has counters or tags-list rejects duplicates.

- [ ] **Step 4: Regenerate, test, and commit**

```bash
templ generate
cd site && go test ./internal/pages/demo/components -count=1
git add site/internal/pages/demo/components
git commit -m "docs: complete atomic input API references"
```

---

### Task 14: Migrate complex Input API references

**Files:**

- Modify:
  `site/internal/pages/demo/components/{combobox,form,schemaform,search,structuredinput}.templ`

- [ ] **Step 1: Add failing complex API expectations**

Assert structured sections exist for every public type below.

- [ ] **Step 2: Document exact public surfaces**

```text
combobox: Config, Option, Source, State, OptionsProvider, Handler
form: Config, HTMXConfig, FooterConfig, CancelHTMXConfig, SectionConfig,
      CollapsibleSectionConfig, FlipSectionConfig, SubSectionConfig,
      FieldGroupConfig, FieldMeta, ValidationConfig, FormErrorsConfig,
      FormErrorItem, plus all seven form constructors
schemaform: FieldsConfig, Field, AllowMode and public transform functions
search: Config, Item, Search, SearchField, SearchModal
structuredinput: Config, Column, Option
```

Correct structured-input copy to describe repeatable rows and typed columns,
not prefixes/suffixes/segmented values.

- [ ] **Step 3: Regenerate, test, and commit**

```bash
templ generate
cd site && go test ./internal/pages/demo/components -count=1
git add site/internal/pages/demo/components
git commit -m "docs: complete complex input API references"
```

---

### Task 15: Migrate Feedback and Navigation API references

**Files:**

- Modify:
  `site/internal/pages/demo/components/{alert,drawer,modal,spinner,steps,toast,tooltip,breadcrumbs,dropdown,link,navbar,pagination,sidebar,tabs}.templ`
- Modify: `site/internal/pages/demo/components/registry.go`

- [ ] **Step 1: Add failing API-registration expectations**

Assert all 14 pages have structured sections and exact catalog Kind mappings.

- [ ] **Step 2: Document exact public surfaces**

```text
alert: Config, LinkConfig, HTMXConfig, ActionConfig
drawer: Config
modal: Config, AlertDialogConfig, ButtonAction, HTMXConfig
spinner: Config
steps: Config, Step
toast: Config, MessageConfig, Sender, HTMXConfig, ContainerConfig and all five constructors
tooltip: constructor plus every functional option
breadcrumbs: Config, Item
dropdown: Config, Item, Section
link: constructor plus every functional option
navbar: Config, NavLink, UserProfile, UserMenuItem, ActionItem
pagination: Config, HTMXConfig, PageItem
sidebar: Config, Item, Section, OverlayConfig
tabs: Config, Tab, TabHTMX
```

Correct effective defaults and stale claims:

- sidebar zero `LogoHref` renders `/`;
- tooltip ID and label are required, and the default trigger label is explicit;
- form nil `PreventEnterSubmit` behavior is already corrected in Task 14;
- spinner does not claim to render a label/busy-state wrapper;
- toast does not claim configurable positions;
- file-input, range, textarea, tags-list, and structured-input catalog
  descriptions match actual behavior.

- [ ] **Step 3: Remove the flat API renderer**

Every component page now supplies structured sections. Delete the old flat
`PropDoc` and `APIReference`; keep `APIPropDoc` and
`StructuredAPIReference`. Make a missing structured API section a contract-test
failure.

- [ ] **Step 4: Regenerate, test, and commit**

```bash
templ generate
cd site && go test ./internal/pages/demo/components -count=1
git add site/internal/pages/demo
git commit -m "docs: complete feedback and navigation API references"
```

---

### Task 16: Enforce documentation and catalog contracts

**Files:**

- Complete: `site/internal/pages/demo/components/api_contract_test.go`
- Modify: `site/internal/pages/catalog/catalog_test.go`
- Modify: `site/tests/e2e/sidebar_test.go`

**Interfaces:**

- Proves: all public struct fields, Kinds, routes, navigation, and docs agree.

- [ ] **Step 1: Add exact reflection coverage checks**

```go
func assertStructSectionComplete(t *testing.T, section demo.APISection) {
    typ := section.StructType
    documented := map[string]demo.APIPropDoc{}
    for _, prop := range section.Props {
        require.NotContains(t, documented, prop.Name)
        documented[prop.Name] = prop
        require.NotEmpty(t, prop.Description)
        require.NotEmpty(t, prop.Default)
    }
    for field := range typ.Fields() {
        if field.IsExported() {
            require.Contains(t, documented, field.Name)
            if isNamedScalar(field.Type) &&
                strings.HasPrefix(
                    field.Type.PkgPath(),
                    "github.com/araihu/goshtoso/components/",
                ) {
                require.NotEmpty(t, documented[field.Name].Allowed)
            }
        }
    }
    require.Len(t, documented, exportedFieldCount(typ))
}
```

`isNamedScalar` returns true for named string, signed-integer, or
unsigned-integer types. Also assert option/function signatures, constructors,
section IDs, descriptions, and every listed allowed value are non-empty.

Add `assertAllExportedStructsRegistered`: starting from the test package's
working directory, walk `../../../../../components`, parse non-test,
non-generated `.go` files with `go/parser`, and collect every exported struct
that has an exported named or embedded field. Compare its
`<import-path>.<type>` key exactly with the set of non-nil
`APISection.StructType` values. Concrete render instances have only private
fields and are therefore covered by the Kind/constructor inventory instead.
This makes a newly exported nested config/data struct fail until it is added to
the appropriate page.

- [ ] **Step 2: Add 1:1 Kind/page/navigation checks**

Flatten catalog Kinds and compare exactly with `components.AllKinds()`. Compare
catalog component paths with `Demos` component keys. Derive sidebar tests from
`catalog.ComponentPages()` and delete the stale hard-coded 37-entry slice.

- [ ] **Step 3: Run contract tests**

```bash
cd site
go test ./internal/pages/catalog ./internal/pages/demo/components -count=1
go test ./tests/e2e -run TestSidebar_AllComponentsPresent -count=1 -timeout 5m
```

Expected: PASS for 42 pages and 74 Kinds.

- [ ] **Step 4: Commit**

```bash
git add site
git commit -m "test: enforce component documentation contracts"
```

---

### Task 17: Add registry-driven direct-load smoke tests

**Files:**

- Create: `site/tests/e2e/component_docs_smoke_test.go`
- Modify: `site/tests/e2e/e2e_test.go`

- [ ] **Step 1: Add a reusable browser-failure collector**

Capture `pageerror`, error-level console events, failed requests, and HTTP
responses with status `>= 400`. Protect callback slices with a mutex. Reuse the
existing `filterIgnorable` policy only for known browser noise; component errors
must not be filtered. Use this exact test-facing shape:

```go
type pageFailures struct {
    mu       sync.Mutex
    messages []string
}

func watchPageFailures(page playwright.Page) *pageFailures
func (failures *pageFailures) RequireEmpty(t *testing.T)
```

`RequireEmpty` copies the slice while holding the mutex, applies
`filterIgnorable`, and calls `require.Empty`.

- [ ] **Step 2: Write the direct-load test**

```go
func TestAllComponentDocsDirectLoad(t *testing.T) {
    for _, entry := range catalog.ComponentPages() {
        entry := entry
        t.Run(entry.Active, func(t *testing.T) {
            page := newPage(t, sharedBrowser)
            failures := watchPageFailures(page)
            response, err := page.Goto(baseURL+entry.Path,
                playwright.PageGotoOptions{
                    WaitUntil: playwright.WaitUntilStateDomcontentloaded,
            })
            require.NoError(t, err)
            require.NotNil(t, response)
            require.Equal(t, http.StatusOK, response.Status())

            heading, err := page.Locator("main h1").TextContent()
            require.NoError(t, err)
            require.Equal(t, entry.Title, strings.TrimSpace(heading))

            descriptionCount, err := page.
                Locator("[data-component-description]").Count()
            require.NoError(t, err)
            require.Equal(t, 1, descriptionCount)

            previewCount, err := page.Locator("[data-demo-preview]").Count()
            require.NoError(t, err)
            require.GreaterOrEqual(t, previewCount, 1)

            codeCount, err := page.Locator("[data-demo-code]").Count()
            require.NoError(t, err)
            require.GreaterOrEqual(t, codeCount, 1)

            apiCount, err := page.Locator("[data-api-reference]").Count()
            require.NoError(t, err)
            require.Equal(t, 1, apiCount)
            failures.RequireEmpty(t)
        })
    }
}
```

- [ ] **Step 3: Run the test and fix claim/runtime failures**

Run:

```bash
go test ./site/tests/e2e/... -run TestAllComponentDocsDirectLoad \
  -count=1 -timeout 10m
```

Expected: PASS for all 42 subtests. Any failure must be fixed in its source
component/page and receive a focused test when it exposes a behavioral claim.

- [ ] **Step 4: Commit**

```bash
git add components site
git commit -m "test: smoke every component documentation page"
```

---

### Task 18: Add HTMX-navigation and representative theme smoke tests

**Files:**

- Modify: `site/tests/e2e/component_docs_smoke_test.go`

- [ ] **Step 1: Write catalog-wide HTMX navigation smoke**

Start at the first component page, then click each sidebar link in catalog
order. Re-query the link after every swap. For every destination assert:

```text
URL equals catalog path
main h1 equals catalog title
[data-api-reference] exists
[data-demo-preview] exists
matching sidebar link has aria-current="page"
window.Alpine is defined
window.htmx is defined
no accumulated browser failures
```

- [ ] **Step 2: Write the representative theme matrix**

Use pages:

```text
/components/button
/components/form
/components/modal
/components/table
/components/toast
```

For each, run:

```text
goshtoso light
goshtoso dark
minimal light
minimal dark
```

Set storage consent and `localStorage.theme`/`darkMode` in
`page.AddInitScript` before navigation:

```go
script := fmt.Sprintf(`
    document.cookie = "gt_storage=allowed; Path=/; SameSite=Lax";
    localStorage.setItem("theme", %q);
    localStorage.setItem("darkMode", %q);
`, themeName, strconv.FormatBool(dark))
require.NoError(t, page.AddInitScript(playwright.Script{
    Content: &script,
}))
```

After navigation, read `data-theme` and the `<html>` class with
`GetAttribute`, then assert they equal the matrix row. For the first
`[data-demo-preview]`, require `IsVisible() == true`, a non-nil bounding box,
and width and height greater than zero. Finally evaluate:

```js
() => getComputedStyle(document.documentElement)
    .getPropertyValue("--color-surface").trim()
```

and require a non-empty token plus no browser failures.

- [ ] **Step 3: Run navigation and theme smoke**

```bash
go test ./site/tests/e2e/... \
  -run 'TestAllComponentDocsFragmentNavigation|TestComponentDocsThemeMatrix' \
  -count=1 -timeout 15m
```

Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add site/tests/e2e
git commit -m "test: smoke component docs navigation and themes"
```

---

### Task 19: Regenerate artifacts and run the completion audit

**Files:**

- Regenerate: all affected `*_templ.go`
- Regenerate: `.claude/skills/using-goshtoso/components-reference.md`
- Regenerate:
  `.agents/skills/using-goshtoso/references/components-reference.md`
- Regenerate: `assets/styles.css` only if `just css` reports a source-driven
  change
- Modify: source files only when a gate exposes a real issue

- [ ] **Step 1: Regenerate from source**

```bash
templ generate
just css
go run ./cmd/skillgen
find components site/internal site/tests -type f -name '*.go' \
  ! -name '*_templ.go' -print0 | xargs -0 gofmt -w
```

Do not keep unrelated formatter churn. Confirm generated changes correspond to
edited sources.

- [ ] **Step 2: Prove API and documentation invariants**

```bash
rg -n '\bVariant\b|Variant[A-Z]|[A-Z][A-Za-z]+Variant\b' \
  components site/internal/pages/demo/components \
  --glob '*.go' --glob '*.templ' --glob '!**/*_templ.go'
```

Expected: no matches.

```bash
rg -n '\bVariant\b' README.md docs/COMPONENT_MODEL.md \
  docs/COMPONENT_API_NAMING.md docs/USAGE.md \
  .agents/skills/using-goshtoso .claude/skills/using-goshtoso
```

Expected: only prose explaining why there is no universal Variant.

Run the public renderable and docs contracts:

```bash
go test ./components -run 'Test(AllKinds|PublicRenderableInventory)' -count=1
cd site
go test ./internal/pages/catalog ./internal/pages/demo/components -count=1
```

Expected: 74 unique Kinds, 42 component pages, full reflected-field coverage.

- [ ] **Step 3: Run formatting, lint, fix, and build gates**

```bash
git diff --check
golangci-lint run
(cd site && golangci-lint run)
go fix ./...
(cd site && go fix ./...)
templ generate
go run ./cmd/skillgen
git diff --check
go build -o bin/server ./site/cmd/server
```

Inspect any `go fix` changes, regenerate again, and rerun `git diff --check`.

- [ ] **Step 4: Run all unit and site tests**

```bash
go test ./... -count=1
(cd site && go test $(go list ./... | grep -v /tests/e2e) -count=1)
```

Expected: PASS.

- [ ] **Step 5: Run the complete E2E suite**

```bash
go test ./site/tests/e2e/... -count=1 -timeout 15m
```

Expected: PASS, including all direct-load, fragment-navigation, and theme
smoke subtests.

- [ ] **Step 6: Perform the requirement-by-requirement completion audit**

Record command output proving:

```text
74/74 Kinds represented by public constructors
74/74 Kinds mapped to documentation
42/42 component routes in catalog, registry, sidebar, and direct smoke
42/42 pages with preview, code, and structured API reference
all reflected public config fields documented exactly once
no stale removed symbol in docs or snippets
all focused and full test gates green
```

- [ ] **Step 7: Commit final generated and verification fixes when present**

```bash
git status --short
git add README.md docs components internal/skillgen site .agents .claude \
  assets/styles.css
if ! git diff --cached --quiet; then
  git commit -m "chore: verify component API documentation contracts"
fi
```

---

### Task 20: Publish the release changelog and migration guide

**Files:**

- Create: `CHANGELOG.md`
- Create: `docs/MIGRATING_COMPONENT_API.md`
- Create: `docs/migration_examples_test.go`
- Modify: `README.md`

**Interfaces:**

- Consumes: the previous released Git tag and the final public API produced by
  Tasks 1–19.
- Produces: an `Unreleased` changelog entry explicitly based on the previous
  release, plus a compile-checked consumer migration guide.

- [ ] **Step 1: Lock the release comparison base**

Resolve and record the previous released version:

```bash
git describe --tags --abbrev=0
git log --oneline "$(git describe --tags --abbrev=0)"..HEAD
git diff --stat "$(git describe --tags --abbrev=0)"..HEAD
```

The changelog heading must say `Unreleased` until the release version and date
are known, and must state the exact previous tag used as the migration base.
Do not invent the next version number.

- [ ] **Step 2: Write failing compile checks for migration examples**

Add `docs/migration_examples_test.go` with external-package examples covering:

- the shared `components.Component` / `Kind()` identity contract;
- `Tone`, `Appearance`, and `Mode` configuration dimensions;
- one split primitive pair, such as `Modal` / `AlertDialog`;
- the Button, Link, Kbd, and Tooltip functional-option constructors;
- one concrete component return value stored and switched on by Kind.

Write the intended new API first and run:

```bash
go test ./docs -count=1
```

Expected before the guide/API names are corrected: compile failure or a failing
example assertion. Keep the final snippets in the guide aligned with these
compile-checked examples.

- [ ] **Step 3: Write the release changelog**

Create `CHANGELOG.md` with an `Unreleased` entry that:

- links to `docs/MIGRATING_COMPONENT_API.md`;
- calls out the release as breaking while the project is alpha;
- summarizes component identity, dimension vocabulary, split primitives,
  functional options, concrete renderable instances, curated helpers, and the
  documentation/smoke-test contract;
- distinguishes source-breaking changes from behavioral/default changes;
- lists removals without implying compatibility aliases exist.

- [ ] **Step 4: Write the consumer migration guide**

Build the exact old-to-new mapping from the final diff and generated consumer
reference, not from memory:

```bash
previous_tag="$(git describe --tags --abbrev=0)"
git diff "$previous_tag"..HEAD -- components internal/skillgen \
  .agents/skills/using-goshtoso/references/components-reference.md
```

The guide must include:

- a searchable old symbol → new symbol/action table;
- before/after Go snippets for each breaking-change family;
- split-primitive selection guidance;
- effective-default notes where zero values or constructor defaults matter;
- a mechanical upgrade checklist with `rg` searches for removed names;
- the shared `Kind()` loop-and-switch example and the rationale for concrete
  dimension names instead of a universal `Variant`.

- [ ] **Step 5: Link and verify the release documentation**

Link the changelog and migration guide from `README.md`, then run:

```bash
go test ./docs -count=1
rg -n 'Unreleased|MIGRATING_COMPONENT_API|Breaking' CHANGELOG.md README.md
rg -n '\bVariant\b|Variant[A-Z]|[A-Z][A-Za-z]+Variant\b' \
  CHANGELOG.md docs/MIGRATING_COMPONENT_API.md
git diff --check
```

Expected: tests pass; `Variant` matches only historical old-symbol mappings or
the rationale explaining why it is no longer universal.

- [ ] **Step 6: Perform a release-note completeness audit**

Compare every breaking public API commit and the final generated component
reference against the changelog. Record evidence that all of these are covered:

```text
identity and Kind
dimension renames
split primitives
functional-option constructors
concrete renderable returns
removed/private helpers
effective default or behavior changes
documentation and smoke-test guarantees
```

- [ ] **Step 7: Commit**

```bash
git add CHANGELOG.md README.md docs/MIGRATING_COMPONENT_API.md \
  docs/migration_examples_test.go
git commit -m "docs: add component API migration changelog"
```
