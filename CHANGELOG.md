# Changelog

All notable changes to goshtoso are documented in this file.

## Unreleased

### Added
- `selectfield.ToOptions[T]` — generic helper to convert any slice to `[]Option` with value/label accessor functions
- `table.LinkMode` — `LinkSPA` (default), `LinkBoost`, `LinkFull` options for controlling row navigation strategy
- `docs/THEMING.md` — CSS token reference with all design tokens, available themes, and custom theme guide
- Godoc comments on all exported struct fields across all 22+ components
- "Known Pitfalls" section in `docs/USAGE.md` covering HTMX history cache + Alpine.js state and IntersectionObserver in nested scroll containers

### Changed
- `textinput.ContainerClasses()` — removed `max-w-xs` default; width is now determined by parent layout
- `textarea.ContainerClasses()` — removed `max-w-md` default; width is now determined by parent layout
- `selectfield.ContainerClasses()` — removed conditional `max-w-xs` when Label is set; width is now determined by parent layout

### Migration
- If you relied on the default `max-w-xs` constraint on text inputs or selects, add `Class: "max-w-xs"` to your Config to restore the previous behavior
- Table `Row.Link` now defaults to `LinkSPA` mode (swaps `#main-content-area`). Previous behavior was `LinkBoost` (full body swap). Set `LinkMode: table.LinkBoost` explicitly if you need full-body navigation.

## v0.0.1-alpha

Initial release with 22 components:

Accordion, Alert, Avatar, Badge, Banner, Button, Card, Carousel, Checkbox, Codeblock, Combobox, Dropdown, Modal, Navbar, Pagination, Select, Sidebar, Spinner, Table, Tabs, Textarea, Text Input, Toast, Toggle, Tooltip

Plus form composition components: Form, Section, CollapsibleSection, FlipSection, FieldGroup, SubSection

Plus multi-value input components: TagsList, KeyValue, Triplet

Features:
- 16 themes (including TOTVS and Dracula) with full dark mode support
- E2E test suite with 381 Playwright tests
- HTMX integration (sorting, pagination, infinite scroll, lazy loading)
- Alpine.js client-side interactivity (combobox search, expandable rows, form validation)
- Tailwind CSS v4 with CSS custom property design tokens
