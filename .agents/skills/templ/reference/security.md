# templ security / injection model

Verbatim from templ.guide. templ v0.3.x.

templ is designed to stop user-provided data from injecting vulnerabilities. Know what is auto-protected vs. what you must handle.

## What's auto-escaped

- **Node text** `{ expr }` → escaped with `templ.EscapeString` (context-aware HTML encoding). XSS-inert.
- Plain text node content (no expression) is **not** modified.

```templ
<div>Node text is not modified at all.</div>
<div>{ "will be escaped using templ.EscapeString" }</div>
```

## `<script>` / `<style>` bodies forbid variables

Variables are **not permitted** inside `<script>`/`<style>` tag bodies — only constant content. This blocks the most direct injection path.
```templ
<script> function showAlert(){ alert("hello"); } </script>
<style type="text/css"> /* Only CSS is allowed */ </style>
```
Pass dynamic data via attributes, `templ.JSONScript`, `templ.JSFuncCall`, or `{{ }}` interpolation (see components.md).

## `on*` attributes → `templ.ComponentScript`

`onClick` etc. accept a `script` template / `ComponentScript`, so user data is escaped via `templ.Escape`.
```templ
script onClickHandler(msg string) { alert(msg); }
templ Example(msg string) {
	<div onClick={ onClickHandler(msg) }>{ "will be HTML encoded" }</div>
}
```

## Style attributes — constants only

`style={ expr }` with a non-constant **throws an error**. Use `css` style templates instead.
```templ
<div style={ "will throw an error" }></div>   // ❌
```

## Class sanitization

Class names are sanitized by default; a failed name becomes `--templ-css-class-safe-name`. Bypass with `templ.SafeClass` (result is still escaped).
```templ
<div class={ "unsafe</style>-will-sanitized", templ.SafeClass("&sanitization bypassed") }></div>
// → <div class="--templ-css-class-safe-name &amp;sanitization bypassed"></div>
```

## URL sanitization

`href` must be a `templ.SafeURL`; templ strips JavaScript URLs unless bypassed. **String constants are not sanitized.**
```templ
<a href="http://constants.example.com/not/sanitized">Text</a>
<a href={ templ.URL("sanitized to remove attacks") }>...</a>
<a href={ templ.SafeURL("NOT sanitized — trusted only") }>...</a>
```

## CSS value sanitization

In `css` blocks, property names and **constant** values are not sanitized/escaped. Expression-based values pass through `templ.SanitizeCSS`, replacing unsafe values with placeholders.
```css
css className() { background-color: #ffffff; }   /* constant: untouched */
css className() { color: { red }; }              /* expression: SanitizeCSS */
```

## Bypass functions (use only with trusted data)

| Function | Bypasses |
|----------|----------|
| `templ.Raw(html)` | all HTML escaping — trusted source only |
| `templ.SafeURL(u)` | URL sanitization |
| `templ.SafeClass(c)` | class-name sanitization (still escaped) |
| `templ.SafeCSS` | CSS sanitization |
| `templ.JSExpression(s)` | JS JSON-encoding (XSS risk if untrusted) |
| `templ.JSUnsafeFuncCall(s)` | script sanitization |

## Content Security Policy (nonces)

For CSP, generate a per-request nonce, put it in the `Content-Security-Policy` header, and set it on `<script nonce=...>` tags (templ supports threading a nonce via context). Lets inline scripts run under a strict policy without `unsafe-inline`. See templ.guide/security/content-security-policy.
