---
name: templ
description: Use when writing, generating, debugging, or reviewing templ (.templ) files — Go's typed HTML templating. Covers the templ keyword, components, expressions { }, control flow (if/else/switch/for), attributes (conditional/boolean/spread/URL/JSON/class/style), template composition and { children... }, css blocks, script templates and passing Go data to JS (templ.JSFuncCall/JSONString/JSONScript/JSExpression), templ.Raw, render-once, context, the templ CLI (generate/fmt/lsp/--watch), code generation into *_templ.go, and templ's escaping/injection security model (SafeURL/SafeCSS/SafeClass). Also covers templ + Alpine.js / htmx attribute-escaping bugs.
---

# templ

## Overview

**templ compiles `.templ` files into Go functions that return a `templ.Component`.** Markup is type-checked Go, not runtime string templates. You edit `<name>.templ`, run `templ generate` → produces `<name>_templ.go`, and call the generated function.

```go
type Component interface {
	Render(ctx context.Context, w io.Writer) error
}
```

Reference docs crawled from templ.guide / templ.guide/llms.md (templ v0.3.x). Authoritative condensed source: <https://templ.guide/llms.md>.

**This repo (Goshtoso) is a templ component library** (Go + templ + Tailwind + HTMX + Alpine.js). Generated `*_templ.go` files are **never hand-edited** — regenerate. See `reference/gotchas.md` and the project `AGENTS.md` for the repo's hard-won escaping rules before generating any Alpine/htmx markup.

## When to use

- Writing or editing any `.templ` file: components, expressions, control flow, attributes.
- Composing components, passing `{ children... }`, passing components as parameters.
- Styling with `css` blocks / dynamic `class`/`style` attributes.
- Running JavaScript: `<script>` tags, `script` templates, passing Go data to JS, Alpine `x-data`.
- Running the CLI (`templ generate`, `fmt`, `lsp`, `--watch`) or wiring it into dev/CI.
- Debugging why a component renders wrong, an Alpine/htmx attribute silently breaks, or output is unexpectedly escaped/sanitized.

## Core syntax

```templ
package main

import "strings"

var greeting = "Welcome!"   // outside templ blocks = ordinary Go

templ headerTemplate(name string) {
	<header data-testid="header">
		<h1>{ name }</h1>                      // { } = interpolation expression
		<p>{ strings.ToUpper(name) }</p>       // funcs, (value, error) tuples OK
	</header>
}
```

| Rule | Detail |
|------|--------|
| Component def | `templ Name(args) { ... }` — uppercase = exported, lowercase = private (Go rules) |
| Interpolation | `{ expr }` outputs **HTML-escaped** text. Strings, numbers, bools, `fmt.Stringer`, `(T, error)` |
| Tags must close | `<a></a>` or self-close `<hr/>`. templ knows void elements → emits `<br>` not `<br/>` |
| Render a component | `@Component()` / `@pkg.Component()` |
| Children slot | `{ children... }` inside a component; caller nests markup in `@Wrapper() { ... }` |
| Control flow | standard Go `if`/`else`, `switch`, `for` directly inside markup — no escaping |
| Raw Go | `{{ x := 1 }}` for statements; declare vars/funcs outside `templ` blocks |

## Control flow

```templ
templ list(items []Item, isLoggedIn bool) {
	if isLoggedIn {
		<div>Welcome back!</div>
	} else {
		<input type="button" value="Log in"/>
	}
	switch role {
		case "admin": <span>Admin</span>
		default:      <span>User</span>
	}
	<ul>
		for _, item := range items {
			<li>{ item.Name }</li>
		}
	</ul>
}
```

Text that literally starts with `if`/`switch`/`for` is parsed as a statement — wrap in a string expression `{ "if ..." }` or capitalize to emit it as text.

## Attributes (quick reference)

| Need | Syntax |
|------|--------|
| Dynamic value | `<button value={ name }>` |
| Conditional attribute | `<hr if cond { class="x" } />` |
| Boolean attribute | `<hr noshade/>` · variable: `<hr noshade?={ b } />` |
| Spread (`templ.Attributes`) | `<p { attrs... }>` |
| Dynamic key | `<p { "data-" + id }="v">` |
| URL (sanitized) | `<a href={ templ.URL(u) }>` · bypass `templ.SafeURL(u)` |
| Class (multi/conditional) | `class={ "btn", templ.KV("active", isActive) }` |
| Style | `style={ s }` (expr) — but see security: only via `css`/sanitized |
| JSON into attr (Alpine) | `x-data={ templ.JSONString(data) }` |
| JS handler | `onClick={ templ.JSFuncCall("fn", arg) }` or a `script` template |

## CLI

```bash
templ generate              # walk tree, .templ → *_templ.go  (run after EVERY .templ edit)
templ generate -f x.templ   # single file
templ generate --watch --proxy="http://localhost:8080" --cmd="go run ."  # live reload
templ fmt .                 # format; templ fmt -fail . in CI
templ lsp                   # language server (used by IDE extensions, not directly)
```

In this repo: `templ generate` (or `just gp-generate`). If it reports "0 updates" wrongly, `rm components/<n>/<n>_templ.go && templ generate`.

## Detailed references (load when needed)

- **`reference/syntax.md`** — full syntax: elements, every attribute form, expressions/escaping, if/switch/for, composition + `children...` + components-as-params + `templ.Join`, `templ.Raw`, raw Go, render-once (`OnceHandle`), `context`, comments.
- **`reference/components.md`** — component model, code-only components (`templ.ComponentFunc`), method components, view models, `css` blocks/components, passing Go data to JavaScript (`JSFuncCall`/`JSExpression`/`JSONString`/`JSONScript`/`{{ }}` interpolation), `script` templates.
- **`reference/cli.md`** — every `generate`/`fmt`/`lsp` flag, watch/proxy live reload, CI checks, version/timestamp control.
- **`reference/security.md`** — injection model: what's auto-escaped, why `<script>`/`<style>` forbid variables, `on*` → `templ.ComponentScript`, `templ.URL`/`SafeURL`, class sanitization/`SafeClass`, `css` sanitization/`SanitizeCSS`, CSP nonces.
- **`reference/gotchas.md`** — **read before generating Alpine.js/htmx markup.** templ escaping silently breaks `x-data`/`hx-*` (no console error); `json.Marshal` in attributes; null-array crashes; regeneration quirks. Cross-references repo `AGENTS.md`.

## Common mistakes

- **Forgot `templ generate`.** Editing `.templ` does nothing until you regenerate the `*_templ.go`. Build will use stale output.
- **Editing `*_templ.go` by hand.** Generated — overwritten on next run. Edit the `.templ` source.
- **Escaped Alpine/htmx attributes.** `{ }` and `json.Marshal` produce `&quot;` inside attributes → Alpine swallows the parse error silently. Use single-quoted JS builders or `templ.JSONString`. See `reference/gotchas.md`.
- **Variable in `<script>`/`<style>` body.** Not permitted (injection guard). Pass data via attributes/`JSONScript`/`{{ }}` or `script` templates.
- **`style={ expr }` with a non-constant** throws — use a `css` block/component instead.
- **Unclosed tags.** Every element must close or self-close; templ won't infer it.
