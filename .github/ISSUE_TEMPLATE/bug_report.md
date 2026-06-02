---
name: Bug report
about: Report a defect or visual-parity issue in a component
title: "[Bug]: "
labels: bug
assignees: ''
---

## Description

A clear description of what's wrong.

## Component affected

Which component (e.g. Table, Combobox, Dropdown)? Which variant?

## Reproduction steps

1.
2.
3.

## Expected vs actual

- **Expected:**
- **Actual:**

## Environment

- **Theme:** (e.g. Minimal, ...)
- **Mode:** light / dark
- **Browser:** (E2E runs Chromium via Playwright)
- **Go version:** `go version`
- **templ version:**
- **Goshtoso version / commit:**

## Screenshots

If a visual-parity issue, attach Goshtoso vs PenguinUI screenshots.

## Notes

Did you check the rendered HTML in devtools for `&quot;` inside an `x-data`
(templ-escaping bug)?
