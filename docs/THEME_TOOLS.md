# Public theme tools

`themes.Definition` is a complete version 1 snapshot. `themes.Preset(key)`
materializes every registered token, typography reference and radius from the
installed catalog. `themes.Materialize` applies typed overrides and recomputes
derived roles. A saved definition does not inherit future preset changes.

Use `themes.Decode` at a JSON boundary and `themes.Validate` before persistence.
Both reject unsupported schemas, incomplete tokens and invalid values.
`themes.RenderCSS(definition, resolver)` produces literal CSS scoped to
`:root[data-theme="custom"]`. The application sets that attribute and its
`dark` class. `themes.CatalogRevision()` identifies the materializer and source
palette. `themes.Version()` reports the selected Go module version when present.

## Editor and isolated preview

Compose public packages `components/themeeditor` and `components/themeplayground`.
The site example at `site/internal/pages/demo/contentpages/start/landing.templ`
uses the same public API as external consumers.

```go
initial, err := themes.Preset(themes.KeyGoshtoso)
if err != nil { return err }
editor := themeeditor.Config{
    ID: "appearance-editor", Initial: initial,
    Fonts: themes.DefaultFonts(), PreviewID: "appearance-preview",
}
preview := themeplayground.Config{
    ID: "appearance-preview", Src: "/appearance/preview", Title: "Appearance preview",
}
```

Render `themeeditor.ThemeEditor(editor)` and
`themeplayground.ThemePlayground(preview)` in the application page. Serve the
preview route with `themeplayground.Frame(FrameConfig{ID: "appearance-preview",
Content: examples, Fonts: fonts})`. `Content` and optional `Head` are trusted,
application-owned templ components. Host and frame must share an origin. Mount
`assets.Handler()` and load `head.Dependencies()` in the host document; the frame
loads its own Goshtoso dependencies.

The editor owns no storage, HTTP mutation, application color mode or forms.
Its hidden input defaults to `name="definition"`; `InputName` changes that name.
The caller owns its save form and can add buttons through `Actions`. Preview
receives only typed definitions. Messages check origin, source window, schema,
instance and increasing sequence. The iframe has a separate document; preview
colors and typography never restyle the host. Use representative static examples
without live operational actions in `Content`.

Events bubble from the editor root. `themeeditor:ready` and `themeeditor:change`
carry `id`, `definition`, `previewMode`, `dirty` and `valid`. Read the complete
snapshot from `detail.definition`. Failed edits retain the last valid snapshot
and report `valid: false`; do not enable Save until validity returns.

Dispatch `themeeditor:load` on the root with `{id, definition, dirty}` to load a
new baseline. A `{id, previewMode}` event changes only light/dark inspection.
A `{id, fonts}` event refreshes only the font catalog, preserving
raw invalid text and the draft. Dispatch `themeeditor:committed` with `{id}` after
successful persistence. The caller confirms its own navigation or state changes;
the editor confirms its reset and preset replacement through Goshtoso Dialog.
Cancel/Escape preserves input. CSS export downloads the current complete draft.
Contrast is advisory. Surface, Primary and Secondary panels show foreground and
background Aa samples, their inverse, the live ratio and AAA/AA/AA large/Fail
badges using the site checker thresholds. Translucent pairs are Not rated because
their contrast depends on the backdrop. Preview-mode changes inspect that mode's
on-* and background tokens without changing the host theme.

Color roles compose `palette.Palette` inside a `Select` shell. Each trigger shows
the literal color and its swatch. Named Tailwind picks resolve immediately to
color values; saved definitions never retain palette-variable references. The
per-color Reset restores that role from the loaded baseline. The dropdown keeps
an Exact value input for hex, sRGB and OKLCH, including alpha. Palette grids mount
lazily when opened. Picks and Escape return focus to the trigger.

The inspection control is an icon button with an accessible next-action label
and `aria-pressed` state. It changes only the isolated preview mode.
`Config.PresetOptions()` supplies Select options with primary, secondary and
surface swatches and the title face of each definition. Options carry their font-preview metadata, so on-demand private font previews
also load on a read-only selection page without a mounted editor.
Named preset stacks use a matching curated Google preview face when available;
this does not change the stored typography reference.

## Font catalogs

`FontRef` stores only `source` and `id`. `FontOption` provides trusted display and
asset metadata. System, Google, packaged and upload sources share one picker.
Each font option and selected trigger render in that typeface. Google option
previews use private aliases and load on demand, avoiding registration of raw
families in the host document. URL assets use immutable opaque family aliases.
`themes.RenderFontCSS` provides the same validated declarations and family
expression to applications.

Use `themes.ResolveFont(catalog, ref)` as the server resolver. Applications own
upload validation, durable asset identities and serving bytes. Supply an optional
same-origin `FrameConfig.FontCatalogURL` when uploads must become available to an
already mounted preview. The response is a JSON array of `FontOption` objects.
Font messages do not supply URLs. Unknown or unavailable assets show feedback;
applications keep saved references and decide their recovery UI.

## Controlled Combobox updates

`combobox.Option.LabelAttrs` attaches public attributes to an option's label.
ThemeEditor uses `style="font-family:..."` for font previews. Static search with
no options endpoint filters locally.

For a client Combobox, dispatch `combobox:load` on its root with
`{id, values: [value], options: [{value, label, disabled, labelAttrs}]}`.
`values` and `options` are optional; present values must exist in the resulting
catalog. Replacement validates before changing state. It emits no user-change
event and writes no preferences. Dynamic label attributes allow `style`, `title`,
`data-*` and `aria-*`, never event handlers. Server Comboboxes continue to use
server-rendered bodies.

## Rich Select options

`selectfield.Option.Leading` accepts non-interactive templ content.
`Option.LabelAttrs` applies attributes to each option label. Its `style` follows
the selected trigger label. Leading content appears in both the option and the
selected trigger. The options remain a fixed catalog for that mount; replace the
Select root when the catalog changes. Rich options keep the component's native
change event, submitted value and arrow/Enter/Space keyboard behavior.

The editor's Preview mode fieldset uses labelled Light/Dark icon radio segments.
Selection changes the preview and visible color roles. It does not change the
host page, global mode preference, theme definition or draft dirty state.

## CSS import

ThemeEditor imports CSS automatically when a file is selected in its Goshtoso
FileInput. Import one complete
Goshtoso theme export, up to 1 MiB. Supported input contains literal hex/sRGB/OKLCH
colors for every role, rem radius, and title/body font stacks in :root or a single
[data-theme] block, optionally inside @layer. Same-file var aliases are resolved.
Multiple named themes, conditional overrides, missing roles, unresolved colors,
unavailable fonts and out-of-range values are rejected without replacing the draft.

CSS is parsed in a detached CSSStyleSheet and converted into a validated Definition.
It is never attached to the host or frame. Arbitrary style rules and font-face URLs
are ignored; @import rules are discarded by the browser without loading them.
Font references resolve against Config.Fonts, and exports include a comment that
preserves source identity when a system and Google font share the same family.
A font file must already be in the caller's catalog; import does not upload fonts.

Replacing editor changes uses the existing confirmation Dialog. Cancel preserves
all draft values. Successful import marks the draft dirty and emits the ordinary
themeeditor:change event; persistence remains the caller's responsibility. Reset
returns to the pre-import baseline. The caller's name, branding and global mode
are outside the imported Definition. Errors use FileInput's native aria-invalid
state, a persistent associated message, and notify with danger tone when the caller
has mounted ToastContainer.

The chooser clears after import, cancellation or error so the same file can be
selected again. Selecting no file does not import or show an error.
